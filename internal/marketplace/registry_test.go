package marketplace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilterRegistrySearch(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	results, err := registry.Search(context.Background(), SearchOptions{Query: "jest"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result for 'jest'")
	}

	found := false
	for _, r := range results {
		if r.Name == "jest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'jest' in results")
	}
}

func TestFilterRegistrySearchByCategory(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	results, err := registry.Search(context.Background(), SearchOptions{Category: "testing"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result for 'testing' category")
	}

	for _, r := range results {
		if r.Category != "testing" {
			t.Errorf("Expected category 'testing', got '%s'", r.Category)
		}
	}
}

func TestFilterRegistrySearchByTags(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	results, err := registry.Search(context.Background(), SearchOptions{Tags: []string{"python"}})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result for 'python' tag")
	}
}

func TestFilterRegistrySearchMinRating(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	results, err := registry.Search(context.Background(), SearchOptions{MinRating: 4.5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range results {
		if r.Rating < 4.5 {
			t.Errorf("Expected rating >= 4.5, got %f", r.Rating)
		}
	}
}

func TestFilterRegistryGetCategories(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	cats := registry.GetCategories()
	if len(cats) == 0 {
		t.Error("Expected at least one category")
	}

	expectedCats := []string{"build", "container", "infra", "linting", "package", "security", "testing"}
	for _, expected := range expectedCats {
		found := false
		for _, c := range cats {
			if c == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category '%s' not found", expected)
		}
	}
}

func TestFilterRegistryGetTrending(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	trending, err := registry.GetTrending(context.Background())
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	if len(trending) != 5 {
		t.Errorf("Expected 5 trending filters, got %d", len(trending))
	}

	for i := 1; i < len(trending); i++ {
		if trending[i-1].Downloads < trending[i].Downloads {
			t.Error("Should be sorted by downloads descending")
		}
	}
}

func TestFilterRegistryRecommendations(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	installed := []string{"jest", "eslint"}
	recs, err := registry.GetRecommendations(context.Background(), installed)
	if err != nil {
		t.Fatalf("GetRecommendations failed: %v", err)
	}

	for _, r := range recs {
		if r.Name == "jest" || r.Name == "eslint" {
			t.Errorf("Recommendations should not include installed filters")
		}
	}
}

func TestFilterRegistryDownload(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	_, err := registry.DownloadFilter(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent filter")
	}

	_, err = registry.DownloadFilter(context.Background(), "go")
	if err != nil {
		t.Logf("Download failed (expected - URL may not be valid): %v", err)
	}
}

func TestFilterStoreLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	store, err := registry.LoadStore()
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}

	if store.Installed == nil || store.Ratings == nil {
		t.Error("Expected non-nil maps in store")
	}

	store.Installed["test-filter"] = InstalledFilter{
		Name:    "test-filter",
		Version: "1.0.0",
		Source:  "test",
	}

	err = registry.SaveStore(store)
	if err != nil {
		t.Fatalf("SaveStore failed: %v", err)
	}

	loaded, err := registry.LoadStore()
	if err != nil {
		t.Fatalf("LoadStore after save failed: %v", err)
	}

	if _, ok := loaded.Installed["test-filter"]; !ok {
		t.Error("Expected 'test-filter' in loaded store")
	}
}

func TestFilterRegistryRateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	err := registry.RateFilter("jest", 5.0)
	if err != nil {
		t.Fatalf("RateFilter failed: %v", err)
	}

	rating := registry.GetRating("jest")
	if rating != 5.0 {
		t.Errorf("Expected rating 5.0, got %f", rating)
	}
}

func TestFilterRegistryRateFilterInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	err := registry.RateFilter("jest", 6.0)
	if err == nil {
		t.Error("Expected error for invalid rating")
	}

	err = registry.RateFilter("jest", 0)
	if err == nil {
		t.Error("Expected error for invalid rating")
	}
}

func TestFilterRegistrySearchWithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	results, err := registry.Search(context.Background(), SearchOptions{Limit: 3})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) > 3 {
		t.Errorf("Expected max 3 results, got %d", len(results))
	}
}

func TestFilterRegistryAllFiltersHaveVersions(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	for name, f := range registry.builtin {
		if f.Version == "" {
			t.Errorf("Filter '%s' has no version", name)
		}
		if f.Author == "" {
			t.Errorf("Filter '%s' has no author", name)
		}
	}
}

func TestFilterRegistryBuiltinFiltersCount(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	if len(registry.builtin) < 10 {
		t.Errorf("Expected at least 10 builtin filters, got %d", len(registry.builtin))
	}
}

func TestFilterStorePath(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	expectedPath := filepath.Join(tmpDir, "store.json")
	store, err := registry.LoadStore()
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}

	store.Installed["test"] = InstalledFilter{Name: "test", Version: "1.0"}
	if err := registry.SaveStore(store); err != nil {
		t.Fatalf("SaveStore failed: %v", err)
	}

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected store file at %s", expectedPath)
	}
}
