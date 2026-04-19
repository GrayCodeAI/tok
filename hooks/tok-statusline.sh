#!/usr/bin/env bash

set -euo pipefail

CONFIG_DIR="${TOK_CONFIG_DIR:-$HOME/.config/tok}"
FLAG_FILE="$CONFIG_DIR/.tok-active"

if [[ ! -f "$FLAG_FILE" ]]; then
  exit 0
fi

MODE="$(tr -d '[:space:]' < "$FLAG_FILE")"
if [[ -z "$MODE" || "$MODE" == "full" ]]; then
  printf "[TOK]"
  exit 0
fi

printf "[TOK:%s]" "$(echo "$MODE" | tr '[:lower:]' '[:upper:]')"
