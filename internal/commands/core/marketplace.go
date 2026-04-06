package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Browse and install community TOML filters",
	Long:  `Search, install, and manage community-contributed TOML filters.`,
}

var marketplaceSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available filters",
	Args:  cobra.ExactArgs(1),
	RunE:  runMarketplaceSearch,
}

var marketplaceInstallCmd = &cobra.Command{
	Use:   "install <filter-name>",
	Short: "Install a community filter",
	Args:  cobra.ExactArgs(1),
	RunE:  runMarketplaceInstall,
}

var marketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed community filters",
	RunE:  runMarketplaceList,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceSearchCmd)
	marketplaceCmd.AddCommand(marketplaceInstallCmd)
	marketplaceCmd.AddCommand(marketplaceListCmd)
	registry.Add(func() { registry.Register(marketplaceCmd) })
}

// CommunityFilters is a registry of known community filters.
var CommunityFilters = map[string]string{
	"jest":       "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/jest.toml",
	"vitest":     "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/vitest.toml",
	"playwright": "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/playwright.toml",
	"cypress":    "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/cypress.toml",
	"mocha":      "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/mocha.toml",
	"eslint":     "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/eslint.toml",
	"biome":      "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/biome.toml",
	"swc":        "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/swc.toml",
	"webpack":    "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/webpack.toml",
	"vite":       "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/vite.toml",
	"rollup":     "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/rollup.toml",
	"trivy":      "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/trivy.toml",
	"snyk":       "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/snyk.toml",
	"opentofu":   "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/opentofu.toml",
	"pulumi":     "https://raw.githubusercontent.com/GrayCodeAI/tokman/main/filters/pulumi.toml",
}

func runMarketplaceSearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])
	fmt.Printf("Searching for filters matching '%s'...\n\n", query)

	found := 0
	for name, url := range CommunityFilters {
		if strings.Contains(strings.ToLower(name), query) {
			fmt.Printf("  %-20s %s\n", name, url)
			found++
		}
	}

	if found == 0 {
		fmt.Println("No filters found. Try a different search term.")
		fmt.Println("\nAvailable filters:")
		for name := range CommunityFilters {
			fmt.Printf("  %s\n", name)
		}
	}

	return nil
}

func runMarketplaceInstall(cmd *cobra.Command, args []string) error {
	name := args[0]
	url, ok := CommunityFilters[name]
	if !ok {
		return fmt.Errorf("unknown filter: %s (use 'tokman marketplace search' to find filters)", name)
	}

	// Download filter with security controls
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download filter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("filter not found (HTTP %d)", resp.StatusCode)
	}

	// Validate Content-Type (allow octet-stream for raw GitHub)
	ct := resp.Header.Get("Content-Type")
	if ct != "" && ct != "text/plain" && ct != "application/octet-stream" && !strings.HasPrefix(ct, "text/") {
		return fmt.Errorf("unexpected content type: %s", ct)
	}

	// Limit download size to 1MB to prevent memory exhaustion
	const maxFilterSize = 1 * 1024 * 1024
	limitedReader := http.MaxBytesReader(nil, resp.Body, maxFilterSize)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read filter (exceeded %d bytes): %w", maxFilterSize, err)
	}

	// Validate the downloaded content is valid TOML
	var dummy map[string]any
	if _, err := toml.Decode(string(content), &dummy); err != nil {
		return fmt.Errorf("downloaded content is not a valid TOML filter: %w", err)
	}

	// Save to user filters directory
	filterDir := config.FiltersDir()
	if err := os.MkdirAll(filterDir, 0700); err != nil {
		return fmt.Errorf("failed to create filter directory: %w", err)
	}

	filterPath := filepath.Join(filterDir, name+".toml")
	if err := os.WriteFile(filterPath, content, 0600); err != nil {
		return fmt.Errorf("failed to save filter: %w", err)
	}

	fmt.Printf("Installed '%s' to %s\n", name, filterPath)
	return nil
}

func runMarketplaceList(cmd *cobra.Command, args []string) error {
	filterDir := config.FiltersDir()

	entries, err := os.ReadDir(filterDir)
	if err != nil || len(entries) == 0 {
		fmt.Println("No community filters installed.")
		fmt.Println("Install with: tokman marketplace install <name>")
		return nil
	}

	fmt.Println("Installed community filters:")
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".toml") {
			name := strings.TrimSuffix(e.Name(), ".toml")
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}
