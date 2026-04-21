package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
)

var cheatsheetCmd = &cobra.Command{
	Use:     "cheatsheet",
	Aliases: []string{"modes", "quickref"},
	Short:   "One-shot reference card for tok compression modes and commands",
	Long: `Print a terse reference card covering:
  • the six compression modes (lite/full/ultra/wenyan-*)
  • the skill family (commit, review, compress, help)
  • activation from env and config file

Non-persistent — does not change mode or write flag files.`,
	RunE: runCheatsheet,
}

const cheatsheetBody = `
TOK CHEATSHEET
==============

MODES
-----
  lite          drop filler. keep grammar.
  full          drop articles + filler + hedging. fragments OK. (default)
  ultra         extreme compression. abbreviations. arrows for causality.
  wenyan-lite   classical register, filler stripped, grammar kept.
  wenyan-full   fragments + arrow causality + tech-term abbreviations.
  wenyan-ultra  extreme abbreviation, no connectives, symbolic chains.

  Env var:   export TOK_DEFAULT_MODE=ultra
  Config:    ~/.config/tok/config.json  →  {"defaultMode":"lite"}
  Resolution order: env > config file > "full"

COMMANDS YOU INVOKE
-------------------
  tok <cmd> ...           wrap any supported CLI (git, kubectl, npm, etc.)
  tok md <file>           compress a markdown file in place; --mode=<mode>
  tok md <file> --restore revert a compressed file from its .original.md
  tok commit-msg          read staged diff, emit Conventional Commits subject
  tok review-diff         scan a diff, emit one-line review comments
  tok learn               inspect/export the learned-rule database
  tok savings             token-savings report
  tok doctor              diagnose filter TOML files
  tok cheatsheet          this card

SKILL FAMILY (Claude Code plugin)
---------------------------------
  /tok [mode]          set persistent compression mode
  /tok-commit          caveman-style commit-message persona
  /tok-review          caveman-style PR-review persona
  /tok:compress <file> invoke tok-compress skill (calls tok md)
  /tok-help            long-form help card

DEACTIVATE
----------
  "stop tok" / "normal mode"   or unset TOK_DEFAULT_MODE
  /tok off                     turn off for the session

MORE
----
  Plugin authoring:  docs/plugin-dev.md
  Benchmarks:        evals/bench.sh
  Source:            https://github.com/GrayCodeAI/tok
`

func runCheatsheet(cmd *cobra.Command, args []string) error {
	fmt.Print(cheatsheetBody)
	return nil
}

func init() {
	registry.Add(func() { registry.Register(cheatsheetCmd) })
}
