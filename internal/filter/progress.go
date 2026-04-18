package filter

// ProgressCallback is called during pipeline processing to report real-time progress.
// Set by the commands/shared package to integrate with the status line.
// Parameters: layerName, inputTokens, outputTokens, progressPercent (0-100)
var ProgressCallback func(layerName string, inputTokens, outputTokens int, progressPercent float64)

// reportProgress invokes the registered progress callback if set.
func reportProgress(layer string, inTokens, outTokens int, progress float64) {
	if ProgressCallback != nil {
		ProgressCallback(layer, inTokens, outTokens, progress)
	}
}
