package filter

import (
	"strings"
	"testing"
)

func TestNewBodyFilter(t *testing.T) {
	f := NewBodyFilter()
	if f == nil {
		t.Fatal("NewBodyFilter returned nil")
	}
	if f.Name() != "body" {
		t.Errorf("Name() = %q, want 'body'", f.Name())
	}
}

func TestBodyFilter_Apply_None(t *testing.T) {
	f := NewBodyFilter()
	input := "func main() { return 42 }"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestBodyFilter_Apply_Minimal(t *testing.T) {
	f := NewBodyFilter()
	input := "func main() { return 42 }"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("ModeMinimal should not modify input (only aggressive strips bodies)")
	}
	if saved != 0 {
		t.Errorf("ModeMinimal should save 0, got %d", saved)
	}
}

func TestBodyFilter_Apply_Aggressive_Go(t *testing.T) {
	f := NewBodyFilter()
	input := `package main

func main() {
	fmt.Println("hello")
	fmt.Println("world")
}

func helper() {
	x := 1
	y := 2
	z := x + y
}
`
	output, saved := f.Apply(input, ModeAggressive)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	// Should preserve function signatures
	if !strings.Contains(output, "func main()") {
		t.Error("should preserve main function signature")
	}
	if !strings.Contains(output, "func helper()") {
		t.Error("should preserve helper function signature")
	}
}

func TestBodyFilter_Apply_Aggressive_Python(t *testing.T) {
	f := NewBodyFilter()
	input := `def hello():
    print("hello")
    print("world")
    return True

def goodbye():
    pass
`
	output, saved := f.Apply(input, ModeAggressive)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	if !strings.Contains(output, "def hello():") {
		t.Error("should preserve hello function signature")
	}
}

func TestBodyFilter_Apply_Aggressive_JS(t *testing.T) {
	f := NewBodyFilter()
	input := `function hello() {
    console.log("hello");
    console.log("world");
    return true;
}

const x = () => {
    return 42;
};
`
	output, saved := f.Apply(input, ModeAggressive)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	_ = output
}

func TestBodyFilter_Apply_NonCode(t *testing.T) {
	f := NewBodyFilter()
	input := "This is just plain text without any code patterns.\nIt should not be modified at all."
	output, saved := f.Apply(input, ModeAggressive)
	if output != input {
		t.Error("non-code input should not be modified")
	}
	if saved != 0 {
		t.Errorf("non-code input should save 0, got %d", saved)
	}
}

func TestBodyFilter_SingleLineFunction(t *testing.T) {
	f := NewBodyFilter()
	input := `package main

func add(a, b int) int { return a + b }
`
	output, saved := f.Apply(input, ModeAggressive)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	_ = output
}

func BenchmarkBodyFilter_Apply(b *testing.B) {
	f := NewBodyFilter()
	input := strings.Repeat("func test() {\n    x := 1\n    y := 2\n    z := x + y\n}\n", 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeAggressive)
	}
}
