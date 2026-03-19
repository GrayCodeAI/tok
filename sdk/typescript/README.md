# @tokman/sdk

> **TokMan** - The World's Best Token Reduction System for Node.js/TypeScript

A production-ready SDK for compressing LLM context with 14 research-backed compression layers.

## Features

- 🚀 **14-Layer Compression Pipeline** - Research-backed algorithms for maximum reduction
- 🎯 **Adaptive Compression** - Auto-detects content type (code, conversation, logs)
- 📊 **Full TypeScript Support** - Complete type definitions included
- 🔄 **Streaming Support** - Handle large contexts efficiently
- ⚡ **Zero Dependencies** - Lightweight client for any Node.js environment

## Installation

```bash
npm install @tokman/sdk
# or
yarn add @tokman/sdk
# or
pnpm add @tokman/sdk
```

## Quick Start

```typescript
import { TokMan } from '@tokman/sdk';

// Initialize client (defaults to http://localhost:8080)
const client = new TokMan({
  baseUrl: 'http://localhost:8080',
  timeout: 30000,
  defaultMode: 'balanced',
});

// Basic compression
const result = await client.compress('Your long text here...');
console.log(`Reduced from ${result.originalTokens} to ${result.finalTokens} tokens`);
console.log(`Reduction: ${result.reductionPercent}%`);

// Adaptive compression (auto-detects content type)
const adaptive = await client.compressAdaptive(`
  function main() {
    return 42;
  }
`);
console.log(`Detected: ${adaptive.detectedContentType}`);
```

## API Reference

### `new TokMan(config?)`

Create a new TokMan client.

```typescript
const client = new TokMan({
  baseUrl: 'http://localhost:8080',  // Server URL
  timeout: 30000,                     // Request timeout (ms)
  defaultMode: 'balanced',            // 'conservative' | 'balanced' | 'aggressive'
  debug: false,                       // Enable logging
  headers: {},                        // Custom headers
});
```

### `compress(input, options?)`

Compress text with optional mode.

```typescript
const result = await client.compress('text', {
  mode: 'aggressive',      // Override default mode
  targetTokens: 1000,      // Target token count
});
```

### `compressAdaptive(input, options?)`

Compress with automatic content detection.

```typescript
const result = await client.compressAdaptive('text', {
  contentType: 'code',     // Hint content type (optional)
  targetTokens: 500,
});
```

### `analyze(input)`

Analyze content without compression.

```typescript
const analysis = await client.analyze('function main() {}');
console.log(analysis.contentType);      // 'code'
console.log(analysis.confidence);       // 0.95
console.log(analysis.recommendedMode);  // 'conservative'
console.log(analysis.characteristics);  // { hasCode: true, ... }
```

### `health()`

Check server health.

```typescript
const { status, version } = await client.health();
```

### `stats()`

Get server statistics.

```typescript
const { version, layerCount } = await client.stats();
```

## Compression Modes

| Mode | Description | Typical Reduction |
|------|-------------|-------------------|
| `conservative` | Preserve all critical information | 10-25% |
| `balanced` | Good balance of reduction and preservation | 25-45% |
| `aggressive` | Maximum reduction, may lose some detail | 45-70% |

## Content Types

The adaptive compression automatically detects:

- **`code`** - Source code, scripts, config files
- **`conversation`** - Chat logs, dialogue, Q&A
- **`logs`** - System logs, error messages
- **`documents`** - Articles, documentation
- **`mixed`** - Mixed content types

## Error Handling

```typescript
import { TokMan, TokManError } from '@tokman/sdk';

try {
  const result = await client.compress('text');
} catch (error) {
  if (error instanceof TokManError) {
    console.log(`Error: ${error.message}`);
    console.log(`Code: ${error.code}`);
    console.log(`Status: ${error.statusCode}`);
  }
}
```

## Server Setup

This SDK requires the TokMan server. Start it with:

```bash
# Build the server
go build -o tokman ./cmd/server

# Run with default settings
./tokman

# Or with custom port
PORT=9090 ./tokman
```

## TypeScript Types

All types are exported for TypeScript users:

```typescript
import type {
  CompressionResult,
  AnalysisResult,
  ServerStats,
  CompressionMode,
  ContentType,
  LayerStats,
  TokManConfig,
} from '@tokman/sdk';
```

## License

MIT © GrayCodeAI

## Links

- [GitHub Repository](https://github.com/GrayCodeAI/tokman)
- [Documentation](https://github.com/GrayCodeAI/tokman#readme)
- [Go SDK](../go)
- [Python SDK](../python)
