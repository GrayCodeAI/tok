<div align="center">

# 🚀 TokMan

**Token-aware CLI proxy & AI gateway for coding assistants**

*Practical 20-layer pipeline focused on real-world compression quality*

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![CI](https://github.com/GrayCodeAI/tokman/workflows/CI/badge.svg)](https://github.com/GrayCodeAI/tokman/actions)
[![Security](https://github.com/GrayCodeAI/tokman/workflows/Security/badge.svg)](https://github.com/GrayCodeAI/tokman/actions)
[![codecov](https://codecov.io/gh/GrayCodeAI/tokman/branch/main/graph/badge.svg)](https://codecov.io/gh/GrayCodeAI/tokman)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrayCodeAI/tokman)](https://goreportcard.com/report/github.com/GrayCodeAI/tokman)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 Website](https://tokman.dev) · [📖 Documentation](./docs) · [💬 Discord](https://discord.gg/HrVA7ePyV) · [🐛 Issues](https://github.com/GrayCodeAI/tokman/issues)

</div>

---

## 💡 What is TokMan?

TokMan intercepts CLI commands and applies an intelligent **20-layer compression pipeline** to drastically reduce token usage for AI coding assistants with practical, high-impact defaults.

```
┌──────────────────────────────────────────────────────────────┐
│  Input: 10,000 tokens  →  TokMan Pipeline  →  Output: 1,500  │
│                                                                │
│  💰 Cost Savings:    $0.085 → $0.013  (85% reduction)        │
│  ⚡ Speed Boost:     Faster AI responses                       │
│  🎯 Quality:         Preserves critical information            │
└──────────────────────────────────────────────────────────────┘
```

## ✨ Key Features

<table>
<tr>
<td width="50%">

### 🔥 Performance
- **60-90% token reduction** on typical dev workflows
- **2-3x speedup** with SIMD optimization (Go 1.26+)
- **Sub-millisecond processing** for most commands
- **Zero configuration** needed

</td>
<td width="50%">

### 🎯 Intelligence
- **20 practical compression layers** from academic research
- **Content-aware** filtering (code, logs, JSON, etc.)
- **Context preservation** - keeps what matters
- **Semantic understanding** of command output

</td>
</tr>
<tr>
<td width="50%">

### 🔌 Extensibility
- **WASM plugin system** for custom filters
- **97+ TOML filters** for popular tools
- **Scriptable** via CLI or HTTP proxy
- **API access** for programmatic use

</td>
<td width="50%">

### 🛡️ Enterprise Ready
- **Production tested** on large codebases
- **Privacy first** - all processing local
- **Audit logs** and analytics dashboard
- **Team cost tracking** and budgets

</td>
</tr>
</table>

## 📊 Real-World Impact

### Token Savings (30-minute Claude Code session)

| Command | Uses | Before | After | Savings |
|---------|------|--------|-------|---------|
| 📁 `ls` / `tree` | 10× | 2,000 | 400 | **80%** ↓ |
| 📄 `cat` / `read` | 20× | 40,000 | 12,000 | **70%** ↓ |
| 🔍 `grep` / `rg` | 8× | 16,000 | 3,200 | **80%** ↓ |
| 🎯 `git status` | 10× | 3,000 | 600 | **80%** ↓ |
| 📝 `git diff` | 5× | 10,000 | 2,500 | **75%** ↓ |
| 📜 `git log` | 5× | 2,500 | 500 | **80%** ↓ |
| ✅ `git commit` | 8× | 1,600 | 120 | **92%** ↓ |
| 🧪 `npm test` | 5× | 25,000 | 2,500 | **90%** ↓ |
| 🔬 `pytest` | 4× | 8,000 | 800 | **90%** ↓ |
| 📦 `npm ls` | 3× | 900 | 180 | **80%** ↓ |
| **📊 Total** | | **~118,000** | **~23,500** | **🎉 80%** ↓ |

### 💰 Cost Reduction

| Usage Pattern | Without TokMan | With TokMan | Monthly Savings |
|---------------|----------------|-------------|-----------------|
| 🧑‍💻 Individual (30 min/day) | $15 | $2.25 | **$12.75** |
| 👥 Small Team (5 devs) | $75 | $11.25 | **$63.75** |
| 🏢 Team (20 devs) | $300 | $45 | **$255** |
| 🏭 Enterprise (100 devs) | $1,500 | $225 | **$1,275** |

*Based on Claude Sonnet 3.5 pricing ($3/MTok input, $15/MTok output)*

## 🚀 Quick Start

### Installation

#### 🍺 Homebrew (Recommended - macOS/Linux)

```bash
brew tap GrayCodeAI/tokman
brew install tokman
```

#### 🚀 Install Script (Linux/macOS/Windows)

```bash
curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh
```

#### 📦 Pre-built Binaries

Download from [GitHub Releases](https://github.com/GrayCodeAI/tokman/releases/latest):
- macOS: `tokman-darwin-amd64.tar.gz`, `tokman-darwin-arm64.tar.gz`
- Linux: `tokman-linux-amd64.tar.gz`, `tokman-linux-arm64.tar.gz`
- Windows: `tokman-windows-amd64.zip`

#### 🐹 Go Install

```bash
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
```

#### 🔨 Build from Source

```bash
git clone https://github.com/GrayCodeAI/tokman.git
cd tokman
make build

# Or build for all platforms
make build-all
```

### Setup for Your AI Tool

```bash
# Claude Code / GitHub Copilot
tokman init -g

# Cursor
tokman init --agent cursor

# Windsurf
tokman init --agent windsurf

# Cline / Roo Code
tokman init --agent cline

# Gemini CLI
tokman init -g --gemini
```

### Verify Installation

```bash
tokman --version      # Check version
tokman doctor         # Verify setup
tokman gain           # View savings stats
```

### Usage

Once installed, TokMan automatically intercepts commands:

```bash
# These are automatically compressed:
git status
npm ls
npm test
cat large-file.json

# Or use standalone:
tokman compress < input.txt
tokman benchmark --suite git-status
tokman dashboard  # Launch analytics dashboard
```

## 🧠 How It Works

TokMan uses a **20-layer pipeline** inspired by cutting-edge research:

```
Input → Content Detection → Pipeline Selection → Compression → Output
         ↓                    ↓                   ↓
      [JSON, Code,        [Surface, Trim,    [20 layers:
       Logs, etc.]         Extract, Core]     Entropy, H2O,
                                               AST, Gist, etc.]
```

### Compression Tiers

| Tier | Layers | Reduction | Use Case |
|------|--------|-----------|----------|
| 🟢 **Surface** | 3 | 30-50% | Quick cleanup, preserve everything |
| 🟡 **Trim** | 12 | 50-70% | Balanced compression |
| 🟠 **Extract** | 24 | 70-90% | Aggressive, preserve essence |
| 🔴 **Core** | 20 | 90%+ | Maximum practical compression |

### Specialized Profiles

- 💻 **Code**: Syntax-aware, preserves structure (50-70%)
- 📝 **Log**: Deduplication, pattern grouping (60-80%)
- 💬 **Thread**: Conversation-aware, context preservation (55-75%)

## 📦 Supported Tools

TokMan has built-in filters for **97+ development tools**:

<details>
<summary><b>🔧 Version Control</b></summary>

- Git, GitHub CLI, GitLab CLI
- Mercurial, SVN

</details>

<details>
<summary><b>🐳 Containers & Orchestration</b></summary>

- Docker, Docker Compose
- Kubernetes (kubectl), Helm
- Podman, containerd

</details>

<details>
<summary><b>📦 Package Managers</b></summary>

- npm, yarn, pnpm, bun
- pip, uv, poetry
- cargo, go mod
- maven, gradle

</details>

<details>
<summary><b>🧪 Testing & Linting</b></summary>

- Jest, Vitest, Playwright
- pytest, unittest
- cargo test, go test
- ESLint, Ruff, golangci-lint

</details>

<details>
<summary><b>☁️ Cloud & Infrastructure</b></summary>

- AWS CLI, gcloud, az
- Terraform, Ansible
- PostgreSQL, MySQL

</details>

[**See full list →**](./docs/TOML_FILTERS.md)

## 🆕 New Features

### Generic Test Runner
Auto-detect and run project tests with a single command:
```bash
tokman test-runner              # Auto-detect test runner
tokman test-runner cargo test   # Run Rust tests
tokman test-runner npm test     # Run Node.js tests
tokman test-runner pytest       # Run Python tests
```

**Supported test runners:** Cargo, Go, Vitest, Jest, npm, pnpm, Pytest, RSpec, Rake Test, Playwright

### Quota Estimation
Estimate subscription tier usage based on your token consumption:
```bash
tokman gain --quota pro         # Estimate 'pro' tier usage
tokman gain --quota 5x          # Estimate '5x' tier usage
tokman gain --quota 20x         # Estimate '20x' tier usage
```

Shows projected monthly usage, tier limits, and upgrade recommendations.

### Session Adoption Tracking
View TokMan adoption across your Claude Code sessions:
```bash
tokman adoption                 # Show last 10 sessions
tokman adoption --limit 20      # Show last 20 sessions
```

### Errors-Only Mode
Run any command and show only errors/warnings:
```bash
tokman err npm run build
tokman err cargo build
tokman err go test ./...
```

### Smart File Summaries
Generate 2-line summaries of any file:
```bash
tokman smart main.go            # Go file summary
tokman smart package.json       # NPM package summary
tokman smart README.md          # Documentation summary
```

## 🔬 Technical Details

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         TokMan CLI                           │
├─────────────────────────────────────────────────────────────┤
│  🎯 Command Router  →  📊 Content Detector  →  ⚙️  Pipeline  │
├─────────────────────────────────────────────────────────────┤
│                    20 Compression Layers                     │
│  ├─ Entropy Filtering        ├─ H2O (Heavy-Hitter Oracle)  │
│  ├─ Perplexity Pruning       ├─ Attention Sink             │
│  ├─ AST Preservation         ├─ Meta-Token Compression     │
│  ├─ Goal-Driven Selection    ├─ Semantic Chunking          │
│  └─ ... 23 more layers                                      │
├─────────────────────────────────────────────────────────────┤
│  💾 Cache Layer  →  📈 Analytics  →  🔌 Plugin System       │
└─────────────────────────────────────────────────────────────┘
```

### Performance Optimizations

- **SIMD acceleration**: AVX2, AVX-512, ARM NEON support
- **Fingerprint caching**: Skip redundant processing
- **Streaming mode**: Handle large inputs (>500K tokens)
- **Parallel execution**: Multi-threaded pipeline
- **Early exit**: Stop when budget met

### Research Foundation

Built on 30+ academic papers including:

- **Selective Context** (Mila 2023) - Entropy filtering
- **LLMLingua** (Microsoft 2023) - Perplexity pruning
- **H2O** (NeurIPS 2023) - Heavy-hitter oracle
- **StreamingLLM** (2023) - Attention sink
- **AutoCompressor** (Princeton/MIT 2023) - Hierarchical compression

[**Full research list →**](./docs/LAYERS.md)

## 📖 Documentation

- [📘 Quick Start Guide](./docs/QUICK_START.md)
- [🔧 TOML Filter Reference](./docs/TOML_FILTERS.md)
- [🧪 Benchmark Results](./docs/BENCHMARKS.md)
- [🛡️ Security Guide](./docs/SECURITY.md)
- [🔌 API Reference](./docs/API.md)
- [🎯 Tuning Guide](./docs/TUNING.md)
- [🤖 Agent Integration](./docs/AGENT_INTEGRATION.md)
- [❓ Troubleshooting](./docs/TROUBLESHOOTING.md)

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Quick Contribution Guide

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/tokman.git
cd tokman

# 2. Create a branch
git checkout -b feature/my-new-feature

# 3. Make changes and test
make test
make lint

# 4. Commit and push
git commit -m "feat: add amazing feature"
git push origin feature/my-new-feature

# 5. Open a Pull Request
```

### Development Tools

```bash
make build          # Build binary
make test           # Run tests
make test-cover     # Generate coverage report
make lint           # Run linters
make bench          # Run benchmarks
make check          # Run all checks
```

## 📊 Project Stats

- **Language**: Go 1.26+
- **Packages**: 150+ internal packages
- **Tests**: 144 packages with tests
- **Lines of Code**: ~50,000
- **Built-in Filters**: 97 TOML filters
- **Compression Layers**: 20
- **Platforms**: Linux, macOS, Windows (amd64/arm64)

## 🗺️ Roadmap

- [x] Core compression pipeline (20 layers)
- [x] TOML filter system
- [x] Agent integration (Claude, Cursor, Copilot, etc.)
- [x] Analytics dashboard
- [x] SIMD optimization
- [x] WASM plugin system
- [ ] Cloud sync for team settings
- [ ] Browser extension
- [ ] IDE plugins (VS Code, JetBrains)
- [ ] Real-time collaboration features
- [ ] Advanced ML-based compression

## ❓ FAQ

<details>
<summary><b>How does TokMan reduce tokens?</b></summary>

TokMan applies a layered compression pipeline that removes noise, groups similar content, truncates redundancy, and preserves critical information. Core stages are production-oriented, with additional experimental layers available.
</details>

<details>
<summary><b>Does it lose important information?</b></summary>

TokMan uses quality metrics (6-metric grading, A+ to F) to ensure compression preserves signal. The goal is to remove noise while keeping everything the AI needs.
</details>

<details>
<summary><b>Which AI tools does it support?</b></summary>

Claude Code, Cursor, GitHub Copilot, Windsurf, Cline/Roo Code, Gemini CLI, Codex, and Aider. Basically any tool that runs shell commands.
</details>

<details>
<summary><b>Is it fast enough for real-time use?</b></summary>

Yes. Most commands complete in <20ms of overhead. SIMD optimizations planned for Go 1.26+ will reduce this further.
</details>

<details>
<summary><b>Can I add custom filters?</b></summary>

Yes! Create `.toml` filter files in `~/.config/tokman/filters/`. See the [filter writing guide](docs/DEVELOPMENT.md) for details.
</details>

<details>
<summary><b>Is my data safe?</b></summary>

TokMan processes everything locally. No data is sent externally. Telemetry is opt-in and never collects file contents. See [SECURITY.md](SECURITY.md).
</details>

<details>
<summary><b>Can I use it in CI/CD?</b></summary>

Yes! See the [deployment guide](docs/DEPLOYMENT.md) for GitHub Actions, GitLab CI, and Docker integration.
</details>

---

## 🔧 Troubleshooting

<details>
<summary><b>TokMan not found after installation</b></summary>

```bash
# Check if in PATH
which tokman

# Add Go bin to PATH
export PATH="$HOME/go/bin:$PATH"
# Add to ~/.bashrc or ~/.zshrc for persistence
```
</details>

<details>
<summary><b>Hooks not intercepting commands</b></summary>

```bash
# Reinstall hooks
tokman init --uninstall
tokman init -g

# Verify with doctor
tokman doctor
```
</details>

<details>
<summary><b>Database errors</b></summary>

```bash
# Reset database
rm ~/.local/share/tokman/tokman.db
tokman status  # Recreates automatically
```
</details>

<details>
<summary><b>High memory usage on large files</b></summary>

```bash
# Use streaming mode (auto-enabled for >500K tokens)
# Or set a budget to limit output
tokman --budget 2000 cat large_file.txt
```
</details>

For more help, see the [full troubleshooting guide](docs/DEPLOYMENT.md#troubleshooting) or [open an issue](https://github.com/GrayCodeAI/tokman/issues).

---

## 📄 License

TokMan is released under the [MIT License](LICENSE).

## 🙏 Acknowledgments

Built with research from:
- Microsoft Research (LLMLingua, LongLLMLingua)
- Stanford University (Gist Compression)
- MIT CSAIL (AutoCompressor)
- Princeton University (AutoCompressor)
- UC Berkeley (MemGPT, H2O)
- Tsinghua University (EHPC)
- Mila (Selective Context)
- NUS (LongCodeZip)
- Shanghai Jiao Tong University (SWE-Pruner)
- LinkedIn (ProCut)
- And 20+ other institutions

See [CITATION.cff](CITATION.cff) for academic citation information.

Special thanks to the open-source community and all contributors. See [AUTHORS.md](AUTHORS.md).

## 💬 Community & Support

- 💬 [Discord Server](https://discord.gg/HrVA7ePyV) - Chat with the community
- 🐛 [Issue Tracker](https://github.com/GrayCodeAI/tokman/issues) - Report bugs
- 📧 [Email](mailto:hello@tokman.dev) - Contact the team
- 🐦 [Twitter](https://twitter.com/tokman_dev) - Follow updates

---

<div align="center">

**⭐ Star us on GitHub if TokMan helps you save tokens!**

Made with ❤️ by the TokMan team

</div>
