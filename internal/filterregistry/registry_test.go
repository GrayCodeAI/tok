package filterregistry

import "testing"

func TestNewFilterRegistry(t *testing.T) {
	reg := NewFilterRegistry()
	if reg == nil {
		t.Fatal("Expected non-nil registry")
	}
}

func TestPublishFilter(t *testing.T) {
	reg := NewFilterRegistry()

	err := reg.Publish(&FilterEntry{
		ID:      "test-filter",
		Name:    "Test Filter",
		Content: "match = '^test'",
		Tags:    []string{"test"},
	})
	if err != nil {
		t.Fatalf("Publish error: %v", err)
	}
}

func TestInstallFilter(t *testing.T) {
	reg := NewFilterRegistry()

	reg.Publish(&FilterEntry{
		ID:      "my-filter",
		Name:    "My Filter",
		Content: "match = '^my-tool'",
	})

	filter, err := reg.Install("my-filter")
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if filter.Name != "My Filter" {
		t.Errorf("Expected 'My Filter', got %s", filter.Name)
	}
}

func TestSafetyCheck(t *testing.T) {
	errors := SafetyCheck("os.execute('bad')")
	if len(errors) == 0 {
		t.Error("Expected safety errors for dangerous content")
	}

	errors = SafetyCheck("match = '^safe'")
	if len(errors) != 0 {
		t.Error("Expected no errors for safe content")
	}
}
