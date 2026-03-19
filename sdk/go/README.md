# TokMan Go SDK

Native Go SDK for token compression with 14-layer research-based pipeline.

## Installation

```bash
go get github.com/GrayCodeAI/tokman/sdk/go/tokman
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/GrayCodeAI/tokman/sdk/go/tokman"
)

func main() {
    client := tokman.DefaultClient()
    
    result, err := client.Compress("Long text to compress...")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Saved %d tokens (%.1f%%)\n", result.TokensSaved, result.ReductionPercent)
}
```

## API Reference

### `tokman.New(config) *Client`

Create a new client with custom configuration.

```go
client := tokman.New(tokman.Config{
    Mode:               tokman.ModeAggressive,
    Budget:             4000,
    EnableStreaming:    true,
    EnableCompaction:   true,
    EnableH2O:          true,
    EnableAttentionSink: true,
})
```

### `client.Compress(input) (*Result, error)`

Compress text and return stats.

```go
result, err := client.Compress(text)
// result.Output          - compressed text
// result.OriginalTokens  - input token count
// result.FinalTokens     - output token count
// result.TokensSaved     - tokens saved
// result.ReductionPercent - compression ratio
// result.LayersApplied   - layers that contributed
```

### `client.CompressWithBudget(input, budget) (*Result, error)`

Compress with a specific token budget.

### `client.CompressAdaptive(input) (*Result, error)`

Auto-detect content type and optimize layer selection.

### `client.Analyze(input) ContentType`

Detect content type: `code`, `logs`, `conversation`, `git`, `test`, `docker/infra`, `mixed`.

### `client.Stream() *Stream`

Create a streaming compressor.

```go
stream := client.Stream()
stream.Write([]byte("chunk 1"))
stream.Write([]byte("chunk 2"))
output := stream.Flush()
```

## Modes

- `ModeMinimal` - Preserves more content, lower compression
- `ModeAggressive` - Maximizes compression

## Content Types

- `ContentTypeCode` - Source code
- `ContentTypeLogs` - Log output
- `ContentTypeConversation` - Chat/dialogue
- `ContentTypeGit` - Git command output
- `ContentTypeTest` - Test runner output
- `ContentTypeDocker` - Container/infra output
- `ContentTypeMixed` - Multiple types

## License

MIT
