# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `tok completion` command for shell completion generation (bash, zsh, fish, powershell)
- `tok man` command for man page generation
- `tok self-update` command for easy version upgrades
- Fuzz tests for TOML parsing and pipeline input processing
- Input size guard (50MB limit) for stdin reading
- Signal handling for graceful shutdown (SIGINT, SIGTERM)
- Dockerfile with multi-stage Alpine build
- SBOM generation (CycloneDX) in release pipeline
- Output abstraction layer (`internal/output`) replacing direct fmt.Print* calls
- Shell script path validation (absolute path, traversal, permissions checks)

### Changed
- Consolidated dual architecture into single cobra-based entry point
- Unified module path to `github.com/GrayCodeAI/tok` throughout codebase
- Coverage thresholds unified to 60% across all CI files
- All `GrayCodeAI` references updated from legacy `tokman` naming

### Fixed
- DAGOptimizer topological sort (Kahn's algorithm)
- DifferentialCompressor line-level diff implementation
- PerplexityOptimizer frequency-based token ranking
- HeatmapGenerator Output method
- Registry init ordering for late-registered commands

## [0.29.0] - 2025-04-19

### Added
- Merged tokman (output filtering) and tork (input compression) into unified tok CLI
- 12 AI agent rule files (cursor, windsurf, cline, copilot, claude-code, aider, continue, roo-code, cody, code-whisperer, tabnine, codeium)
- Shell hook scripts (bash + PowerShell) for prompt integration
- 31-layer compression pipeline with research-backed algorithms
- 100+ built-in command wrappers with intelligent filtering
- SQLite-based token usage tracking
- TOML-based declarative filter configuration
- MCP server support
- TUI dashboard

### Changed
- Complete rebrand from tokman/tork to tok
- Module path changed from `github.com/GrayCodeAI/tokman` to `github.com/GrayCodeAI/tok`
- Config paths updated to `~/.config/tok/` and `~/.local/share/tok/`

[Unreleased]: https://github.com/GrayCodeAI/tok/compare/v0.29.0...HEAD
[0.29.0]: https://github.com/GrayCodeAI/tok/releases/tag/v0.29.0
