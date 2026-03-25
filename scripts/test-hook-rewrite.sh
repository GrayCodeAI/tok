#!/bin/bash
# Test suite for tokman-rewrite.sh hook
# Feeds mock JSON through the hook and verifies the rewritten commands.
#
# Usage: bash scripts/test-hook-rewrite.sh
# Requires: tokman, jq

set -euo pipefail

HOOK="${HOOK:-$HOME/.claude/hooks/tokman-rewrite.sh}"
PASS=0
FAIL=0
TOTAL=0

# Colors
GREEN='\033[32m'
RED='\033[31m'
DIM='\033[2m'
RESET='\033[0m'

# Check prerequisites
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required. Install: https://jqlang.github.io/jq/download/" >&2
  exit 1
fi

if ! command -v tokman &>/dev/null; then
  echo "ERROR: tokman is not in PATH. Build with: make build" >&2
  exit 1
fi

if [ ! -f "$HOOK" ]; then
  echo "ERROR: Hook not found at $HOOK" >&2
  echo "Install with: tokman init -g" >&2
  exit 1
fi

# Ensure tokman is enabled for hook to work
ENABLED_DIR="$HOME/.local/share/tokman"
mkdir -p "$ENABLED_DIR"
touch "$ENABLED_DIR/.enabled"

test_rewrite() {
  local description="$1"
  local input_cmd="$2"
  local expected_cmd="$3"  # empty string = expect no rewrite
  TOTAL=$((TOTAL + 1))

  local input_json
  input_json=$(jq -n --arg cmd "$input_cmd" '{"tool_input":{"command":$cmd}}')
  local output
  output=$(echo "$input_json" | bash "$HOOK" 2>/dev/null) || true

  if [ -z "$expected_cmd" ]; then
    # Expect no rewrite (hook exits 0 with no output)
    if [ -z "$output" ]; then
      printf "  ${GREEN}PASS${RESET} %s ${DIM}(no rewrite)${RESET}\n" "$description"
      PASS=$((PASS + 1))
    else
      local actual
      actual=$(echo "$output" | jq -r '.hookSpecificOutput.updatedInput.command // empty' 2>/dev/null || echo "")
      printf "  ${RED}FAIL${RESET} %s\n" "$description"
      printf "       expected: (no rewrite)\n"
      printf "       actual:   %s\n" "$actual"
      FAIL=$((FAIL + 1))
    fi
  else
    local actual
    actual=$(echo "$output" | jq -r '.hookSpecificOutput.updatedInput.command // empty' 2>/dev/null || echo "")
    if [ "$actual" = "$expected_cmd" ]; then
      printf "  ${GREEN}PASS${RESET} %s ${DIM}→ %s${RESET}\n" "$description" "$actual"
      PASS=$((PASS + 1))
    else
      printf "  ${RED}FAIL${RESET} %s\n" "$description"
      printf "       expected: %s\n" "$expected_cmd"
      printf "       actual:   %s\n" "$actual"
      FAIL=$((FAIL + 1))
    fi
  fi
}

echo "============================================"
echo "  TokMan Rewrite Hook Test Suite"
echo "============================================"
echo ""

# ---- Git commands ----
echo "--- Git commands ---"
test_rewrite "git status" "git status" "tokman git status"
test_rewrite "git log --oneline -10" "git log --oneline -10" "tokman git log --oneline -10"
test_rewrite "git diff HEAD" "git diff HEAD" "tokman git diff HEAD"
test_rewrite "git show abc123" "git show abc123" "tokman git show abc123"
test_rewrite "git add ." "git add ." ""

# ---- GitHub CLI ----
echo ""
echo "--- GitHub CLI ---"
test_rewrite "gh pr list" "gh pr list" "tokman gh pr list"
test_rewrite "gh issue view 123" "gh issue view 123" "tokman gh issue view 123"
test_rewrite "gh pr diff 456" "gh pr diff 456" "tokman gh pr diff 456"

# ---- Docker ----
echo ""
echo "--- Docker ---"
test_rewrite "docker ps" "docker ps" "tokman docker ps"
test_rewrite "docker images" "docker images" "tokman docker images"
test_rewrite "docker logs container" "docker logs container" "tokman docker logs container"
test_rewrite "docker run ubuntu" "docker run ubuntu" ""

# ---- Kubernetes ----
echo ""
echo "--- Kubernetes ---"
test_rewrite "kubectl get pods" "kubectl get pods" "tokman kubectl get pods"
test_rewrite "kubectl describe svc" "kubectl describe svc" "tokman kubectl describe svc"

# ---- System commands ----
echo ""
echo "--- System commands ---"
test_rewrite "ls -la" "ls -la" "tokman ls -la"
test_rewrite "find . -name '*.go'" "find . -name '*.go'" "tokman find . -name '*.go'"
test_rewrite "grep -r TODO src/" "grep -r TODO src/" "tokman grep -r TODO src/"
test_rewrite "tree src/" "tree src/" "tokman tree src/"

# ---- Package managers ----
echo ""
echo "--- Package managers ---"
test_rewrite "npm install" "npm install" "tokman npm install"
test_rewrite "npm test" "npm test" "tokman npm test"
test_rewrite "cargo test" "cargo test" "tokman cargo test"
test_rewrite "pip install flask" "pip install flask" "tokman pip install flask"

# ---- Test runners ----
echo ""
echo "--- Test runners ---"
test_rewrite "jest" "jest" "tokman jest"
test_rewrite "pytest" "pytest" "tokman pytest"
test_rewrite "vitest" "vitest" "tokman vitest"
test_rewrite "rspec" "rspec" "tokman rspec"
test_rewrite "rake test" "rake test" "tokman rake test"

# ---- Build tools ----
echo ""
echo "--- Build tools ---"
test_rewrite "tsc --noEmit" "tsc --noEmit" "tokman tsc --noEmit"
test_rewrite "next build" "next build" "tokman next build"
test_rewrite "golangci-lint run" "golangci-lint run" "tokman golangci-lint run"

# ---- AWS ----
echo ""
echo "--- AWS ---"
test_rewrite "aws s3 ls" "aws s3 ls" "tokman aws s3 ls"
test_rewrite "aws sts get-caller-identity" "aws sts get-caller-identity" "tokman aws sts get-caller-identity"

# ---- Linters ----
echo ""
echo "--- Linters ---"
test_rewrite "rubocop" "rubocop" "tokman rubocop"
test_rewrite "ruff check ." "ruff check ." "tokman ruff check ."
test_rewrite "prettier --check ." "prettier --check ." "tokman prettier --check ."
test_rewrite "mypy src/" "mypy src/" "tokman mypy src/"

# ---- Edge cases ----
echo ""
echo "--- Edge cases ---"
test_rewrite "already prefixed" "tokman git status" "tokman git status"
test_rewrite "unknown command" "vim file.txt" ""
test_rewrite "piped command" "cat file.txt | grep foo" ""

# ---- Summary ----
echo ""
echo "============================================"
echo "  Results: $PASS/$TOTAL passed"
if [ "$FAIL" -gt 0 ]; then
  echo "  ${RED}$FAIL FAILED${RESET}"
  echo "============================================"
  exit 1
else
  echo "  ${GREEN}ALL PASSED${RESET}"
  echo "============================================"
  exit 0
fi
