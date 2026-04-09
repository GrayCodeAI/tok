# TokMan HTTP API Reference

This document describes the REST HTTP API endpoints for TokMan.

## Base URL

```
http://localhost:8080/api/v1
```

## Endpoints

### Health

#### GET /health

Returns the health status of the TokMan service.

**Response (200)**
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "timestamp": "2026-04-09T12:00:00Z",
  "components": [
    {
      "name": "config",
      "status": "healthy",
      "message": "Configuration valid",
      "timestamp": "2026-04-09T12:00:00Z"
    }
  ]
}
```

#### GET /health/live

Kubernetes liveness probe.

#### GET /health/ready

Kubernetes readiness probe.

---

### Compression

#### POST /compress

Compresses text using the TokMan pipeline.

**Request**
```json
{
  "text": "text to compress",
  "mode": "minimal",
  "budget": 2000,
  "preset": "balanced"
}
```

**Response (200)**
```json
{
  "original": "text to compress",
  "compressed": "compressed output",
  "stats": {
    "original_tokens": 100,
    "output_tokens": 30,
    "tokens_saved": 70,
    "compression_ratio": 0.3
  }
}
```

---

### Metrics

#### GET /metrics

Returns current metrics snapshot.

**Response (200)**
```json
{
  "commands_processed": 1000,
  "compression_runs": 5000,
  "cache_hits": 3000,
  "cache_misses": 2000,
  "active_connections": 2,
  "memory_usage_mb": 150
}
```

---

### Configuration

#### GET /config

Returns current configuration.

#### PUT /config

Updates configuration.

---

### Errors

Error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human readable message"
}
```

| Code | Status | Description |
|------|--------|-------------|
| ERR_INVALID_INPUT | 400 | Invalid input |
| ERR_NOT_FOUND | 404 | Resource not found |
| ERR_INTERNAL | 500 | Internal error |
