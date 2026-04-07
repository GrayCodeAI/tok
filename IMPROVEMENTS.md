# TokMan Technical Improvements - Detailed Guide

This document provides specific, actionable code improvements with before/after examples.

---

## 1. Remove Panics from Production Code

### Issue Location
`internal/filter/content_detect.go`

### Current Code Pattern
```go
// AVOID: This can crash the application
panic("invalid content type")
```

### Fixed Version
```go
// Recommended approach
type DetectionError struct {
    Input string
    Reason string
}

func (e *DetectionError) Error() string {
    return fmt.Sprintf("detection failed: %s (input: %q)", e.Reason, e.Input[:min(50, len(e.Input))])
}

// Instead of panic, return error
if invalidCondition {
    return DetectionResult{}, &DetectionError{
        Input: input,
        Reason: "failed to match content patterns",
    }
}
```

### Test Case
```go
func TestDetectorErrors(t *testing.T) {
    d := NewDetector()
    
    tests := []struct {
        name     string
        input    string
        wantErr  bool
        errType  string
    }{
        {
            name:     "empty input",
            input:    "",
            wantErr:  false,  // Empty is valid
            errType:  "",
        },
        {
            name:     "invalid UTF-8",
            input:    "\xff\xfe",
            wantErr:  true,
            errType:  "DetectionError",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := d.Detect(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && err != nil {
                if _, ok := err.(*DetectionError); !ok {
                    t.Errorf("Expected DetectionError, got %T", err)
                }
            }
        })
    }
}
```

---

## 2. Standardize Error Context Across Modules

### Issue: Missing Context in Errors
**Problem**: When errors occur, callers don't know which operation failed.

### Before
```go
func (r *OSCommandRunner) Run(ctx context.Context, args []string) (string, int, error) {
    if len(args) == 0 {
        return "", 0, nil  // Silently returns nothing!
    }
    // ...
}
```

### After
```go
// Step 1: Define error types
var (
    ErrEmptyCommand = errors.New("command arguments cannot be empty")
    ErrInvalidCommand = errors.New("command name contains invalid characters")
    ErrCommandNotFound = errors.New("command not found in PATH")
)

// Step 2: Use error wrapping with context
func (r *OSCommandRunner) Run(ctx context.Context, args []string) (string, int, error) {
    // Validate inputs
    if len(args) == 0 {
        return "", 1, fmt.Errorf("Run: %w", ErrEmptyCommand)
    }
    
    if err := validateCommandName(args[0]); err != nil {
        return "", 126, fmt.Errorf("Run: validate command %q: %w", args[0], err)
    }
    
    // Look up command
    cmdPath, err := exec.LookPath(args[0])
    if err != nil {
        return "", 127, fmt.Errorf("Run: lookup %q: %w", args[0], ErrCommandNotFound)
    }
    
    // Execute
    cmd := exec.CommandContext(ctx, cmdPath, args[1:]...)
    cmd.Env = r.Env
    
    output, err := cmd.CombinedOutput()
    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        } else {
            exitCode = 1
            return string(output), exitCode, fmt.Errorf("Run: execute %q: %w", args[0], err)
        }
    }
    
    return string(output), exitCode, nil
}

// Step 3: Test the error handling
func TestRunErrors(t *testing.T) {
    r := NewOSCommandRunner()
    ctx := context.Background()
    
    tests := []struct {
        name    string
        args    []string
        want    error
    }{
        {
            name: "empty args",
            args: []string{},
            want: ErrEmptyCommand,
        },
        {
            name: "invalid command",
            args: []string{"cmd;rm -rf /"},  // Shell metachar
            want: ErrInvalidCommand,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, _, err := r.Run(ctx, tt.args)
            if !errors.Is(err, tt.want) {
                t.Errorf("Run() error = %v, want %v", err, tt.want)
            }
        })
    }
}
```

---

## 3. Fix Resource Cleanup

### Issue: Multiple Defers Can Hide Errors

**Locations affected**:
- `internal/tracking/tracker.go`
- `internal/mcp/cache.go`
- `internal/teamcosts/allocator.go`

### Before: Problematic Pattern
```go
type Tracker struct {
    db *sql.DB
    cache *Cache
    metrics *Metrics
}

func (t *Tracker) Close() error {
    defer t.cache.Close()      // Might return error
    defer t.metrics.Close()    // Might return error
    return t.db.Close()        // Only this error is returned!
}
```

### After: Proper Cleanup
```go
// Step 1: Define cleanup helper
type CloseErrors []error

func (ce CloseErrors) Error() string {
    if len(ce) == 0 {
        return ""
    }
    msgs := make([]string, len(ce))
    for i, err := range ce {
        msgs[i] = err.Error()
    }
    return "close errors: " + strings.Join(msgs, "; ")
}

// Step 2: Implement proper cleanup
func (t *Tracker) Close() error {
    var closeErrs CloseErrors
    
    // Close in reverse order of opening
    if err := t.metrics.Close(); err != nil {
        closeErrs = append(closeErrs, fmt.Errorf("metrics close: %w", err))
    }
    
    if err := t.cache.Close(); err != nil {
        closeErrs = append(closeErrs, fmt.Errorf("cache close: %w", err))
    }
    
    if err := t.db.Close(); err != nil {
        closeErrs = append(closeErrs, fmt.Errorf("db close: %w", err))
    }
    
    if len(closeErrs) > 0 {
        return closeErrs
    }
    return nil
}

// Step 3: Test cleanup
func TestTrackerCloseErrors(t *testing.T) {
    tracker := NewTracker(t.TempDir())
    
    // Inject failure in cache
    tracker.cache = &FailingCache{} // Mock
    
    err := tracker.Close()
    if err == nil {
        t.Error("Close() should return error when cache fails")
    }
    
    // Verify all close methods were called
    if !tracker.cache.closed || !tracker.db.closed || !tracker.metrics.closed {
        t.Error("Not all resources were closed")
    }
}
```

---

## 4. Add Structured Logging

### Issue: No Structured Logging for Observability

### Before: Ad-hoc Logging
```go
func (e *Engine) Process(input string) (string, int) {
    output := input
    totalSaved := 0
    
    for _, filter := range e.filters {
        filtered, saved := filter.Apply(output, e.mode)
        output = filtered
        totalSaved += saved
        fmt.Printf("Applied %s: saved %d tokens\n", filter.Name(), saved)  // Unstructured!
    }
    
    return output, totalSaved
}
```

### After: Structured Logging (Go 1.21+)
```go
import "log/slog"

type Engine struct {
    filters []Filter
    mode    Mode
    logger  *slog.Logger
}

func NewEngine(mode Mode, logger *slog.Logger) *Engine {
    if logger == nil {
        logger = slog.Default()
    }
    return &Engine{
        filters: registerFilters(),
        mode:    mode,
        logger:  logger,
    }
}

func (e *Engine) Process(ctx context.Context, input string) (string, int, error) {
    output := input
    totalSaved := 0
    startTime := time.Now()
    
    e.logger.InfoContext(ctx, "processing started",
        slog.Int("input_length", len(input)),
        slog.String("mode", string(e.mode)),
    )
    
    for _, filter := range e.filters {
        filterStart := time.Now()
        
        if ec, ok := filter.(EnableCheck); ok && !ec.IsEnabled() {
            e.logger.DebugContext(ctx, "filter disabled",
                slog.String("filter", filter.Name()),
            )
            continue
        }
        
        filtered, saved := filter.Apply(output, e.mode)
        output = filtered
        totalSaved += saved
        
        e.logger.DebugContext(ctx, "filter applied",
            slog.String("filter", filter.Name()),
            slog.Int("tokens_saved", saved),
            slog.Duration("duration", time.Since(filterStart)),
        )
    }
    
    e.logger.InfoContext(ctx, "processing completed",
        slog.Int("total_saved", totalSaved),
        slog.Duration("duration", time.Since(startTime)),
        slog.Int("compression_ratio", int(100*float64(totalSaved)/float64(len(input)))),
    )
    
    return output, totalSaved, nil
}

// Setup in main
func main() {
    // JSON logging for production
    opts := &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }
    handler := slog.NewJSONHandler(os.Stderr, opts)
    logger := slog.New(handler)
    slog.SetDefault(logger)
    
    engine := NewEngine(ModeMinimal, logger)
    // ...
}
```

---

## 5. Improve Configuration Validation

### Issue: Errors Happen at Runtime, Not Startup

### Before: Late Validation
```go
type Config struct {
    Budget   int
    Mode     string
    MaxLayer int
    Timeout  time.Duration
}

func (cfg *Config) Use() {
    // Error discovered here during operation!
    if cfg.Budget < 0 {
        fmt.Println("Warning: invalid budget")
    }
}
```

### After: Early Validation
```go
import (
    "fmt"
    "time"
)

// Define validation rules as constants
const (
    MinBudget    = 0
    MaxBudget    = 1000000
    DefaultBudget = 2000
    
    DefaultTimeout = 30 * time.Second
    MaxTimeout     = 5 * time.Minute
)

var ValidModes = map[string]bool{
    "none":       true,
    "minimal":    true,
    "aggressive": true,
}

type ConfigError struct {
    Field   string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config validation: %s: %s", e.Field, e.Message)
}

type Config struct {
    Budget   int
    Mode     string
    MaxLayer int
    Timeout  time.Duration
}

// Validate checks all configuration constraints
func (cfg *Config) Validate() error {
    // Validate Budget
    if cfg.Budget < MinBudget || cfg.Budget > MaxBudget {
        return &ConfigError{
            Field:   "Budget",
            Message: fmt.Sprintf("must be between %d and %d, got %d", MinBudget, MaxBudget, cfg.Budget),
        }
    }
    
    // Validate Mode
    if !ValidModes[cfg.Mode] {
        modes := []string{}
        for m := range ValidModes {
            modes = append(modes, m)
        }
        return &ConfigError{
            Field:   "Mode",
            Message: fmt.Sprintf("must be one of %v, got %q", modes, cfg.Mode),
        }
    }
    
    // Validate MaxLayer
    if cfg.MaxLayer < 1 || cfg.MaxLayer > 31 {
        return &ConfigError{
            Field:   "MaxLayer",
            Message: fmt.Sprintf("must be between 1 and 31, got %d", cfg.MaxLayer),
        }
    }
    
    // Validate Timeout
    if cfg.Timeout < 1*time.Second || cfg.Timeout > MaxTimeout {
        return &ConfigError{
            Field:   "Timeout",
            Message: fmt.Sprintf("must be between 1s and %s, got %s", MaxTimeout, cfg.Timeout),
        }
    }
    
    return nil
}

// LoadConfig loads and validates configuration
func LoadConfig(path string) (*Config, error) {
    cfg := &Config{
        Budget:   DefaultBudget,
        Mode:     "minimal",
        MaxLayer: 31,
        Timeout:  DefaultTimeout,
    }
    
    // Parse from file/env
    if err := cfg.parseFromFile(path); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    
    // Validate before returning
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}

// Test validation
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        cfg     *Config
        wantErr bool
    }{
        {
            name: "valid config",
            cfg: &Config{
                Budget:   2000,
                Mode:     "minimal",
                MaxLayer: 15,
                Timeout:  30 * time.Second,
            },
            wantErr: false,
        },
        {
            name: "negative budget",
            cfg: &Config{
                Budget: -1,
            },
            wantErr: true,
        },
        {
            name: "invalid mode",
            cfg: &Config{
                Budget: 1000,
                Mode:   "invalid",
            },
            wantErr: true,
        },
        {
            name: "timeout too long",
            cfg: &Config{
                Budget:  1000,
                Mode:    "minimal",
                Timeout: 10 * time.Minute,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.cfg.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## 6. Performance: Pre-compile Regexes

### Issue: Regexes Compiled in Hot Path

**Location**: `internal/cortex/detect.go` and other detection modules

### Before: Inefficient
```go
func (d *Detector) Detect(content string) DetectionResult {
    // These regexes are compiled EVERY TIME
    patterns := []*regexp.Regexp{
        regexp.MustCompile(`func\s+\w+`),
        regexp.MustCompile(`package\s+\w+`),
        regexp.MustCompile(`import\s*\(`),
        // ... many more
    }
    
    for _, p := range patterns {
        if p.MatchString(content) {
            // ...
        }
    }
}
```

### After: Efficient
```go
// Compile once at package init
var (
    goFuncPattern     = regexp.MustCompile(`func\s+\w+`)
    goPackagePattern  = regexp.MustCompile(`package\s+\w+`)
    goImportPattern   = regexp.MustCompile(`import\s*\(`)
    pythonDefPattern  = regexp.MustCompile(`def\s+\w+\s*\(`)
    pythonClassPattern = regexp.MustCompile(`class\s+\w+`)
    // ... rest of patterns
)

type Detector struct {
    patterns map[string][]*regexp.Regexp
}

func NewDetector() *Detector {
    return &Detector{
        patterns: map[string][]*regexp.Regexp{
            "go": {goFuncPattern, goPackagePattern, goImportPattern},
            "python": {pythonDefPattern, pythonClassPattern},
            // ...
        },
    }
}

func (d *Detector) Detect(content string) DetectionResult {
    result := DetectionResult{
        Features: make(map[string]bool),
    }
    
    for lang, patterns := range d.patterns {
        for _, p := range patterns {
            if p.MatchString(content) {
                result.Features[lang] = true
                break
            }
        }
    }
    
    return result
}
```

### Benchmark to Prove Impact
```go
func BenchmarkDetectRegexAllocation(b *testing.B) {
    content := loadLargeFile()
    
    b.Run("Before", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            // Old way: compile every time
            re := regexp.MustCompile(`func\s+\w+`)
            re.MatchString(content)
        }
    })
    
    b.Run("After", func(b *testing.B) {
        re := regexp.MustCompile(`func\s+\w+`)
        for i := 0; i < b.N; i++ {
            // New way: reuse compiled regex
            re.MatchString(content)
        }
    })
}
```

---

## 7. Reduce Allocations in Hot Path

### Issue: Unnecessary Map Allocation

**Location**: `internal/filter/filter.go` - `DetectLanguage()` function

### Before
```go
func DetectLanguage(output string) string {
    // Allocates map EVERY CALL
    scores := map[string]int{
        "go": 0, "python": 0, "rust": 0, "javascript": 0,
        "typescript": 0, "java": 0, "c": 0, "cpp": 0,
        "ruby": 0, "sql": 0, "shell": 0,
    }
    
    // ... lots of code checking conditions
    
    bestLang := "unknown"
    bestScore := 0
    for lang, score := range scores {
        if score > bestScore {
            bestScore = score
            bestLang = lang
        }
    }
    
    return bestLang
}
```

### After: Use Array Instead of Map
```go
type LangScore struct {
    Lang string
    Score int
}

var languages = [11]string{
    "go", "python", "rust", "javascript", "typescript",
    "java", "c", "cpp", "ruby", "sql", "shell",
}

const (
    langGo = iota
    langPython
    langRust
    langJavaScript
    langTypeScript
    langJava
    langC
    langCpp
    langRuby
    langSQL
    langShell
    numLangs
)

func DetectLanguageEfficient(output string) string {
    // Allocate once (11 elements, fixed size)
    scores := [numLangs]int{}
    
    // Go indicators
    if strings.Contains(output, "func ") {
        scores[langGo] = 10
    }
    if strings.Contains(output, "package ") {
        scores[langGo] += 5
    }
    // ... etc
    
    // Find max score
    bestIdx := 0
    bestScore := 0
    for i, score := range scores {
        if score > bestScore {
            bestScore = score
            bestIdx = i
        }
    }
    
    if bestScore == 0 {
        return "unknown"
    }
    return languages[bestIdx]
}

// Prove with allocation benchmark
func BenchmarkDetectLanguageAllocations(b *testing.B) {
    content := `
func main() {
    fmt.Println("hello")
}
`
    
    b.Run("Before (map)", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            DetectLanguage(content)
        }
    })
    
    b.Run("After (array)", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            DetectLanguageEfficient(content)
        }
    })
}
```

---

## 8. Add Context Throughout

### Issue: Missing Context Support in Core Functions

### Before
```go
func (e *Engine) Process(input string) (string, int) {
    // No way to cancel long-running operations
    // No way to set timeouts
    // No way to trace requests
}
```

### After
```go
func (e *Engine) Process(ctx context.Context, input string) (string, int, error) {
    // Support cancellation
    select {
    case <-ctx.Done():
        return input, 0, ctx.Err()
    default:
    }
    
    output := input
    totalSaved := 0
    
    for _, filter := range e.filters {
        // Allow cancellation between filters
        select {
        case <-ctx.Done():
            return output, totalSaved, ctx.Err()
        default:
        }
        
        filtered, saved := filter.Apply(output, e.mode)
        output = filtered
        totalSaved += saved
    }
    
    return output, totalSaved, nil
}

// In commands, add timeout
func (cmd *CompressCommand) Run(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    output, saved, err := engine.Process(ctx, input)
    if err != nil {
        return fmt.Errorf("compression timeout or cancelled: %w", err)
    }
    
    return nil
}
```

---

## Implementation Checklist

- [ ] **Remove Panics** (1-2 hours)
  - [ ] Audit all packages for panic calls
  - [ ] Replace with error returns
  - [ ] Add tests for error cases

- [ ] **Add Error Context** (2-3 hours)
  - [ ] Define sentinel errors in core modules
  - [ ] Wrap errors with context
  - [ ] Add error type assertions in tests

- [ ] **Fix Resource Cleanup** (2-3 hours)
  - [ ] Review all defer patterns
  - [ ] Implement CloseErrors helper
  - [ ] Test cleanup failure scenarios

- [ ] **Add Structured Logging** (3-4 hours)
  - [ ] Choose logging library (slog recommended)
  - [ ] Add logger to Engine and main components
  - [ ] Replace printf with structured logs

- [ ] **Improve Config Validation** (1-2 hours)
  - [ ] Define ConfigError type
  - [ ] Move validation to Load time
  - [ ] Add validation tests

- [ ] **Performance: Regexes** (1 hour)
  - [ ] Identify all runtime regex compilations
  - [ ] Move to package initialization
  - [ ] Benchmark improvement

- [ ] **Performance: Allocations** (1-2 hours)
  - [ ] Profile hot paths
  - [ ] Replace maps with arrays where possible
  - [ ] Benchmark improvements

- [ ] **Add Context Support** (2-3 hours)
  - [ ] Add context.Context parameter to core functions
  - [ ] Implement cancellation checks
  - [ ] Add timeout tests

**Total Estimated Effort**: 15-20 hours (can be parallelized)

---

## Testing Strategy

For each improvement, include:

1. **Happy path test** - Normal operation works
2. **Error case test** - Error handling works
3. **Edge case test** - Boundary conditions handled
4. **Benchmark test** - Performance improvement measured

Example:
```go
func TestImprovement(t *testing.T) {
    // Happy path
    t.Run("success", func(t *testing.T) {
        result, err := operation()
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        // Assert result
    })
    
    // Error case
    t.Run("error", func(t *testing.T) {
        result, err := operationWithError()
        if err == nil {
            t.Fatalf("expected error, got nil")
        }
        // Assert error type/message
    })
    
    // Edge case
    t.Run("edge case", func(t *testing.T) {
        result, err := operationEdgeCase()
        // Assert behavior
    })
}

func BenchmarkImprovement(b *testing.B) {
    for i := 0; i < b.N; i++ {
        operation()
    }
}
```

---

## References

- Go Error Handling: https://golang.org/doc/effective_go#errors
- Structured Logging: https://pkg.go.dev/log/slog
- Context Package: https://pkg.go.dev/context
- Regex Performance: https://golang.org/pkg/regexp/#Regexp

