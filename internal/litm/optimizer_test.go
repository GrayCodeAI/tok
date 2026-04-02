package litm

import "testing"

func TestLITMOptimizer(t *testing.T) {
	opt := NewLITMOptimizer(100)

	content := "func main() {\n// comment\nreturn 0\nlog.Println(\"hello\")\n}"
	result := opt.Optimize(content)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Content == "" {
		t.Error("Expected non-empty content")
	}
}

func TestLITMOptimizerSmall(t *testing.T) {
	opt := NewLITMOptimizer(1000)

	content := "short\ncontent"
	result := opt.Optimize(content)
	if result.Placement != "full" {
		t.Errorf("Expected full placement for small content, got %s", result.Placement)
	}
}
