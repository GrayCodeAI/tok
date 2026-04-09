# ADR-002: Error Handling Strategy

**Date:** 2026-04-09
**Status:** Accepted

## Decision

We will use domain-specific errors with error wrapping:

```go
var (
    ErrConfigInvalid     = errors.New("configuration invalid")
    ErrCommandNotFound   = errors.New("command not found")
    ErrCompressionFailed = errors.New("compression failed")
    ErrBudgetExceeded    = errors.New("token budget exceeded")
)
```

## Rationale

1. **Debugging:** Domain errors make it easier to identify issue source
2. **Recovery:** Different errors can trigger different recovery strategies
3. **Testing:** Easier to verify error conditions in tests
4. **Exit codes:** Domain errors map to appropriate exit codes

## Consequences

- **Positive:** Clear error messages
- **Positive:** Structured error handling
- **Negative:** More error types to maintain
