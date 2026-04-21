package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/GrayCodeAI/tok/internal/hooks"
)

// HookDiagnosis is the return value of hooks.diagnose — both used as
// a tea.Msg value and as the structured payload returned from the
// action. OK is true when the hook is in a healthy, attributing state.
type HookDiagnosis struct {
	OK      bool
	Summary string
}

// diagnoseHooks walks every environmental signal tok uses to tag a
// command with its agent / provider / model / session and reports what
// it found. The goal is a one-shot explanation the user can act on:
// "your hook isn't active", "TOK_AGENT isn't being set by Claude
// Code", etc.
//
// We deliberately keep this function dep-free of the cobra layer so it
// can run from inside the TUI without re-entering the command pipeline.
func diagnoseHooks() HookDiagnosis {
	var issues []string
	var ok []string

	// 1. Is the hook flag file present? Without it, shells won't
	//    auto-prepend `tok` to wrapped commands and nothing gets
	//    tagged at all.
	if hooks.IsActive() {
		ok = append(ok, "Hook active ("+hooks.GetMode()+" mode)")
	} else {
		issues = append(issues,
			"Hook flag file missing — run `tok init -g` to activate.")
	}

	// 2. Check the env vars the hook relies on to emit context.
	//    TOK_AGENT is the most common — Claude Code sets it, but
	//    piped shells + some terminal muxers strip it.
	for _, v := range []string{"TOK_AGENT", "TOK_PROVIDER", "TOK_MODEL", "TOK_SESSION_ID"} {
		if val := strings.TrimSpace(os.Getenv(v)); val != "" {
			ok = append(ok, fmt.Sprintf("%s=%s", v, val))
		} else {
			issues = append(issues,
				fmt.Sprintf("%s unset — commands won't be tagged with this dimension.", v))
		}
	}

	// 3. Sanity-check the flag path so users can inspect it manually
	//    if they suspect a stale file or permission problem.
	flagPath := hooks.GetFlagPath()
	ok = append(ok, "Flag path: "+flagPath)

	var b strings.Builder
	if len(issues) == 0 {
		b.WriteString("Hook looks healthy.\n")
	} else {
		b.WriteString(fmt.Sprintf("Found %d issue(s):\n", len(issues)))
		for _, i := range issues {
			b.WriteString("• " + i + "\n")
		}
	}
	if len(ok) > 0 {
		b.WriteString("Confirmed:\n")
		for _, o := range ok {
			b.WriteString("✓ " + o + "\n")
		}
	}
	return HookDiagnosis{
		OK:      len(issues) == 0,
		Summary: strings.TrimSpace(b.String()),
	}
}
