package mcptools

import "testing"

func TestMCPToolRegistry(t *testing.T) {
	reg := NewMCPToolRegistry()

	if reg.Count() < 20 {
		t.Errorf("Expected at least 20 tools, got %d", reg.Count())
	}

	tool := reg.Get("ctx_read")
	if tool == nil {
		t.Error("Expected ctx_read tool")
	}

	tool = reg.Get("fetch_clean")
	if tool == nil {
		t.Error("Expected fetch_clean tool")
	}

	results := reg.Search("read")
	if len(results) == 0 {
		t.Error("Expected search results for 'read'")
	}
}

func TestMCPToolRegistrySearch(t *testing.T) {
	reg := NewMCPToolRegistry()

	results := reg.Search("compress")
	if len(results) == 0 {
		t.Error("Expected results for 'compress'")
	}

	results = reg.Search("nonexistent_tool_xyz")
	if len(results) != 0 {
		t.Error("Expected no results for nonexistent tool")
	}
}

func TestEntityProtector(t *testing.T) {
	p := NewEntityProtector()

	protected, data := p.Protect("Buy $AAPL at $150.00 today")
	if protected == "" {
		t.Error("Expected non-empty output")
	}
	_ = data
}
