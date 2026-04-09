# ADR-001: Pipeline Architecture Decision

**Date:** 2026-04-09
**Status:** Accepted
**Context:** Designing the compression pipeline architecture

## Decision

We will use a 31-layer pipeline where each layer implements a common interface:

```go
type FilterLayer interface {
    Apply(input string, mode Mode) (string, int)
    Name() string
    Enabled() bool
}
```

## Rationale

1. **Modularity:** Each layer is independent and can be enabled/disabled
2. **Extensibility:** New layers can be added without modifying existing code
3. **Testability:** Each layer can be tested in isolation
4. **Research-backed:** Multiple layers based on published research papers

## Consequences

- **Positive:** Easy to add new compression techniques
- **Positive:** Clear separation of concerns
- **Negative:** Potential performance overhead from layer chaining
- **Negative:** More complex configuration

## Notes

Each layer should implement stage gates (`shouldSkip*()`) for early exit optimization.
