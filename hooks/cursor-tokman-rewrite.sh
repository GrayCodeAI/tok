#!/usr/bin/env bash
# tok-hook-version: 1
# Tok Cursor Agent hook — rewrites shell commands to use tok for token savings.
# Works with both Cursor editor and cursor-cli (they share ~/.cursor/hooks.json).
# Cursor preToolUse hook format: receives JSON on stdin, returns JSON on stdout.
# Requires: tok, jq

if ! command -v jq &>/dev/null; then
  echo "[tok] WARNING: jq is not installed. Hook cannot rewrite commands." >&2
  exit 0
fi

if ! command -v tok &>/dev/null; then
  echo "[tok] WARNING: tok is not installed or not in PATH." >&2
  exit 0
fi

INPUT=$(cat)
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -z "$CMD" ]; then
  echo '{}'
  exit 0
fi

# Delegate rewrite logic to tok rewrite.
REWRITTEN=$(tok rewrite "$CMD" 2>/dev/null) || { echo '{}'; exit 0; }

if [ "$CMD" = "$REWRITTEN" ]; then
  echo '{}'
  exit 0
fi

jq -n --arg cmd "$REWRITTEN" '{
  "permission": "allow",
  "updated_input": { "command": $cmd }
}'
