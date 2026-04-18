# TokMan Complete Code Analysis - Part 5: Configuration & TOML System

## 5. Configuration System

### Configuration Hierarchy

```
Priority (highest to lowest):
1. CLI Flags (--budget 2000)
2. Environment Variables (TOKMAN_BUDGET=2000)
3. Config File (~/.config/tokman/config.toml)
4. Default Values (hardcoded)
```

### Config File Structure

```toml
# ~/.config/tokman/config.toml

[tracking]
enabled = true
database_path = "~/.local/share/tokman/tokman.db"
retention_days = 90

[filter]
mode = "minimal"  # or "aggressive"
budget = 2000
query_intent = ""

[pipeline]
max_context_tokens = 2000000
preset = "balanced"  # fast, balanced, full

# Core layers (1-9)
enable_entropy = true
enable_perplexity = true
enable_goal_driven = true
enable_ast = true
enable_contrastive = true
enable_ngram = true
enable_evaluator = true
enable_gist = true
enable_hierarchical = true

# Advanced layers (10-20)
enable_compaction = true
compaction_threshold = 500
compaction_preserve_turns = 10
compaction_max_tokens = 5000

enable_attribution = true
attribution_threshold = 0.3

enable_h2o = true
h2o_sink_size = 4
h2o_recent_size = 512
h2o_heavy_hitter_size = 256

enable_attention_sink = true
attention_sink_count = 4
attention_recent_count = 512

enable_meta_token = true
meta_token_window = 8
meta_token_min_size = 3

enable_semantic_chunk = true
semantic_chunk_method = "sliding"
semantic_chunk_min_size = 128
semantic_chunk_threshold = 0.7

enable_sketch_store = true
sketch_budget_ratio = 0.1
sketch_max_size = 1000
sketch_heavy_hitter = 0.8

enable_lazy_pruner = true
lazy_base_budget = 1000
lazy_decay_rate = 0.95
lazy_revival_budget = 100

enable_semantic_anchor = true
semantic_anchor_ratio = 0.2
semantic_anchor_spacing = 50

enable_agent_memory = true
agent_knowledge_retention = 0.8
agent_history_prune = 0.5
agent_consolidation_max = 10

[hooks]
excluded_commands = ["vim", "nano", "emacs"]

[dashboard]
port = 8080
enabled = true

[remote]
enabled = false
compression_addr = "localhost:50051"
analytics_addr = "localhost:50053"
timeout = 30

[cache]
enabled = true
max_size = 10000
ttl = 3600  # seconds

[quality]
guardrail_enabled = false
min_quality_score = 0.7
```

### Environment Variables

```bash
# Core settings
export TOKMAN_BUDGET=2000
export TOKMAN_MODE=aggressive
export TOKMAN_PRESET=balanced
export TOKMAN_QUERY="debug"

# Layer toggles
export TOKMAN_LLM=1
export TOKMAN_COMPACTION=1
export TOKMAN_H2O=1
export TOKMAN_ATTENTION_SINK=1

# Behavior
export TOKMAN_VERBOSE=1
export TOKMAN_DRY_RUN=0
export TOKMAN_SILENT=0
export TOKMAN_JSON=0

# AI Agent attribution
export TOKMAN_AGENT="Claude Code"
export TOKMAN_MODEL="claude-3-opus"
export TOKMAN_PROVIDER="Anthropic"
```

### Configuration Loading (`internal/config/config.go`)

```go
type Config struct {
    Tracking  TrackingConfig
    Filter    FilterConfig
    Pipeline  PipelineConfig
    Hooks     HooksConfig
    Dashboard DashboardConfig
    Remote    RemoteConfig
    Cache     CacheConfig
    Quality   QualityConfig
}

func Load(cfgFile string) (*Config, error) {
    // 1. Set defaults
    viper.SetDefault("tracking.enabled", true)
    viper.SetDefault("filter.mode", "minimal")
    viper.SetDefault("pipeline.preset", "balanced")
    // ... 100+ defaults
    
    // 2. Load config file
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigName("config")
        viper.SetConfigType("toml")
        viper.AddConfigPath(ConfigPath())
    }
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }
    
    // 3. Environment variables
    viper.SetEnvPrefix("TOKMAN")
    viper.AutomaticEnv()
    
    // 4. Unmarshal
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

### Issues & Improvements

#### Issue 1: 100+ Default Values

**Problem**: Hardcoded defaults scattered everywhere
```go
// Current: Defaults in multiple places
viper.SetDefault("tracking.enabled", true)
viper.SetDefault("filter.mode", "minimal")
viper.SetDefault("pipeline.preset", "balanced")
// ... 97 more
```

**Fix**: Centralized defaults
```go
// defaults.go
func DefaultConfig() *Config {
    return &Config{
        Tracking: TrackingConfig{
            Enabled:       true,
            DatabasePath:  "~/.local/share/tokman/tokman.db",
            RetentionDays: 90,
        },
        Filter: FilterConfig{
            Mode:        ModeMinimal,
            Budget:      0,
            QueryIntent: "",
        },
        Pipeline: PipelineConfig{
            MaxContextTokens: 2000000,
            Preset:           "balanced",
            Core: CoreLayersConfig{
                Entropy:      LayerConfig{Enabled: true},
                Perplexity:   LayerConfig{Enabled: true},
                GoalDriven:   LayerConfig{Enabled: true},
                AST:          LayerConfig{Enabled: true},
                Contrastive:  LayerConfig{Enabled: true},
                Ngram:        LayerConfig{Enabled: true},
                Evaluator:    LayerConfig{Enabled: true},
                Gist:         LayerConfig{Enabled: true},
                Hierarchical: LayerConfig{Enabled: true},
            },
            Advanced: AdvancedLayersConfig{
                Compaction: CompactionConfig{
                    Enabled:       true,
                    Threshold:     500,
                    PreserveTurns: 10,
                    MaxTokens:     5000,
                },
                H2O: H2OConfig{
                    Enabled:         true,
                    SinkSize:        4,
                    RecentSize:      512,
                    HeavyHitterSize: 256,
                },
                // ... more layers
            },
        },
        // ... more sections
    }
}

func Load(cfgFile string) (*Config, error) {
    // Start with defaults
    cfg := DefaultConfig()
    
    // Override with file
    if err := loadFile(cfgFile, cfg); err != nil {
        return nil, err
    }
    
    // Override with env vars
    if err := loadEnv(cfg); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

#### Issue 2: No Validation

**Problem**: Invalid configs silently fail
```go
// Current: No validation
cfg, err := Load(cfgFile)
// cfg.Pipeline.H2OSinkSize could be negative!
```

**Fix**: Config validation
```go
func (c *Config) Validate() error {
    var errs []error
    
    // Validate tracking
    if c.Tracking.RetentionDays < 1 {
        errs = append(errs, fmt.Errorf("tracking.retention_days must be >= 1"))
    }
    
    // Validate filter
    if c.Filter.Budget < 0 {
        errs = append(errs, fmt.Errorf("filter.budget must be >= 0"))
    }
    
    // Validate pipeline
    if c.Pipeline.Advanced.H2O.SinkSize < 0 {
        errs = append(errs, fmt.Errorf("h2o.sink_size must be >= 0"))
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("config validation failed: %v", errs)
    }
    
    return nil
}

func Load(cfgFile string) (*Config, error) {
    cfg := DefaultConfig()
    
    if err := loadFile(cfgFile, cfg); err != nil {
        return nil, err
    }
    
    if err := loadEnv(cfg); err != nil {
        return nil, err
    }
    
    // Validate before returning
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

#### Issue 3: No Config Hot Reload

**Problem**: Must restart to apply config changes
```go
// Current: Config loaded once at startup
cfg, _ := config.Load(cfgFile)
```

**Fix**: Watch for config changes
```go
type ConfigWatcher struct {
    cfg      *Config
    mu       sync.RWMutex
    watchers []func(*Config)
}

func NewConfigWatcher(cfgFile string) (*ConfigWatcher, error) {
    cfg, err := Load(cfgFile)
    if err != nil {
        return nil, err
    }
    
    w := &ConfigWatcher{cfg: cfg}
    
    // Watch for file changes
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        newCfg, err := Load(cfgFile)
        if err != nil {
            slog.Error("config reload failed", "error", err)
            return
        }
        
        w.mu.Lock()
        w.cfg = newCfg
        w.mu.Unlock()
        
        // Notify watchers
        for _, watcher := range w.watchers {
            watcher(newCfg)
        }
        
        slog.Info("config reloaded")
    })
    
    return w, nil
}

func (w *ConfigWatcher) Get() *Config {
    w.mu.RLock()
    defer w.mu.RUnlock()
    return w.cfg
}

func (w *ConfigWatcher) OnChange(fn func(*Config)) {
    w.watchers = append(w.watchers, fn)
}
```

---

## TOML Filter System

### Custom Filter Definition

```toml
# ~/.config/tokman/filters/my_tool.toml

[my_command]
match = "^my-tool (build|test)"
description = "Custom filter for my-tool"

# Output patterns to preserve
output_patterns = [
    "^Building...",
    "^Testing...",
    "^✓ Success",
    "^✗ Failed"
]

# Lines to strip
strip_lines_matching = [
    "^INFO:",
    "^DEBUG:",
    "^Downloading"
]

# Lines to always preserve
preserve_patterns = [
    "^ERROR:",
    "^WARN:",
    "^FATAL:"
]

# Token budget
budget = 2000

# Compression mode
mode = "aggressive"
```

### TOML Loader (`internal/toml/loader.go`)

```go
type Loader struct {
    filters map[string]*FilterConfig
    mu      sync.RWMutex
}

func NewLoader() *Loader {
    return &Loader{
        filters: make(map[string]*FilterConfig),
    }
}

func (l *Loader) LoadFilters(dir string) error {
    // Find all .toml files
    files, err := filepath.Glob(filepath.Join(dir, "*.toml"))
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if err := l.loadFile(file); err != nil {
            slog.Warn("failed to load filter", "file", file, "error", err)
            continue
        }
    }
    
    return nil
}

func (l *Loader) loadFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    var filters map[string]*FilterConfig
    if err := toml.Unmarshal(data, &filters); err != nil {
        return err
    }
    
    l.mu.Lock()
    defer l.mu.Unlock()
    
    for name, cfg := range filters {
        l.filters[name] = cfg
    }
    
    return nil
}

func (l *Loader) Match(command string) *FilterConfig {
    l.mu.RLock()
    defer l.mu.RUnlock()
    
    for _, cfg := range l.filters {
        if cfg.Matches(command) {
            return cfg
        }
    }
    
    return nil
}
```

### Filter Application

```go
type TOMLFilter struct {
    config *FilterConfig
}

func (f *TOMLFilter) Apply(input string, mode Mode) (string, int) {
    lines := strings.Split(input, "\n")
    var filtered []string
    
    for _, line := range lines {
        // Check preserve patterns first
        if f.shouldPreserve(line) {
            filtered = append(filtered, line)
            continue
        }
        
        // Check strip patterns
        if f.shouldStrip(line) {
            continue
        }
        
        // Check output patterns
        if f.matchesOutputPattern(line) {
            filtered = append(filtered, line)
            continue
        }
        
        // Default: keep line
        filtered = append(filtered, line)
    }
    
    output := strings.Join(filtered, "\n")
    return output, estimateTokens(input) - estimateTokens(output)
}

func (f *TOMLFilter) shouldPreserve(line string) bool {
    for _, pattern := range f.config.PreservePatterns {
        if matched, _ := regexp.MatchString(pattern, line); matched {
            return true
        }
    }
    return false
}

func (f *TOMLFilter) shouldStrip(line string) bool {
    for _, pattern := range f.config.StripLinesMatching {
        if matched, _ := regexp.MatchString(pattern, line); matched {
            return true
        }
    }
    return false
}
```

### Issues & Improvements

#### Issue 1: Regex Compilation on Every Call

**Problem**: Compiling regex patterns repeatedly
```go
// Current: Compile on every match
func (f *TOMLFilter) shouldStrip(line string) bool {
    for _, pattern := range f.config.StripLinesMatching {
        if matched, _ := regexp.MatchString(pattern, line); matched {
            return true
        }
    }
    return false
}
```

**Fix**: Pre-compile regex patterns
```go
type TOMLFilter struct {
    config          *FilterConfig
    preserveRegex   []*regexp.Regexp
    stripRegex      []*regexp.Regexp
    outputRegex     []*regexp.Regexp
    compileOnce     sync.Once
}

func (f *TOMLFilter) compile() {
    f.compileOnce.Do(func() {
        f.preserveRegex = compilePatterns(f.config.PreservePatterns)
        f.stripRegex = compilePatterns(f.config.StripLinesMatching)
        f.outputRegex = compilePatterns(f.config.OutputPatterns)
    })
}

func compilePatterns(patterns []string) []*regexp.Regexp {
    regexes := make([]*regexp.Regexp, 0, len(patterns))
    for _, pattern := range patterns {
        if re, err := regexp.Compile(pattern); err == nil {
            regexes = append(regexes, re)
        }
    }
    return regexes
}

func (f *TOMLFilter) shouldStrip(line string) bool {
    f.compile()
    for _, re := range f.stripRegex {
        if re.MatchString(line) {
            return true
        }
    }
    return false
}
```

**Speedup**: 10-100x for repeated calls

#### Issue 2: No Filter Composition

**Problem**: Can't combine multiple filters
```go
// Current: Only one filter per command
filter := loader.Match(command)
```

**Fix**: Filter composition
```go
type CompositeFilter struct {
    filters []Filter
}

func (c *CompositeFilter) Apply(input string, mode Mode) (string, int) {
    output := input
    totalSaved := 0
    
    for _, filter := range c.filters {
        output, saved := filter.Apply(output, mode)
        totalSaved += saved
    }
    
    return output, totalSaved
}

// Usage
func (l *Loader) MatchAll(command string) Filter {
    var filters []Filter
    
    for _, cfg := range l.filters {
        if cfg.Matches(command) {
            filters = append(filters, NewTOMLFilter(cfg))
        }
    }
    
    if len(filters) == 0 {
        return nil
    }
    
    if len(filters) == 1 {
        return filters[0]
    }
    
    return &CompositeFilter{filters: filters}
}
```

### Built-in TOML Filters (97+)

TokMan includes 97+ pre-built filters for popular tools:

```
internal/toml/builtin/
├── vcs/
│   ├── git.toml
│   ├── gh.toml
│   └── gitlab.toml
├── container/
│   ├── docker.toml
│   ├── kubectl.toml
│   └── helm.toml
├── cloud/
│   ├── aws.toml
│   ├── gcloud.toml
│   └── az.toml
├── pkgmgr/
│   ├── npm.toml
│   ├── pip.toml
│   └── cargo.toml
└── test/
    ├── jest.toml
    ├── pytest.toml
    └── vitest.toml
```

### Configuration Best Practices

1. **Start with defaults**: Use `balanced` preset
2. **Enable caching**: Set `cache.enabled = true`
3. **Set reasonable budgets**: 2000-5000 tokens
4. **Use TOML filters**: For command-specific optimization
5. **Monitor performance**: Use `tokman benchmark`
6. **Validate configs**: Run `tokman config validate`
