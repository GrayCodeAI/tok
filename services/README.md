# TokMan Microservice Architecture

This directory contains the microservice architecture implementation of TokMan.

## Architecture Overview

```
┌─────────────────┐
│   API Gateway   │  ← HTTP/REST API entry point
│    (Port 8080)  │
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┬──────────┐
    │         │          │          │          │
    ▼         ▼          ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│Compression│ │Analytics│ │ Security│ │  Config │ │  Other  │
│Service    │ │ Service │ │ Service │ │ Service │ │Services │
│(gRPC)     │ │ (gRPC)  │ │ (gRPC)  │ │ (gRPC)  │ │ (gRPC)  │
│:50051     │ │ :50052  │ │ :50053  │ │ :50054  │ │         │
└────────┘ └────────┘ └────────┘ └────────┘ └────────┘
```

## Services

### 1. API Gateway (`api-gateway/`)
- **Port**: 8080
- **Protocol**: HTTP/REST
- **Purpose**: Single entry point for all client requests
- **Features**:
  - Request routing to backend services
  - Rate limiting (100 req/sec)
  - CORS support
  - Request logging
  - Panic recovery
  - Health checks

### 2. Compression Service (`compression-service/`)
- **Port**: 50051
- **Protocol**: gRPC
- **Purpose**: Core text compression pipeline
- **Features**:
  - 20+ layer compression pipeline
  - Budget enforcement
  - Layer selection
  - Streaming support
  - Result caching

### 3. Analytics Service (`analytics-service/`)
- **Port**: 50052
- **Protocol**: gRPC
- **Purpose**: Command tracking and analytics
- **Features**:
  - Command recording
  - Token savings tracking
  - Daily/weekly reports
  - Checkpoint telemetry
  - SQLite persistence

### 4. Security Service (`security-service/`)
- **Port**: 50053
- **Protocol**: gRPC
- **Purpose**: Security scanning and validation
- **Features**:
  - PII detection (AWS keys, tokens, credit cards)
  - Content redaction
  - Suspicious pattern detection
  - Path validation
  - Budget validation

### 5. Config Service (`config-service/`)
- **Port**: 50054
- **Protocol**: gRPC
- **Purpose**: Configuration management
- **Features**:
  - Pipeline configuration
  - Layer presets
  - Service discovery
  - Dynamic config updates

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health
curl http://localhost:8080/health/services

# Compress text
curl -X POST http://localhost:8080/api/v1/compress \
  -H "Content-Type: application/json" \
  -d '{"text": "Your text here", "mode": "minimal"}'

# Scan for security issues
curl -X POST http://localhost:8080/api/v1/scan \
  -H "Content-Type: text/plain" \
  -d 'AKIAIOSFODNN7EXAMPLE'

# Redact PII
curl -X POST http://localhost:8080/api/v1/redact \
  -H "Content-Type: text/plain" \
  -d 'Contact: user@example.com, Key: AKIA...'

# View Prometheus metrics
open http://localhost:9090

# View Grafana dashboards
open http://localhost:3000  # admin/admin
```

## API Endpoints

### Compression
- `POST /api/v1/compress` - Compress text
- `POST /api/v1/compress/stream` - Stream compression (SSE)
- `GET /api/v1/filters` - List available filters

### Security
- `POST /api/v1/scan` - Scan content for security issues
- `POST /api/v1/redact` - Redact PII from content

### Analytics
- `GET /api/v1/stats` - Get compression statistics
- `GET /api/v1/commands` - Get recent commands
- `GET /api/v1/savings` - Get token savings

### Health
- `GET /health` - Gateway health
- `GET /health/services` - All services health

## Protocol Buffers

Service definitions are in `shared/proto/` and service-specific protos:

```
services/
├── shared/proto/common.proto
├── compression-service/proto/compression.proto
├── analytics-service/proto/analytics.proto
├── security-service/proto/security.proto
└── config-service/proto/config.proto
```

Generate Go code:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       services/shared/proto/common.proto
```

## Development

### Running Locally

```bash
# Terminal 1: Compression Service
cd services/compression-service
go run cmd/main.go

# Terminal 2: Security Service
cd services/security-service
go run cmd/main.go

# Terminal 3: API Gateway
cd services/api-gateway
go run cmd/main.go
```

### Running Tests

```bash
# Test all services
go test ./services/...

# Test with race detector
go test -race ./services/...

# Run benchmarks
go test -bench=. ./services/...
```

## Deployment

### Kubernetes

See `k8s/` directory for Kubernetes manifests:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/services/
kubectl apply -f k8s/ingress.yaml
```

### Scaling

```bash
# Scale compression service
kubectl scale deployment compression-service --replicas=5

# Scale security service
kubectl scale deployment security-service --replicas=3
```

## Monitoring

### Metrics

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

### Key Metrics

- Request rate/latency by endpoint
- Compression ratio per layer
- Token savings per command
- Security scans per second
- Service health status

## Architecture Decisions

### Why gRPC for Internal Communication?

- **Performance**: Binary protocol, HTTP/2 multiplexing
- **Type Safety**: Protocol buffer definitions
- **Streaming**: Native support for bidirectional streaming
- **Code Generation**: Auto-generated client/server code

### Why API Gateway Pattern?

- **Single Entry Point**: Clients only need to know one endpoint
- **Cross-Cutting Concerns**: Auth, rate limiting, logging in one place
- **Protocol Translation**: HTTP/REST to gRPC
- **Service Aggregation**: Combine multiple service responses

### Service Boundaries

- **Compression Service**: Stateless, CPU-intensive, horizontally scalable
- **Analytics Service**: Stateful (SQLite), single instance or external DB
- **Security Service**: Stateless, horizontally scalable
- **Config Service**: Read-heavy, cached, single instance

## Future Enhancements

- [ ] Service mesh (Istio/Linkerd)
- [ ] Distributed tracing (Jaeger/Zipkin)
- [ ] Circuit breaker pattern
- [ ] Blue/green deployments
- [ ] Auto-scaling based on metrics
- [ ] Multi-region deployment
- [ ] GraphQL federation
