# TokMan Python SDK

Token reduction with 14-layer research-based compression pipeline achieving 95-99% reduction.

## Installation

```bash
pip install tokman
```

Requires the `tokman` binary. Install with:

```bash
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
```

## Quick Start

```python
from tokman import TokMan, Mode

# Initialize
tm = TokMan(mode=Mode.AGGRESSIVE, budget=4000)

# Compress text
result = tm.process("Long text to compress...")

print(f"Original: {result.original_tokens} tokens")
print(f"Final: {result.final_tokens} tokens")
print(f"Saved: {result.tokens_saved} tokens ({result.reduction_percent}%)")
print(f"Output: {result.output}")
```

## Convenience Functions

```python
from tokman import compress

# Quick one-liner
output = compress("Long text...", mode=Mode.AGGRESSIVE)
```

## Streaming

```python
from tokman import TokMan, StreamChunk

def on_chunk(chunk: StreamChunk):
    print(f"Compressed: {chunk.content[:50]}...")

tm = TokMan()
writer = tm.stream(on_chunk)

writer.write("First chunk...")
writer.write("Second chunk...")
writer.close()
```

## Content Analysis

```python
tm = TokMan()
content_type = tm.analyze("func main() { ... }")
print(content_type)  # 'code'
```

## API Reference

### `TokMan(mode, budget, tokman_path)`

Main compression client.

### `process(text) -> CompressionResult`

Compress text and return stats.

### `analyze(text) -> str`

Detect content type: 'code', 'logs', 'conversation', 'git', 'test', 'docker/infra', 'mixed'.

### `compress(text, mode, budget) -> str`

Quick compression function.

## License

MIT
