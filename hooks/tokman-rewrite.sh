#!/usr/bin/env bash
# tokman-hook-version: 2
# TokMan Claude Code hook — rewrites commands to use tokman for token savings.
# Requires: tokman >= 0.2.0, jq
#
# This is a thin delegating hook: all rewrite logic lives in `tokman rewrite`,
# which is the single source of truth (internal/discover/registry.go).
# To add or change rewrite rules, edit the Go registry — not this file.

if ! command -v jq &>/dev/null; then
  exit 0
fi

if ! command -v tokman &>/dev/null; then
  exit 0
fi

# Version guard: tokman rewrite was added in 0.2.0.
# Older binaries: warn once and exit cleanly (no silent failure).
TOKMAN_VERSION=$(tokman --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$TOKMAN_VERSION" ]; then
  MAJOR=$(echo "$TOKMAN_VERSION" | cut -d. -f1)
  MINOR=$(echo "$TOKMAN_VERSION" | cut -d. -f2)
  # Require >= 0.2.0
  if [ "$MAJOR" -eq 0 ] && [ "$MINOR" -lt 2 ]; then
    echo "[tokman] WARNING: tokman $TOKMAN_VERSION is too old (need >= 0.2.0). Upgrade: go install github.com/GrayCodeAI/tokman@latest" >&2
    exit 0
  fi
fi

INPUT=$(cat)
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -z "$CMD" ]; then
  exit 0
fi

# Delegate all rewrite logic to the Go binary.
# tokman rewrite exits 1 when there's no rewrite — hook passes through silently.
REWRITTEN=$(tokman rewrite "$CMD" 2>/dev/null) || exit 0

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
      "permissionDecisionReason": "TokMan auto-rewrite",
      "updatedInput": $updated
    }
  }'
