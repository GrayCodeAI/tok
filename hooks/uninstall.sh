#!/usr/bin/env bash

set -euo pipefail

SNIPPET_START="# >>> tok statusline >>>"
SNIPPET_END="# <<< tok statusline <<<"

remove_from_file() {
  local rc_file="$1"
  [[ -f "$rc_file" ]] || return

  if [[ "$(<"$rc_file")" != *"tok statusline"* ]]; then
    return
  fi

  awk -v start="$SNIPPET_START" -v end="$SNIPPET_END" '
    $0 == start { skip=1; next }
    $0 == end { skip=0; next }
    !skip { print }
  ' "$rc_file" > "$rc_file.tmp" && mv "$rc_file.tmp" "$rc_file"
  echo "Removed: $rc_file"
}

remove_from_file "$HOME/.zshrc"
remove_from_file "$HOME/.bashrc"
echo "Done."
