package compression

import (
	"context"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

func TestServiceCompress(t *testing.T) {
	cfg := filter.PipelineConfig{
		Mode:            filter.ModeMinimal,
		Budget:          4000,
		SessionTracking: true,
		NgramEnabled:    true,
	}
	
	svc := NewService(cfg)
	ctx := context.Background()
	
	tests := []struct {
		name    string
		input   string
		mode    filter.Mode
		budget  int
	}{
		{"simple input", "hello world", filter.ModeMinimal, 4000},
		{"empty input", "", filter.ModeMinimal, 4000},
		{"large input", generateLargeInput(5000), filter.ModeMinimal, 4000},
		{"aggressive mode", "test data with some content", filter.ModeAggressive, 2000},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.Compress(ctx, &CompressRequest{
				Input:  tt.input,
				Mode:   tt.mode,
				Budget: tt.budget,
			})
			
			if err != nil {
				t.Errorf("Compress returned error: %v", err)
				return
			}
			
			if resp == nil {
				t.Error("Compress returned nil response")
				return
			}
			
			// For non-empty input, we should have some output
			if tt.input != "" && resp.Output == "" {
				t.Error("Compress returned empty output for non-empty input")
			}
			
			// Original size should be >= compressed size
			if resp.OriginalSize < resp.CompressedSize {
				t.Logf("Warning: compressed size (%d) > original size (%d)", resp.CompressedSize, resp.OriginalSize)
			}
		})
	}
}

func TestServiceGetLayers(t *testing.T) {
	svc := NewService(filter.PipelineConfig{})
	ctx := context.Background()
	
	layers, err := svc.GetLayers(ctx)
	if err != nil {
		t.Fatalf("GetLayers returned error: %v", err)
	}
	
	if len(layers) == 0 {
		t.Error("GetLayers returned no layers")
	}
	
	// Verify layer structure
	for i, layer := range layers {
		if layer.Name == "" {
			t.Errorf("Layer %d has empty name", i)
		}
		if layer.Number == 0 {
			t.Errorf("Layer %d has zero number", i)
		}
		if layer.Research == "" {
			t.Errorf("Layer %d has empty research", i)
		}
	}
}

func TestServiceGetStats(t *testing.T) {
	svc := NewService(filter.PipelineConfig{})
	ctx := context.Background()
	
	stats, err := svc.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}
	
	// Stats should return a valid (even if empty) response
	if stats == nil {
		t.Error("GetStats returned nil")
	}
}

func generateLargeInput(size int) string {
	result := ""
	for len(result) < size {
		result += "This is a test line with some content for compression. "
	}
	return result[:size]
}
