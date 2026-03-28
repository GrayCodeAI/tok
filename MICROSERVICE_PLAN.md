# TokMan Microservice Architecture Plan

**Date:** 2026-03-28
**Current:** Monolithic CLI with 27 internal packages
**Target:** Domain-driven modular architecture with microservice-ready boundaries

---

## Current Structure Analysis

| Package | Files | Purpose | Microservice Candidate |
|---------|-------|---------|------------------------|
| `filter/` | 230 | 31-layer compression pipeline | **Compression Service** |
| `commands/` | 178 | CLI command handlers | **Command Service** |
| `tracking/` | 4 | SQLite metrics | **Analytics Service** |
| `config/` | 5 | Configuration | **Config Service** |
| `agents/` | 2 | Agent integrations | **Agent Service** |
| `llm/` | 3 | LLM summarization | **LLM Service** |
| `dashboard/` | 7 | Web UI | **Dashboard Service** |
| `server/` | 3 | HTTP server | **API Gateway** |
| `core/` | 7 | Command runner | Shared library |
| `toml/` | 5 | TOML filters | Shared library |

---

## Proposed Architecture

```
tokman/
├── cmd/
│   ├── tokman/           # CLI entry point
│   ├── tokman-server/    # HTTP server entry point
│   └── tokman-worker/    # Background worker entry point
│
├── services/             # Domain services (microservice-ready)
│   ├── compression/      # Compression service (filter pipeline)
│   │   ├── service.go    # Service interface
│   │   ├── grpc/         # gRPC server
│   │   └── http/         # HTTP handler
│   │
│   ├── commands/         # Command execution service
│   │   ├── service.go
│   │   └── handlers/
│   │
│   ├── analytics/        # Tracking/analytics service
│   │   ├── service.go
│   │   └── repository/
│   │
│   ├── agent/            # Agent integration service
│   │   ├── service.go
│   │   └── providers/
│   │
│   └── llm/              # LLM service
│       ├── service.go
│       └── providers/
│
├── internal/             # Shared internal packages
│   ├── core/             # Core utilities (runner, estimator)
│   ├── config/           # Configuration
│   ├── toml/             # TOML filter loader
│   └── utils/            # Utilities
│
├── pkg/                  # Public packages
│   ├── api/              # Public API types
│   └── client/           # Client SDK
│
├── proto/                # Protocol buffer definitions
│   ├── compression.proto
│   ├── analytics.proto
│   └── agent.proto
│
├── deployments/          # Deployment configs
│   ├── docker/
│   ├── kubernetes/
│   └── docker-compose.yml
│
└── api/                  # OpenAPI specs
    └── openapi.yaml
```

---

## Service Definitions

### 1. Compression Service
**Port:** 50051 (gRPC), 8081 (HTTP)
**Responsibility:** 31-layer compression pipeline

```go
type CompressionService interface {
    Compress(ctx context.Context, req *CompressRequest) (*CompressResponse, error)
    GetStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error)
    GetLayers(ctx context.Context) ([]LayerInfo, error)
}
```

### 2. Command Service
**Port:** 50052 (gRPC), 8082 (HTTP)
**Responsibility:** Execute and filter CLI commands

```go
type CommandService interface {
    Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
    GetFilters(ctx context.Context) ([]FilterInfo, error)
}
```

### 3. Analytics Service
**Port:** 50053 (gRPC), 8083 (HTTP)
**Responsibility:** Token tracking, metrics, economics

```go
type AnalyticsService interface {
    Record(ctx context.Context, req *RecordRequest) error
    GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)
    GetEconomics(ctx context.Context) (*EconomicsResponse, error)
}
```

### 4. Agent Service
**Port:** 50054 (gRPC), 8084 (HTTP)
**Responsibility:** Agent integrations (Claude, Cursor, etc.)

```go
type AgentService interface {
    Install(ctx context.Context, agent string) error
    Uninstall(ctx context.Context, agent string) error
    List(ctx context.Context) ([]AgentInfo, error)
}
```

### 5. LLM Service
**Port:** 50055 (gRPC), 8085 (HTTP)
**Responsibility:** LLM-based compression

```go
type LLMService interface {
    Summarize(ctx context.Context, req *SummarizeRequest) (*SummarizeResponse, error)
    IsAvailable(ctx context.Context) bool
}
```

---

## Migration Phases

### Phase 1: Create Service Interfaces ✅ COMPLETE
- ✅ Defined service interfaces in `services/*/service.go`
- ✅ Services wrap existing `internal/` packages
- ✅ No breaking changes to existing code

### Phase 2: Add gRPC/HTTP Handlers ✅ COMPLETE
- ✅ Created `proto/` definitions (compression.proto, analytics.proto)
- ✅ Generated gRPC code in `pkg/api/proto/*/`
- ✅ Created gRPC servers in `services/*/grpc/`
- ✅ Integrated gRPC into `cmd/tokman-server/main.go`
- ✅ HTTP handlers via existing `internal/server/`

### Phase 3: Docker & Kubernetes ✅ COMPLETE
- ✅ Created Dockerfile for each service
- ✅ Created docker-compose for local dev
- ✅ Created Kubernetes manifests

### Phase 4: CLI → Service Migration ✅ COMPLETE
- ✅ CLI calls services via gRPC
- ✅ Support both local and remote services
- ✅ Backward compatible
- ✅ Remote mode flags (--remote, --compression-addr, --analytics-addr)

### Phase 5: Service Discovery & Scaling ✅ COMPLETE
- ✅ Service discovery interface (internal/discovery/discovery.go)
- ✅ Load balancers (round-robin, weighted, least-connection)
- ✅ Service resolver with automatic instance selection
- ✅ Health checker with periodic monitoring
- ✅ Scaling policies for auto-scaling

### Phase 6: Observability & Metrics ✅ COMPLETE
- ✅ Prometheus metrics package (internal/metrics/metrics.go)
- ✅ Compression metrics: requests, tokens, duration, savings
- ✅ gRPC metrics: requests, latency per method
- ✅ Discovery metrics: instances, health checks
- ✅ Load balancer metrics: selections per instance
- ✅ Cache metrics: hits, misses, size
- ✅ /metrics endpoint integrated via promhttp
- ✅ All gRPC servers instrumented with metrics

---

## Benefits

| Current | After Microservices |
|---------|---------------------|
| Single binary | Modular services |
| Shared memory | Isolated processes |
| Single deployment | Independent deployment |
| Fixed scaling | Horizontal scaling |
| Single point of failure | Fault isolation |

---

## Approval Required

1. **Scope:** Full microservices or just modular architecture?
2. **Communication:** gRPC only, HTTP only, or both?
3. **Deployment:** Docker Compose only or Kubernetes ready?
4. **Timeline:** All phases or Phase 1-2 first?

**Recommendation:** Start with Phase 1-2 (modular with gRPC/HTTP interfaces), defer Phase 3-5.

