package filter

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/GrayCodeAI/tok/internal/simd"
)

// BenchmarkSIMDOperations benchmarks SIMD-optimized functions
func BenchmarkSIMDOperations(b *testing.B) {
	// Test data
	smallData := "Hello World"
	mediumData := makeString(1000)
	largeData := makeString(10000)
	hugeData := makeString(100000)

	b.Run("FastHasANSI_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastHasANSI(smallData)
		}
	})

	b.Run("FastHasANSI_Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastHasANSI(mediumData)
		}
	})

	b.Run("FastHasANSI_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastHasANSI(largeData)
		}
	})

	b.Run("FastHasANSI_Huge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastHasANSI(hugeData)
		}
	})

	b.Run("FastCountBytes_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastCountBytes(smallData, 'e')
		}
	})

	b.Run("FastCountBytes_Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastCountBytes(mediumData, 'e')
		}
	})

	b.Run("FastCountBytes_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastCountBytes(largeData, 'e')
		}
	})

	b.Run("FastLower_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastLower(smallData)
		}
	})

	b.Run("FastLower_Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastLower(mediumData)
		}
	})

	b.Run("FastLower_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastLower(largeData)
		}
	})

	b.Run("FastEqual_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastEqual(smallData, smallData)
		}
	})

	b.Run("FastEqual_Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastEqual(mediumData, mediumData)
		}
	})

	b.Run("FastEqual_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.FastEqual(largeData, largeData)
		}
	})

	b.Run("SplitWords_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.SplitWords(smallData)
		}
	})

	b.Run("SplitWords_Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.SplitWords(mediumData)
		}
	})

	b.Run("SplitWords_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simd.SplitWords(largeData)
		}
	})
}

// BenchmarkBytePool benchmarks the byte pool operations
func BenchmarkBytePool(b *testing.B) {
	pool := NewBytePool()

	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("GetPut_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := pool.Get(size)
				pool.Put(buf)
			}
		})
	}
}

// BenchmarkFastStringBuilder benchmarks the fast string builder
func BenchmarkFastStringBuilder(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("FastStringBuilder_%d", size), func(b *testing.B) {
			data := makeString(size)
			for i := 0; i < b.N; i++ {
				builder := NewFastStringBuilder(size)
				builder.WriteString(data)
				_ = builder.String()
			}
		})

		b.Run(fmt.Sprintf("StdStringBuilder_%d", size), func(b *testing.B) {
			data := makeString(size)
			for i := 0; i < b.N; i++ {
				var builder string
				builder += data
				_ = builder
			}
		})
	}
}

// BenchmarkParallelProcessor benchmarks parallel processing
func BenchmarkParallelProcessor(b *testing.B) {
	processor := NewParallelProcessor()

	// Simple processing function
	processFn := func(input string) (string, int) {
		// Simple transformation: uppercase
		output := simd.FastLower(input)
		return output, len(input) - len(output)
	}

	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		items := make([]string, size)
		for i := range items {
			items[i] = makeString(100)
		}

		b.Run(fmt.Sprintf("Parallel_%d_items", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				processor.ProcessItems(items, processFn)
			}
		})

		b.Run(fmt.Sprintf("Sequential_%d_items", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, item := range items {
					processFn(item)
				}
			}
		})
	}
}

// BenchmarkPipeline benchmarks the full pipeline
func BenchmarkPipeline(b *testing.B) {
	configs := []struct {
		name string
		cfg  PipelineConfig
	}{
		{
			name: "Fast",
			cfg:  TierConfig(TierSurface, ModeMinimal),
		},
		{
			name: "Balanced",
			cfg:  TierConfig(TierTrim, ModeMinimal),
		},
		{
			name: "Full",
			cfg:  TierConfig(TierExtract, ModeMinimal),
		},
	}

	sizes := []int{100, 1000, 10000}

	for _, config := range configs {
		for _, size := range sizes {
			input := makeString(size)
			pipeline := NewPipelineCoordinator(config.cfg)

			b.Run(fmt.Sprintf("%s_%d", config.name, size), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					pipeline.Process(input)
				}
			})
		}
	}
}

// BenchmarkCompressionQuality benchmarks compression quality vs speed
func BenchmarkCompressionQuality(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{
			name:  "Code",
			input: makeCodeString(1000),
		},
		{
			name:  "Logs",
			input: makeLogString(1000),
		},
		{
			name:  "JSON",
			input: makeJSONString(1000),
		},
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("%s_Fast", input.name), func(b *testing.B) {
			pipeline := NewPipelineCoordinator(TierConfig(TierSurface, ModeMinimal))
			for i := 0; i < b.N; i++ {
				pipeline.Process(input.input)
			}
		})

		b.Run(fmt.Sprintf("%s_Balanced", input.name), func(b *testing.B) {
			pipeline := NewPipelineCoordinator(TierConfig(TierTrim, ModeMinimal))
			for i := 0; i < b.N; i++ {
				pipeline.Process(input.input)
			}
		})

		b.Run(fmt.Sprintf("%s_Full", input.name), func(b *testing.B) {
			pipeline := NewPipelineCoordinator(TierConfig(TierExtract, ModeMinimal))
			for i := 0; i < b.N; i++ {
				pipeline.Process(input.input)
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("PipelineMemory", func(b *testing.B) {
		var m1, m2 runtime.MemStats

		runtime.GC()
		runtime.ReadMemStats(&m1)

		pipeline := NewPipelineCoordinator(TierConfig(TierTrim, ModeMinimal))
		input := makeString(10000)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline.Process(input)
		}
		b.StopTimer()

		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
		b.ReportMetric(float64(m2.HeapAlloc-m1.HeapAlloc)/float64(b.N), "heap_bytes/op")
	})
}

// BenchmarkLatency benchmarks end-to-end latency
func BenchmarkLatency(b *testing.B) {
	pipeline := NewPipelineCoordinator(TierConfig(TierTrim, ModeMinimal))

	inputs := []struct {
		name  string
		size  int
		input string
	}{
		{"Small", 100, makeString(100)},
		{"Medium", 1000, makeString(1000)},
		{"Large", 10000, makeString(10000)},
	}

	for _, tc := range inputs {
		b.Run(tc.name, func(b *testing.B) {
			latencies := make([]time.Duration, b.N)

			for i := 0; i < b.N; i++ {
				start := time.Now()
				pipeline.Process(tc.input)
				latencies[i] = time.Since(start)
			}

			// Calculate percentiles
			var total time.Duration
			for _, lat := range latencies {
				total += lat
			}
			avg := total / time.Duration(b.N)

			b.ReportMetric(float64(avg.Nanoseconds()), "ns/op")
			b.ReportMetric(float64(avg.Microseconds()), "us/op")
		})
	}
}

// Helper functions

func makeString(size int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	result := make([]byte, size)
	for i := range result {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

func makeCodeString(size int) string {
	code := `
func CalculateTotal(price float64, qty int) float64 {
	return price * float64(qty)
}

func main() {
	result := CalculateTotal(10.5, 3)
	fmt.Println(result)
}
`
	for len(code) < size {
		code += code
	}
	return code[:size]
}

func makeLogString(size int) string {
	log := "2024-01-15 10:30:45 INFO Application started successfully\n"
	log += "2024-01-15 10:30:46 DEBUG Loading configuration\n"
	log += "2024-01-15 10:30:47 INFO Server listening on port 8080\n"
	log += "2024-01-15 10:30:48 WARN High memory usage detected\n"
	log += "2024-01-15 10:30:49 ERROR Failed to connect to database\n"

	for len(log) < size {
		log += log
	}
	return log[:size]
}

func makeJSONString(size int) string {
	json := `{"users":[{"id":1,"name":"Alice","email":"alice@example.com"},{"id":2,"name":"Bob","email":"bob@example.com"}],"meta":{"total":2,"page":1}}`
	for len(json) < size {
		json += json
	}
	return json[:size]
}
