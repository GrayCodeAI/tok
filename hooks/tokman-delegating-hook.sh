#!/usr/bin/env bash
# tok-hook-version: 2.0
# Tok Delegating Hook - All logic in binary, not shell script
#
# This is a thin delegating hook: all rewrite logic lives in `tok rewrite`,
# which is the single source of truth (internal/commands/core/rewrite.go).
# To add or change rewrite rules, edit the Go code — not this file.
#
# Exit code protocol for `tok rewrite`:
#   0 + stdout  Rewrite found, no deny/ask rule matched → auto-allow
#   1           No tok equivalent → pass through unchanged  
#   2           Deny rule matched → pass through (let AI assistant deny)
#   3 + stdout  Ask rule matched → rewrite but let AI assistant prompt user
#   4           Invalid input
#   5           Command is disabled
#   6           Unsafe operation detected
#   7           Resource-intensive operation

set -euo pipefail

# Check dependencies
if ! command -v jq &>/dev/null; then
  echo "[tok] WARNING: jq is not installed. Hook cannot rewrite commands." >&2
  echo "[tok] Install jq: https://jqlang.github.io/jq/download/" >&2
  exit 0
fi

if ! command -v tok &>/dev/null; then
  echo "[tok] WARNING: tok is not installed or not in PATH." >&2
  echo "[tok] Install: https://github.com/GrayCodeAI/tok#installation" >&2
  exit 0
fi

# Version guard: tok rewrite was added in 0.1.0
TOK_VERSION=$(tok --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$TOK_VERSION" ]; then
  MAJOR=$(echo "$TOK_VERSION" | cut -d. -f1)
  MINOR=$(echo "$TOK_VERSION" | cut -d. -f2)
  
  # Require >= 0.1.0
  if [ "$MAJOR" -eq 0 ] && [ "$MINOR" -lt 1 ]; then
    echo "[tok] WARNING: tok $TOK_VERSION is too old (need >= 0.1.0)." >&2
    echo "[tok] Upgrade: brew upgrade tok or curl ... | sh" >&2
    exit 0
  fi
fi

# Read input from stdin
INPUT=$(cat)
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Check if we have a command to process
if [ -z "$CMD" ]; then
  exit 0
fi

# Delegate all rewrite + permission logic to the Rust binary
REWRITTEN=$(tok rewrite "$CMD" 2>/dev/null)
EXIT_CODE=$?

case $EXIT_CODE in
  0)
    # Rewrite found, no permission rules matched — safe to auto-allow.
    # If the output is identical, the command was already using tok.
    [ "$CMD" = "$REWRITTEN" ] && exit 0
    
    # Update the command and auto-allow
    ORIGINAL_INPUT=$(echo "$INPUT" | jq -c '.tool_input')
    UPDATED_INPUT=$(echo "$ORIGINAL_INPUT" | jq --arg cmd "$REWRITTEN" '.command = $cmd')
    
    jq -n \
      --argjson updated "$UPDATED_INPUT" \
      '{
        "hookSpecificOutput": {
          "hookEventName": "PreToolUse",
          "permissionDecision": "allow",
          "permissionDecisionReason": "Tok auto-rewrite for token optimization",
          "updatedInput": $updated
        }
      }'
    ;;
    
  1)
    # No tok equivalent — pass through unchanged.
    exit 0
    ;;
    
  2)
    # Deny rule matched — let AI assistant's native deny rule handle it.
    exit 0
    ;;
    
  3)
    # Ask rule matched — rewrite the command but do NOT auto-allow so that
    # AI assistant prompts the user for confirmation.
    ORIGINAL_INPUT=$(echo "$INPUT" | jq -c '.tool_input')
    UPDATED_INPUT=$(echo "$ORIGINAL_INPUT" | jq --arg cmd "$REWRITTEN" '.command = $cmd')
    
    jq -n \
      --argjson updated "$UPDATED_INPUT" \
      '{
        "hookSpecificOutput": {
          "hookEventName": "PreToolUse",
          "updatedInput": $updated
        }
      }'
    ;;
    
  4|5|6|7)
    # Invalid input, disabled command, unsafe operation, or resource-intensive
    # Pass through unchanged
    exit 0
    ;;
    
  *)
    # Unknown exit code - pass through unchanged
    exit 0
    ;;
esac
