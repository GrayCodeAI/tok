package toml

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TrustEntry holds trust metadata for a project filter file.
type TrustEntry struct {
	SHA256    string `json:"sha256"`
	TrustedAt string `json:"trusted_at"`
}

// TrustStatus represents the result of a trust check.
type TrustStatus int

const (
	TrustStatusTrusted TrustStatus = iota
	TrustStatusUntrusted
	TrustStatusContentChanged
	TrustStatusEnvOverride
)

// TrustStore holds all trusted project filter entries.
type TrustStore struct {
	Version int                   `json:"version"`
	Trusted map[string]TrustEntry `json:"trusted"`
}

// NewTrustStore creates an empty trust store.
func NewTrustStore() *TrustStore {
	return &TrustStore{
		Version: 1,
		Trusted: make(map[string]TrustEntry),
	}
}

// trustStorePath returns the path to the trust store file.
func trustStorePath(configDir string) string {
	return filepath.Join(configDir, "trusted_filters.json")
}

// computeFileHash computes the SHA-256 hash of a file.
func computeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// TrustManager manages project filter trust with SHA-256 verification.
type TrustManager struct {
	mu        sync.RWMutex
	store     *TrustStore
	storePath string
	configDir string
}

// NewTrustManager creates a new trust manager.
func NewTrustManager(configDir string) *TrustManager {
	return &TrustManager{
		store:     NewTrustStore(),
		storePath: trustStorePath(configDir),
		configDir: configDir,
	}
}

// Load loads the trust store from disk.
func (tm *TrustManager) Load() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	data, err := os.ReadFile(tm.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read trust store: %w", err)
	}

	var store TrustStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse trust store: %w", err)
	}
	tm.store = &store
	return nil
}

// Save persists the trust store to disk.
func (tm *TrustManager) Save() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.saveLocked()
}

func (tm *TrustManager) saveLocked() error {
	dir := filepath.Dir(tm.storePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(tm.store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize trust store: %w", err)
	}

	return os.WriteFile(tm.storePath, data, 0600)
}

// CheckTrust verifies if a project filter file is trusted.
// Returns TrustStatus and any error.
func (tm *TrustManager) CheckTrust(filterPath string) (TrustStatus, error) {
	// Fast path: env var override for CI pipelines
	if os.Getenv("TOKMAN_TRUST_PROJECT_FILTERS") == "1" {
		if isCIEnvironment() {
			return TrustStatusEnvOverride, nil
		}
		fmt.Fprintf(os.Stderr, "[tokman] WARNING: TOKMAN_TRUST_PROJECT_FILTERS=1 ignored (CI environment not detected)\n")
	}

	absPath, err := filepath.Abs(filterPath)
	if err != nil {
		return TrustStatusUntrusted, fmt.Errorf("failed to resolve path: %w", err)
	}

	tm.mu.RLock()
	entry, ok := tm.store.Trusted[absPath]
	tm.mu.RUnlock()

	if !ok {
		return TrustStatusUntrusted, nil
	}

	actualHash, err := computeFileHash(filterPath)
	if err != nil {
		return TrustStatusUntrusted, fmt.Errorf("failed to hash filter file: %w", err)
	}

	if actualHash == entry.SHA256 {
		return TrustStatusTrusted, nil
	}

	return TrustStatusContentChanged, nil
}

// TrustProject stores the current SHA-256 hash as trusted.
func (tm *TrustManager) TrustProject(filterPath string) error {
	absPath, err := filepath.Abs(filterPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	hash, err := computeFileHash(filterPath)
	if err != nil {
		return fmt.Errorf("failed to hash filter file: %w", err)
	}

	tm.mu.Lock()
	tm.store.Trusted[absPath] = TrustEntry{
		SHA256:    hash,
		TrustedAt: time.Now().Format(time.RFC3339),
	}
	err = tm.saveLocked()
	tm.mu.Unlock()

	return err
}

// UntrustProject removes trust for a project filter file.
func (tm *TrustManager) UntrustProject(filterPath string) (bool, error) {
	absPath, err := filepath.Abs(filterPath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve path: %w", err)
	}

	tm.mu.Lock()
	_, ok := tm.store.Trusted[absPath]
	if ok {
		delete(tm.store.Trusted, absPath)
	}
	err = tm.saveLocked()
	tm.mu.Unlock()

	return ok, err
}

// ListTrusted returns all trusted project filter paths.
func (tm *TrustManager) ListTrusted() map[string]TrustEntry {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make(map[string]TrustEntry, len(tm.store.Trusted))
	for k, v := range tm.store.Trusted {
		result[k] = v
	}
	return result
}

// RiskSummary holds risk analysis results for a filter file.
type RiskSummary struct {
	FilterCount    int
	HasReplace     bool
	HasMatchOutput bool
	HasCatchAll    bool
}

// AnalyzeFilterRisk analyzes a TOML filter file for risky patterns.
func AnalyzeFilterRisk(content string) RiskSummary {
	return RiskSummary{
		FilterCount:    strings.Count(content, "[filters."),
		HasReplace:     strings.Contains(content, "replace"),
		HasMatchOutput: strings.Contains(content, "match_output"),
		HasCatchAll:    strings.Contains(content, `pattern = "."`) || strings.Contains(content, `pattern = '.'`),
	}
}

// isCIEnvironment checks if running in a known CI environment.
func isCIEnvironment() bool {
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "BUILDKITE", "CIRCLECI", "TRAVIS"}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}
