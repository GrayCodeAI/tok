package smarttool

import (
	"context"
	"regexp"
	"testing"
)

func TestSmartToolManager(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewSmartToolManager(config)

	if len(manager.tools) == 0 {
		t.Error("Expected tools to be registered")
	}

	tool, ok := manager.GetTool("smart_read")
	if !ok {
		t.Error("Expected smart_read tool to exist")
	}

	if tool.Name != "smart_read" {
		t.Errorf("Expected smart_read, got %s", tool.Name)
	}
}

func TestSmartToolExecute(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewSmartToolManager(config)

	result, err := manager.Execute(context.Background(), "smart_tail", []string{"-n", "10", "file.txt"}, ToolOptions{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.LinesKept != 10 {
		t.Errorf("Expected 10 lines, got %d", result.LinesKept)
	}
}

func TestSmartToolIntercept(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewSmartToolManager(config)

	manager.RegisterInterceptor(&Interceptor{
		Pattern:     regexp.MustCompile("git status"),
		Replacement: "git status --short",
		Enabled:     true,
		Priority:    1,
	})

	replaced, shouldReplace := manager.Intercept("git status")
	if !shouldReplace {
		t.Error("Expected interception")
	}

	if replaced != "git status --short" {
		t.Errorf("Expected git status --short, got %s", replaced)
	}
}

func TestSmartToolStats(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewSmartToolManager(config)

	stats := manager.GetStats()

	if stats["total_tools"].(int) == 0 {
		t.Error("Expected non-zero total tools")
	}
}
