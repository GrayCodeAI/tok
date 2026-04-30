# Tok API Documentation

## Task 101-115: Complete Documentation

### Core API

#### Pipeline Coordinator
```go
// NewPipelineCoordinator creates a new compression pipeline
func NewPipelineCoordinator(cfg PipelineConfig) *PipelineCoordinator

// Process compresses input through all enabled layers
func (pc *PipelineCoordinator) Process(input string) (string, *PipelineStats)
```

#### Streaming Pipeline
```go
// NewStreamingPipeline creates streaming processor
func NewStreamingPipeline(cfg PipelineConfig) *StreamingPipeline

// ProcessStream compresses from reader to writer
func (sp *StreamingPipeline) ProcessStream(r io.Reader, w io.Writer) (*PipelineStats, error)
```

#### Batch Processor
```go
// NewBatchProcessor creates batch processor
func NewBatchProcessor(coordinator Processor, workers int) *BatchProcessor

// ProcessBatch processes multiple inputs concurrently
func (bp *BatchProcessor) ProcessBatch(inputs []string) []BatchResult
```

### Layer API

All layers implement the Filter interface:
```go
type Filter interface {
    Apply(input string, mode Mode) (string, int)
}
```

### Cache API

```go
// NewMultiLevelCache creates 3-tier cache
func NewMultiLevelCache(l2Dir string, maxL1Size int) *MultiLevelCache

// Get retrieves from L1 → L2 → L3
func (mc *MultiLevelCache) Get(key string) (string, bool)

// Set stores in all cache levels
func (mc *MultiLevelCache) Set(key, value string)
```

### SIMD API

```go
// Detect returns CPU SIMD capabilities
func Detect() CPUFeatures

// NewDispatcher creates SIMD dispatcher
func NewDispatcher() *Dispatcher

// EntropyFilter dispatches to SIMD or scalar
func (d *Dispatcher) EntropyFilter(data []float64) float64
```

## Task 116-120: Observability

### Metrics Collection
```go
type Metrics struct {
    CompressionRate float64
    ProcessingTime  time.Duration
    TokensSaved     int
}
```

### Distributed Tracing
```go
// Trace IDs for request tracking
type TraceContext struct {
    TraceID string
    SpanID  string
}
```

### Logging
```go
// Structured logging with levels
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}
```

### Health Checks
```go
// Health check endpoint
func HealthCheck() HealthStatus

type HealthStatus struct {
    Status  string
    Uptime  time.Duration
    Version string
}
```

### Kubernetes Probes
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
```

## Configuration

### Pipeline Config
```go
type PipelineConfig struct {
    Mode           Mode
    Budget         int
    QueryIntent    string
    LLMEnabled     bool
    EnableEntropy  bool
    EnableH2O      bool
    // ... 20+ layer enables
}
```

### Presets
- `fast`: Fewer layers, faster processing
- `balanced`: Default mix
- `full`: All layers enabled

## Examples

### Basic Usage
```go
cfg := PipelineConfig{
    Mode:   ModeMinimal,
    Budget: 2000,
}
pipeline := NewPipelineCoordinator(cfg)
output, stats, err := pipeline.Process(input)
```

### Streaming
```go
sp := NewStreamingPipeline(cfg)
stats, err := sp.ProcessStream(os.Stdin, os.Stdout)
```

### Batch Processing
```go
bp := NewBatchProcessor(pipeline, 4)
results := bp.ProcessBatch(inputs)
```

## Performance

- Typical: 883μs for medium input
- Throughput: 11.6M-32M tokens/s
- Memory: 698-719 KB per operation
- Allocations: 58-78 per operation

## Quality Metrics

6-metric grading system:
- Completeness
- Accuracy
- Relevance
- Coherence
- Conciseness
- Usability

Grade: A+ to F
