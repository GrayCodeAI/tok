#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNIPPET_START="# >>> tok statusline >>>"
SNIPPET_END="# <<< tok statusline <<<"
SNIPPET_BODY='if command -v tok >/dev/null 2>&1; then export PS1="$('"$SCRIPT_DIR"'/tok-statusline.sh 2>/dev/null) $PS1"; fi'

CLAUDE_CONFIG_DIR="$HOME/.claude"
CLAUDE_SETTINGS="$CLAUDE_CONFIG_DIR/settings.json"

install_in_file() {
  local rc_file="$1"
  [[ -f "$rc_file" ]] || touch "$rc_file"
  if [[ "$(<"$rc_file")" == *"tok statusline"* ]]; then
    echo "Already configured: $rc_file"
    return
  fi
  {
    echo ""
    echo "$SNIPPET_START"
    echo "$SNIPPET_BODY"
    echo "$SNIPPET_END"
  } >> "$rc_file"
  echo "Configured: $rc_file"
}

install_claude_code_statusline() {
  if [[ ! -d "$CLAUDE_CONFIG_DIR" ]]; then
    echo "Claude Code not detected (no $CLAUDE_CONFIG_DIR)"
    return
  fi

  mkdir -p "$CLAUDE_CONFIG_DIR"

  local existing="{}"
  if [[ -f "$CLAUDE_SETTINGS" ]]; then
    existing="$(cat "$CLAUDE_SETTINGS")"
  fi

  if echo "$existing" | grep -q "tok-statusline"; then
    echo "Claude Code statusline already configured"
    return
  fi

  local statusline_hook
  statusline_hook=$(cat <<'HOOK'
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "Always",
        "hooks": [
          {
            "type": "command",
            "command": "sh -c 'if command -v tok >/dev/null 2>&1; then mkdir -p ~/.config/tok && echo full > ~/.config/tok/.tok-active; fi'"
          }
        ]
      }
    ]
  }
}
HOOK
)

  if echo "$existing" | grep -q '"hooks"'; then
    echo "Merging tok statusline into existing Claude Code settings..."
    python3 -c "
import json, sys
with open('$CLAUDE_SETTINGS') as f:
    cfg = json.load(f)
cfg.setdefault('hooks', {}).setdefault('SessionStart', []).append({
    'matcher': 'Always',
    'hooks': [{
        'type': 'command',
        'command': \"sh -c 'if command -v tok >/dev/null 2>&1; then mkdir -p ~/.config/tok && echo full > ~/.config/tok/.tok-active; fi'\"
    }]
})
with open('$CLAUDE_SETTINGS', 'w') as f:
    json.dump(cfg, f, indent=2)
" 2>/dev/null || {
      echo "Warning: could not merge settings, writing new $CLAUDE_SETTINGS.tok.json"
      echo "$statusline_hook" > "$CLAUDE_SETTINGS.tok.json"
    }
  else
    echo "$statusline_hook" > "$CLAUDE_SETTINGS"
  fi

  echo "Configured: Claude Code statusline ($CLAUDE_SETTINGS)"
}

install_js_mode_hooks() {
  if ! command -v node >/dev/null 2>&1; then
    echo "node not found — skipping JS mode-tracking hooks (install Node.js to enable natural-language activation)"
    return
  fi
  if [[ ! -d "$CLAUDE_CONFIG_DIR" ]]; then
    return
  fi

  local dest="$CLAUDE_CONFIG_DIR/hooks"
  mkdir -p "$dest"
  cp "$SCRIPT_DIR/tok-mode-config.js" "$dest/tok-mode-config.js"
  cp "$SCRIPT_DIR/tok-mode-activate.js" "$dest/tok-mode-activate.js"
  cp "$SCRIPT_DIR/tok-mode-tracker.js" "$dest/tok-mode-tracker.js"
  chmod 0644 "$dest"/tok-mode-*.js

  if command -v python3 >/dev/null 2>&1; then
    python3 - "$CLAUDE_SETTINGS" "$dest" <<'PY' 2>/dev/null || echo "Warning: could not merge JS hook settings"
import json, os, sys
settings, dest = sys.argv[1], sys.argv[2]
if os.path.exists(settings):
    with open(settings) as f:
        cfg = json.load(f)
else:
    cfg = {}
hooks = cfg.setdefault('hooks', {})
def has(group, marker):
    for entry in hooks.get(group, []):
        for h in entry.get('hooks', []):
            if marker in h.get('command', ''):
                return True
    return False
if not has('SessionStart', 'tok-mode-activate.js'):
    hooks.setdefault('SessionStart', []).append({
        'matcher': 'Always',
        'hooks': [{'type': 'command',
                    'command': f'node "{dest}/tok-mode-activate.js"',
                    'timeout': 5}],
    })
if not has('UserPromptSubmit', 'tok-mode-tracker.js'):
    hooks.setdefault('UserPromptSubmit', []).append({
        'hooks': [{'type': 'command',
                    'command': f'node "{dest}/tok-mode-tracker.js"',
                    'timeout': 5}],
    })
with open(settings, 'w') as f:
    json.dump(cfg, f, indent=2)
PY
    echo "Configured: JS mode-tracking hooks ($dest)"
  fi
}

install_in_file "$HOME/.zshrc"
install_in_file "$HOME/.bashrc"
install_claude_code_statusline
install_js_mode_hooks

echo "Done. Restart shell or source your rc file."
echo ""
echo "Statusline badges:"
echo "  [TOK]      — full mode active"
echo "  [TOK:LITE] — lite mode active"
echo "  [TOK:ULTRA]— ultra mode active"
echo "  (nothing)  — tok inactive"
