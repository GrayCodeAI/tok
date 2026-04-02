// Package benchmarking provides DSL for defining benchmarks
package benchmarking

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// DSLParser parses benchmark DSL
type DSLParser struct {
	benchmarks []BenchmarkDefinition
}

// BenchmarkDefinition represents a benchmark defined via DSL
type BenchmarkDefinition struct {
	Name        string
	Type        string
	Description string
	Iterations  int
	Duration    time.Duration
	Warmup      int
	Parameters  map[string]interface{}
	Setup       string
	Teardown    string
	Validators  []ValidatorDefinition
}

// ValidatorDefinition represents a validation rule
type ValidatorDefinition struct {
	Metric    string
	Operator  string
	Threshold float64
	Message   string
}

// NewDSLParser creates a new DSL parser
func NewDSLParser() *DSLParser {
	return &DSLParser{
		benchmarks: make([]BenchmarkDefinition, 0),
	}
}

// Parse parses DSL text into benchmark definitions
func (p *DSLParser) Parse(dsl string) ([]BenchmarkDefinition, error) {
	lines := strings.Split(dsl, "\n")
	var currentDef *BenchmarkDefinition

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse directive
		if strings.HasPrefix(line, "benchmark ") {
			// Save previous benchmark if exists
			if currentDef != nil {
				p.benchmarks = append(p.benchmarks, *currentDef)
			}

			// Start new benchmark
			name := strings.TrimSpace(strings.TrimPrefix(line, "benchmark"))
			currentDef = &BenchmarkDefinition{
				Name:       name,
				Parameters: make(map[string]interface{}),
				Iterations: 100,
				Duration:   30 * time.Second,
				Warmup:     1,
			}
			continue
		}

		if currentDef == nil {
			return nil, fmt.Errorf("line %d: benchmark directive must come first", lineNum+1)
		}

		// Parse properties
		if err := p.parseProperty(currentDef, line); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}
	}

	// Save last benchmark
	if currentDef != nil {
		p.benchmarks = append(p.benchmarks, *currentDef)
	}

	return p.benchmarks, nil
}

func (p *DSLParser) parseProperty(def *BenchmarkDefinition, line string) error {
	// Handle nested blocks
	if strings.HasPrefix(line, "validate {") {
		return nil // Start of validate block
	}

	if line == "}" {
		return nil // End of block
	}

	// Parse key-value pairs
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid property format: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "type":
		def.Type = value
	case "description":
		def.Description = strings.Trim(value, `"`)
	case "iterations":
		iterations, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid iterations: %w", err)
		}
		def.Iterations = iterations
	case "duration":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		def.Duration = duration
	case "warmup":
		warmup, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid warmup: %w", err)
		}
		def.Warmup = warmup
	case "setup":
		def.Setup = strings.Trim(value, `"`)
	case "teardown":
		def.Teardown = strings.Trim(value, `"`)
	default:
		// Store as parameter
		if intVal, err := strconv.Atoi(value); err == nil {
			def.Parameters[key] = intVal
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			def.Parameters[key] = floatVal
		} else if boolVal, err := strconv.ParseBool(value); err == nil {
			def.Parameters[key] = boolVal
		} else {
			def.Parameters[key] = strings.Trim(value, `"`)
		}
	}

	return nil
}

// DSLBuilder helps build benchmark DSL programmatically
type DSLBuilder struct {
	definitions []BenchmarkDefinition
}

// NewDSLBuilder creates a new DSL builder
func NewDSLBuilder() *DSLBuilder {
	return &DSLBuilder{
		definitions: make([]BenchmarkDefinition, 0),
	}
}

// AddBenchmark adds a benchmark definition
func (b *DSLBuilder) AddBenchmark(name string) *BenchmarkBuilder {
	def := BenchmarkDefinition{
		Name:       name,
		Parameters: make(map[string]interface{}),
		Iterations: 100,
		Duration:   30 * time.Second,
		Warmup:     1,
	}
	b.definitions = append(b.definitions, def)
	return &BenchmarkBuilder{def: &b.definitions[len(b.definitions)-1]}
}

// Build generates DSL text
func (b *DSLBuilder) Build() string {
	var output strings.Builder

	for _, def := range b.definitions {
		output.WriteString(fmt.Sprintf("benchmark %s\n", def.Name))

		if def.Type != "" {
			output.WriteString(fmt.Sprintf("  type = %s\n", def.Type))
		}

		if def.Description != "" {
			output.WriteString(fmt.Sprintf("  description = \"%s\"\n", def.Description))
		}

		if def.Iterations != 100 {
			output.WriteString(fmt.Sprintf("  iterations = %d\n", def.Iterations))
		}

		if def.Duration != 30*time.Second {
			output.WriteString(fmt.Sprintf("  duration = %s\n", def.Duration))
		}

		if def.Warmup != 1 {
			output.WriteString(fmt.Sprintf("  warmup = %d\n", def.Warmup))
		}

		for key, value := range def.Parameters {
			switch v := value.(type) {
			case string:
				output.WriteString(fmt.Sprintf("  %s = \"%s\"\n", key, v))
			default:
				output.WriteString(fmt.Sprintf("  %s = %v\n", key, v))
			}
		}

		output.WriteString("\n")
	}

	return output.String()
}

// BenchmarkBuilder builds a single benchmark definition
type BenchmarkBuilder struct {
	def *BenchmarkDefinition
}

// WithType sets the benchmark type
func (b *BenchmarkBuilder) WithType(benchmarkType string) *BenchmarkBuilder {
	b.def.Type = benchmarkType
	return b
}

// WithDescription sets the description
func (b *BenchmarkBuilder) WithDescription(desc string) *BenchmarkBuilder {
	b.def.Description = desc
	return b
}

// WithIterations sets the iterations
func (b *BenchmarkBuilder) WithIterations(iterations int) *BenchmarkBuilder {
	b.def.Iterations = iterations
	return b
}

// WithDuration sets the duration
func (b *BenchmarkBuilder) WithDuration(duration time.Duration) *BenchmarkBuilder {
	b.def.Duration = duration
	return b
}

// WithWarmup sets the warmup iterations
func (b *BenchmarkBuilder) WithWarmup(warmup int) *BenchmarkBuilder {
	b.def.Warmup = warmup
	return b
}

// WithParameter adds a parameter
func (b *BenchmarkBuilder) WithParameter(key string, value interface{}) *BenchmarkBuilder {
	b.def.Parameters[key] = value
	return b
}

// ToBenchmark converts definition to actual benchmark
func (def *BenchmarkDefinition) ToBenchmark() (Benchmark, error) {
	switch def.Type {
	case "compression":
		size := 1024
		if s, ok := def.Parameters["size"].(int); ok {
			size = s
		}
		return NewCompressionBenchmark(def.Name, size), nil

	case "pipeline":
		mode := filter.ModeMinimal
		input := "test data"
		if m, ok := def.Parameters["mode"].(string); ok {
			switch m {
			case "aggressive":
				mode = filter.ModeAggressive
			case "minimal":
				mode = filter.ModeMinimal
			}
		}
		if i, ok := def.Parameters["input"].(string); ok {
			input = i
		}
		return NewPipelineBenchmark(def.Name, mode, input), nil

	case "memory":
		size := 1024
		if s, ok := def.Parameters["size"].(int); ok {
			size = s
		}
		return NewMemoryBenchmark(def.Name, size), nil

	case "concurrency":
		workers := 10
		tasks := 100
		if w, ok := def.Parameters["workers"].(int); ok {
			workers = w
		}
		if t, ok := def.Parameters["tasks"].(int); ok {
			tasks = t
		}
		return NewConcurrencyBenchmark(def.Name, workers, tasks), nil

	default:
		return nil, fmt.Errorf("unknown benchmark type: %s", def.Type)
	}
}

// StandardDSLTemplates provides common benchmark templates
func StandardDSLTemplates() map[string]string {
	return map[string]string{
		"quick": `
# Quick benchmark suite
benchmark compression-small
  type = compression
  size = 1024
  iterations = 10

benchmark compression-medium
  type = compression
  size = 10240
  iterations = 10
`,
		"thorough": `
# Thorough benchmark suite
benchmark compression-small
  type = compression
  size = 1024
  iterations = 100
  warmup = 5

benchmark compression-medium
  type = compression
  size = 10240
  iterations = 100
  warmup = 5

benchmark compression-large
  type = compression
  size = 102400
  iterations = 50
  warmup = 3

benchmark memory-small
  type = memory
  size = 1024
  iterations = 1000

benchmark memory-large
  type = memory
  size = 102400
  iterations = 1000

benchmark concurrency-low
  type = concurrency
  workers = 10
  tasks = 100

benchmark concurrency-high
  type = concurrency
  workers = 100
  tasks = 1000
`,
		"pipeline": `
# Pipeline benchmark suite
benchmark pipeline-minimal
  type = pipeline
  mode = minimal
  iterations = 50

benchmark pipeline-aggressive
  type = pipeline
  mode = aggressive
  iterations = 50
`,
	}
}

// ParseDSLFile parses a DSL file
func ParseDSLFile(filename string) ([]BenchmarkDefinition, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	parser := NewDSLParser()
	return parser.Parse(string(data))
}
