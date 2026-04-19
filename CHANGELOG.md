# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- SECURITY.md with vulnerability disclosure process
- CODE_OF_CONDUCT.md for community standards
- CONTRIBUTING.md with comprehensive development guide
- GitHub issue templates (bug report, feature request)
- Pull request template with detailed checklist
- MASTER_TASK_LIST.md with 1100+ development tasks
- market and ecosystem research documentation

### Changed

- Added `tok` binary to `.gitignore`

### Documentation

- Created broader product research and roadmap notes

## [0.28.2] - 2026-04-03

### Added

- Multiple AI agent integration support (Claude Code, Cursor, Copilot, Windsurf, Cline, Gemini, Codex)
- 31-layer compression pipeline with research-backed filters
- Quality metrics system with 6-metric grading (A+ to F)
- Visual diff output with HTML export
- Multi-file intelligence with dependency-aware ordering
- TOML filter system with 97+ built-in filters
- SQLite-based command tracking and analytics
- Dashboard web interface for analytics
- Session management with snapshot/restore
- Hook integrity verification system
- Cost analysis and budget enforcement
- Telemetry collection (opt-in)

### Filter Layers

1. Entropy Filtering (Selective Context)
2. Perplexity Pruning (LLMLingua)
3. Goal-Driven Selection (SWE-Pruner)
4. AST Preservation (LongCodeZip)
5. Contrastive Ranking (LongLLMLingua)
6. N-gram Abbreviation (CompactPrompt)
7. Evaluator Heads (EHPC)
8. Gist Compression (Stanford/Berkeley)
9. Hierarchical Summary (AutoCompressor)
10. Budget Enforcement
11. Compaction (MemGPT)
12. Attribution Filter (ProCut)
13. H2O Filter (Heavy-Hitter Oracle)
14. Attention Sink (StreamingLLM)
15. Meta-Token
16. Semantic Chunk (ChunkKV-style)
17. Sketch Store (KVReviver)
18. Lazy Pruner (LazyLLM)
19. Semantic Anchor
20. Agent Memory (Focus-inspired)
- Plus question-aware and density-adaptive layers

## [0.28.1] - 2026-03-20

### Added

- Initial release of Tok token-aware CLI proxy
- Basic filter pipeline implementation
- Core command runner and token estimation
- Configuration system with Viper + TOML
- Basic hook system for Claude Code integration
- CLI commands:
  - `tok init` - Initialize with AI tools
  - `tok doctor` - System health check
  - `tok status` - Show current status
  - `tok help` - Display help information
  - `tok version` - Show version information

## [0.28.0] - 2026-03-01

### Added

- Project initialization
- Core architecture design
- Research compilation from 120+ papers
- 31-layer filter pipeline design
- Proof of concept implementation

---

<!-- Version History Template for Future Releases -->
<!--
## [x.y.z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing functionality

### Deprecated
- Features that will be removed

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security improvements

### Performance
- Performance improvements

### Documentation
- Documentation updates
-->

<!-- Links -->
[Unreleased]: https://github.com/lakshmanpatel/tok/compare/v0.28.2...HEAD
[0.28.2]: https://github.com/lakshmanpatel/tok/compare/v0.28.1...v0.28.2
[0.28.1]: https://github.com/lakshmanpatel/tok/compare/v0.28.0...v0.28.1
[0.28.0]: https://github.com/lakshmanpatel/tok/releases/tag/v0.28.0
