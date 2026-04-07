# TokMan API Documentation

## Overview

TokMan provides a comprehensive gRPC and REST API for token analysis, compression, and management.

- **Base URL**: `https://api.tokman.dev`
- **gRPC Endpoint**: `api.tokman.dev:8083`
- **Auth**: Bearer token in `Authorization` header

## Authentication

All requests require authentication via Bearer token:

```bash
Authorization: Bearer your-api-key
```

Obtain an API key from your [dashboard](https://tokman.dev/dashboard).

## Response Format

### Success (2xx)

```json
{
  "success": true,
  "data": {
    "tokens_saved": 5000,
    "compression_ratio": 0.75
  }
}
```

### Error (4xx, 5xx)

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Retry after 30 seconds.",
    "details": {
      "limit": 1000,
      "used": 1001,
      "reset_at": "2024-01-15T12:00:00Z"
    }
  }
}
```

## API Endpoints

### Analysis

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/analyze` | Analyze code for compression |
| `POST` | `/analyze-batch` | Analyze multiple files |
| `POST` | `/analyze-stream` | Stream large file analysis |
| `GET` | `/analyze/{id}` | Get analysis result |

### Analytics

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/analytics/dashboard` | Get dashboard data |
| `GET` | `/analytics/stats` | Get aggregated stats |
| `GET` | `/analytics/trends` | Get historical trends |
| `GET` | `/analytics/filters` | Get filter performance |

### Team Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/team` | Get team info |
| `PUT` | `/team` | Update team settings |
| `GET` | `/team/users` | List team members |
| `POST` | `/team/users` | Add team member |
| `DELETE` | `/team/users/{id}` | Remove team member |

### Accounting

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/billing/usage` | Get usage metrics |
| `GET` | `/billing/invoices` | List invoices |
| `PUT` | `/billing/payment-method` | Update payment method |

## Rate Limiting

- **Free Tier**: 100 requests/day
- **Pro Tier**: 10,000 requests/day
- **Enterprise**: Unlimited

Rate limit headers:

```
X-RateLimit-Limit: 10000
X-RateLimit-Remaining: 9950
X-RateLimit-Reset: 1705330800
```

## Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `AUTHENTICATION_ERROR` | 401 | Invalid or missing API key |
| `AUTHORIZATION_ERROR` | 403 | Permission denied |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `VALIDATION_ERROR` | 400 | Invalid request parameters |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists |
| `SERVER_ERROR` | 500 | Internal server error |

## Pagination

Endpoints returning lists support pagination:

```
GET /team/users?page=1&limit=50
```

Response:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 125,
    "pages": 3
  }
}
```

## Filters

Query parameters for filtering:

```
GET /analytics/stats?start_date=2024-01-01&end_date=2024-01-31&group_by=day
```

## Timestamps

All timestamps are in ISO 8601 format (UTC):

```
2024-01-15T12:30:00Z
```

## Versioning

API versions are specified in the URL:

```
https://api.tokman.dev/v1/analyze
```

Current stable version: **v1**

## Webhooks

Real-time notifications via webhooks:

```json
{
  "event": "analysis_complete",
  "timestamp": "2024-01-15T12:30:00Z",
  "data": {
    "analysis_id": "ana_123",
    "tokens_saved": 5000
  }
}
```

Register webhooks in your [dashboard](https://tokman.dev/dashboard).

## Rate Limits By Tier

### Free Tier
- 100 API requests per day
- 1M tokens per month
- Max 5MB request size

### Pro Tier
- 10,000 API requests per day
- 50M tokens per month
- Max 100MB request size
- Webhook support

### Enterprise Tier
- Unlimited API requests
- Unlimited tokens
- Custom rate limits
- Priority support

## Best Practices

1. **Use Batch API** for multiple files to reduce API calls
2. **Implement Caching** to avoid re-analyzing same content
3. **Handle Rate Limits** gracefully with exponential backoff
4. **Monitor Quotas** and upgrade if nearing limits
5. **Use gRPC** for low-latency, streaming use cases

## SDK Support

Official SDKs available for:
- Python (`tokman-py`)
- Node.js/TypeScript (`tokman-js`)
- Go (`tokman-go`)
- Rust (`tokman-rs`)

See [SDK Documentation](../sdks/README.md) for usage.

## Support

- 📧 support@tokman.dev
- 💬 [Discord Community](https://discord.gg/tokman)
- 📖 [Full API Reference](./reference.md)

## Examples

See [examples/](../examples/) for complete code samples in multiple languages.
