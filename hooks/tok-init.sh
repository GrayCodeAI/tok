#!/usr/bin/env bash
# tok-init.sh - Installation script for tok transparent rewriting hook
# tok-init-version: 1.0
#
# Sets up the tok rewrite hook for various AI coding assistants.
# Similar to rtk's `rtk init -g` command.
#
# Usage:
#   tok init -g                    # Global install for all agents
#   tok init --agent claude-code   # Install for specific agent
#   tok init --agent cursor        # Install for Cursor
#   tok init --list                # List supported agents
#   tok init --uninstall           # Remove hook from all agents
#   bash tok-init.sh --agent <name> # Direct script usage

set -eo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOK_HOME="${TOK_HOME:-$(cd "$SCRIPT_DIR/.." && pwd)}"
HOOK_SCRIPT="$TOK_HOME/hooks/tok-rewrite-hook.sh"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Agent configs: name|hook_dir
AGENT_LIST=(
  "claude-code|Claude Code|~/.claude/hooks"
  "cursor|Cursor IDE|~/.cursor/hooks"
  "windsurf|Windsurf IDE|~/.windsurf/hooks"
  "cline|Cline / Roo Code|~/.cline/hooks"
  "roo-code|Roo Code|~/.roo-code/hooks"
  "codex|OpenAI Codex|~/.codex/hooks"
  "gemini|Google Gemini CLI|~/.gemini/hooks"
  "kilocode|Kilo Code|~/.kilocode/hooks"
  "antigravity|Google Antigravity|~/.antigravity/hooks"
  "copilot|GitHub Copilot|project:.github/hooks"
  "opencode|OpenCode|~/.config/opencode/plugins"
  "openclaw|OpenClaw|~/.openclaw/hooks"
)

# Print colored message
print_info() {
  echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
  echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
  echo -e "${RED}✗${NC} $1"
}

# Resolve path (handle ~)
resolve_path() {
  local path="$1"
  if [[ "$path" == ~* ]]; then
    echo "${HOME}${path:1}"
  else
    echo "$path"
  fi
}

# Get agent config by name
get_agent_config() {
  local name="$1"
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    if [[ "$agent_name" == "$name" ]]; then
      echo "$entry"
      return 0
    fi
  done
  return 1
}

# Get agent display name
get_agent_display_name() {
  local name="$1"
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    local rest="${entry#*|}"
    local display="${rest%%|*}"
    if [[ "$agent_name" == "$name" ]]; then
      echo "$display"
      return 0
    fi
  done
  echo "$name"
}

# Get agent hook dir
get_agent_hook_dir() {
  local name="$1"
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    local rest="${entry#*|}"
    rest="${rest#*|}"
    if [[ "$agent_name" == "$name" ]]; then
      echo "$rest"
      return 0
    fi
  done
  return 1
}

# Check if hook is already installed for an agent
is_hook_installed() {
  local agent="$1"
  local hook_dir
  hook_dir=$(resolve_path "$(get_agent_hook_dir "$agent")")
  
  # Check for project-relative paths
  if [[ "$hook_dir" == project:* ]]; then
    hook_dir="./${hook_dir#project:}"
  fi
  
  [[ -f "$hook_dir/tok-rewrite-hook.sh" ]] || [[ -f "$hook_dir/tok-rewrite.sh" ]]
}

# Install hook for an agent
install_hook() {
  local agent="$1"
  local hook_dir
  hook_dir=$(resolve_path "$(get_agent_hook_dir "$agent")")
  
  # Handle project-relative paths
  if [[ "$hook_dir" == project:* ]]; then
    hook_dir="./${hook_dir#project:}"
  fi
  
  # Create hooks directory
  mkdir -p "$hook_dir" || {
    print_warning "Cannot create directory: $hook_dir"
    return 1
  }
  
  # Copy hook script
  local dest="$hook_dir/tok-rewrite-hook.sh"
  if [[ -f "$HOOK_SCRIPT" ]]; then
    cp "$HOOK_SCRIPT" "$dest"
    chmod +x "$dest"
    print_success "Hook installed for $(get_agent_display_name "$agent")"
  else
    print_warning "Hook script not found at $HOOK_SCRIPT"
    return 1
  fi
  
  # Create agent-specific config if needed
  create_agent_config "$agent" "$hook_dir"
  
  return 0
}

# Create agent-specific configuration
create_agent_config() {
  local agent="$1"
  local hook_dir="$2"
  
  case "$agent" in
    claude-code)
      local claude_md="$HOME/.claude/CLAUDE.md"
      if [[ ! -f "$claude_md" ]] || ! grep -q "tok-rewrite-hook" "$claude_md" 2>/dev/null; then
        cat >> "$claude_md" << 'EOF'

## tok Command Rewriting
The tok rewrite hook is installed. Commands are automatically rewritten to use tok for token optimization.
EOF
      fi
      ;;
    copilot)
      local hook_config="$hook_dir/tok-rewrite.json"
      cat > "$hook_config" << 'EOF'
{
  "hooks": {
    "PreToolUse": [
      {
        "type": "command",
        "command": "tok hook copilot",
        "cwd": ".",
        "timeout": 5
      }
    ]
  }
}
EOF
      ;;
  esac
}

# Uninstall hook for an agent
uninstall_hook() {
  local agent="$1"
  local hook_dir
  hook_dir=$(resolve_path "$(get_agent_hook_dir "$agent")")
  
  if [[ "$hook_dir" == project:* ]]; then
    hook_dir="./${hook_dir#project:}"
  fi
  
  local removed=0
  
  for hook_file in "tok-rewrite-hook.sh" "tok-rewrite.sh"; do
    if [[ -f "$hook_dir/$hook_file" ]]; then
      rm "$hook_dir/$hook_file"
      removed=$((removed + 1))
    fi
  done
  
  if [[ $removed -gt 0 ]]; then
    print_success "Hook removed for $(get_agent_display_name "$agent")"
  else
    print_info "No hook found for $(get_agent_display_name "$agent")"
  fi
  
  return 0
}

# List supported agents
list_agents() {
  echo "Supported AI agents:"
  echo ""
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    local rest="${entry#*|}"
    local display="${rest%%|*}"
    local hook_dir_raw="${rest#*|}"
    local hook_dir
    hook_dir=$(resolve_path "$hook_dir_raw")
    
    local status="not detected"
    if [[ "$hook_dir" != project:* ]]; then
      if [[ -d "$hook_dir" ]] || [[ -d "$(dirname "$hook_dir")" ]]; then
        status="detected"
      fi
      if is_hook_installed "$agent_name"; then
        status="hook installed"
      fi
    else
      status="project-level"
    fi
    
    printf "  %-15s %-20s [%s]\n" "$agent_name" "$display" "$status"
  done
  echo ""
  echo "Usage:"
  echo "  tok init --agent <name>    # Install for specific agent"
  echo "  tok init -g                # Install for all detected agents"
  echo "  tok init --uninstall       # Remove from all agents"
}

# Install for all detected agents
install_all() {
  local installed=0
  local failed=0
  
  print_info "Installing tok rewrite hook for all detected agents..."
  echo ""
  
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    local hook_dir_raw
    hook_dir_raw="${entry##*|}"
    local hook_dir
    hook_dir=$(resolve_path "$hook_dir_raw")
    
    # Skip project-relative paths for global install
    if [[ "$hook_dir" == project:* ]]; then
      continue
    fi
    
    # Check if agent directory exists
    local parent_dir
    parent_dir=$(dirname "$hook_dir")
    if [[ -d "$parent_dir" ]] || [[ -d "$hook_dir" ]]; then
      if install_hook "$agent_name"; then
        installed=$((installed + 1))
      else
        failed=$((failed + 1))
      fi
    fi
  done
  
  echo ""
  print_success "Installed for $installed agent(s)"
  [[ $failed -gt 0 ]] && print_warning "$failed installation(s) failed"
}

# Uninstall from all agents
uninstall_all() {
  print_info "Removing tok rewrite hook from all agents..."
  echo ""
  
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    uninstall_hook "$agent_name"
  done
  
  echo ""
  print_success "Cleanup complete"
}

# Show usage
show_usage() {
  cat << EOF
tok-init.sh - Install tok transparent rewriting hook

Usage:
  tok init -g                    # Global install for all detected agents
  tok init --agent <name>        # Install for specific agent
  tok init --uninstall           # Remove hook from all agents
  tok init --list                # List supported agents
  bash tok-init.sh --agent <name> # Direct script usage

Supported agents:
EOF
  
  for entry in "${AGENT_LIST[@]}"; do
    local agent_name="${entry%%|*}"
    local rest="${entry#*|}"
    local display="${rest%%|*}"
    echo "  $agent_name - $display"
  done
  
  echo ""
  echo "Environment variables:"
  echo "  TOK_NO_REWRITE=1    Disable rewriting"
  echo "  TOK_ULTRA_COMPACT=1 Enable ultra-compact output"
}

# Parse arguments
ACTION=""
TARGET_AGENT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -g|--global)
      ACTION="install-all"
      shift
      ;;
    --agent)
      ACTION="install-agent"
      TARGET_AGENT="${2:-}"
      shift 2
      ;;
    --uninstall)
      ACTION="uninstall-all"
      shift
      ;;
    --list)
      ACTION="list"
      shift
      ;;
    --help|-h)
      show_usage
      exit 0
      ;;
    *)
      print_error "Unknown option: $1"
      show_usage
      exit 1
      ;;
  esac
done

# Execute action
case "${ACTION:-}" in
  install-all)
    install_all
    ;;
  install-agent)
    if [[ -z "$TARGET_AGENT" ]]; then
      print_error "Agent name required"
      show_usage
      exit 1
    fi
    
    if ! get_agent_config "$TARGET_AGENT" >/dev/null 2>&1; then
      print_error "Unknown agent: $TARGET_AGENT"
      echo "Supported agents:"
      for entry in "${AGENT_LIST[@]}"; do
        local agent_name="${entry%%|*}"
        echo "  $agent_name"
      done
      exit 1
    fi
    
    install_hook "$TARGET_AGENT"
    ;;
  uninstall-all)
    uninstall_all
    ;;
  list)
    list_agents
    ;;
  *)
    show_usage
    ;;
esac
