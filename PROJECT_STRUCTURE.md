# TokMan Project Structure

This document describes the organization of the TokMan codebase following best practices from top Go projects (Stripe, HashiCorp, Kubernetes, etc.).

## Directory Layout

```
tokman/
├── api/                    # API definitions (Protocol Buffers, OpenAPI)
│   ├── proto/             # Common proto files
│   └── v1/                # API version 1 proto files
├── build/                 # Build artifacts and outputs
│   └── dist/              # Distribution binaries
├── cmd/                   # Main applications (main packages)
│   └── tokman/            # Main CLI entry point
├── configs/               # Configuration files and templates
│   ├── filters/           # TOML filter definitions
│   └── tokman.yaml        # Default configuration
├── docs/                  # Documentation
├── examples/              # Example usage and configurations
├── internal/              # Private application code
│   ├── commands/          # CLI command implementations
│   ├── config/            # Configuration loading
│   ├── core/              # Core utilities (runner, estimator)
│   ├── errors/            # Error definitions
│   ├── filter/            # Compression filters
│   ├── security/          # Security scanning
│   ├── server/            # HTTP server
│   ├── session/           # Session management
│   ├── tracking/          # Analytics and tracking
│   └── utils/             # Utility functions
├── pkg/                   # Public library code (can be imported)
│   └── lib/               # Shared libraries
├── scripts/               # Build and automation scripts
├── services/              # Microservices
│   ├── api-gateway/       # HTTP API Gateway
│   ├── compression-service/  # Compression pipeline service
│   ├── analytics-service/    # Analytics and tracking service
│   ├── security-service/     # Security scanning service
│   └── config-service/       # Configuration service
├── test/                  # Test files
│   ├── benchmarks/        # Performance benchmarks
│   ├── e2e/              # End-to-end tests
│   ├── fixtures/         # Test fixtures
│   ├── integration/      # Integration tests
│   └── unit/             # Unit tests
├── web/                   # Web assets and handlers
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile               # Build automation
├── README.md              # Main documentation
└── docker-compose.yml     # Docker orchestration
```

## Key Directories

### `/api`
API definitions and Protocol Buffer files.
- Follows the pattern used by Kubernetes and gRPC projects
- Versioned APIs under `v1/`, `v2/`, etc.

### `/cmd`
Main applications for this project.
- Each subdirectory is a binary (e.g., `cmd/tokman/main.go`)
- Follows the golang-standards/project-layout

### `/configs`
Configuration files and templates.
- Default configurations
- Example configs
- TOML filter definitions

### `/internal`
Private application code.
- Cannot be imported by other projects
- Contains business logic, services, repositories

### `/pkg`
Public library code.
- Can be imported by other projects
- Stable APIs with versioning

### `/services`
Microservices architecture.
- Each service has its own `cmd/`, `internal/`, `proto/`
- Follows microservice best practices from Stripe, HashiCorp

### `/test`
Test files organized by type.
- `unit/` - Unit tests
- `integration/` - Integration tests
- `e2e/` - End-to-end tests
- `benchmarks/` - Performance benchmarks
- `fixtures/` - Test data

### `/scripts`
Build and automation scripts.
- `build.sh` - Build automation
- `test.sh` - Test runner
- `deploy.sh` - Deployment scripts

### `/build`
Build artifacts (gitignored).
- Compiled binaries in `dist/`
- Temporary build files

## Design Principles

1. **Clear Boundaries**: Public (`pkg/`) vs Private (`internal/`)
2. **Separation of Concerns**: Commands, services, repositories separated
3. **Testability**: Tests organized by type, mirrors source structure
4. **Scalability**: Microservices can scale independently
5. **Maintainability**: Consistent structure, clear naming

## Comparison with Top Projects

| Pattern | TokMan | Kubernetes | Docker | HashiCorp |
|---------|--------|------------|--------|-----------|
| Main entry | `cmd/` | `cmd/` | `cmd/` | `cmd/` |
| Private code | `internal/` | `pkg/` | `internal/` | `internal/` |
| Public libs | `pkg/` | `staging/` | `pkg/` | `sdk/` |
| API defs | `api/` | `api/` | `api/` | `proto/` |
| Configs | `configs/` | N/A | N/A | N/A |
| Tests | `test/` | `test/` | N/A | N/A |

## Adding New Code

### New CLI Command
```
cmd/tokman/main.go -> internal/commands/<category>/<command>.go
```

### New Microservice
```
services/<service-name>/
├── cmd/                  # Service entry point
├── internal/
│   ├── handler/         # HTTP/gRPC handlers
│   ├── service/         # Business logic
│   └── repository/      # Data access
└── proto/               # Service proto definitions
```

### New Filter
```
internal/filter/<filter-name>.go
test/unit/filter/<filter-name>_test.go
```

## Build Commands

```bash
# Build main binary
make build

# Build all services
./scripts/build.sh all

# Run tests
make test

# Run specific test type
make test-unit
make test-integration
make test-e2e

# Run benchmarks
make benchmark
```

## References

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Kubernetes Project Structure](https://github.com/kubernetes/kubernetes)
- [Docker Project Structure](https://github.com/docker/docker)
- [Stripe CLI Structure](https://github.com/stripe/stripe-cli)
