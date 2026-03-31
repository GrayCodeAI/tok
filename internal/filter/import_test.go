package filter

import (
	"strings"
	"testing"
)

func TestNewImportFilter(t *testing.T) {
	f := NewImportFilter()
	if f == nil {
		t.Fatal("NewImportFilter returned nil")
	}
	if f.Name() != "import" {
		t.Errorf("Name() = %q, want 'import'", f.Name())
	}
}

func TestImportFilter_Apply_NonCode(t *testing.T) {
	f := NewImportFilter()
	input := "just some plain text without imports"
	output, saved := f.Apply(input, ModeMinimal)
	if output != input {
		t.Error("non-code input should not be modified")
	}
	if saved != 0 {
		t.Errorf("non-code input should save 0, got %d", saved)
	}
}

func TestImportFilter_Apply_GoImports(t *testing.T) {
	f := NewImportFilter()
	input := `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("hello")
}
`
	output, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	if !strings.Contains(output, "imports condensed") {
		t.Error("should condense import block")
	}
	if !strings.Contains(output, "func main()") {
		t.Error("should preserve function body")
	}
}

func TestImportFilter_Apply_PythonImports(t *testing.T) {
	f := NewImportFilter()
	input := `import os
import sys
from pathlib import Path

def main():
    print("hello")
`
	output, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	if !strings.Contains(output, "imports condensed") {
		t.Error("should condense python imports")
	}
}

func TestImportFilter_Apply_JSImports(t *testing.T) {
	f := NewImportFilter()
	input := `import React from 'react';
import { useState } from 'react';
const fs = require('fs');

function App() {
    return null;
}
`
	output, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	if !strings.Contains(output, "imports condensed") {
		t.Error("should condense JS imports")
	}
}

func TestImportFilter_Apply_RustImports(t *testing.T) {
	f := NewImportFilter()
	input := `use std::io;
use std::fs::File;
use serde::{Deserialize, Serialize};

fn main() {
    println!("hello");
}
`
	output, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	if !strings.Contains(output, "imports condensed") {
		t.Error("should condense rust imports")
	}
}

func BenchmarkImportFilter_Apply(b *testing.B) {
	f := NewImportFilter()
	input := strings.Repeat("import \"fmt\"\nimport \"os\"\nimport \"strings\"\n", 30) + "\nfunc main() {}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeMinimal)
	}
}
