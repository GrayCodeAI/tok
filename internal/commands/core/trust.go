package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var trustList bool

// trustCmd represents the trust command
var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust project-local TOML filters in current directory",
	Long: `Review and trust the .tok/filters.toml file in the current directory.

Trust-before-load security model:
- Untrusted filters are SKIPPED (not loaded with warning)
- This command stores the SHA-256 hash after user review
- Content changes invalidate trust (requires re-review)
- TOK_TRUST_PROJECT_FILTERS=1 overrides for CI pipelines

Examples:
  tok trust          # Review and trust .tok/filters.toml
  tok trust --list   # List all trusted projects`,
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
	},
	RunE: runTrust,
}

func init() {
	registry.Add(func() { registry.Register(trustCmd) })
	trustCmd.Flags().BoolVarP(&trustList, "list", "l", false, "List all trusted projects")
}

// TrustStore represents the stored trust entries
type TrustStore struct {
	Version uint32                `json:"version"`
	Trusted map[string]TrustEntry `json:"trusted"`
}

// TrustEntry represents a single trusted filter
type TrustEntry struct {
	SHA256    string `json:"sha256"`
	TrustedAt string `json:"trusted_at"`
}

// TrustStatus represents the trust state
type TrustStatus int

const (
	TrustStatusUntrusted TrustStatus = iota
	TrustStatusTrusted
	TrustStatusContentChanged
	TrustStatusEnvOverride
)

func runTrust(cmd *cobra.Command, args []string) error {
	if trustList {
		return listTrusted()
	}

	filterPath := ".tok/filters.toml"

	// Check if file exists
	content, err := os.ReadFile(filterPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no .tok/filters.toml found in current directory")
		}
		return fmt.Errorf("failed to read .tok/filters.toml: %w", err)
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// Display file content
	out.Global().Println()
	out.Global().Println(cyan("=== .tok/filters.toml ==="))
	out.Global().Println(string(content))
	out.Global().Println(cyan("==========================="))
	out.Global().Println()

	// Print risk summary
	printRiskSummary(string(content))
	out.Global().Println()

	// Calculate hash
	hash := computeHash(content)

	// Ask for confirmation
	out.Global().Print("Trust this filter file? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		out.Global().Println("Trust canceled.")
		return nil
	}

	// Store trust
	if err := trustFilter(filterPath, hash); err != nil {
		return fmt.Errorf("failed to store trust: %w", err)
	}

	out.Global().Println()
	out.Global().Printf("%s Trusted .tok/filters.toml (sha256:%s)\n",
		green("✓"), hash[:16])
	out.Global().Println("Project-local filters will now be applied.")

	return nil
}

func listTrusted() error {
	store, err := readTrustStore()
	if err != nil {
		return fmt.Errorf("failed to read trust store: %w", err)
	}

	if len(store.Trusted) == 0 {
		out.Global().Println("No trusted project filters.")
		return nil
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	out.Global().Println()
	out.Global().Println(cyan("Trusted project filters:"))
	out.Global().Println(strings.Repeat("═", 60))

	for path, entry := range store.Trusted {
		date := entry.TrustedAt
		if len(date) > 10 {
			date = date[:10]
		}
		out.Global().Printf("  %s (trusted %s)\n", yellow(path), date)
		out.Global().Printf("    sha256:%s\n", entry.SHA256)
	}

	return nil
}

func printRiskSummary(content string) {
	filterCount := strings.Count(content, "[[filters")
	hasReplace := strings.Contains(content, "replace")
	hasMatchOutput := strings.Contains(content, "match_output")
	hasDotPattern := strings.Contains(content, `pattern = "."`) || strings.Contains(content, "pattern = '.'")

	yellow := color.New(color.FgYellow).SprintFunc()

	out.Global().Println("Risk summary:")
	out.Global().Printf("  Filters: %d\n", filterCount)

	if hasReplace {
		out.Global().Printf("  %s Contains 'replace' rules (can rewrite output)\n", yellow("[!]"))
	}
	if hasMatchOutput {
		out.Global().Printf("  %s Contains 'match_output' rules (can replace entire output)\n", yellow("[!]"))
	}
	if hasDotPattern {
		out.Global().Printf("  %s Contains catch-all pattern '.' (matches everything)\n", yellow("[!]"))
	}
	if !hasReplace && !hasMatchOutput && !hasDotPattern {
		out.Global().Println("  No high-risk patterns detected.")
	}
}

func computeHash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func trustStorePath() string {
	dataDir := config.DataPath()
	return filepath.Join(dataDir, "trusted_filters.json")
}

func readTrustStore() (*TrustStore, error) {
	path := trustStorePath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &TrustStore{
			Version: 1,
			Trusted: make(map[string]TrustEntry),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var store TrustStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	if store.Trusted == nil {
		store.Trusted = make(map[string]TrustEntry)
	}

	return &store, nil
}

func writeTrustStore(store *TrustStore) error {
	path := trustStorePath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func canonicalKey(filterPath string) (string, error) {
	absPath, err := filepath.Abs(filterPath)
	if err != nil {
		return "", err
	}

	// Try to resolve symlinks
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = resolved
	}

	return absPath, nil
}

func trustFilter(filterPath string, hash string) error {
	key, err := canonicalKey(filterPath)
	if err != nil {
		return err
	}

	store, err := readTrustStore()
	if err != nil {
		return err
	}

	store.Version = 1
	store.Trusted[key] = TrustEntry{
		SHA256:    hash,
		TrustedAt: time.Now().UTC().Format(time.RFC3339),
	}

	return writeTrustStore(store)
}

// CheckTrust checks if a filter file is trusted
func CheckTrust(filterPath string) TrustStatus {
	// Fast path: env var override for CI pipelines
	if os.Getenv("TOK_TRUST_PROJECT_FILTERS") == "1" {
		// Require CI environment to prevent .envrc injection
		if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" ||
			os.Getenv("GITLAB_CI") != "" || os.Getenv("JENKINS_URL") != "" {
			return TrustStatusEnvOverride
		}
	}

	key, err := canonicalKey(filterPath)
	if err != nil {
		return TrustStatusUntrusted
	}

	store, err := readTrustStore()
	if err != nil {
		return TrustStatusUntrusted
	}

	entry, ok := store.Trusted[key]
	if !ok {
		return TrustStatusUntrusted
	}

	// Verify hash
	content, err := os.ReadFile(filterPath)
	if err != nil {
		return TrustStatusUntrusted
	}

	actualHash := computeHash(content)
	if actualHash != entry.SHA256 {
		return TrustStatusContentChanged
	}

	return TrustStatusTrusted
}

// UntrustFilter removes trust for a filter file
func UntrustFilter(filterPath string) (bool, error) {
	key, err := canonicalKey(filterPath)
	if err != nil {
		return false, err
	}

	store, err := readTrustStore()
	if err != nil {
		return false, err
	}

	_, existed := store.Trusted[key]
	if existed {
		delete(store.Trusted, key)
		if err := writeTrustStore(store); err != nil {
			return false, err
		}
	}

	return existed, nil
}
