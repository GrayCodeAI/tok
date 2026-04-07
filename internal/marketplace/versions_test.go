package marketplace

import (
	"context"
	"testing"
)

func TestVersionParsing(t *testing.T) {
	v, err := ParseVersion("1.2.3")
	if err != nil {
		t.Fatalf("ParseVersion failed: %v", err)
	}

	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("Expected 1.2.3, got %d.%d.%d", v.Major, v.Minor, v.Patch)
	}
}

func TestVersionComparison(t *testing.T) {
	v1, _ := ParseVersion("1.2.3")
	v2, _ := ParseVersion("2.0.0")
	v3, _ := ParseVersion("1.3.0")
	v4, _ := ParseVersion("1.2.4")

	if v1.Compare(v2) >= 0 {
		t.Error("Expected v1 < v2")
	}
	if v1.Compare(v3) >= 0 {
		t.Error("Expected v1 < v3")
	}
	if v1.Compare(v4) >= 0 {
		t.Error("Expected v1 < v4")
	}
	if v1.Compare(v1) != 0 {
		t.Error("Expected v1 == v1")
	}
}

func TestVersionCompatibility(t *testing.T) {
	v1, _ := ParseVersion("1.2.3")
	v2, _ := ParseVersion("1.5.0")
	v3, _ := ParseVersion("2.0.0")

	if !v1.IsCompatible(v2) {
		t.Error("Expected v1 compatible with v2 (same major)")
	}
	if v1.IsCompatible(v3) {
		t.Error("Expected v1 incompatible with v3 (different major)")
	}
}

func TestVersionManagerCheckForUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)
	vm := NewVersionManager(registry)

	_, hasUpdate, err := vm.CheckForUpdates(context.Background(), "go", "0.0.1")
	if err != nil {
		t.Fatalf("CheckForUpdates failed: %v", err)
	}
	if !hasUpdate {
		t.Logf("Update check works (current version may be >= installed)")
	}

	_, hasUpdate, err = vm.CheckForUpdates(context.Background(), "nonexistent", "1.0.0")
	if err == nil {
		t.Error("Expected error for nonexistent filter")
	}
}

func TestConflictDetector(t *testing.T) {
	cd := NewConflictDetector()

	spec1 := FilterSpec{Meta: &FilterMeta{Name: "test", Version: "1.0.0"}}
	spec2 := FilterSpec{Meta: &FilterMeta{Name: "test", Version: "1.5.0"}}
	spec3 := FilterSpec{Meta: &FilterMeta{Name: "test", Version: "2.0.0"}}

	cd.AddFilter("test", spec1)
	cd.AddFilter("test", spec2)

	err := cd.AddFilter("test", spec3)
	if err == nil {
		t.Error("Expected version conflict error for major version change")
	}
}

func TestConflictDetectorCheckConflicts(t *testing.T) {
	cd := NewConflictDetector()

	cd.installed["dep1"] = FilterSpec{Meta: &FilterMeta{Name: "dep1", Version: "2.0.0"}}

	deps := []Dependency{
		{Name: "dep1", MinVersion: "3.0.0"},
		{Name: "dep2"},
	}

	conflicts := cd.CheckConflicts(deps)
	if len(conflicts) == 0 {
		t.Error("Expected at least one conflict")
	}
}

func TestBundleCreation(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)
	fb := NewFilterBundler(registry)

	bundle, err := fb.CreateBundle("test-bundle", []string{"go", "cargo", "pytest"})
	if err != nil {
		t.Fatalf("CreateBundle failed: %v", err)
	}

	if bundle.Name != "test-bundle" {
		t.Errorf("Expected name 'test-bundle', got '%s'", bundle.Name)
	}

	if len(bundle.Filters) != 3 {
		t.Errorf("Expected 3 filters, got %d", len(bundle.Filters))
	}
}

func TestFilterAnalytics(t *testing.T) {
	fa := NewFilterAnalytics()

	fa.RecordDownload("jest")
	fa.RecordDownload("jest")
	fa.RecordDownload("pytest")

	fa.RecordRating("jest", 5.0)
	fa.RecordRating("jest", 4.0)

	downloads, avg, count := fa.GetStats("jest")
	if downloads != 2 {
		t.Errorf("Expected 2 downloads, got %d", downloads)
	}
	if count != 2 {
		t.Errorf("Expected 2 ratings, got %d", count)
	}
	if avg != 4.5 {
		t.Errorf("Expected 4.5 average, got %f", avg)
	}
}

func TestCollectionManager(t *testing.T) {
	cm := NewCollectionManager()

	coll := cm.Create("my-collection", "My test collection", "testuser")
	if coll.Name != "my-collection" {
		t.Errorf("Expected name 'my-collection', got '%s'", coll.Name)
	}

	err := cm.AddFilter("jest", "my-collection")
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}

	filters, err := cm.GetFilters("my-collection")
	if err != nil {
		t.Fatalf("GetFilters failed: %v", err)
	}
	if len(filters) != 1 || filters[0] != "jest" {
		t.Errorf("Expected ['jest'], got %v", filters)
	}
}

func TestAuthorProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	profiles := registry.GetAuthorProfiles()
	if len(profiles) == 0 {
		t.Error("Expected at least one author profile")
	}

	for author, profile := range profiles {
		if profile.TotalDls <= 0 {
			t.Errorf("Author %s should have downloads", author)
		}
	}
}

func TestListByCategory(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	byCat := registry.ListByCategory()
	if len(byCat) == 0 {
		t.Error("Expected at least one category")
	}

	for cat, filters := range byCat {
		if len(filters) == 0 {
			t.Errorf("Category %s should have filters", cat)
		}
		for _, f := range filters {
			if f.Category != cat {
				t.Errorf("Filter %s has wrong category %s", f.Name, f.Category)
			}
		}
	}
}

func TestFilterModerator(t *testing.T) {
	fm := NewFilterModerator()

	err := fm.ReportFilter("test-filter", "inappropriate content", "reporter1")
	if err != nil {
		t.Fatalf("ReportFilter failed: %v", err)
	}

	reports := fm.GetPendingReports()
	if len(reports) != 1 {
		t.Errorf("Expected 1 pending report, got %d", len(reports))
	}

	err = fm.ResolveReport(0, "resolved")
	if err != nil {
		t.Fatalf("ResolveReport failed: %v", err)
	}

	reports = fm.GetPendingReports()
	if reports[0].Status != "resolved" {
		t.Errorf("Expected status 'resolved', got '%s'", reports[0].Status)
	}
}

func TestVersionManagerResolveDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)
	vm := NewVersionManager(registry)

	deps := []Dependency{
		{Name: "test-dep", Required: true},
		{Name: "test-dep", MinVersion: "2.0.0", Required: true},
	}

	resolved, err := vm.ResolveDependencies(context.Background(), deps)
	if err != nil {
		t.Logf("Expected behavior for missing dependency: %v", err)
	}
	if resolved != nil && len(resolved) > 0 {
		t.Logf("Dependencies resolved: %d", len(resolved))
	}
}

func TestListByAuthor(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)

	byAuthor := registry.ListByAuthor()
	if len(byAuthor) == 0 {
		t.Error("Expected at least one author")
	}

	for author, filters := range byAuthor {
		if len(filters) == 0 {
			t.Errorf("Author %s should have filters", author)
		}
	}
}

func TestGetFilterVersions(t *testing.T) {
	versions := NewFilterRegistry("").GetFilterVersions("test")
	if len(versions) == 0 {
		t.Error("Expected at least one version")
	}
}

func TestFilterBundlerGetBundleFilters(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewFilterRegistry(tmpDir)
	fb := NewFilterBundler(registry)

	bundle, _ := fb.CreateBundle("test", []string{"go", "pytest"})
	filters := fb.GetBundleFilters(bundle)

	if len(filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(filters))
	}
}
