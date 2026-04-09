# ADR-003: HTTP API Design

**Date:** 2026-04-09
**Status:** Accepted

## Decision

The HTTP API will follow REST principles with JSON payloads:

- `/health` - Health checks (liveness, readiness)
- `/api/v1/compress` - Compression endpoint
- `/api/v1/metrics` - Metrics endpoint
- `/api/v1/config` - Configuration management

## Rationale

1. **Familiarity:** REST + JSON is widely understood
2. **Tooling:** Excellent client library support
3. **Standards:** Follows Kubernetes probe conventions
4. **Simplicity:** No complex authentication initially

## API Response Format

```json
{
  "error": "ERR_CODE",
  "message": "Human readable message",
  "data": {}
}
```

## Consequences

- **Positive:** Easy to integrate with other tools
- **Positive:** Good client library support
- **Negative:** Not as efficient as gRPC for high-frequency calls
