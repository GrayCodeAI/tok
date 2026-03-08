#!/usr/bin/env bash
# tokman-hook-version: 1
# TokMan Command Rewriter for Shell Integration
# Source this file in your .bashrc or .zshrc:
#   source /path/to/tokman/hooks/tokman-rewrite.sh
#
# This hook intercepts commands and rewrites them to use TokMan wrappers
# when appropriate, reducing token usage in LLM interactions.

# Path to tokman binary (auto-detect)
_TOKMAN_BIN="${TOKMAN_BIN:-tokman}"

# Check if tokman is available
if ! command -v "$_TOKMAN_BIN" &> /dev/null; then
    # TokMan not found, skip rewriting
    return 0 2>/dev/null || exit 0
fi

# Version guard: tokman rewrite requires >= 0.1.0
_TOKMAN_VERSION=$("$_TOKMAN_BIN" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$_TOKMAN_VERSION" ]; then
    _TOKMAN_MAJOR=$(echo "$_TOKMAN_VERSION" | cut -d. -f1)
    _TOKMAN_MINOR=$(echo "$_TOKMAN_VERSION" | cut -d. -f2)
    if [ "$_TOKMAN_MAJOR" -eq 0 ] && [ "$_TOKMAN_MINOR" -lt 1 ]; then
        echo "[tokman] WARNING: tokman $_TOKMAN_VERSION is too old (need >= 0.1.0). Upgrade: go install github.com/GrayCodeAI/tokman@latest" >&2
        return 0 2>/dev/null || exit 0
    fi
fi

# ============================================
# ZSH-Specific Integration
# ============================================
if [[ -n "$ZSH_VERSION" ]]; then
    # ZSH: Use precmd for post-command tracking
    _tokman_zsh_precmd() {
        local last_cmd="${history[$HISTCMD]}"
        
        # Skip if empty or tokman command itself
        if [[ -z "$last_cmd" ]] || [[ "$last_cmd" == tokman* ]]; then
            return 0
        fi
        
        # Ask tokman for rewrite (silent unless debug)
        if [[ -n "$TOKMAN_DEBUG" ]]; then
            local rewritten
            rewritten=$("$_TOKMAN_BIN" rewrite "$last_cmd" 2>/dev/null)
            if [[ -n "$rewritten" ]] && [[ "$rewritten" != "$last_cmd" ]]; then
                echo "[tokman] Would rewrite: $last_cmd → $rewritten" >&2
            fi
        fi
    }
    
    # ZSH: Use preexec for pre-command interception
    _tokman_zsh_preexec() {
        # Store command for potential rewriting
        _TOKMAN_LAST_CMD="$1"
    }
    
    # Register ZSH hooks
    autoload -Uz add-zsh-hook
    add-zsh-hook precmd _tokman_zsh_precmd
    add-zsh-hook preexec _tokman_zsh_preexec
    
    # ZSH-specific completion support
    _tokman_zsh_completion() {
        local -a commands
        commands=(
            'init:Initialize TokMan configuration'
            'status:Show token savings summary'
            'rewrite:Rewrite a command to use TokMan wrappers'
            'verify:Verify hook integrity'
            'economics:Show spending vs savings analysis'
            'ls:List directory with filtered output'
            'git:Git commands with filtered output'
            'gh:GitHub CLI with filtered output'
            'docker:Docker CLI with filtered output'
            'kubectl:Kubernetes CLI with filtered output'
        )
        _describe 'command' commands
    }
    
    compdef _tokman_zsh_completion tokman

# ============================================
# BASH Integration
# ============================================
elif [[ -n "$BASH_VERSION" ]]; then
    # BASH: Use PROMPT_COMMAND for post-command tracking
    _tokman_bash_prompt_command() {
        local last_cmd
        last_cmd=$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')
        
        # Skip if empty or tokman command itself
        if [[ -z "$last_cmd" ]] || [[ "$last_cmd" == tokman* ]]; then
            return 0
        fi
        
        # Ask tokman for rewrite (silent unless debug)
        if [[ -n "$TOKMAN_DEBUG" ]]; then
            local rewritten
            rewritten=$("$_TOKMAN_BIN" rewrite "$last_cmd" 2>/dev/null)
            if [[ -n "$rewritten" ]] && [[ "$rewritten" != "$last_cmd" ]]; then
                echo "[tokman] Would rewrite: $last_cmd → $rewritten" >&2
            fi
        fi
    }
    
    # Append to existing PROMPT_COMMAND
    if [[ -z "$PROMPT_COMMAND" ]]; then
        PROMPT_COMMAND="_tokman_bash_prompt_command"
    else
        PROMPT_COMMAND="${PROMPT_COMMAND};_tokman_bash_prompt_command"
    fi
    
    # BASH completion support
    _tokman_bash_completion() {
        local cur prev commands
        COMPREPLY=()
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
        commands="init status rewrite verify economics ls git gh docker kubectl"
        
        if [[ ${COMP_CWORD} -eq 1 ]]; then
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        fi
    }
    
    complete -F _tokman_bash_completion tokman
fi

# ============================================
# Claude Code JSON Hook Integration
# ============================================
# For Claude Code's PreToolUse hook mechanism
# Reads JSON from stdin, extracts command, rewrites if needed
_tokman_rewrite_json() {
    local json_input
    local command
    local rewritten
    
    # Read JSON from stdin
    json_input=$(cat)
    
    # Extract command field using basic parsing (no jq dependency)
    # Matches: "command": "value"
    command=$(echo "$json_input" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"command"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    
    if [[ -z "$command" ]]; then
        # No command field, return as-is
        echo "$json_input"
        return 0
    fi
    
    # Ask tokman for rewrite
    rewritten=$("$_TOKMAN_BIN" rewrite "$command" 2>/dev/null)
    
    if [[ -n "$rewritten" ]] && [[ "$rewritten" != "$command" ]]; then
        # Replace command in JSON
        echo "$json_input" | sed "s/\"command\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"command\": \"$rewritten\"/"
    else
        # No rewrite, return original
        echo "$json_input"
    fi
}

# ============================================
# Installation and Status Functions
# ============================================

# Installation function - adds hook to shell config
tokman_install_hook() {
    local shell_rc
    local hook_source
    
    # Detect shell config file
    if [[ -n "$ZSH_VERSION" ]]; then
        shell_rc="$HOME/.zshrc"
    elif [[ -n "$BASH_VERSION" ]]; then
        shell_rc="$HOME/.bashrc"
    else
        echo "Unsupported shell" >&2
        return 1
    fi
    
    # Get absolute path to this script
    if [[ -n "$ZSH_VERSION" ]]; then
        hook_source="${0:A}"
    else
        hook_source="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/tokman-rewrite.sh"
    fi
    
    # Check if already installed
    if grep -q "source.*tokman-rewrite.sh" "$shell_rc" 2>/dev/null; then
        echo "✓ TokMan hook already installed in $shell_rc"
        return 0
    fi
    
    # Add to shell config
    echo "" >> "$shell_rc"
    echo "# TokMan shell integration" >> "$shell_rc"
    echo "source \"$hook_source\"" >> "$shell_rc"
    
    echo "✓ TokMan hook installed to $shell_rc"
    echo "  Run 'source $shell_rc' or restart your shell to activate"
}

# Show current status
tokman_status() {
    echo "🌸 TokMan Shell Integration"
    echo "─────────────────────────────"
    echo "Binary: $_TOKMAN_BIN"
    
    if command -v "$_TOKMAN_BIN" &> /dev/null; then
        echo "Status: ✓ Installed"
        echo "Version: $_TOKMAN_VERSION"
        if [[ -n "$ZSH_VERSION" ]]; then
            echo "Shell: ZSH ($ZSH_VERSION)"
            echo "Hooks: precmd, preexec"
        elif [[ -n "$BASH_VERSION" ]]; then
            echo "Shell: BASH ($BASH_VERSION)"
            echo "Hooks: PROMPT_COMMAND"
        fi
        "$_TOKMAN_BIN" rewrite list 2>/dev/null
    else
        echo "Status: ✗ Not found"
        echo "  Set TOKMAN_BIN environment variable to specify path"
    fi
}

# ============================================
# Convenience Aliases
# ============================================
# These are lightweight and only active when tokman is available
if command -v "$_TOKMAN_BIN" &> /dev/null; then
    alias ts='tokman status'
    alias tr='tokman rewrite'
    alias te='tokman economics'
    alias tv='tokman verify'
fi

# Export functions for subshells (bash only)
if [[ -n "$BASH_VERSION" ]]; then
    export -f _tokman_rewrite_json tokman_install_hook tokman_status 2>/dev/null || true
fi

# Print status on source (optional, comment out if undesired)
if [[ -n "$TOKMAN_VERBOSE" ]]; then
    tokman_status
fi