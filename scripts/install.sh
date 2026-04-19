#!/usr/bin/env bash
# tok installer — curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tok/main/scripts/install.sh | sh
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tok/main/scripts/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tok/main/scripts/install.sh | sh -s -- --version v0.29.0
#   curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tok/main/scripts/install.sh | sh -s -- --dry-run

set -euo pipefail

# --- Config ---
REPO="GrayCodeAI/tok"
BINARY="tok"
INSTALL_DIR="${HOME}/.local/bin"
GITHUB_API="https://api.github.com/repos/${REPO}"
GITHUB_RELEASES="https://github.com/${REPO}/releases"

# --- Flags ---
VERSION=""
DRY_RUN=false

# --- Parse args ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      echo "Usage: install.sh [--version <tag>] [--dry-run]"
      echo "  --version   Install a specific version (e.g. v0.29.0)"
      echo "  --dry-run   Show what would be done without making changes"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# --- Helpers ---
log()  { echo "==> $*"; }
warn() { echo "WARN: $*" >&2; }
die()  { echo "ERROR: $*" >&2; exit 1; }

dry() {
  if $DRY_RUN; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

# --- Detect OS & Arch ---
detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    darwin)  echo "darwin" ;;
    linux)   echo "linux" ;;
    *)       die "Unsupported OS: $os (only darwin and linux are supported)" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)             die "Unsupported arch: $arch (only amd64 and arm64 are supported)" ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

# --- Resolve version ---
if [[ -z "$VERSION" ]]; then
  log "Fetching latest release..."
  if command -v curl >/dev/null 2>&1; then
    VERSION="$(curl -fsSL "${GITHUB_API}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')"
  elif command -v wget >/dev/null 2>&1; then
    VERSION="$(wget -qO- "${GITHUB_API}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')"
  else
    die "Neither curl nor wget found. Install one and try again."
  fi
  [[ -z "$VERSION" ]] && die "Could not determine latest release version."
fi

log "Installing tok ${VERSION} for ${OS}/${ARCH}..."

# --- Build download URL ---
ASSET="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="${GITHUB_RELEASES}/download/${VERSION}/${ASSET}"

if $DRY_RUN; then
  log "Would download: ${DOWNLOAD_URL}"
  log "Would extract to: ${INSTALL_DIR}/${BINARY}"
  log "Would verify with: ${INSTALL_DIR}/${BINARY} --version"
  exit 0
fi

# --- Ensure install dir exists ---
if [[ ! -d "$INSTALL_DIR" ]]; then
  log "Creating ${INSTALL_DIR}"
  mkdir -p "$INSTALL_DIR"
fi

# --- Download & extract ---
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

log "Downloading ${ASSET}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${TMPDIR}/${ASSET}" "$DOWNLOAD_URL" || die "Download failed. Check that version ${VERSION} exists at ${GITHUB_RELEASES}"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "${TMPDIR}/${ASSET}" "$DOWNLOAD_URL" || die "Download failed. Check that version ${VERSION} exists at ${GITHUB_RELEASES}"
fi

log "Extracting..."
tar -xzf "${TMPDIR}/${ASSET}" -C "$TMPDIR"

# Find the binary (may be in a subdirectory or at top level)
BIN_SRC=""
if [[ -f "${TMPDIR}/${BINARY}" ]]; then
  BIN_SRC="${TMPDIR}/${BINARY}"
elif [[ -f "${TMPDIR}/${BINARY}_${VERSION#v}_${OS}_${ARCH}/${BINARY}" ]]; then
  BIN_SRC="${TMPDIR}/${BINARY}_${VERSION#v}_${OS}_${ARCH}/${BINARY}"
else
  # Try to find it anywhere in the tmpdir
  BIN_SRC="$(find "$TMPDIR" -name "$BINARY" -type f | head -1)"
  [[ -z "$BIN_SRC" ]] && die "Could not find ${BINARY} binary in archive."
fi

chmod +x "$BIN_SRC"
dry mv "$BIN_SRC" "${INSTALL_DIR}/${BINARY}"

# --- Add to PATH if needed ---
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  log "Adding ${INSTALL_DIR} to PATH..."
  SHELL_RC=""
  if [[ -n "${ZSH_VERSION:-}" ]] || [[ "$SHELL" == *zsh ]]; then
    SHELL_RC="${HOME}/.zshrc"
  elif [[ -n "${BASH_VERSION:-}" ]] || [[ "$SHELL" == *bash ]]; then
    SHELL_RC="${HOME}/.bashrc"
  fi

  if [[ -n "$SHELL_RC" ]]; then
    if ! grep -q "${INSTALL_DIR}" "$SHELL_RC" 2>/dev/null; then
      dry sh -c "echo '' >> '${SHELL_RC}'"
      dry sh -c "echo '# Added by tok installer' >> '${SHELL_RC}'"
      dry sh -c "echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> '${SHELL_RC}'"
      log "Added PATH entry to ${SHELL_RC}. Run 'source ${SHELL_RC}' or restart your shell."
    else
      log "${INSTALL_DIR} already in ${SHELL_RC}."
    fi
  else
    warn "Could not detect shell config file. Add this to your shell rc manually:"
    warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
fi

# --- Verify ---
if [[ -x "${INSTALL_DIR}/${BINARY}" ]]; then
  log "Verifying installation..."
  "${INSTALL_DIR}/${BINARY}" --version
  log "tok ${VERSION} installed successfully to ${INSTALL_DIR}/${BINARY}"
else
  die "Installation failed — binary not found at ${INSTALL_DIR}/${BINARY}"
fi
