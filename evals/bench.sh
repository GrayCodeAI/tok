#!/usr/bin/env bash
# tok benchmark harness.
# Runs a fixture set of commands through: raw, tok, and (optionally) rtk.
# Reports tokens, saved-%, and wall time per engine.
#
# Usage:
#   evals/bench.sh [--no-rtk]
#
# Exit codes:
#   0 success, 1 missing tok, 2 fixtures missing.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TOK="${TOK:-$ROOT/tok}"
RTK="${RTK:-rtk}"
RUN_RTK=1

for arg in "$@"; do
  case "$arg" in
    --no-rtk) RUN_RTK=0 ;;
    -h|--help)
      sed -n '2,12p' "$0"
      exit 0
      ;;
  esac
done

if [[ ! -x "$TOK" ]]; then
  echo "error: tok binary not executable at $TOK" >&2
  exit 1
fi

# Fixtures: label | command producing output.
fixtures=(
  "git-log|git log --oneline -50"
  "git-diff|git diff HEAD~10 HEAD"
  "ls-wide|ls -la /usr/bin"
  "go-vet|true"
  "find-go|find . -name '*.go' -not -path './.gomodcache/*' -not -path './node_modules/*'"
)

est_tokens() {
  # ~4 chars per token heuristic (matches tok/core.EstimateTokens default).
  local n
  n=$(wc -c < "$1")
  echo $(( n / 4 ))
}

time_ms() {
  local t
  t=$(date +%s%N)
  echo $(( t / 1000000 ))
}

printf "%-12s %-10s %10s %10s %8s %8s\n" "fixture" "engine" "bytes" "tokens" "saved%" "ms"
printf "%s\n" "-------------------------------------------------------------------------"

for f in "${fixtures[@]}"; do
  label="${f%%|*}"
  cmd="${f#*|}"

  raw=$(mktemp)
  tok_out=$(mktemp)
  rtk_out=$(mktemp)
  trap 'rm -f "$raw" "$tok_out" "$rtk_out"' EXIT

  # raw
  t0=$(time_ms)
  bash -c "$cmd" > "$raw" 2>&1 || true
  t1=$(time_ms)
  raw_bytes=$(wc -c < "$raw")
  raw_tok=$(est_tokens "$raw")
  printf "%-12s %-10s %10d %10d %8s %8d\n" "$label" "raw" "$raw_bytes" "$raw_tok" "-" "$((t1 - t0))"

  # tok (Unix pipe filter — tok compress)
  t0=$(time_ms)
  "$TOK" compress --mode aggressive < "$raw" > "$tok_out" 2>/dev/null || cp "$raw" "$tok_out"
  t1=$(time_ms)
  tok_bytes=$(wc -c < "$tok_out")
  tok_tok=$(est_tokens "$tok_out")
  if (( raw_tok > 0 )); then
    saved=$(( (raw_tok - tok_tok) * 100 / raw_tok ))
  else
    saved=0
  fi
  printf "%-12s %-10s %10d %10d %7d%% %8d\n" "$label" "tok" "$tok_bytes" "$tok_tok" "$saved" "$((t1 - t0))"

  # rtk (optional)
  if [[ "$RUN_RTK" = "1" ]] && command -v "$RTK" >/dev/null 2>&1; then
    t0=$(time_ms)
    "$RTK" < "$raw" > "$rtk_out" 2>/dev/null || cp "$raw" "$rtk_out"
    t1=$(time_ms)
    rtk_bytes=$(wc -c < "$rtk_out")
    rtk_tok=$(est_tokens "$rtk_out")
    if (( raw_tok > 0 )); then
      rsaved=$(( (raw_tok - rtk_tok) * 100 / raw_tok ))
    else
      rsaved=0
    fi
    printf "%-12s %-10s %10d %10d %7d%% %8d\n" "$label" "rtk" "$rtk_bytes" "$rtk_tok" "$rsaved" "$((t1 - t0))"
  fi

  rm -f "$raw" "$tok_out" "$rtk_out"
  trap - EXIT
  echo
done
