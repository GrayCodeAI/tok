# TokMan Architecture

## Overview

TokMan is a comprehensive token optimization system designed to reduce LLM API costs by 60-99% through intelligent compression of command output.

## Core Components

### 1. MCP Context Server
- **Location**: `internal/mcp/`
- **Features**: 27 MCP tools, JSON-RPC 2.0, SQLite persistence
- **Purpose**: Lean-ctx inspired context management with 99% token savings

### 2. Compression Pipeline (31 Layers)
- **Location**: `internal/core/`
- **Categories**:
  - **Statistical**: Entropy, Perplexity, N-gram deduplication
  - **Semantic**: Goal-driven, contrastive learning, AST parsing
  - **Structural**: Code folding, import collapse, comment removal
  - **Adaptive**: Budget enforcement, lazy pruner, semantic anchors

### 3. Filter System
- **Location**: `internal/filter/`, `internal/toml/`
- **Count**: 114+ filters (85 Go, 94 TOML built-in)
- **Features**: 8-stage pipeline, schema validation, safety checks

### 4. Security Layer
- **Location**: `internal/security/`
- **Features**: PII detection, prompt injection detection, secret scanning

### 5. Dashboard
- **Location**: `internal/dashboard/`
- **Features**: WebSocket live updates, 36 API endpoints, cost projections

### 6. Tracking & Analytics
- **Location**: `internal/tracking/`
- **Features**: Per-layer statistics, cost estimation, CSV/JSON export

### 7. Multi-Provider Gateway
- **Location**: `internal/gateway/`
- **Features**: Anthropic, OpenAI support with fallback chains

### 8. TUI
- **Location**: `internal/tui/`
- **Features**: Bubble Tea dashboard with 4 tabs

### 9. Internationalization
- **Location**: `internal/i18n/`
- **Languages**: 9 (en, fr, zh, ja, ko, es, de, pt, it)

### 10. HTML Extraction
- **Location**: `internal/html/`
- **Features**: Site-specific extractors (GitHub, Wikipedia, HN)

## Stats

| Metric | Value |
|--------|-------|
| Total Go Files | 537 |
| Total Filters | 114+ (94 TOML + 20 Go) |
| Compression Layers | 31 |
| API Endpoints | 36 |
| Languages | 9 |
| Test Coverage | 147 test files |
| Completion | 100% |
