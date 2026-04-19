// tok - unified token optimization CLI
//
// tok combines input compression (for human-written text) and output filtering
// (for terminal/tool output) into a single CLI tool.
//
// Usage:
//   tok <command> [args]
//
// Examples:
//   tok compress -mode ultra -input "text to compress"
//   tok git status
//   tok doctor
package main

import (
	"fmt"
	"os"

	"github.com/lakshmanpatel/tok/internal/app"
)

func main() {
	app := app.New()
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "tok error: %v\n", err)
		os.Exit(1)
	}
}
