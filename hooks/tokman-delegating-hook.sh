#!/usr/bin/env bash
# tokman-hook-version: 2.0
# TokMan Delegating Hook - All logic in binary, not shell script
#
# This is a thin delegating hook: all rewrite logic lives in `tokman rewrite`,
# which is the single source of truth (internal/commands/core/rewrite.go).
# To add or change rewrite rules, edit the Go code — not this file.
#
# Exit code protocol for `tokman rewrite`:
#   0 + stdout  Rewrite found, no deny/ask rule matched → auto-allow
#   1           No tokman equivalent → pass through unchanged  
#   2           Deny rule matched → pass through (let AI assistant deny)
#   3 + stdout  Ask rule matched → rewrite but let AI assistant prompt user
#   4           Invalid input
#   5           Command is disabled
#   6           Unsafe operation detected
#   7           Resource-intensive operation

set -euo pipefail

# Check dependencies
if ! command -v jq &>/dev/null; then
  echo "[tokman] WARNING: jq is not installed. Hook cannot rewrite commands." >&2
  echo "[tokman] Install jq: https://jqlang.github.io/jq/download/" >&2
  exit 0
fi

if ! command -v tokman &>/dev/null; then
  echo "[tokman] WARNING: tokman is not installed or not in PATH." >&2
  echo "[tokman] Install: https://github.com/GrayCodeAI/tokman#installation" >&2
  exit 0
fi

# Version guard: tokman rewrite was added in 0.1.0
TOKMAN_VERSION=$(tokman --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$TOKMAN_VERSION" ]; then
  MAJOR=$(echo "$TOKMAN_VERSION" | cut -d. -f1)
  MINOR=$(echo "$TOKMAN_VERSION" | cut -d. -f2)
  
  # Require >= 0.1.0
  if [ "$MAJOR" -eq 0 ] && [ "$MINOR" -lt 1 ]; then
    echo "[tokman] WARNING: tokman $TOKMAN_VERSION is too old (need >= 0.1.0)." >&2
    echo "[tokman] Upgrade: brew upgrade tokman or curl ... | sh" >&2
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
REWRITTEN=$(tokman rewrite "$CMD" 2>/dev/null)
EXIT_CODE=$?

case $EXIT_CODE in
  0)
    # Rewrite found, no permission rules matched — safe to auto-allow.
    # If the output is identical, the command was already using tokman.
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
          "permissionDecisionReason": "TokMan auto-rewrite for token optimization",
          "updatedInput": $updated
        }
      }'
    ;;
    
  1)
    # No tokman equivalent — pass through unchanged.
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
