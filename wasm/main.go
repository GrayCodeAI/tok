//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
	
	"github.com/GrayCodeAI/tokman/internal/filter"
)

func main() {
	c := make(chan struct{}, 0)
	
	// Create global TokMan object
	tokman := map[string]interface{}{
		"process":    processFunc(),
		"analyze":    analyzeFunc(),
		"stream":     streamFunc(),
		"version":    "1.2.0",
		"layerCount": 14,
	}
	
	js.Global().Set("TokMan", js.ValueOf(tokman))
	
	<-c
}

// processFunc returns the main compression function
func processFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return errorResult("input required")
		}
		
		input := args[0].String()
		mode := filter.ModeMinimal
		budget := 4000
		
		if len(args) > 1 {
			opts := args[1]
			if opts.Type() == js.TypeObject {
				if m := opts.Get("mode"); m.Type() == js.TypeString {
					if m.String() == "aggressive" {
						mode = filter.ModeAggressive
					}
				}
				if b := opts.Get("budget"); b.Type() == js.TypeNumber {
					budget = b.Int()
				}
			}
		}
		
		config := filter.PipelineConfig{
			Mode:   mode,
			Budget: budget,
		}
		
		coordinator := filter.NewPipelineCoordinator(config)
		output, stats := coordinator.Process(input)
		
		result := map[string]interface{}{
			"output":           output,
			"originalTokens":   stats.OriginalTokens,
			"finalTokens":      stats.FinalTokens,
			"tokensSaved":      stats.TotalSaved,
			"reductionPercent": stats.ReductionPercent,
			"layersApplied":    stats.LayersApplied,
		}
		
		return js.ValueOf(result)
	})
}

// analyzeFunc returns content type analysis
func analyzeFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return errorResult("input required")
		}
		
		input := args[0].String()
		selector := filter.NewAdaptiveLayerSelector()
		ct := selector.AnalyzeContent(input)
		
		result := map[string]interface{}{
			"contentType": ct.String(),
		}
		
		return js.ValueOf(result)
	})
}

// streamFunc returns streaming compression for chunked input
func streamFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return errorResult("callback required")
		}
		
		callback := args[0]
		mode := filter.ModeMinimal
		budget := 4000
		
		if len(args) > 1 {
			opts := args[1]
			if opts.Type() == js.TypeObject {
				if m := opts.Get("mode"); m.Type() == js.TypeString {
					if m.String() == "aggressive" {
						mode = filter.ModeAggressive
					}
				}
				if b := opts.Get("budget"); b.Type() == js.TypeNumber {
					budget = b.Int()
				}
			}
		}
		
		config := filter.PipelineConfig{
			Mode:   mode,
			Budget: budget,
		}
		
		inputChan, outputChan := filter.StreamChannel(config)
		
		// Handle output in goroutine
		go func() {
			for chunk := range outputChan {
				result := map[string]interface{}{
					"content":      chunk.Content,
					"isCompressed": chunk.IsCompressed,
					"tokensSaved":  chunk.TokensSaved,
				}
				callback.Invoke(js.ValueOf(result))
			}
		}()
		
		// Return send function
		sendFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return nil
			}
			inputChan <- args[0].String()
			return nil
		})
		
		closeFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			close(inputChan)
			return nil
		})
		
		return js.ValueOf(map[string]interface{}{
			"send":  sendFunc,
			"close": closeFunc,
		})
	})
}

func errorResult(msg string) js.Value {
	return js.ValueOf(map[string]interface{}{
		"error": msg,
	})
}
