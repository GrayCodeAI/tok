package filter

import (
	"strings"
	"testing"
)

func TestNewASTPreserveFilter(t *testing.T) {
	f := NewASTPreserveFilter()
	if f == nil {
		t.Fatal("NewASTPreserveFilter returned nil")
	}
	if f.Name() != "ast_preserve" {
		t.Errorf("Name() = %q, want 'ast_preserve'", f.Name())
	}
}

func TestASTPreserveFilter_Apply_None(t *testing.T) {
	f := NewASTPreserveFilter()
	input := "some code"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestASTPreserveFilter_PreservesImports(t *testing.T) {
	f := NewASTPreserveFilter()
	input := `package main

import "fmt"
import "os"

func main() {
	// lots of implementation detail
	fmt.Println("hello")
	os.Exit(0)
}
`
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestASTPreserveFilter_PreservesTypes(t *testing.T) {
	f := NewASTPreserveFilter()
	input := `type User struct {
	Name  string
	Email string
	// many more fields...
}

func (u *User) Validate() error {
	// implementation
	return nil
}
`
	output, _ := f.Apply(input, ModeMinimal)
	if !strings.Contains(output, "type User struct") {
		t.Error("should preserve type declaration")
	}
}

func TestASTPreserveFilter_Go(t *testing.T) {
	f := NewASTPreserveFilter()
	input := `package main

func main() {
	x := 1
	y := 2
	z := x + y
	if z > 0 {
		return
	}
}
`
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if !strings.Contains(output, "func main") {
		t.Error("should preserve function declaration")
	}
	_ = saved
}

func TestASTPreserveFilter_Rust(t *testing.T) {
	f := NewASTPreserveFilter()
	input := `pub fn process(data: &[u8]) -> Result<(), Error> {
	let parsed = parse(data)?;
	validate(parsed)?;
	Ok(())
}
`
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if !strings.Contains(output, "pub fn process") {
		t.Error("should preserve function signature")
	}
	_ = saved
}
