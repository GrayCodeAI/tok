// Package tracking provides per-layer statistics tracking.
package tracking

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// LayerStat tracks statistics for a single layer.
type LayerStat struct {
	Name         string        `json:"name"`
	Invocations  int64         `json:"invocations"`
	TokensIn     int64         `json:"tokens_in"`
	TokensOut    int64         `json:"tokens_out"`
	TokensSaved  int64         `json:"tokens_saved"`
	TotalTime    time.Duration `json:"total_time"`
	AvgTime      time.Duration `json:"avg_time"`
	LastUsed     time.Time     `json:"last_used"`
	Errors       int64         `json:"errors"`
}

// SavingsPercent returns the percentage of tokens saved.
func (ls *LayerStat) SavingsPercent() float64 {
	if ls.TokensIn == 0 {
		return 0
	}
	return float64(ls.TokensSaved) / float64(ls.TokensIn) * 100
}

// CompressionRatio returns the compression ratio.
func (ls *LayerStat) CompressionRatio() float64 {
	if ls.TokensIn == 0 {
		return 1.0
	}
	return float64(ls.TokensOut) / float64(ls.TokensIn)
}

// LayerTracker tracks statistics for all layers.
type LayerTracker struct {
	mu     sync.RWMutex
	layers map[string]*LayerStat
	onUpdate func(string, *LayerStat)
}

// NewLayerTracker creates a new layer tracker.
func NewLayerTracker() *LayerTracker {
	return &LayerTracker{
		layers: make(map[string]*LayerStat),
	}
}

// SetUpdateCallback sets a callback for updates.
func (lt *LayerTracker) SetUpdateCallback(cb func(string, *LayerStat)) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.onUpdate = cb
}

// Record records a layer invocation.
func (lt *LayerTracker) Record(layerName string, tokensIn, tokensOut int, duration time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	stat, ok := lt.layers[layerName]
	if !ok {
		stat = &LayerStat{Name: layerName}
		lt.layers[layerName] = stat
	}

	stat.Invocations++
	stat.TokensIn += int64(tokensIn)
	stat.TokensOut += int64(tokensOut)
	stat.TokensSaved += int64(tokensIn - tokensOut)
	stat.TotalTime += duration
	stat.LastUsed = time.Now()

	if stat.Invocations > 0 {
		stat.AvgTime = stat.TotalTime / time.Duration(stat.Invocations)
	}

	if lt.onUpdate != nil {
		lt.onUpdate(layerName, stat)
	}
}

// RecordError records a layer error.
func (lt *LayerTracker) RecordError(layerName string) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	stat, ok := lt.layers[layerName]
	if !ok {
		stat = &LayerStat{Name: layerName}
		lt.layers[layerName] = stat
	}
	stat.Errors++
}

// Get returns statistics for a layer.
func (lt *LayerTracker) Get(layerName string) (*LayerStat, bool) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	stat, ok := lt.layers[layerName]
	if !ok {
		return nil, false
	}
	// Return a copy
	statCopy := *stat
	return &statCopy, true
}

// GetAll returns all layer statistics.
func (lt *LayerTracker) GetAll() map[string]*LayerStat {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	result := make(map[string]*LayerStat, len(lt.layers))
	for name, stat := range lt.layers {
		statCopy := *stat
		result[name] = &statCopy
	}
	return result
}

// GetTopSavers returns the top N layers by tokens saved.
func (lt *LayerTracker) GetTopSavers(n int) []*LayerStat {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	all := make([]*LayerStat, 0, len(lt.layers))
	for _, stat := range lt.layers {
		statCopy := *stat
		all = append(all, &statCopy)
	}

	// Sort by tokens saved (descending)
	sort.Slice(all, func(i, j int) bool {
		return all[i].TokensSaved > all[j].TokensSaved
	})

	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

// GetMostUsed returns the top N most used layers.
func (lt *LayerTracker) GetMostUsed(n int) []*LayerStat {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	all := make([]*LayerStat, 0, len(lt.layers))
	for _, stat := range lt.layers {
		statCopy := *stat
		all = append(all, &statCopy)
	}

	// Sort by invocations (descending)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Invocations > all[j].Invocations
	})

	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

// Reset resets all statistics.
func (lt *LayerTracker) Reset() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.layers = make(map[string]*LayerStat)
}

// TotalStats returns aggregated statistics.
func (lt *LayerTracker) TotalStats() LayerStat {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	total := LayerStat{Name: "total"}
	for _, stat := range lt.layers {
		total.Invocations += stat.Invocations
		total.TokensIn += stat.TokensIn
		total.TokensOut += stat.TokensOut
		total.TokensSaved += stat.TokensSaved
		total.TotalTime += stat.TotalTime
		total.Errors += stat.Errors
	}

	if total.Invocations > 0 {
		total.AvgTime = total.TotalTime / time.Duration(total.Invocations)
	}

	return total
}

// Report generates a statistics report.
func (lt *LayerTracker) Report() *LayerReport {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	report := &LayerReport{
		GeneratedAt: time.Now(),
		Layers:      make(map[string]LayerStat),
	}

	for name, stat := range lt.layers {
		report.Layers[name] = *stat
	}

	report.Total = lt.TotalStats()
	report.TopSavers = lt.GetTopSavers(5)
	report.MostUsed = lt.GetMostUsed(5)

	return report
}

// LayerReport contains a full statistics report.
type LayerReport struct {
	GeneratedAt time.Time            `json:"generated_at"`
	Layers      map[string]LayerStat `json:"layers"`
	Total       LayerStat            `json:"total"`
	TopSavers   []*LayerStat         `json:"top_savers"`
	MostUsed    []*LayerStat         `json:"most_used"`
}

// String returns a formatted report.
func (r *LayerReport) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Layer Statistics Report (generated %s)\n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintln(&b, strings.Repeat("=", 60))

	fmt.Fprintf(&b, "\nTotal: %d invocations, %d tokens saved (%.1f%%)\n",
		r.Total.Invocations, r.Total.TokensSaved, r.Total.SavingsPercent())

	if len(r.TopSavers) > 0 {
		fmt.Fprintln(&b, "\nTop Savers:")
		for i, ls := range r.TopSavers {
			fmt.Fprintf(&b, "  %d. %s: %d tokens (%.1f%%)\n",
				i+1, ls.Name, ls.TokensSaved, ls.SavingsPercent())
		}
	}

	if len(r.MostUsed) > 0 {
		fmt.Fprintln(&b, "\nMost Used:")
		for i, ls := range r.MostUsed {
			fmt.Fprintf(&b, "  %d. %s: %d invocations\n",
				i+1, ls.Name, ls.Invocations)
		}
	}

	return b.String()
}

// Global layer tracker instance.
var globalLayerTracker = NewLayerTracker()

// GetGlobalLayerTracker returns the global tracker.
func GetGlobalLayerTracker() *LayerTracker {
	return globalLayerTracker
}

// InstrumentedProcessor wraps a processor with tracking.
type InstrumentedProcessor struct {
	name      string
	processor func(string) (string, int)
	tracker   *LayerTracker
}

// NewInstrumentedProcessor creates an instrumented processor.
func NewInstrumentedProcessor(name string, processor func(string) (string, int), tracker *LayerTracker) *InstrumentedProcessor {
	if tracker == nil {
		tracker = globalLayerTracker
	}
	return &InstrumentedProcessor{
		name:      name,
		processor: processor,
		tracker:   tracker,
	}
}

// Process processes content with tracking.
func (ip *InstrumentedProcessor) Process(input string) (string, int) {
	start := time.Now()
	tokensIn := len(input) / 4 // Approximation

	output, saved := ip.processor(input)

	tokensOut := len(output) / 4
	duration := time.Since(start)

	ip.tracker.Record(ip.name, tokensIn, tokensOut, duration)

	return output, saved
}

