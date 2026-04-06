#!/usr/bin/env bash
# TokMan installer - Cross-platform installation script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh
#
# Or with custom install directory:
#   INSTALL_DIR=/usr/local/bin curl -fsSL ... | sh

set -e

# Configuration
REPO="GrayCodeAI/tokman"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="tokman"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv6l)
            ARCH="arm"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    info "Detected platform: ${OS}-${ARCH}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download and extract binary
install_binary() {
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    info "Downloading TokMan for ${OS}-${ARCH}..."

    # Construct download URL
    local release_url="https://github.com/${REPO}/releases/latest/download/tokman-${OS}-${ARCH}"
    
    if [ "$OS" = "windows" ]; then
        release_url="${release_url}.exe"
        BINARY_NAME="tokman.exe"
    fi

    # Try to download with curl
    if command_exists curl; then
        if ! curl -fsSL -o "${tmp_dir}/${BINARY_NAME}" "$release_url"; then
            # Try with .tar.gz extension (GitHub releases format)
            release_url="https://github.com/${REPO}/releases/latest/download/tokman-${OS}-${ARCH}.tar.gz"
            info "Trying compressed archive..."
            if curl -fsSL "$release_url" | tar -xz -C "$tmp_dir"; then
                success "Downloaded and extracted TokMan"
            else
                error "Failed to download TokMan. Check that releases exist at: $release_url"
            fi
        else
            success "Downloaded TokMan"
        fi
    elif command_exists wget; then
        if ! wget -q -O "${tmp_dir}/${BINARY_NAME}" "$release_url"; then
            release_url="https://github.com/${REPO}/releases/latest/download/tokman-${OS}-${ARCH}.tar.gz"
            info "Trying compressed archive..."
            if wget -q -O - "$release_url" | tar -xz -C "$tmp_dir"; then
                success "Downloaded and extracted TokMan"
            else
                error "Failed to download TokMan"
            fi
        else
            success "Downloaded TokMan"
        fi
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    # Create install directory
    info "Installing to ${INSTALL_DIR}..."
    mkdir -p "$INSTALL_DIR"

    # Copy binary
    cp "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    success "TokMan installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Check if PATH contains install directory
check_path() {
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        warn "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "Add to PATH by running:"
        echo ""
        
        # Detect shell
        if [ -n "$BASH_VERSION" ]; then
            echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
            echo "  source ~/.bashrc"
        elif [ -n "$ZSH_VERSION" ]; then
            echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc"
            echo "  source ~/.zshrc"
        else
            echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        fi
        echo ""
    fi
}

# Verify installation
verify_installation() {
    info "Verifying installation..."
    
    if [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        local version
        version=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null || echo "unknown")
        success "Installation verified! Version: $version"
        return 0
    else
        error "Installation verification failed"
        return 1
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ TokMan installed successfully!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Verify installation:"
    echo "   tokman --version"
    echo ""
    echo "2. Initialize for your AI coding assistant:"
    echo "   tokman init -g               # Claude Code (default)"
    echo "   tokman init --cursor         # Cursor"
    echo "   tokman init --windsurf       # Windsurf"
    echo "   tokman init --copilot        # GitHub Copilot"
    echo ""
    echo "3. Test it out:"
    echo "   tokman git status"
    echo "   tokman stats                 # Show token savings"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo "Discord: Coming soon!"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  TokMan Installer"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    detect_platform
    install_binary
    verify_installation
    check_path
    print_next_steps
}

# Run main
main
