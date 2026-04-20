#!/usr/bin/env bash
# Bundle the skills/ tree into a single tok.skill zip archive.
#
# The .skill file is a flat ZIP consumable by skill registries (Claude Code,
# Codex, etc.) that accept bundled skill packages. Mirrors caveman.skill.
#
# Usage:
#   scripts/build-skill.sh [OUT]
#
#   OUT  Output path. Defaults to tok.skill in the repo root.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${1:-$ROOT/tok.skill}"
SKILLS_DIR="$ROOT/skills"

if [[ ! -d "$SKILLS_DIR" ]]; then
  echo "error: $SKILLS_DIR not found" >&2
  exit 1
fi

zipper=""
if command -v zip >/dev/null 2>&1; then
  zipper="zip"
elif command -v python3 >/dev/null 2>&1; then
  zipper="python"
else
  echo "error: need zip(1) or python3" >&2
  exit 1
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# Stage: copy skills/ tree minus editor cruft.
cp -R "$SKILLS_DIR/." "$tmp/"
find "$tmp" -type f \( -name ".DS_Store" -o -name "*.swp" -o -name "*~" \) -delete

# Include top-level metadata if present.
for f in LICENSE README.md; do
  [[ -f "$ROOT/$f" ]] && cp "$ROOT/$f" "$tmp/"
done

# Write a manifest with version + checksum.
version="$(git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || echo "dev")"
count=$(find "$tmp" -type f | wc -l | tr -d ' ')
cat > "$tmp/MANIFEST" <<EOF
name: tok
version: $version
files: $count
built_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)
EOF

rm -f "$OUT"
if [[ "$zipper" = "zip" ]]; then
  ( cd "$tmp" && zip -qr "$OUT" . )
else
  python3 - "$tmp" "$OUT" <<'PY'
import os, sys, zipfile
src, out = sys.argv[1], sys.argv[2]
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for root, _, files in os.walk(src):
        for f in files:
            full = os.path.join(root, f)
            arc = os.path.relpath(full, src)
            z.write(full, arc)
PY
fi

bytes=$(wc -c < "$OUT" | tr -d ' ')
echo "built $OUT ($bytes bytes, $count files, version $version)"
