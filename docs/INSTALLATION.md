# 📥 Installation Guide

Multiple installation methods available for TokMan:

---

## 🍺 Homebrew (Recommended for macOS/Linux)

### Install
```bash
# Add TokMan tap (one-time)
brew tap GrayCodeAI/tokman

# Install TokMan
brew install tokman
```

### Update
```bash
brew upgrade tokman
```

### Uninstall
```bash
brew uninstall tokman
```

---

## 🚀 Install Script (Linux/macOS/Windows)

Quick one-liner installation:

```bash
curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh
```

### Custom install directory
```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh
```

### Verify installation
```bash
tokman --version
```

---

## 📦 Pre-built Binaries

Download from [GitHub Releases](https://github.com/GrayCodeAI/tokman/releases/latest):

### Linux
```bash
# AMD64
wget https://github.com/GrayCodeAI/tokman/releases/latest/download/tokman-linux-amd64.tar.gz
tar -xzf tokman-linux-amd64.tar.gz
sudo mv tokman-linux-amd64 /usr/local/bin/tokman

# ARM64
wget https://github.com/GrayCodeAI/tokman/releases/latest/download/tokman-linux-arm64.tar.gz
tar -xzf tokman-linux-arm64.tar.gz
sudo mv tokman-linux-arm64 /usr/local/bin/tokman
```

### macOS
```bash
# Apple Silicon (M1/M2/M3)
wget https://github.com/GrayCodeAI/tokman/releases/latest/download/tokman-darwin-arm64.tar.gz
tar -xzf tokman-darwin-arm64.tar.gz
sudo mv tokman-darwin-arm64 /usr/local/bin/tokman

# Intel (x86_64)
wget https://github.com/GrayCodeAI/tokman/releases/latest/download/tokman-darwin-amd64.tar.gz
tar -xzf tokman-darwin-amd64.tar.gz
sudo mv tokman-darwin-amd64 /usr/local/bin/tokman
```

### Windows
Download `tokman-windows-amd64.zip` from releases and extract to your PATH.

---

## 🔨 Build from Source

### Prerequisites
- Go 1.21 or later
- make (optional, but recommended)

### Clone and build
```bash
git clone https://github.com/GrayCodeAI/tokman.git
cd tokman
make build
```

### Install locally
```bash
make install  # Installs to ~/.local/bin
```

### Or install globally
```bash
sudo make install-global  # Installs to /usr/local/bin
```

### Build for all platforms
```bash
make build-all
```

---

## 🐹 Go Install

If you have Go installed:

```bash
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
```

---

## 🔐 Verify Installation

### Check version
```bash
tokman --version
```

### Run help
```bash
tokman --help
```

### Test with a command
```bash
tokman git status
```

---

## ⚙️ Setup AI Assistant Integration

After installation, initialize TokMan for your AI coding assistant:

### Claude Code (default)
```bash
tokman init -g
```

### Other assistants
```bash
tokman init --cursor      # Cursor
tokman init --windsurf    # Windsurf
tokman init --copilot     # GitHub Copilot
tokman init --cline       # Cline
tokman init --codex       # Codex
tokman init --all         # All detected assistants
```

### Restart your AI assistant
After running `init`, restart your AI coding assistant for changes to take effect.

---

## 🔄 Updates

### Homebrew
```bash
brew upgrade tokman
```

### Install script
Re-run the install script:
```bash
curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh
```

### From source
```bash
cd tokman
git pull
make build
make install
```

---

## 🗑️ Uninstallation

### Homebrew
```bash
brew uninstall tokman
```

### Manual
```bash
# Remove binary
rm ~/.local/bin/tokman  # or /usr/local/bin/tokman

# Remove configuration (optional)
rm -rf ~/.config/tokman

# Remove AI assistant hooks (if installed)
tokman init --remove
```

---

## 🐚 Shell Completions

TokMan supports shell completions for Bash, Zsh, Fish, and PowerShell.

### Bash
```bash
# Load once
source <(tokman completion bash)

# Add to ~/.bashrc for permanent effect
echo 'source <(tokman completion bash)' >> ~/.bashrc
```

### Zsh
```bash
# Load once
source <(tokman completion zsh)

# Add to ~/.zshrc for permanent effect
echo 'source <(tokman completion zsh)' >> ~/.zshrc
```

### Fish
```bash
# Load once
tokman completion fish | source

# Add to ~/.config/fish/config.fish for permanent effect
tokman completion fish > ~/.config/fish/completions/tokman.fish
```

### PowerShell
```powershell
# Load once
tokman completion powershell | Out-String | Invoke-Expression

# Add to profile for permanent effect
tokman completion powershell >> $PROFILE
```

---

## 🆘 Troubleshooting

### "tokman: command not found"

**Solution:** Add install directory to PATH:

```bash
# For ~/.local/bin
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For /usr/local/bin (usually already in PATH)
# Check with: echo $PATH
```

### Permission denied

**Solution:** Make binary executable:
```bash
chmod +x /path/to/tokman
```

### Install script fails to download

**Solution:** Check your internet connection and GitHub availability:
```bash
curl -I https://github.com/GrayCodeAI/tokman/releases/latest
```

### Homebrew installation fails

**Solution:** Update Homebrew and try again:
```bash
brew update
brew install tokman
```

---

## 📚 Next Steps

After installation:

1. **Initialize for your AI assistant:** `tokman init -g`
2. **Test it out:** `tokman git status`
3. **View token savings:** `tokman stats` (or `tokman gain`)
4. **Explore features:** `tokman --help`

---

## 🔗 Links

- **GitHub:** https://github.com/GrayCodeAI/tokman
- **Releases:** https://github.com/GrayCodeAI/tokman/releases
- **Issues:** https://github.com/GrayCodeAI/tokman/issues
- **Discussions:** https://github.com/GrayCodeAI/tokman/discussions

---

<div align="center">

**Need help?** Check our [Troubleshooting Guide](TROUBLESHOOTING.md) or [open an issue](https://github.com/GrayCodeAI/tokman/issues)

</div>
