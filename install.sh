#!/bin/bash
# TokMan Installer Script
# Usage: curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | bash
#
# Options:
#   TOKMAN_VERSION=v0.28.2  Install specific version
#   TOKMAN_DIR=~/.local/bin Install to specific directory

set -euo pipefail

# Configuration
REPO="GrayCodeAI/tokman"
BINARY="tokman"
DEFAULT_DIR="${HOME}/.local/bin"
INSTALL_DIR="${TOKMAN_DIR:-$DEFAULT_DIR}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log()    { echo -e "${GREEN}[tokman]${NC} $*"; }
warn()   { echo -e "${YELLOW}[tokman]${NC} $*"; }
error()  { echo -e "${RED}[tokman]${NC} $*" >&2; }
info()   { echo -e "${BLUE}[tokman]${NC} $*"; }

# Detect OS and architecture
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)   os="linux" ;;
        Darwin*)  os="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        FreeBSD*) os="freebsd" ;;
        *)        error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        i386|i686)     arch="386" ;;
        *)             error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

# Get latest version from GitHub
get_latest_version() {
    if [ -n "${TOKMAN_VERSION:-}" ]; then
        echo "$TOKMAN_VERSION"
        return
    fi

    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        error "Could not determine latest version"
        error "Set TOKMAN_VERSION manually or check https://github.com/${REPO}/releases"
        exit 1
    fi

    echo "$version"
}

# Download and install
install_tokman() {
    local platform version url tmp_dir archive_name

    platform=$(detect_platform)
    version=$(get_latest_version)

    log "Installing TokMan ${version} for ${platform}..."

    # Determine archive format
    local ext="tar.gz"
    if [[ "$platform" == windows_* ]]; then
        ext="zip"
    fi

    archive_name="${BINARY}_${version#v}_${platform}.${ext}"
    url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

    # Create temp directory
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    # Download
    info "Downloading from ${url}..."
    if ! curl -fsSL -o "${tmp_dir}/${archive_name}" "$url"; then
        error "Download failed. Check the URL or try:"
        error "  TOKMAN_VERSION=v0.28.2 bash install.sh"
        exit 1
    fi

    # Verify checksum if available
    local checksum_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"
    if curl -fsSL -o "${tmp_dir}/checksums.txt" "$checksum_url" 2>/dev/null; then
        info "Verifying checksum..."
        local expected actual
        expected=$(grep "${archive_name}" "${tmp_dir}/checksums.txt" | awk '{print $1}')
        if [ -n "$expected" ]; then
            if command -v sha256sum &>/dev/null; then
                actual=$(sha256sum "${tmp_dir}/${archive_name}" | awk '{print $1}')
            elif command -v shasum &>/dev/null; then
                actual=$(shasum -a 256 "${tmp_dir}/${archive_name}" | awk '{print $1}')
            fi

            if [ -n "${actual:-}" ] && [ "$expected" != "$actual" ]; then
                error "Checksum verification failed!"
                error "Expected: ${expected}"
                error "Actual:   ${actual}"
                exit 1
            fi
            log "Checksum verified ✓"
        fi
    fi

    # Extract
    info "Extracting..."
    if [[ "$ext" == "tar.gz" ]]; then
        tar xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"
    else
        unzip -q "${tmp_dir}/${archive_name}" -d "$tmp_dir"
    fi

    # Install
    mkdir -p "$INSTALL_DIR"
    mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"

    log "Installed to ${INSTALL_DIR}/${BINARY}"
}

# Check if install dir is in PATH
check_path() {
    if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
        warn "${INSTALL_DIR} is not in your PATH"
        echo ""
        info "Add it to your shell config:"

        local shell_name
        shell_name=$(basename "${SHELL:-bash}")

        case "$shell_name" in
            bash)
                echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
                echo "  source ~/.bashrc"
                ;;
            zsh)
                echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc"
                echo "  source ~/.zshrc"
                ;;
            fish)
                echo "  fish_add_path ${INSTALL_DIR}"
                ;;
            *)
                echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
                ;;
        esac
        echo ""
    fi
}

# Post-install verification
verify_install() {
    if "${INSTALL_DIR}/${BINARY}" --version &>/dev/null; then
        log "TokMan installed successfully! ✓"
        echo ""
        info "Version: $("${INSTALL_DIR}/${BINARY}" --version 2>&1 || echo 'unknown')"
        echo ""
        log "Quick start:"
        echo "  tokman init -g          # Set up for Claude Code"
        echo "  tokman doctor           # Verify installation"
        echo "  tokman --help           # See all commands"
        echo ""
        info "Documentation: https://github.com/${REPO}"
    else
        error "Installation verification failed"
        error "Binary exists at: ${INSTALL_DIR}/${BINARY}"
        error "Try running: ${INSTALL_DIR}/${BINARY} --version"
        exit 1
    fi
}

# Main
main() {
    echo ""
    echo "  ╔══════════════════════════════════════╗"
    echo "  ║        TokMan Installer              ║"
    echo "  ║  Token-aware CLI proxy for AI tools   ║"
    echo "  ╚══════════════════════════════════════╝"
    echo ""

    # Check for required tools
    if ! command -v curl &>/dev/null; then
        error "curl is required but not installed"
        exit 1
    fi

    install_tokman
    check_path
    verify_install
}

main "$@"
