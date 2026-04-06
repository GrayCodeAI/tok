# ⚡ Quick Wins - This Week

**Can complete in 1-2 days for massive impact**

---

## 1. 🍺 Homebrew Formula (2-3 hours)

**Impact:** 🔴 MASSIVE - Makes installation 1 command

### Task:
Create `Formula/tokman.rb`:

```ruby
class Tokman < Formula
  desc "Token-aware CLI proxy with advanced quality analysis"
  homepage "https://github.com/GrayCodeAI/tokman"
  url "https://github.com/GrayCodeAI/tokman/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "CALCULATE_AFTER_TAG"
  license "MIT"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "tokman"
  end

  test do
    assert_match "tokman", shell_output("#{bin}/tokman --version")
  end
end
```

### Steps:
1. Create `Formula/` directory
2. Add `tokman.rb`
3. Tag release: `git tag v0.1.0 && git push --tags`
4. Calculate sha256: `shasum -a 256 v0.1.0.tar.gz`
5. Update formula with sha256
6. Test locally: `brew install --build-from-source Formula/tokman.rb`
7. Create homebrew-tokman tap repository
8. Document in README

---

## 2. 📥 Install Script (3-4 hours)

**Impact:** 🔴 HIGH - Easy cross-platform installation

### Task:
Create `install.sh`:

```bash
#!/bin/bash
set -e

REPO="GrayCodeAI/tokman"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Downloading tokman for ${OS}-${ARCH}..."
RELEASE_URL="https://github.com/${REPO}/releases/latest/download/tokman-${OS}-${ARCH}.tar.gz"

mkdir -p "$INSTALL_DIR"
curl -fsSL "$RELEASE_URL" | tar -xz -C "$INSTALL_DIR"
chmod +x "${INSTALL_DIR}/tokman"

echo "✅ TokMan installed to ${INSTALL_DIR}/tokman"
echo ""
echo "Add to PATH:"
echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
```

### Steps:
1. Create install.sh
2. Set up GitHub release workflow
3. Test on macOS (x86_64, arm64)
4. Test on Linux
5. Add to README

---

## 3. 🔨 Improved Makefile (1 hour)

**Impact:** 🟡 MEDIUM - Better developer experience

### Task:
Update `Makefile`:

```makefile
.PHONY: build build-small build-all test lint install clean

BINARY=tokman
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X 'github.com/GrayCodeAI/tokman/internal/commands/shared.Version=$(VERSION)'"

build:
	CGO_ENABLED=0 go build -o $(BINARY) $(LDFLAGS) ./cmd/tokman

build-small:
	CGO_ENABLED=0 go build -o $(BINARY) $(LDFLAGS) -gcflags="-trimpath" ./cmd/tokman

build-all:
	GOOS=linux GOARCH=amd64 go build -o tokman-linux-amd64 $(LDFLAGS) ./cmd/tokman
	GOOS=linux GOARCH=arm64 go build -o tokman-linux-arm64 $(LDFLAGS) ./cmd/tokman
	GOOS=darwin GOARCH=amd64 go build -o tokman-darwin-amd64 $(LDFLAGS) ./cmd/tokman
	GOOS=darwin GOARCH=arm64 go build -o tokman-darwin-arm64 $(LDFLAGS) ./cmd/tokman
	GOOS=windows GOARCH=amd64 go build -o tokman-windows-amd64.exe $(LDFLAGS) ./cmd/tokman

test:
	go test -cover ./...

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/

clean:
	rm -f $(BINARY) tokman-*
```

---

## 4. 🏷️ Version Injection (30 minutes)

**Impact:** 🟢 LOW - Professional polish

### Task:
Add version variable:

```go
// internal/commands/shared/version.go
package shared

var Version = "dev" // Overridden by ldflags at build time
```

Update version command:
```go
func init() {
    rootCmd.Version = shared.Version
}
```

---

## 5. 🔗 Command Aliases (1 hour)

**Impact:** 🟡 MEDIUM - Better UX

### Task:
Add aliases to commands:

```go
// internal/commands/analysis/stats.go
var statsCmd = &cobra.Command{
    Use:     "stats",
    Aliases: []string{"gain", "savings"}, // Like RTK
    Short:   "Show token savings statistics",
}

// internal/commands/analysis/quality.go
var qualityCmd = &cobra.Command{
    Use:     "quality",
    Aliases: []string{"grade", "score"},
    Short:   "Analyze compression quality",
}
```

---

## 📋 Checklist

### Day 1 (4-5 hours):
- [ ] Update Makefile with version injection
- [ ] Create version.go
- [ ] Add command aliases
- [ ] Create install.sh script
- [ ] Test install script locally

### Day 2 (4-5 hours):
- [ ] Set up GitHub release workflow
- [ ] Tag v0.1.0
- [ ] Create Homebrew formula
- [ ] Test Homebrew installation
- [ ] Update README with installation methods
- [ ] Commit and push everything

---

## 🎯 Success Criteria

After these quick wins:
- ✅ Users can install via: `brew install tokman`
- ✅ Users can install via: `curl ... | sh`
- ✅ Users can download pre-built binaries
- ✅ Version shows correctly: `tokman --version`
- ✅ Command aliases work: `tokman gain`

---

## 📈 Expected Impact

**Before:**
- Installation: Build from source only
- Distribution: Manual
- Version: Hardcoded
- UX: Basic

**After:**
- Installation: 1 command (`brew install` or `curl | sh`)
- Distribution: Automated releases
- Version: Dynamic from git tags
- UX: Professional with aliases

**Result:** TokMan becomes as easy to install as RTK and Snip!

---

<div align="center">

**Total Time: ~8-10 hours**

**Total Impact: MASSIVE** 🚀

**Let's do this!**

</div>
