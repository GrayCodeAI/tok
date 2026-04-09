# ADR-005: Resilience Patterns

**Date:** 2026-04-09
**Status:** Accepted

## Decision

Implement circuit breaker pattern for external service calls:

```go
cb := resilience.NewCircuitBreaker("compression",
    resilience.WithMaxFailures(5),
    resilience.WithTimeout(30*time.Second),
)
```

## Rationale

1. **Fault isolation:** Prevent cascade failures
2. **Graceful degradation:** Service remains available in degraded mode
3. **Recovery:** Automatic recovery after timeout

## Implementation

- **State:** Closed (normal) → Open (failing) → Half-open (testing)
- **Failure threshold:** 5 consecutive failures
- **Recovery timeout:** 30 seconds

## Consequences

- **Positive:** Handles external service failures
- **Positive:** Prevents system overload
- **Negative:** Additional complexity
