#!/usr/bin/env bash
# tok-hook-version: 2
# Tok Claude Code hook — rewrites commands to use tok for token savings.
# Requires: tok >= 0.2.0, jq
#
# This is a thin delegating hook: all rewrite logic lives in `tok rewrite`,
# which is the single source of truth (internal/discover/registry.go).
# To add or change rewrite rules, edit the Go registry — not this file.

if ! command -v jq &>/dev/null; then
  exit 0
fi

if ! command -v tok &>/dev/null; then
  exit 0
fi

# Version guard: tok rewrite was added in 0.2.0.
# Older binaries: warn once and exit cleanly (no silent failure).
TOK_VERSION=$(tok --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$TOK_VERSION" ]; then
  MAJOR=$(echo "$TOK_VERSION" | cut -d. -f1)
  MINOR=$(echo "$TOK_VERSION" | cut -d. -f2)
  # Require >= 0.2.0
  if [ "$MAJOR" -eq 0 ] && [ "$MINOR" -lt 2 ]; then
    echo "[tok] WARNING: tok $TOK_VERSION is too old (need >= 0.2.0). Upgrade: go install github.com/GrayCodeAI/tok/cmd/tok@latest" >&2
    exit 0
  fi
fi

INPUT=$(cat)
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -z "$CMD" ]; then
  exit 0
fi

# Delegate all rewrite logic to the Go binary.
# tok rewrite exits 1 when there's no rewrite — hook passes through silently.
REWRITTEN=$(tok rewrite "$CMD" 2>/dev/null) || exit 0

# No change — nothing to do.
if [ "$CMD" = "$REWRITTEN" ]; then
  exit 0
fi

ORIGINAL_INPUT=$(echo "$INPUT" | jq -c '.tool_input')
UPDATED_INPUT=$(echo "$ORIGINAL_INPUT" | jq --arg cmd "$REWRITTEN" '.command = $cmd')

jq -n \
  --argjson updated "$UPDATED_INPUT" \
  '{
    "hookSpecificOutput": {
      "hookEventName": "PreToolUse",
      "permissionDecision": "allow",
      "permissionDecisionReason": "Tok auto-rewrite",
      "updatedInput": $updated
    }
  }'
