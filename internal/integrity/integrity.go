// Package integrity provides hook integrity verification via SHA-256.
//
// TokMan installs a PreToolUse hook (tokman-rewrite.sh) that auto-approves
// rewritten commands. Because this hook bypasses Claude Code's permission
// prompts, any unauthorized modification represents a command injection vector.
//
// This module provides:
//   - SHA-256 hash computation and storage at install time
//   - Runtime verification before command execution
//   - Manual verification via `tokman verify`
package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HashFilename is the filename for the stored hash (dotfile alongside hook)
const HashFilename = ".tokman-hook.sha256"

// HookFilename is the expected hook script filename
const HookFilename = "tokman-rewrite.sh"

// CurrentHookVersion is the minimum supported generated hook version.
const CurrentHookVersion = 1

// IntegrityStatus represents the result of hook integrity verification
type IntegrityStatus int

const (
	// StatusVerified indicates hash matches - hook is unmodified since last install
	StatusVerified IntegrityStatus = iota
	// StatusTampered indicates hash mismatch - hook has been modified outside of tokman init
	StatusTampered
	// StatusNoBaseline indicates hook exists but no stored hash (installed before integrity checks)
	StatusNoBaseline
	// StatusNotInstalled indicates neither hook nor hash file exist (TokMan not installed)
	StatusNotInstalled
	// StatusOrphanedHash indicates hash file exists but hook was deleted
	StatusOrphanedHash
	// StatusOutdated indicates hook exists but is older than the current generated version
	StatusOutdated
)

// String returns a human-readable status name
func (s IntegrityStatus) String() string {
	switch s {
	case StatusVerified:
		return "VERIFIED"
	case StatusTampered:
		return "TAMPERED"
	case StatusNoBaseline:
		return "NO_BASELINE"
	case StatusNotInstalled:
		return "NOT_INSTALLED"
	case StatusOrphanedHash:
		return "ORPHANED_HASH"
	case StatusOutdated:
		return "OUTDATED"
	default:
		return "UNKNOWN"
	}
}

// VerificationResult contains detailed verification results
type VerificationResult struct {
	Status          IntegrityStatus
	Expected        string // Expected hash (for StatusTampered)
	Actual          string // Actual hash (for StatusTampered)
	HookPath        string // Path to the hook file
	HashPath        string // Path to the hash file
	HookVersion     int
	RequiredVersion int
}

// ComputeHash computes SHA-256 hash of a file, returned as lowercase hex
func ComputeHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// HashPath derives the hash file path from the hook path
func HashPath(hookPath string) string {
	dir := filepath.Dir(hookPath)
	return filepath.Join(dir, HashFilename)
}

// StoreHash stores SHA-256 hash of the hook script after installation.
//
// Format is compatible with `sha256sum -c`:
//
//	<hex_hash>  tokman-rewrite.sh
//
// The hash file is set to read-only (0444) as a speed bump
// against casual modification. Not a security boundary — an
// attacker with write access can chmod it — but forces a
// deliberate action rather than accidental overwrite.
func StoreHash(hookPath string) error {
	hash, err := ComputeHash(hookPath)
	if err != nil {
		return err
	}

	hashFile := HashPath(hookPath)
	filename := filepath.Base(hookPath)
	if filename == "" {
		filename = HookFilename
	}

	// Format: "<hash>  <filename>\n" (sha256sum format)
	content := fmt.Sprintf("%s  %s\n", hash, filename)

	// If hash file exists and is read-only, make it writable first
	if info, err := os.Stat(hashFile); err == nil {
		if info.Mode().Perm()&0200 == 0 {
			if err := os.Chmod(hashFile, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to make hash file writable: %v\n", err)
			}
		}
	}

	if err := os.WriteFile(hashFile, []byte(content), 0444); err != nil {
		return fmt.Errorf("failed to write hash to %s: %w", hashFile, err)
	}

	return nil
}

// RemoveHash removes the stored hash file (called during uninstall)
func RemoveHash(hookPath string) (bool, error) {
	hashFile := HashPath(hookPath)

	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		return false, nil
	}

	// Make writable before removing
	if err := os.Chmod(hashFile, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to make hash file writable: %v\n", err)
	}

	if err := os.Remove(hashFile); err != nil {
		return false, fmt.Errorf("failed to remove hash file %s: %w", hashFile, err)
	}

	return true, nil
}

// readStoredHash reads the stored hash from the hash file.
//
// Expects exact `sha256sum -c` format: `<64 hex>  <filename>\n`
// Rejects malformed files rather than silently accepting them.
func readStoredHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read hash file %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", fmt.Errorf("empty hash file: %s", path)
	}

	line := lines[0]

	// sha256sum format uses two-space separator: "<hash>  <filename>"
	parts := strings.SplitN(line, "  ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid hash format in %s (expected 'hash  filename')", path)
	}

	hash := parts[0]
	if len(hash) != 64 {
		return "", fmt.Errorf("invalid SHA-256 hash length in %s: expected 64, got %d", path, len(hash))
	}

	// Verify it's valid hex
	if _, err := hex.DecodeString(hash); err != nil {
		return "", fmt.Errorf("invalid SHA-256 hash in %s: %w", path, err)
	}

	return hash, nil
}

// ResolveHookPath resolves the default hook path (~/.claude/hooks/tokman-rewrite.sh)
func ResolveHookPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".claude", "hooks", HookFilename), nil
}

// VerifyHookAt verifies hook integrity for a specific hook path (testable)
func VerifyHookAt(hookPath string) (*VerificationResult, error) {
	hashFile := HashPath(hookPath)

	hookExists := fileExists(hookPath)
	hashExists := fileExists(hashFile)

	result := &VerificationResult{
		HookPath:        hookPath,
		HashPath:        hashFile,
		RequiredVersion: CurrentHookVersion,
	}

	switch {
	case !hookExists && !hashExists:
		result.Status = StatusNotInstalled
	case !hookExists && hashExists:
		result.Status = StatusOrphanedHash
	case hookExists && !hashExists:
		version, err := readHookVersion(hookPath)
		if err == nil {
			result.HookVersion = version
			if version < CurrentHookVersion {
				result.Status = StatusOutdated
				break
			}
		}
		result.Status = StatusNoBaseline
	default:
		// Both exist - compare hashes
		stored, err := readStoredHash(hashFile)
		if err != nil {
			return nil, err
		}

		actual, err := ComputeHash(hookPath)
		if err != nil {
			return nil, err
		}
		version, err := readHookVersion(hookPath)
		if err == nil {
			result.HookVersion = version
		}

		if stored == actual {
			if result.HookVersion < CurrentHookVersion {
				result.Status = StatusOutdated
			} else {
				result.Status = StatusVerified
			}
		} else {
			result.Status = StatusTampered
			result.Expected = stored
			result.Actual = actual
		}
	}

	return result, nil
}

// VerifyHook verifies hook integrity against stored hash using default path
func VerifyHook() (*VerificationResult, error) {
	hookPath, err := ResolveHookPath()
	if err != nil {
		return nil, err
	}
	return VerifyHookAt(hookPath)
}

// RuntimeCheck performs a runtime integrity gate.
//
// Behavior:
//   - Verified / NotInstalled / NoBaseline: silent, continue (returns nil)
//   - Tampered: returns error with details
//   - OrphanedHash: logs warning, continues (returns nil)
//
// No env-var bypass is provided — if the hook is legitimately modified,
// re-run `tokman init` to re-establish the baseline.
func RuntimeCheck() error {
	result, err := VerifyHook()
	if err != nil {
		return err
	}

	switch result.Status {
	case StatusVerified, StatusNotInstalled, StatusNoBaseline:
		// All good, proceed
		return nil

	case StatusOutdated:
		return fmt.Errorf(`hook version is outdated
  Installed version: %d
  Required version:  %d

  The hook at ~/.claude/hooks/tokman-rewrite.sh is from an older TokMan install.
  Re-run TokMan setup before executing commands.

  To restore:  tokman init --claude
  To inspect:  tokman verify`,
			result.HookVersion,
			result.RequiredVersion)

	case StatusTampered:
		return fmt.Errorf(`hook integrity check FAILED
  Expected hash: %s...
  Actual hash:   %s...

  The hook at ~/.claude/hooks/tokman-rewrite.sh has been modified.
  This may indicate tampering. TokMan will not execute.

  To restore:  tokman init
  To inspect:  tokman verify`,
			truncateHash(result.Expected, 16),
			truncateHash(result.Actual, 16))

	case StatusOrphanedHash:
		// Log warning but don't block - hook is gone, nothing to exploit
		fmt.Fprintln(os.Stderr, "tokman: warning: hash file exists but hook is missing")
		fmt.Fprintln(os.Stderr, "  Run `tokman init` to reinstall.")
		return nil

	default:
		return nil
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// truncateHash truncates a hash string for display
func truncateHash(hash string, maxLen int) string {
	if len(hash) <= maxLen {
		return hash
	}
	return hash[:maxLen]
}

func readHookVersion(path string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read hook file %s: %w", path, err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# tokman-hook-version:") {
			var version int
			_, err := fmt.Sscanf(line, "# tokman-hook-version: %d", &version)
			if err != nil {
				return 0, fmt.Errorf("invalid hook version marker in %s: %w", path, err)
			}
			return version, nil
		}
	}
	return 0, nil
}
