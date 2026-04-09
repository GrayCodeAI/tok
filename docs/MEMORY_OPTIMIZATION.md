# Memory Optimization Guide

This document provides recommendations for optimizing memory usage in TokMan.

## Field Alignment

Go automatically pads struct fields for alignment. Use fieldalignment tool to detect inefficient struct layouts.

```bash
# Install
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

# Check a package
fieldalignment ./internal/config/...

# Fix issues
fieldalignment -fix ./internal/config/...
```

## Recommended Struct Ordering

### PipelineConfig

Current order (may have padding issues):
```go
type PipelineConfig struct {
    MaxContextTokens int     // 8 bytes
    ChunkSize         int    // 8 bytes
    EnableEntropy     bool   // 1 byte + 7 padding
    // ...
}
```

Optimized order:
```go
type PipelineConfig struct {
    // 8-byte fields first
    MaxContextTokens int
    ChunkSize        int
    
    // 4-byte fields
    EnableEntropy    bool   // bools can be packed together
    
    // 1-byte fields grouped
    // ...
}
```

## sync.Pool Usage

Use `sync.Pool` for frequently allocated objects:

```go
var byteSlicePool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func GetBuffer() []byte {
    return byteSlicePool.Get().([]byte)[:0]
}

func PutBuffer(buf []byte) {
    byteSlicePool.Put(buf)
}
```

## String Interning

For frequently repeated strings, use interning:

```go
type stringInterner struct {
    mu   sync.RWMutex
    strs map[string]string
}

func (i *stringInterner) Intern(s string) string {
    i.mu.RLock()
    if existing, ok := i.strs[s]; ok {
        i.mu.RUnlock()
        return existing
    }
    i.mu.RUnlock()
    
    i.mu.Lock()
    if existing, ok := i.strs[s]; ok {
        i.mu.Unlock()
        return existing
    }
    i.strs[s] = s
    i.mu.Unlock()
    return s
}
```

## Buffer Pooling

Use buffer pooling in hot paths:

```go
// In filter/bytes.go - already implemented
type ByteSlicePool struct {
    pool sync.Pool
}

func (p *ByteSlicePool) Get() []byte {
    return p.pool.Get().([]byte)[:0]
}

func (p *ByteSlicePool) Put(b []byte) {
    p.pool.Put(b[:0])
}
```
