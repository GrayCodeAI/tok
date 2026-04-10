# TokMan API Reference

This document describes the internal package APIs for developers.

## Core Packages

### `internal/core/`

Core functionality for command execution and token estimation.

#### CommandRunner Interface

```go
type CommandRunner interface {
    Run(ctx context.Context, cmd string, args ...string) (string, error)
    RunWithStdin(ctx context.Context, input string, cmd string, args ...string) (string, error)
}
```

#### OSCommandRunner

```go
type OSCommandRunner struct {
    logger  *utils.Logger
    timeout time.Duration
    env     []string
    dir     string
}

func NewOSCommandRunner(opts ...Option) *OSCommandRunner
func (r *OSCommandRunner) Run(ctx context.Context, cmd string, args ...string) (string, error)
func (r *OSCommandRunner) RunWithStdin(ctx context.Context, input string, cmd string, args ...string) (string, error)
func (r *OSCommandRunner) WithTimeout(timeout time.Duration) *OSCommandRunner
func (r *OSCommandRunner) WithEnv(env []string) *OSCommandRunner
func (r *OSCommandRunner) WithDir(dir string) *OSCommandRunner
```

#### Token Estimator

```go
func EstimateTokens(text string) int
func EstimateTokensDetailed(text string) TokenEstimate

type TokenEstimate struct {
    Tiktoken    int     // tiktoken count
    Approximate int     // len(text) / approximation
    Characters  int     // Total characters
    Words       int     // Total words
    Lines       int     // Total lines
}
```

#### Execution Result

```go
type ExecutionResult struct {
    Command      string
    Args         []string
    Output       string
    Stderr       string
    ExitCode     int
    Duration     time.Duration
    TokensBefore int
    TokensAfter  int
    TokensSaved  int
    Error        error
}
```

---

### `internal/filter/`

Practical 20-layer compression pipeline.

#### Filter Types

```go
type Mode int

const (
    ModeNone      Mode = iota
    ModeMinimal
    ModeAggressive
)

func ParseMode(s string) (Mode, error)
func (m Mode) String() string
```

#### Layer Interface

```go
type Layer interface {
    Name() string
    Apply(input string, mode Mode) (string, int)
    ShouldSkip(input string, config PipelineConfig) bool
    Priority() int
}
```

#### Pipeline Configuration

```go
type PipelineConfig struct {
    Mode            Mode
    QueryIntent     string
    Budget          int
    LLMEnabled      bool
    SessionTracking bool
    
    // Layer enables (boolean flags for each layer)
    EnableEntropy           bool
    EnablePerplexity        bool
    EnableGoalDriven        bool
    EnableAST               bool
    EnableContrastive       bool
    EnableNgram             bool
    EnableEvaluatorHeads    bool
    EnableGist              bool
    EnableHierarchical      bool
    EnableBudget            bool
    EnableCompaction        bool
    EnableAttribution       bool
    EnableH2O               bool
    EnableAttentionSink     bool
    EnableMetaToken         bool
    EnableSemanticChunk     bool
    EnableSketchStore       bool
    EnableLazyPruner        bool
    EnableSemanticAnchor    bool
    EnableAgentMemory       bool
    
    // Additional layers
    EnableQuestionAware     bool
    EnableDensityAdaptive   bool
}

func DefaultPipelineConfig() PipelineConfig
func (c *PipelineConfig) WithMode(mode Mode) *PipelineConfig
func (c *PipelineConfig) WithBudget(budget int) *PipelineConfig
func (c *PipelineConfig) WithQueryIntent(intent string) *PipelineConfig
```

#### Pipeline Coordinator

```go
type PipelineCoordinator struct {
    config  PipelineConfig
    layers  []Layer
    stats   PipelineStats
}

func NewPipelineCoordinator(config PipelineConfig) *PipelineCoordinator
func (p *PipelineCoordinator) Process(input string) (string, PipelineStats)
func (p *PipelineCoordinator) ProcessStream(ctx context.Context, reader io.Reader, writer io.Writer) error
func (p *PipelineCoordinator) Stats() PipelineStats
func (p *PipelineCoordinator) Reset()
```

#### Pipeline Statistics

```go
type PipelineStats struct {
    TotalTokens      int
    SavedTokens      int
    CompressionRatio float64
    LayersApplied    int
    LayersSkipped    int
    Duration         time.Duration
    QualityMetrics   QualityMetrics
}
```

#### Quality Metrics

```go
type QualityMetrics struct {
    SemanticScore     float64  // 0.0 - 1.0
    SignalToNoise     float64  // 0.0 - 1.0
    ContextRetention  float64  // 0.0 - 1.0
    Readability       float64  // 0.0 - 1.0
    Completeness      float64  // 0.0 - 1.0
    CompressionRatio  float64  // 0.0 - 1.0
}

func CalculateQualityMetrics(input, output string) QualityMetrics
func (m QualityMetrics) OverallScore() float64
func (m QualityMetrics) Grade() string  // A+ to F
func (m QualityMetrics) IsAcceptable() bool
func (m QualityMetrics) Recommendations() []string
```

#### Presets

```go
type Preset struct {
    Name        string
    Mode        Mode
    Layers      map[string]bool
    Description string
}

func GetPreset(name string) (Preset, error)
func ListPresets() []string
```

---

### `internal/config/`

Configuration management.

#### Config Structure

```go
type Config struct {
    Tracking  TrackingConfig  `mapstructure:"tracking"`
    Filter    FilterConfig    `mapstructure:"filter"`
    Pipeline  PipelineConfig  `mapstructure:"pipeline"`
    Hooks     HooksConfig     `mapstructure:"hooks"`
    Dashboard DashboardConfig `mapstructure:"dashboard"`
}

func LoadConfig() (*Config, error)
func LoadConfigFromFile(path string) (*Config, error)
func (c *Config) Validate() error
func (c *Config) Save() error
```

#### Tracking Configuration

```go
type TrackingConfig struct {
    Enabled      bool   `mapstructure:"enabled"`
    DatabasePath string `mapstructure:"database_path"`
    Retention    string `mapstructure:"retention"`
}
```

#### Filter Configuration

```go
type FilterConfig struct {
    Mode            string `mapstructure:"mode"`
    Preset          string `mapstructure:"preset"`
    DefaultBudget   int    `mapstructure:"default_budget"`
    EnableAggressive bool  `mapstructure:"enable_aggressive"`
}
```

#### Hooks Configuration

```go
type HooksConfig struct {
    ExcludedCommands []string `mapstructure:"excluded_commands"`
    AutoInstall      bool     `mapstructure:"auto_install"`
    VerifyIntegrity  bool     `mapstructure:"verify_integrity"`
}
```

#### Dashboard Configuration

```go
type DashboardConfig struct {
    Enabled  bool   `mapstructure:"enabled"`
    Port     int    `mapstructure:"port"`
    BindAddr string `mapstructure:"bind_addr"`
}
```

---

### `internal/tracking/`

Command tracking and analytics.

#### Tracker

```go
type Tracker struct {
    db *sql.DB
}

func NewTracker(dbPath string) (*Tracker, error)
func (t *Tracker) Close() error
func (t *Tracker) Record(record CommandRecord) error
func (t *Tracker) GetStats() (*TrackingStats, error)
func (t *Tracker) GetTopCommands(n int) ([]CommandStats, error)
func (t *Tracker) GetHistory(limit int) ([]CommandRecord, error)
func (t *Tracker) Clear() error
```

#### Command Record

```go
type CommandRecord struct {
    ID           string    `json:"id"`
    Command      string    `json:"command"`
    Args         []string  `json:"args"`
    OutputTokens int       `json:"output_tokens"`
    SavedTokens  int       `json:"saved_tokens"`
    Timestamp    time.Time `json:"timestamp"`
    Duration     time.Duration `json:"duration"`
}
```

#### Tracking Statistics

```go
type TrackingStats struct {
    TotalCommands  int
    TotalTokens    int
    TotalSaved     int
    AvgCompression float64
    TopCommand     string
    TotalDuration  time.Duration
}
```

---

### `internal/toml/`

TOML filter configuration.

#### Filter Definition

```go
type FilterDef struct {
    Name         string
    Match        string
    OutputPatterns []string
    StripLines     []string
    MaxLines       int
    Description    string
}

func ParseFilterDef(data []byte) (*FilterDef, error)
func (f *FilterDef) MatchCommand(cmd string) bool
func (f *FilterDef) Apply(output string) string
```

#### Filter Registry

```go
type FilterRegistry struct {
    filters map[string]*FilterDef
}

func NewFilterRegistry() *FilterRegistry
func (r *FilterRegistry) Register(name string, def *FilterDef)
func (r *FilterRegistry) Match(cmd string) *FilterDef
func (r *FilterRegistry) LoadFromFile(path string) error
func (r *FilterRegistry) LoadFromDir(dir string) error
```

---

### `internal/integrity/`

Hook integrity verification.

```go
func RuntimeCheck() ([]IntegrityResult, error)
func StoreHash(name string, content []byte) error
func RemoveHash(name string) error
func VerifyHash(name string, content []byte) (bool, error)
```

#### Integrity Result

```go
type IntegrityResult struct {
    Name    string
    Status  IntegrityStatus
    Message string
    Hash    string
}

type IntegrityStatus int

const (
    StatusOK IntegrityStatus = iota
    StatusModified
    StatusMissing
    StatusUnknown
)
```

---

### `internal/tee/`

Output tee and logging.

```go
type OutputTee struct {
    writers []io.Writer
}

func NewOutputTee(initial ...io.Writer) *OutputTee
func (t *OutputTee) Write(p []byte) (n int, err error)
func (t *OutputTee) AddWriter(w io.Writer)
func (t *OutputTee) RemoveWriter(w io.Writer)
```

---

### `internal/dashboard/`

Web dashboard for analytics.

```go
type Dashboard struct {
    server  *http.Server
    tracker *tracking.Tracker
}

func NewDashboard(addr string, tracker *tracking.Tracker) *Dashboard
func (d *Dashboard) Start() error
func (d *Dashboard) Stop() error
func (d *Dashboard) Addr() string
```

---

### `internal/economics/`

Cost analysis utilities.

```go
type CostConfig struct {
    Model         string
    InputCost     float64 // Per 1M tokens
    OutputCost    float64 // Per 1M tokens
    Currency      string
}

func CalculateSavings(tokensSaved int, config CostConfig) SavingsReport

type SavingsReport struct {
    TokensSaved       int
    CostSaved         float64
    ProjectedMonthly  float64
    ProjectedYearly   float64
    Efficiency        string
}
```

---

### `internal/telemetry/`

Telemetry collection (opt-in).

```go
type TelemetryClient struct {
    enabled bool
    endpoint string
}

func NewTelemetryClient(endpoint string) *TelemetryClient
func (c *TelemetryClient) Enable()
func (c *TelemetryClient) Disable()
func (c *TelemetryClient) Track(event TelemetryEvent) error
```

#### Telemetry Event

```go
type TelemetryEvent struct {
    EventName   string
    Properties  map[string]interface{}
    Timestamp   time.Time
    Anonymized  bool
}
```

---

## Common Patterns

### Builder Pattern

Most components support builder/option pattern:

```go
runner := core.NewOSCommandRunner(
    core.WithTimeout(30 * time.Second),
    core.WithEnv(os.Environ()),
    core.WithLogger(logger),
)
```

### Error Handling

Use named errors and wrapping:

```go
var ErrFilterNotFound = errors.New("filter not found")
var ErrBudgetExceeded = errors.New("token budget exceeded")

// Usage
if errors.Is(err, ErrFilterNotFound) {
    // Handle specific error
}

if err != nil {
    return fmt.Errorf("failed to process: %w", err)
}
```

### Context Usage

All long-running operations accept context:

```go
func Process(ctx context.Context, input string) (string, error)
```

### Configuration Loading

Priority order (last wins):
1. Default values
2. Config file (~/.config/tokman/config.toml)
3. Environment variables (TOKMAN_*)
4. CLI flags

---

**Note:** This API reference covers the public interfaces. Internal implementations may change without notice. Always use the provided interfaces for stability.
