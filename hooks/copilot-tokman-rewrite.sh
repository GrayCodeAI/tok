#!/usr/bin/env bash
# tok-hook-version: 1
# Tok Copilot hook — rewrites shell commands to use tok for token savings.
# GitHub Copilot PreToolUse hook format: receives JSON on stdin, returns JSON on stdout.
# Supports both VS Code Copilot Chat (updatedInput) and Copilot CLI (deny-with-suggestion).

if ! command -v tok &>/dev/null; then
  echo "[tok] WARNING: tok not in PATH" >&2
  exit 0
fi

INPUT=$(cat)

# Detect VS Code Copilot Chat format (snake_case keys)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty')
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -n "$TOOL_NAME" ] && [ -n "$CMD" ]; then
  case "$TOOL_NAME" in
    runTerminalCommand|Bash|bash)
      REWRITTEN=$(tok rewrite "$CMD" 2>/dev/null) || { echo '{}'; exit 0; }
      if [ "$CMD" != "$REWRITTEN" ] && [ -n "$REWRITTEN" ]; then
        jq -n --arg cmd "$REWRITTEN" '{
          "hookSpecificOutput": {
            "hookEvent": "PreToolUse",
            "updatedInput": { "command": $cmd }
          }
        }'
        exit 0
      fi
      ;;
  esac
  echo '{}'
  exit 0
fi

# Detect Copilot CLI format (camelCase keys)
TOOL_NAME_CAMEL=$(echo "$INPUT" | jq -r '.toolName // empty')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs // empty')

if [ "$TOOL_NAME_CAMEL" = "bash" ] && [ -n "$TOOL_ARGS" ]; then
  CMD=$(echo "$TOOL_ARGS" | jq -r '.command // empty')
  if [ -n "$CMD" ]; then
    REWRITTEN=$(tok rewrite "$CMD" 2>/dev/null) || { echo '{}'; exit 0; }
    if [ "$CMD" != "$REWRITTEN" ] && [ -n "$REWRITTEN" ]; then
      jq -n --arg reason "Token savings: use '$REWRITTEN' instead" '{
        "permissionDecision": "deny",
        "permissionDecisionReason": $reason
      }'
      exit 0
    fi
  fi
fi

echo '{}'
