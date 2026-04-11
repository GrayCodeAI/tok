package filter

import (
	"testing"
)

func FuzzPipeline(f *testing.F) {
	f.Add([]byte("Hello, World!"))
	f.Add([]byte("test content with some repeated content"))
	f.Add([]byte(`{"key": "value", "array": [1, 2, 3]}`))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		input := string(data)

		modes := []Mode{ModeMinimal, ModeAggressive}

		for _, mode := range modes {
			p := NewPipelineCoordinator(PipelineConfig{
				Mode:            mode,
				SessionTracking: false,
				NgramEnabled:    true,
			})

			output, _ := p.Process(input)
			_ = output
		}
	})
}

func FuzzCompressSettings(f *testing.F) {
	f.Add([]byte("sample text"), 1000)
	f.Add([]byte(""), 0)
	f.Add([]byte("long text "), 5000)

	f.Fuzz(func(t *testing.T, data []byte, budget int) {
		input := string(data)

		if budget < 0 {
			budget = 0
		}
		if budget > 1000000 {
			budget = 1000000
		}

		p := NewPipelineCoordinator(PipelineConfig{
			Mode:            ModeMinimal,
			Budget:          budget,
			SessionTracking: false,
		})

		output, _ := p.Process(input)
		_ = output
	})
}

func FuzzModes(f *testing.F) {
	f.Add([]byte("test input"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		input := string(data)

		p := NewPipelineCoordinator(PipelineConfig{
			Mode:             ModeMinimal,
			EnableEntropy:    true,
			EnablePerplexity: true,
			EnableH2O:        true,
			EnableCompaction: true,
		})

		_, _ = p.Process(input)
	})
}
