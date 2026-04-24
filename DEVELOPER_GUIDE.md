# Tok Developer Quick Reference

## 🚀 New Components Usage

### Rate Limiting
```go
import "github.com/GrayCodeAI/tok/internal/ratelimit"

// Check if request allowed
if !ratelimit.CheckGlobal() {
    return ErrRateLimitExceeded
}

// Wait for rate limit (with context)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := ratelimit.WaitGlobal(ctx); err != nil {
    return err
}
```

### Input Validation
```go
import "github.com/GrayCodeAI/tok/internal/validation"

// Validate input size (max 10MB)
if err := validation.ValidateInputSize(input); err != nil {
    return err
}

// Validate command arguments
if err := validation.ValidateCommandArgs(args); err != nil {
    return err
}

// Sanitize paths (prevent traversal)
safePath, err := validation.SanitizePath(userPath)
if err != nil {
    return err
}
```

### Pipeline Coordinator Pooling
```go
import "github.com/GrayCodeAI/tok/internal/filter"

// Use default pool
pool := filter.GetDefaultPool()
coord := pool.Get()
defer pool.Put(coord)

output, stats := coord.Process(input)

// Or create custom pool
customPool := filter.NewCoordinatorPool(myConfig)
```

### Global State Management
```go
import "github.com/GrayCodeAI/tok/internal/state"

// Get global state manager
mgr := state.Global()

// Set/get config
mgr.SetConfig(cfg)
config := mgr.GetConfig()

// Set/get flags
mgr.SetFlags(verbose, dryRun, ultraCompact, queryIntent, budget)
verbose, dryRun, _, _, _ := mgr.GetFlags()

// Check verbose mode
if mgr.IsVerbose() {
    log.Println("Debug info")
}
```

### Database Retry Logic
```go
import "github.com/GrayCodeAI/tok/internal/retry"

// Simple retry
err := retry.Do(ctx, retry.DefaultConfig(), func() error {
    return db.Exec(query, args...)
})

// Retry with result
result, err := retry.DoWithResult(ctx, retry.DefaultConfig(), func() (*Result, error) {
    return db.Query(query)
})

// Custom config
cfg := retry.Config{
    MaxAttempts: 5,
    InitialWait: 200 * time.Millisecond,
    MaxWait:     10 * time.Second,
    Multiplier:  2.0,
}
```

### Circuit Breaker
```go
import "github.com/GrayCodeAI/tok/internal/breaker"

// Create breaker (5 failures, 30s timeout)
breaker := breaker.New(5, 30*time.Second)

// Call with protection
err := breaker.Call(func() error {
    return externalService.Call()
})

if err == breaker.ErrCircuitOpen {
    // Circuit is open, fail fast
    return err
}

// Check state
switch breaker.State() {
case breaker.StateClosed:
    // Normal operation
case breaker.StateOpen:
    // Circuit open, rejecting requests
case breaker.StateHalfOpen:
    // Testing if service recovered
}
```

### TTL Cache
```go
import "github.com/GrayCodeAI/tok/internal/ttlcache"

// Create cache (5min TTL, 100MB max)
cache := ttlcache.New(5*time.Minute, 100*1024*1024)

// Set with size
cache.Set("key", value, sizeInBytes)

// Get
if val, ok := cache.Get("key"); ok {
    // Use value
}

// Delete
cache.Delete("key")

// Clear all
cache.Clear()

// Get stats
items, totalSize := cache.Stats()
```

## 🔧 Migration Guide

### Before (Old Code)
```go
// Creating new coordinator every time
func processCommand(input string) string {
    coord := filter.NewPipelineCoordinator(&config)
    output, _ := coord.Process(input)
    return output
}

// No validation
func handleCommand(args []string) error {
    return executeCommand(args)
}

// No retry
func saveRecord(record *Record) error {
    _, err := db.Exec(query, record)
    return err
}
```

### After (New Code)
```go
// Use pooled coordinator
func processCommand(input string) string {
    // Validate input
    if err := validation.ValidateInputSize(input); err != nil {
        return ""
    }
    
    // Check rate limit
    if !ratelimit.CheckGlobal() {
        return ""
    }
    
    // Use pool
    pool := filter.GetDefaultPool()
    coord := pool.Get()
    defer pool.Put(coord)
    
    output, _ := coord.Process(input)
    return output
}

// With validation
func handleCommand(args []string) error {
    if err := validation.ValidateCommandArgs(args); err != nil {
        return err
    }
    return executeCommand(args)
}

// With retry
func saveRecord(ctx context.Context, record *Record) error {
    return retry.Do(ctx, retry.DefaultConfig(), func() error {
        _, err := db.Exec(query, record)
        return err
    })
}
```

## 📊 Performance Tips

1. **Always use coordinator pooling** - 10-20x faster than creating new
2. **Validate early** - Fail fast on invalid input
3. **Use rate limiting** - Protect against DoS
4. **Add retry logic** - Handle transient failures
5. **Use circuit breakers** - Prevent cascading failures
6. **Cache with TTL** - Prevent memory leaks

## 🧪 Testing

```bash
# Run all tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./internal/filter/
go test -bench=. ./internal/ratelimit/

# Test rate limiting
for i in {1..300}; do tok status & done

# Test input validation
dd if=/dev/zero bs=1M count=20 | tok compress

# Test coordinator pooling
go test -bench=BenchmarkPool -benchmem ./internal/filter/
```

## 🐛 Common Issues

### Rate Limit Exceeded
```go
// Solution: Use Wait instead of Allow
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := ratelimit.WaitGlobal(ctx); err != nil {
    // Handle timeout
}
```

### Input Too Large
```go
// Solution: Stream or chunk large inputs
if len(input) > validation.MaxInputSize {
    // Process in chunks or use streaming mode
}
```

### Circuit Breaker Open
```go
// Solution: Wait for recovery or use fallback
if err == breaker.ErrCircuitOpen {
    // Use cached result or return degraded response
    return fallbackResponse()
}
```

## 📚 Additional Resources

- [FIXES_IMPLEMENTED.md](./FIXES_IMPLEMENTED.md) - Detailed implementation guide
- [internal/ratelimit/ratelimit.go](./internal/ratelimit/ratelimit.go) - Rate limiter source
- [internal/validation/validator.go](./internal/validation/validator.go) - Validator source
- [internal/filter/pool.go](./internal/filter/pool.go) - Pool source
- [internal/retry/retry.go](./internal/retry/retry.go) - Retry source
- [internal/breaker/breaker.go](./internal/breaker/breaker.go) - Circuit breaker source
- [internal/ttlcache/cache.go](./internal/ttlcache/cache.go) - Cache source
