#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACT_DIR="${ROOT_DIR}/artifacts"
REPORT_PATH="${ARTIFACT_DIR}/eval-adaptive-report.md"
ITERATIONS="${1:-5}"

mkdir -p "${ARTIFACT_DIR}"

echo "Running adaptive evaluation with ${ITERATIONS} iterations per sample..."
(
  cd "${ROOT_DIR}"
  GOCACHE="${ROOT_DIR}/.cache/go-build" go run ./cmd/eval-adaptive -iterations "${ITERATIONS}" > "${REPORT_PATH}"
)

echo "Adaptive report written to: ${REPORT_PATH}"
echo
sed -n '1,80p' "${REPORT_PATH}"

