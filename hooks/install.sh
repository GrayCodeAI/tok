#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNIPPET_START="# >>> tok statusline >>>"
SNIPPET_END="# <<< tok statusline <<<"
SNIPPET_BODY='if command -v tok >/dev/null 2>&1; then export PS1="$('"$SCRIPT_DIR"'/tok-statusline.sh 2>/dev/null) $PS1"; fi'

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

install_in_file "$HOME/.zshrc"
install_in_file "$HOME/.bashrc"
echo "Done. Restart shell or source your rc file."
