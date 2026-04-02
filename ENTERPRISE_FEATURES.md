# Enterprise Features

This document describes the enterprise-grade features added to TokMan in Phase 9-13.

## Packages Added

### Performance & Optimization (Phase 9)

#### `internal/benchmarking`
Comprehensive performance benchmarking framework for TokMan.

**Features:**
- Multi-type benchmarks: Compression, Pipeline, Memory, Concurrency, End-to-End
- Configurable suites with warmup, iterations, and duration
- Lifecycle hooks for extensibility
- Memory profiling and allocation tracking
- Statistical analysis (P50, P95, P99 latencies)

**CLI Usage:**
```bash
tokman benchmark run [suite]
tokman benchmark list
```

#### `internal/stress`
Stress testing system for evaluating system performance under load.

**Features:**
- Load, spike, soak, and breakdown test types
- Configurable RPS (requests per second)
- Latency percentile tracking
- Resource metrics collection
- Automatic ramp-up and cooldown phases

**CLI Usage:**
```bash
tokman stress run [scenario]
tokman stress scenarios
```

### Enterprise Security & Reliability (Phase 12)

#### `internal/chaos`
Chaos engineering framework for testing system resilience.

**Features:**
- Multiple fault types: latency, error, memory, CPU, network, disk
- Configurable blast radius (scope)
- Safety mechanisms with auto-rollback
- Health check integration
- Kill switch for immediate abort

**CLI Usage:**
```bash
# Available via programmatic API
```

#### `internal/canary`
Canary deployment system for gradual rollouts.

**Features:**
- Multiple strategies: linear, stepped, analysis, shadow
- Automatic traffic splitting
- Metric-based promotion criteria
- Auto-rollback on failure
- Multi-phase deployments

**CLI Usage:**
```bash
tokman canary create [name] --service [svc] --target [version]
tokman canary start [deployment-id]
tokman canary list
```

#### `internal/abtest`
A/B testing framework for experimentation.

**Features:**
- Multiple experiment types: A/B, multivariate, bandit, switchback
- Multiple randomization methods
- Statistical significance calculation
- Automatic winner selection
- Segment-based targeting

**API Usage:**
```go
manager := abtest.NewManager(storage)
experiment, _ := manager.CreateExperiment(config)
variant, _ := manager.AssignVariant(ctx, experiment.ID, userID)
```

### Cost Intelligence (Phase 1 Completion)

#### `internal/costforecast`
Cost forecasting engine for predictive budgeting.

**Features:**
- Multiple forecasting models: linear regression, moving average, exponential smoothing, seasonal decomposition
- Ensemble forecasting with weighted averaging
- Seasonality detection
- Confidence intervals
- Risk level assessment

**Models:**
- Linear Regression: Trend-based prediction
- Moving Average: Smooth short-term fluctuations
- Exponential Smoothing: Weight recent data more heavily
- Seasonal Decomposition: Account for weekly/monthly patterns

#### `internal/budgetalerts`
Budget alert system with multi-channel notifications.

**Features:**
- Multi-level thresholds (warning, critical, emergency)
- Multiple notification channels (email, Slack, webhook, PagerDuty)
- Automatic and manual resolution
- Alert history and audit trail
- Cooldown periods to prevent spam

**Alert Rules:**
- Metric-based conditions
- Aggregation functions (avg, sum, min, max, count)
- Comparison operators

#### `internal/teamcosts`
Team cost allocation and chargeback system.

**Features:**
- Team-based budget management
- Cost allocation rules
- Usage breakdown by service, model, feature
- Variance reporting
- Trend analysis

**Allocation Methods:**
- Percentage-based
- Fixed amount
- Usage-based

### Infrastructure (Phase 11-12)

#### `internal/mcphost`
MCP (Model Context Protocol) host management.

**Features:**
- Multi-server connection management
- Tool discovery and invocation
- Resource access
- Session management
- Event-driven architecture

**Capabilities:**
- Tools: Execute functions on MCP servers
- Resources: Read structured data
- Prompts: Access prompt templates

#### `internal/iteragent`
Iterative agent framework for multi-step reasoning.

**Features:**
- ReAct-style iteration loop
- Tool usage capability
- Memory management (short-term, long-term, working)
- Reflection and self-improvement
- Pause/resume functionality

**Agent Lifecycle:**
1. Think: Reason about current state
2. Decide: Choose next action
3. Execute: Perform action
4. Observe: Process results
5. Reflect: Learn and adapt

## Integration

All packages integrate with existing TokMan infrastructure:

- **Dashboard**: Visualize benchmark results, stress tests, cost forecasts
- **Tracking**: Store experiment results, alert history, allocation data
- **CLI**: Commands for all major operations
- **API**: gRPC/HTTP endpoints for programmatic access

## Configuration

Each package supports configuration via:
- Environment variables
- Config files (TOML)
- Programmatic API

Example configuration:
```toml
[benchmarking]
default_iterations = 100
warmup_runs = 2

[stress]
default_duration = "5m"
max_concurrency = 1000

[budget_alerts]
default_cooldown = "1h"
enable_auto_resolve = true
```

## Testing

All packages include comprehensive test suites:
- Unit tests for core functionality
- Integration tests for cross-package operations
- Benchmark tests for performance-critical paths

Run tests:
```bash
go test ./internal/benchmarking/...
go test ./internal/stress/...
go test ./internal/chaos/...
# etc.
```

## Future Enhancements

Planned improvements:
- [ ] Distributed stress testing
- [ ] ML-based chaos experiments
- [ ] Automated canary analysis
- [ ] Bayesian A/B testing
- [ ] Real-time cost anomaly detection
- [ ] Multi-region MCP hosting
- [ ] Agent collaboration protocols
