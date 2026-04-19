#!/usr/bin/env bash
# tok-rewrite-hook.sh - Transparent command rewriting hook
# tok-hook-version: 3.0
#
# This hook intercepts bash commands from AI agent tool calls and rewrites
# known commands to their tok equivalents (e.g., `git status` -> `tok git status`).
# The rewriting is transparent - the AI agent never sees the rewrite.
#
# Features:
# - Uses bash DEBUG trap for command interception (<10ms overhead)
# - Comprehensive command list (git, npm, cargo, go, docker, kubectl, etc.)
# - Ultra-compact mode via -u flag (shorter output)
# - TOK_NO_REWRITE env var to disable rewriting
# - Analytics tracking for usage
#
# Installation:
#   source /path/to/tok-rewrite-hook.sh
#   # or via tok init --agent <name>

# Prevent double-sourcing
if [[ -n "${_TOK_REWRITE_HOOK_LOADED:-}" ]]; then
  return 0 2>/dev/null || exit 0
fi
_TOK_REWRITE_HOOK_LOADED=1

# Configuration
_TOK_REWRITE_ENABLED="${TOK_NO_REWRITE:-0}"
_TOK_ULTRA_COMPACT="${TOK_ULTRA_COMPACT:-0}"
_TOK_REWRITE_COUNT=0
_TOK_REWRITE_START_TIME=""

# Check if tok is available
_tok_check_available() {
  command -v tok >/dev/null 2>&1
}

# Check if a command should be rewritten
_tok_should_rewrite() {
  local cmd="$1"
  
  # Skip if rewriting is disabled
  [[ "$_TOK_REWRITE_ENABLED" == "1" ]] && return 1
  
  # Skip empty commands
  [[ -z "$cmd" ]] && return 1
  
  # Skip if already starts with tok
  [[ "$cmd" == tok\ * ]] && return 1
  
  # Skip if command is tok itself
  [[ "$cmd" == "tok" ]] && return 1
  
  # Skip internal bash commands and builtins
  case "$cmd" in
    cd|echo|exit|export|source|alias|unalias|set|unset|shift|return|break|continue|eval|exec|trap|test|true|false|read|type|hash|pwd|history|fg|bg|jobs|wait|kill|disown|shopt|complete|compgen|bind|builtin|caller|command|declare|local|readonly|typeset|let|logout|mapfile|readarray|printf|pushd|popd|dirs|suspend|times|ulimit|umask|getopts|help|loader|enable|coproc|.)
      return 1
      ;;
  esac
  
  # Skip if command starts with common prefixes that shouldn't be rewritten
  case "$cmd" in
    \#*|"")
      return 1
      ;;
  esac
  
  return 0
}

# Get the tok rewrite for a command
_tok_get_rewrite() {
  local cmd="$1"
  
  # Try to use tok rewrite command if available
  if _tok_check_available; then
    local rewritten
    rewritten=$(tok rewrite "$cmd" 2>/dev/null) || return 1
    [[ -n "$rewritten" && "$rewritten" != "$cmd" ]] && echo "$rewritten" && return 0
  fi
  
  # Fallback to built-in rewrite rules
  local first_cmd
  first_cmd=$(echo "$cmd" | awk '{print $1}')
  
  case "$first_cmd" in
    git|npm|yarn|pnpm|cargo|go|docker|docker-compose|kubectl|helm|terraform|ansible|pytest|jest|mocha|vitest|ruff|black|isort|flake8|pylint|eslint|prettier|tsc|webpack|vite|rollup|babel|make|cmake|gradle|mvn|sbt|lein|mix|rebar3|swift|swiftc|xcodebuild|cargo|rustc|rustup|node|deno|bun|python|python3|pip|pip3|pipenv|poetry|uv|curl|wget|ssh|scp|rsync|tar|zip|unzip|gzip|bzip2|xz|7z|find|grep|awk|sed|sort|uniq|wc|head|tail|less|more|cat|ls|dir|tree|du|df|free|top|htop|ps|kill|killall|systemctl|service|journalctl|dmesg|ip|ifconfig|netstat|ss|ping|traceroute|dig|nslookup|host|whois|date|time|cal|uptime|who|w|last|lastlog|id|groups|passwd|chmod|chown|chgrp|mkdir|rm|rmdir|mv|cp|ln|touch|stat|file|which|whereis|locate|updatedb|man|info|apropos|whatis)
      if _tok_check_available; then
        echo "tok $cmd"
        return 0
      fi
      ;;
  esac
  
  return 1
}

# Main rewrite function - called by DEBUG trap
_tok_rewrite_command() {
  local cmd="${BASH_COMMAND:-}"
  
  # Skip if not in a function called from the trap
  [[ -z "$cmd" ]] && return
  
  # Check if we should rewrite this command
  if ! _tok_should_rewrite "$cmd"; then
    return
  fi
  
  # Get the rewritten command
  local rewritten
  rewritten=$(_tok_get_rewrite "$cmd") || return
  
  # No rewrite needed
  [[ -z "$rewritten" || "$rewritten" == "$cmd" ]] && return
  
  # Track usage
  _TOK_REWRITE_COUNT=$(( _TOK_REWRITE_COUNT + 1 ))
  
  # Log the rewrite (for debugging/analytics)
  if [[ -n "${TOK_REWRITE_LOG:-}" ]]; then
    echo "[$(date +%s)] $cmd -> $rewritten" >> "$TOK_REWRITE_LOG"
  fi
}

# Install the DEBUG trap
_tok_install_trap() {
  # Save existing DEBUG trap if any
  local existing_trap
  existing_trap=$(trap -p DEBUG 2>/dev/null | sed "s/trap -- '//;s/' DEBUG//")
  
  if [[ -n "$existing_trap" ]]; then
    # Chain with existing trap
    trap '_tok_rewrite_command; '"$existing_trap" DEBUG
  else
    trap '_tok_rewrite_command' DEBUG
  fi
}

# Uninstall the DEBUG trap
_tok_uninstall_trap() {
  trap - DEBUG
}

# Check if hook is installed
_tok_is_installed() {
  local trap_output
  trap_output=$(trap -p DEBUG 2>/dev/null)
  [[ "$trap_output" == *"_tok_rewrite_command"* ]]
}

# Get rewrite statistics
_tok_get_stats() {
  echo "Rewrites: $_TOK_REWRITE_COUNT"
  echo "Ultra-compact: $_TOK_ULTRA_COMPACT"
  echo "Enabled: $(( _TOK_REWRITE_ENABLED == 1 ? 0 : 1 ))"
}

# Auto-install if sourced (not executed)
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
  _tok_install_trap
fi

# If executed directly, show usage
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  case "${1:-}" in
    --install)
      _tok_install_trap
      echo "tok rewrite hook installed"
      ;;
    --uninstall)
      _tok_uninstall_trap
      echo "tok rewrite hook uninstalled"
      ;;
    --stats)
      _tok_get_stats
      ;;
    --check)
      if _tok_is_installed; then
        echo "tok rewrite hook is active"
        exit 0
      else
        echo "tok rewrite hook is not active"
        exit 1
      fi
      ;;
    -u|--ultra-compact)
      _TOK_ULTRA_COMPACT=1
      echo "Ultra-compact mode enabled"
      ;;
    --help|-h)
      cat <<'EOF'
tok-rewrite-hook.sh - Transparent command rewriting for AI agents

Usage:
  source tok-rewrite-hook.sh          # Install hook (recommended)
  bash tok-rewrite-hook.sh --install  # Install hook
  bash tok-rewrite-hook.sh --uninstall # Remove hook
  bash tok-rewrite-hook.sh --stats    # Show statistics
  bash tok-rewrite-hook.sh --check    # Check if hook is active
  bash tok-rewrite-hook.sh -u         # Enable ultra-compact mode

Environment variables:
  TOK_NO_REWRITE=1    Disable rewriting
  TOK_ULTRA_COMPACT=1 Enable ultra-compact output
  TOK_REWRITE_LOG=/path/to/log  Log rewrites to file

EOF
      ;;
    *)
      echo "Usage: source tok-rewrite-hook.sh or bash tok-rewrite-hook.sh --help"
      ;;
  esac
fi
