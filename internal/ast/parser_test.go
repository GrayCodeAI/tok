package ast

import (
	"go/ast"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser(LangGo)
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if p.language != LangGo {
		t.Errorf("lang = %d, want %d", p.language, LangGo)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		lang     Language
	}{
		{"main.go", LangGo},
		{"lib.rs", LangRust},
		{"script.py", LangPython},
		{"app.js", LangJavaScript},
		{"index.ts", LangTypeScript},
		{"unknown.xyz", LangUnknown},
		{"foo.bar", LangUnknown},
		{".go", LangGo},
	}
	for _, tt := range tests {
		got := DetectLanguage(tt.filename)
		if got != tt.lang {
			t.Errorf("DetectLanguage(%q) = %d, want %d", tt.filename, got, tt.lang)
		}
	}
}

func TestParser_ParseGo(t *testing.T) {
	p := NewParser(LangGo)

	src := `package main

import (
	"fmt"
	"net/http"
)

// MyStruct is a test struct
type MyStruct struct {
	Name string
	Age  int
}

// MyFunc does something
func MyFunc(x int) int {
	return x + 1
}

const MyConst = 42

var myVar = "hello"
`

	result, err := p.Parse("main.go", []byte(src))
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if result.Language != LangGo {
		t.Errorf("Language = %d, want %d", result.Language, LangGo)
	}
	if result.Package != "main" {
		t.Errorf("Package = %q, want %q", result.Package, "main")
	}
	if len(result.Imports) != 2 {
		t.Errorf("Imports = %d, want 2", len(result.Imports))
	}
	if len(result.Functions) != 1 {
		t.Errorf("Functions = %d, want 1", len(result.Functions))
	}
	if len(result.Types) != 1 {
		t.Errorf("Types = %d, want 1", len(result.Types))
	}
	if len(result.Constants) != 1 {
		t.Errorf("Constants = %d, want 1", len(result.Constants))
	}
	if len(result.Variables) != 1 {
		t.Errorf("Variables = %d, want 1", len(result.Variables))
	}
}

func TestParser_ParseInvalidGo(t *testing.T) {
	p := NewParser(LangGo)

	src := `this is not valid go code {{{`
	_, err := p.Parse("bad.go", []byte(src))
	if err == nil {
		t.Error("expected parse error for invalid Go code")
	}
}

func TestParser_ParseUnknownLang(t *testing.T) {
	p := NewParser(LangUnknown)

	src := `func main() {}`
	result, err := p.Parse("unknown.xyz", []byte(src))
	if result != nil || err != nil {
		t.Error("expected (nil, nil) for unknown language")
	}
}

func TestParser_ExtractComments(t *testing.T) {
	src := `package test

// TopLevelComment describes the package
type Foo struct{}

// FunctionComment describes the function
func Bar() {}
`
	p := NewParser(LangGo)
	result, err := p.Parse("test.go", []byte(src))
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if len(result.Comments) < 2 {
		t.Errorf("Comments = %d, want >= 2", len(result.Comments))
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && searchStr(s, sub)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Verify AST integration works

type visitorFunc func(ast.Node) ast.Visitor

func (f visitorFunc) Visit(n ast.Node) ast.Visitor {
	return f(n)
}
