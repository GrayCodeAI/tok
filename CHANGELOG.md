# 📝 Changelog

All notable changes to TokMan will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-04-07

### ✨ New Features
- 🎯 **Delegating Hook System** - All rewrite logic in binary with exit code protocol (RTK-style)
- 🧪 **Inline Filter Tests** - TOML-based test declarations with `[[tests.filtername]]` syntax
- ✅ **Filter Validation Command** - `tokman validate` checks 96 filters for syntax and quality
- 🧪 **Filter Test Command** - `tokman tests` runs inline tests with color-coded output
- 🔒 **Safety Checks** - Deny dangerous commands (rm, dd), warn on unsafe operations
- 📦 **Homebrew Formula** - Easy installation via `brew install GrayCodeAI/tokman/tokman`
- 🚀 **Install Script** - Cross-platform installer (`curl ... | sh`)
- 🏷️ **Dynamic Versioning** - Version injection via ldflags from git tags
- 🔗 **Command Aliases** - `tokman gain` (stats), `tokman grade` (quality), etc.
- 📊 **Quality Scoring** - 6-metric analysis with A+ to F grades
- 🎨 **Visual Diff Tool** - Color-coded comparison with HTML export
- 📁 **Multi-File Merging** - Dependency-aware file merging with budget management

### 🎯 Quality Improvements
- 📝 **7 Validation Checks** - Regex validation, schema matching, test verification
- 🛡️ **19 Hook Unit Tests** - Comprehensive safety and rewrite testing
- 🧪 **6 Inline Filter Tests** - Git filter examples (git_status, git_log, git_diff)
- ⚠️ **Warning System** - Missing tests, conflicting configs, reasonable limits
- 🐛 **Fixed Syntax Errors** - Printf format mismatch, backticks in raw strings, duplicate functions
- 🧹 **Code Cleanup** - Removed unused imports, variables, and dead code

### 🛠️ Developer Experience
- 📋 **Professional Makefile** - 18 targets with version injection and multi-platform builds
- 📖 **Troubleshooting Guide** - Comprehensive TROUBLESHOOTING.md
- 📦 **Release Workflow** - GitHub Actions automated releases
- 📝 **Release Guide** - docs/RELEASE.md with step-by-step instructions
- 🔧 **Gitignore Update** - Comprehensive `.gitignore` for clean repository

### 🐛 Bug Fixes
- Fixed printf format mismatch in `project_map.go`
- Fixed backticks in raw string literals in `bench.go`
- Fixed duplicate `outputJSON` function declarations
- Fixed `match = ` → `match_command = ` in 23 builtin TOML filters
- Fixed hook script compatibility issues

### 📦 New Commands
- `tokman rewrite` - Rewrite commands for hooks (exit code protocol)
- `tokman tests` - Run inline TOML filter tests
- `tokman validate` - Validate filter files
- `tokman quality` - Compression quality analysis (A+ to F)
- `tokman merge` - Intelligent multi-file merging

### 📊 Compatibility
- ✅ **Homebrew** - macOS and Linux
- ✅ **Install Script** - Linux (amd64/arm64), macOS (amd64/arm64), Windows
- ✅ **Go Install** - `go install github.com/GrayCodeAI/tokman/cmd/tokman@latest`
- ✅ **Pre-built Binaries** - GitHub Releases
- ✅ **Hook Support** - Claude Code, Cursor, Windsurf, Copilot, Cline, Codex

---

## [Unreleased] (Development)

### Planned
- 🌐 Website launch (tokman.dev)
- 💬 Discord community
- 🌍 Multi-language documentation
- 📊 Performance benchmarking suite
- 🤖 AI agent integration improvements

---

## [2.0.0] - 2026-03-31

---

## [2.0.0] - 2026-03-31

### 🎉 Major Release - Complete Rewrite

This is a major rewrite of TokMan with significant architectural improvements and new features.

### ✨ Major Features

#### **Compression & Processing**
- 🔥 **HTTP Proxy Mode** - Transparent compression for OpenAI, Anthropic, Gemini APIs
- 🎯 **KV-Cache Alignment** - Provider-level caching optimization
- 🔍 **Cross-Message Deduplication** - SimHash near-duplicate detection
- 🎨 **Content-Type Auto-Detection** - 8 types, 16 languages
- ♻️ **Reversible Compression** - LLM-assisted retrieval support
- 📊 **TOON Columnar Encoding** - 40-80% compression for JSON arrays
- 🔤 **Token Dense Dialect** - Unicode symbol shorthand
- 🤖 **LLM-Based Compression** - External process integration
- 📐 **Adaptive Context Scaling** - Per-model profiles
- 🎪 **Position-Aware/LITM** - Content ordering optimization

#### **Memory & Intelligence**
- 🧠 **Engram Memory System** - Tiered summaries (L0/L1/L2)
- 🔄 **Feedback Loop Learning** - Threshold optimization
- 🎯 **Information Bottleneck Filter** - Information theory-based compression
- 📖 **6 Read Modes** - full, map, signatures, diff, aggressive, entropy
- 🔄 **Incremental Deltas** - File change optimization
- 🕸️ **Project Intelligence Graph** - Dependency analysis
- 💾 **Cross-Session Memory** - Persistent storage
- 🤝 **Multi-Agent Context Sharing** - Collaborative AI

#### **Analytics & Monitoring**
- 📊 **Advanced Analytics** - Anomaly detection, forecasting, right-sizing
- 🗺️ **Heatmap Visualization** - Usage patterns
- 📉 **Waste Detection** - Identify optimization opportunities
- 📱 **Multi-Platform Tracking** - 16+ AI clients
- 🎨 **Analytics TUI Dashboard** - Terminal UI
- 📈 **Cost Intelligence** - Forecasting and budgets

#### **Security & Governance**
- 🛡️ **AI Gateway** - Kill switches, quotas, model aliasing
- 🔒 **Content Guardrails** - PII and injection detection
- 🔐 **Security Scanner** - SQLi, XSS, SSRF, path traversal
- 🔍 **eBPF Monitoring** - System-level observability
- 🌐 **Network Firewall** - Traffic control
- 📋 **SIEM Integration** - OCSF format
- 🔬 **Decision Explainability** - Forensic audit records

#### **Developer Experience**
- 🔧 **Template Pipes** - join, truncate, lines, keep, where, each
- 🎯 **JSONPath Extraction** - RFC 9535 compliant
- ✅ **Filter Variants** - Multiple compression strategies
- 🛡️ **Safety Checks** - Automated validation
- 🌐 **Community Filter Registry** - Shared filters
- 🧪 **Filter Test Suites** - Automated testing
- ✅ **Auto-Validation Pipeline** - Quality assurance
- 🎮 **Developer Playground** - Interactive testing
- 📺 **Live Monitor** - htop-style monitoring
- 🎨 **Color Passthrough** - Preserve terminal colors
- ⚡ **PATH Shim Injection** - Transparent integration

#### **Specialized Features**
- 🖼️ **Photon** - Base64 image compression
- 📝 **LogCrunch** - Log deduplication
- 📊 **DiffCrunch** - Diff compression
- 🏗️ **StructuralCollapse** - Structure-aware compression
- 📚 **Dictionary Encoding** - Auto-learned codebook
- 🏆 **Leaderboard** - Gamification
- 📅 **Wrapped Year-in-Review** - Usage statistics
- 🏅 **GitHub Profile Widget** - Display achievements

### ⚡ Performance

- 🚀 **BPE Tokenization** - tiktoken cl100k_base
- 💾 **O(1) LRU Cache** - Doubly-linked list implementation
- 🔧 **SIMD Support** - Go 1.24+ optimizations
- 🔀 **Parallel Pipeline** - Multi-threaded execution
- 🌊 **Streaming Mode** - Large input handling

### 💥 Breaking Changes

- ⚠️ **Go 1.24+ required** - Updated minimum Go version
- 🔄 **Pipeline configuration format changed** - Migration required

---

## [1.5.0] - 2026-04-03

### 🏢 Enterprise Features

#### **Performance Testing**
- 📊 Benchmarking framework (JSON/CSV/Table export)
- 📈 Trend tracking and regression detection
- 🔀 Parallel execution
- 📝 Custom DSL for test scenarios
- 🧪 15 predefined stress test scenarios
- 🌐 Distributed testing support
- 🎲 Chaos engineering (9 fault types)
- 🎮 Game days and blast radius testing

#### **Cost Intelligence**
- 📊 Cost forecasting (4 ML models + ensemble)
- 🔔 Budget alerts (5 notification channels)
- 💰 Team cost allocation and chargeback
- 🚨 Anomaly detection (4 algorithms)
- 📋 Cost policy enforcement
- 🏢 Cost center hierarchy
- 📊 10-widget dashboard
- 📧 Weekly digest (HTML/MD/JSON)

#### **AI Agent Framework**
- 🤖 Iterative agent (ReAct loop, memory, reflection)
- 🌐 LLM providers (OpenAI, Anthropic, Ollama)
- 🔌 MCP host management
- 🎯 Intelligent filter selection
- 😊 Sentiment analysis
- ⚙️ Auto-tuning
- 📊 Workload prediction

#### **Deployment & Reliability**
- 🚢 Canary deployments (4 strategies)
- 🧪 A/B testing (4 experiment types)
- 📋 Audit logging
- 🔐 RBAC (4 roles)
- 🔒 AES-256-GCM encryption
- 📅 Data retention policies

#### **CLI Enhancements**
- 🎯 Shell completion (bash/zsh/fish)
- 🔗 Command aliases
- 📊 Progress indicators
- 🎨 Color themes
- 🧪 Dry-run mode
- 🔗 Command chaining
- 📦 Batch operations
- ↩️ Undo/redo
- 📜 Command history
- ⭐ Favorites
- ⏰ Scheduling

### 🔄 CI/CD

- ✅ GitHub Actions workflows
- 🧪 Benchmark testing
- 🔍 Stress testing
- 🏗️ Multi-platform builds
- 🔒 Security scanning
- 🚀 Automated releases

### 📊 Testing

- 🧪 163+ tests across 119 packages
- 📈 30% average test coverage across 190 packages

### 🐛 Bug Fixes

- 🔧 Fixed context leak in benchmark CI/CD integration

---

## [1.4.0] - 2026-03-15

### Added

- 🎯 Enhanced compression tiers (Surface, Trim, Extract, Core)
- 📦 50+ new TOML filters for popular tools
- 🔍 Content-aware filtering improvements
- 📊 Analytics dashboard enhancements

### Fixed

- 🐛 Memory leak in long-running sessions
- 🔧 Race condition in pipeline execution
- 📝 Incorrect token counting for certain languages

---

## [1.3.0] - 2026-02-20

### Added

- 🤖 Agent integration for Claude Code, Cursor, Copilot
- 📊 Token tracking and analytics
- 🎨 Terminal UI for dashboard
- 🔧 Configuration file support

### Changed

- ⚡ Improved compression performance by 30%
- 📝 Better error messages

---

## [1.2.0] - 2026-01-15

### Added

- 🎯 TOML filter system
- 📦 Built-in filters for common commands
- 🔍 Pattern matching improvements

### Fixed

- 🐛 Panic on empty input
- 🔧 Incorrect output for certain edge cases

---

## [1.1.0] - 2025-12-10

### Added

- 📊 Basic compression pipeline
- 🎯 Command interception
- 📈 Token estimation

---

## [1.0.0] - 2025-11-01

### 🎉 Initial Release

- ✨ First public release of TokMan
- 🎯 Core compression functionality
- 📦 Support for major development tools
- 📊 Basic analytics

---

## Legend

- ✨ New feature
- 🐛 Bug fix
- 🔧 Improvement
- ⚡ Performance
- 📝 Documentation
- 🎨 UI/UX
- 🔒 Security
- 🏗️ Architecture
- 💥 Breaking change
- ⚠️ Deprecation

---

<div align="center">

**[View all releases on GitHub →](https://github.com/GrayCodeAI/tokman/releases)**

</div>
