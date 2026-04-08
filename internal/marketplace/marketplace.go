// Package marketplace provides a community filter marketplace
package marketplace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/GrayCodeAI/tokman/internal/config"
)

// Marketplace manages community filters
type Marketplace struct {
	registryURL string
	localDir    string
	client      *http.Client
	filters     map[string]*Filter
}

// Filter represents a marketplace filter
type Filter struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Author        string                 `json:"author"`
	Version       string                 `json:"version"`
	Category      string                 `json:"category"`
	Tags          []string               `json:"tags"`
	Downloads     int                    `json:"downloads"`
	Rating        float64                `json:"rating"`
	RatingCount   int                    `json:"rating_count"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Content       string                 `json:"content"`
	Config        map[string]interface{} `json:"config"`
	Compatibility []string               `json:"compatibility"`
	Verified      bool                   `json:"verified"`
	Official      bool                   `json:"official"`
}

// FilterCategory represents filter categories
const (
	CategoryGeneral   = "general"
	CategoryLanguage  = "language"
	CategoryTool      = "tool"
	CategoryFramework = "framework"
	CategoryCustom    = "custom"
)

// NewMarketplace creates a new marketplace instance
func NewMarketplace(registryURL string) (*Marketplace, error) {
	if registryURL == "" {
		registryURL = "https://filters.tokman.dev/api/v1"
	}

	localDir := filepath.Join(config.ConfigDir(), "marketplace")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create marketplace directory: %w", err)
	}

	return &Marketplace{
		registryURL: registryURL,
		localDir:    localDir,
		client:      &http.Client{Timeout: 30 * time.Second},
		filters:     make(map[string]*Filter),
	}, nil
}

// SearchFilters searches for filters in the marketplace
func (m *Marketplace) SearchFilters(query string, category string, tags []string) ([]*Filter, error) {
	// Build search URL
	url := fmt.Sprintf("%s/filters/search?q=%s", m.registryURL, query)
	if category != "" {
		url += fmt.Sprintf("&category=%s", category)
	}
	for _, tag := range tags {
		url += fmt.Sprintf("&tag=%s", tag)
	}

	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search filters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var filters []*Filter
	if err := json.NewDecoder(resp.Body).Decode(&filters); err != nil {
		return nil, fmt.Errorf("failed to decode filters: %w", err)
	}

	return filters, nil
}

// ListFilters lists all available filters
func (m *Marketplace) ListFilters(category string, sortBy string) ([]*Filter, error) {
	url := fmt.Sprintf("%s/filters", m.registryURL)
	if category != "" {
		url += fmt.Sprintf("?category=%s", category)
	}

	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list filters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list failed: %s", resp.Status)
	}

	var filters []*Filter
	if err := json.NewDecoder(resp.Body).Decode(&filters); err != nil {
		return nil, fmt.Errorf("failed to decode filters: %w", err)
	}

	// Sort filters
	switch sortBy {
	case "downloads":
		sort.Slice(filters, func(i, j int) bool {
			return filters[i].Downloads > filters[j].Downloads
		})
	case "rating":
		sort.Slice(filters, func(i, j int) bool {
			return filters[i].Rating > filters[j].Rating
		})
	case "newest":
		sort.Slice(filters, func(i, j int) bool {
			return filters[i].CreatedAt.After(filters[j].CreatedAt)
		})
	case "updated":
		sort.Slice(filters, func(i, j int) bool {
			return filters[i].UpdatedAt.After(filters[j].UpdatedAt)
		})
	}

	return filters, nil
}

// GetFilter gets a specific filter by ID
func (m *Marketplace) GetFilter(id string) (*Filter, error) {
	// Check local cache first
	if filter, exists := m.filters[id]; exists {
		return filter, nil
	}

	// Fetch from registry
	url := fmt.Sprintf("%s/filters/%s", m.registryURL, id)
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get filter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("filter not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get filter failed: %s", resp.Status)
	}

	var filter Filter
	if err := json.NewDecoder(resp.Body).Decode(&filter); err != nil {
		return nil, fmt.Errorf("failed to decode filter: %w", err)
	}

	// Cache locally
	m.filters[id] = &filter

	return &filter, nil
}

// DownloadFilter downloads a filter from the marketplace
func (m *Marketplace) DownloadFilter(id string) (*Filter, error) {
	filter, err := m.GetFilter(id)
	if err != nil {
		return nil, err
	}

	// Download filter content
	url := fmt.Sprintf("%s/filters/%s/download", m.registryURL, id)
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download filter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read filter content: %w", err)
	}

	filter.Content = string(content)

	// Save to local directory
	if err := m.saveFilterLocal(filter); err != nil {
		return nil, fmt.Errorf("failed to save filter locally: %w", err)
	}

	slog.Info("Downloaded filter", "id", id, "name", filter.Name)
	return filter, nil
}

// InstallFilter installs a filter for use
func (m *Marketplace) InstallFilter(id string) error {
	filter, err := m.DownloadFilter(id)
	if err != nil {
		return err
	}

	// Install to filters directory
	installDir := config.FiltersPath()
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create filters directory: %w", err)
	}

	installPath := filepath.Join(installDir, fmt.Sprintf("%s.toml", filter.ID))
	if err := os.WriteFile(installPath, []byte(filter.Content), 0644); err != nil {
		return fmt.Errorf("failed to install filter: %w", err)
	}

	slog.Info("Installed filter", "id", id, "path", installPath)
	return nil
}

// UninstallFilter removes an installed filter
func (m *Marketplace) UninstallFilter(id string) error {
	installPath := filepath.Join(config.FiltersPath(), fmt.Sprintf("%s.toml", id))

	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("filter not installed: %s", id)
	}

	if err := os.Remove(installPath); err != nil {
		return fmt.Errorf("failed to uninstall filter: %w", err)
	}

	slog.Info("Uninstalled filter", "id", id)
	return nil
}

// ListInstalledFilters lists all installed filters
func (m *Marketplace) ListInstalledFilters() ([]*Filter, error) {
	installDir := config.FiltersPath()

	entries, err := os.ReadDir(installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Filter{}, nil
		}
		return nil, fmt.Errorf("failed to read filters directory: %w", err)
	}

	var filters []*Filter
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // Remove .toml
		filter, err := m.loadFilterLocal(id)
		if err != nil {
			slog.Warn("Failed to load filter", "id", id, "error", err)
			continue
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

// UpdateFilter updates an installed filter to the latest version
func (m *Marketplace) UpdateFilter(id string) error {
	// Check if filter is installed
	installed, err := m.IsFilterInstalled(id)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("filter not installed: %s", id)
	}

	// Get latest version from marketplace
	latest, err := m.GetFilter(id)
	if err != nil {
		return err
	}

	// Get local version
	local, err := m.loadFilterLocal(id)
	if err != nil {
		return err
	}

	// Check if update is needed
	if local.Version == latest.Version {
		slog.Info("Filter is up to date", "id", id, "version", local.Version)
		return nil
	}

	// Download and install update
	if err := m.InstallFilter(id); err != nil {
		return err
	}

	slog.Info("Updated filter", "id", id, "from", local.Version, "to", latest.Version)
	return nil
}

// IsFilterInstalled checks if a filter is installed
func (m *Marketplace) IsFilterInstalled(id string) (bool, error) {
	installPath := filepath.Join(config.FiltersPath(), fmt.Sprintf("%s.toml", id))
	_, err := os.Stat(installPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// RateFilter submits a rating for a filter
func (m *Marketplace) RateFilter(id string, rating int, comment string) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	ratingReq := struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment,omitempty"`
	}{
		Rating:  rating,
		Comment: comment,
	}

	data, err := json.Marshal(ratingReq)
	if err != nil {
		return fmt.Errorf("failed to marshal rating: %w", err)
	}

	url := fmt.Sprintf("%s/filters/%s/rate", m.registryURL, id)
	resp, err := m.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to submit rating: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rating submission failed: %s", resp.Status)
	}

	slog.Info("Submitted rating", "filter", id, "rating", rating)
	return nil
}

// PublishFilter publishes a filter to the marketplace
func (m *Marketplace) PublishFilter(filter *Filter) error {
	// Validate filter
	if err := m.validateFilter(filter); err != nil {
		return err
	}

	// Publish to registry
	data, err := json.Marshal(filter)
	if err != nil {
		return fmt.Errorf("failed to marshal filter: %w", err)
	}

	url := fmt.Sprintf("%s/filters", m.registryURL)
	resp, err := m.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to publish filter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("publish failed: %s", resp.Status)
	}

	slog.Info("Published filter", "id", filter.ID, "name", filter.Name)
	return nil
}

// GetCategories returns available filter categories
func (m *Marketplace) GetCategories() ([]string, error) {
	return []string{
		CategoryGeneral,
		CategoryLanguage,
		CategoryTool,
		CategoryFramework,
		CategoryCustom,
	}, nil
}

// GetPopularTags returns popular filter tags
func (m *Marketplace) GetPopularTags(limit int) ([]string, error) {
	url := fmt.Sprintf("%s/tags?limit=%d", m.registryURL, limit)
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get tags failed: %s", resp.Status)
	}

	var tags []string
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %w", err)
	}

	return tags, nil
}

// GetStats returns marketplace statistics
func (m *Marketplace) GetStats() (*MarketplaceStats, error) {
	url := fmt.Sprintf("%s/stats", m.registryURL)
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get stats failed: %s", resp.Status)
	}

	var stats MarketplaceStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &stats, nil
}

// MarketplaceStats holds marketplace statistics
type MarketplaceStats struct {
	TotalFilters    int     `json:"total_filters"`
	TotalDownloads  int     `json:"total_downloads"`
	TotalAuthors    int     `json:"total_authors"`
	AverageRating   float64 `json:"average_rating"`
	VerifiedFilters int     `json:"verified_filters"`
	OfficialFilters int     `json:"official_filters"`
}

// Helper functions

func (m *Marketplace) saveFilterLocal(filter *Filter) error {
	path := filepath.Join(m.localDir, fmt.Sprintf("%s.json", filter.ID))

	data, err := json.MarshalIndent(filter, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m *Marketplace) loadFilterLocal(id string) (*Filter, error) {
	path := filepath.Join(m.localDir, fmt.Sprintf("%s.json", id))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var filter Filter
	if err := json.Unmarshal(data, &filter); err != nil {
		return nil, err
	}

	return &filter, nil
}

func (m *Marketplace) validateFilter(filter *Filter) error {
	if filter.ID == "" {
		return fmt.Errorf("filter ID is required")
	}
	if filter.Name == "" {
		return fmt.Errorf("filter name is required")
	}
	if filter.Content == "" {
		return fmt.Errorf("filter content is required")
	}
	if filter.Author == "" {
		return fmt.Errorf("filter author is required")
	}
	return nil
}
