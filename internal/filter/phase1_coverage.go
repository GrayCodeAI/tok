//go:build ignore
// +build ignore

// Phase 1: Coverage improvement script
package main

import (
	"fmt"
	"os/exec"
)

func main() {
	// Check current coverage
	cmd := exec.Command("go", "test", "./internal/filter", "-cover")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))
}
