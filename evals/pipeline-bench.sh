#!/usr/bin/env bash
# Repro for docs/LAYERS.md compression numbers.
#
# Builds tok, runs `tok benchmark --mode aggressive --json` against
# fixtures of 18/180/900/5400 lines (1×, 10×, 50×, 300× a realistic
# code-review fixture), and prints the measured reduction alongside
# latency. Use this to verify the numbers the docs claim or to update
# them when the pipeline changes.
#
# Usage:
#     ./evals/pipeline-bench.sh
#     TOK_BIN=/path/to/tok ./evals/pipeline-bench.sh
#
# Fails the run if any fixture regresses below its baseline threshold,
# catching silent compression regressions before they ship.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TOK_BIN="${TOK_BIN:-}"

if [ -z "$TOK_BIN" ]; then
  TOK_BIN="$(mktemp -d)/tok"
  echo "==> building tok → $TOK_BIN"
  (cd "$REPO_ROOT" && go build -o "$TOK_BIN" ./cmd/tok)
fi

FIXTURE_BASE="$(mktemp)"
trap 'rm -f "$FIXTURE_BASE" "${FIXTURE_BASE}".*' EXIT

# Realistic code-review prose that exercises dedup, entropy, and
# structural-collapse layers together. Chosen over Lorem Ipsum so the
# "quality" metric in the pipeline (which penalizes semantic-signal
# loss) exercises its actual logic.
cat > "$FIXTURE_BASE" <<'FIXTURE'
# Code review notes for the auth service

The authentication middleware validates JWT tokens on every incoming API request.
Each request carries an Authorization header with a bearer token. The middleware
checks the token's signature against the public key loaded at service startup,
then verifies the expiry claim has not passed.

Current known issues:
- The middleware does not check the nbf (not-before) claim, which means a token
  signed with a future nbf would be accepted immediately. This is a bug.
- Token revocation is not implemented. A stolen token remains valid until expiry.
- Rate limiting is per-IP, not per-user. A user behind a shared NAT can be
  collectively rate-limited by other users on the same network.

Recommended fixes:
- Validate the nbf claim at auth/validate.go:47 by comparing with time.Now().Unix().
- Introduce a revocation list in Redis with TTL matching the longest token expiry.
- Switch rate limiting to per-user-id, using the token's sub claim as the key.
FIXTURE

# multiplier → minimum acceptable reduction %
declare -a CASES=(
  "1:50"
  "10:75"
  "50:80"
  "300:80"
)

fail=0
printf "%-12s %-10s %-14s %-10s %s\n" lines reduction latency_ms quality verdict
printf "%-12s %-10s %-14s %-10s %s\n" ----- --------- ---------- ------- -------

for case in "${CASES[@]}"; do
  mult="${case%%:*}"
  threshold="${case##*:}"
  fixture="${FIXTURE_BASE}.${mult}"
  python3 -c "import sys; sys.stdout.write(open('$FIXTURE_BASE').read() * $mult)" > "$fixture"
  lines=$(wc -l < "$fixture")

  json=$("$TOK_BIN" benchmark "$fixture" --mode aggressive --json 2>&1)
  read -r reduction latency quality <<< "$(python3 -c "
import sys, json
d = json.loads(sys.stdin.read())
print(f'{d[\"avg_compression\"]:.1f} {d[\"avg_latency\"]:.2f} {d[\"avg_quality\"]:.2f}')
" <<< "$json")"

  verdict="ok"
  if (( $(python3 -c "print(1 if $reduction < $threshold else 0)") )); then
    verdict="REGRESSION (< ${threshold}%)"
    fail=1
  fi

  printf "%-12s %-10s %-14s %-10s %s\n" \
    "$lines" "${reduction}%" "$latency" "$quality" "$verdict"
done

echo
if [ "$fail" -ne 0 ]; then
  echo "pipeline-bench: one or more fixtures regressed below threshold"
  exit 1
fi
echo "pipeline-bench: all fixtures within threshold"
