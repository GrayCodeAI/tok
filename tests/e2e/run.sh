#!/usr/bin/env bash
# tests/e2e/run.sh — driver for Docker-based end-to-end tests.
#
# Runs each fixture/<name>/ scenario: builds its Dockerfile, executes
# case.sh inside the container, pipes the output through `tok`, and
# diffs against golden/<name>.out.
#
# Usage:
#   run.sh                 all scenarios
#   run.sh name1 name2     subset
#   TOK_E2E_UPDATE=1 run.sh rewrite golden files
#
# Exit: 0 if all scenarios pass, 1 if any fail, 2 on infra error.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
HERE="$ROOT/tests/e2e"
TOK_BIN="${TOK_BIN:-$ROOT/tok}"

if [[ ! -x "$TOK_BIN" ]]; then
  echo "error: tok binary not executable at $TOK_BIN" >&2
  echo "  build with: go build -o tok ./cmd/tok" >&2
  exit 2
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "note: docker not found — skipping e2e suite" >&2
  exit 0
fi

mkdir -p "$HERE/golden"

scenarios=()
if [[ $# -gt 0 ]]; then
  scenarios=("$@")
else
  shopt -s nullglob
  for d in "$HERE/fixtures"/*/; do
    scenarios+=("$(basename "$d")")
  done
  shopt -u nullglob
fi

if [[ ${#scenarios[@]} -eq 0 ]]; then
  echo "no scenarios under $HERE/fixtures — exiting 0"
  exit 0
fi

pass=0
fail=0
for s in "${scenarios[@]}"; do
  dir="$HERE/fixtures/$s"
  if [[ ! -d "$dir" ]]; then
    echo "SKIP $s (no fixture dir)"
    continue
  fi

  img="tok-e2e-$s:$(date +%s)"
  docker build -q -t "$img" "$dir" >/dev/null

  raw=$(docker run --rm "$img")
  filtered=$(printf '%s\n' "$raw" | "$TOK_BIN" --mode full)

  golden="$HERE/golden/$s.out"
  if [[ "${TOK_E2E_UPDATE:-}" = "1" ]]; then
    printf '%s\n' "$filtered" > "$golden"
    echo "UPDATE $s"
    continue
  fi

  if [[ ! -f "$golden" ]]; then
    echo "MISS $s (no golden; run with TOK_E2E_UPDATE=1 to seed)"
    fail=$((fail + 1))
    continue
  fi

  if diff -u "$golden" <(printf '%s\n' "$filtered") >/dev/null; then
    echo "PASS $s"
    pass=$((pass + 1))
  else
    echo "FAIL $s"
    diff -u "$golden" <(printf '%s\n' "$filtered") | head -40
    fail=$((fail + 1))
  fi

  docker image rm -f "$img" >/dev/null 2>&1 || true
done

echo
echo "$pass pass, $fail fail"
[[ $fail -eq 0 ]]
