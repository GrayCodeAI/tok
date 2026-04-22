package filter

import "sync"

var (
	progressCallback   func(layerName string, inputTokens, outputTokens int, progressPercent float64)
	progressCallbackMu sync.RWMutex
)

// SetProgressCallback registers the pipeline progress callback.
// Called by the commands/shared package to integrate with the status line.
func SetProgressCallback(cb func(layerName string, inputTokens, outputTokens int, progressPercent float64)) {
	progressCallbackMu.Lock()
	progressCallback = cb
	progressCallbackMu.Unlock()
}

// GetProgressCallback returns the currently registered progress callback.
func GetProgressCallback() func(layerName string, inputTokens, outputTokens int, progressPercent float64) {
	progressCallbackMu.RLock()
	defer progressCallbackMu.RUnlock()
	return progressCallback
}

// reportProgress invokes the registered progress callback if set.
func reportProgress(layer string, inTokens, outTokens int, progress float64) {
	if cb := GetProgressCallback(); cb != nil {
		cb(layer, inTokens, outTokens, progress)
	}
}
