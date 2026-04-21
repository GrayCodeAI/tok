package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

func TestPipelineSectionRendersBarsAndTable(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.TopLayers = []tracking.DashboardLayerSummary{
		{LayerName: "entropy_prune", CallCount: 100, TotalSaved: 5000, AvgSaved: 50},
		{LayerName: "dedup", CallCount: 80, TotalSaved: 3000, AvgSaved: 37.5},
		{LayerName: "ansi_strip", CallCount: 120, TotalSaved: 1500, AvgSaved: 12.5},
	}

	s := newPipelineSection()
	s.Update(ctx, tea.KeyMsg{})
	view := s.View(ctx)
	for _, want := range []string{"Pipeline", "Top contributors", "entropy_prune", "dedup", "ansi_strip", "Total saved"} {
		if !strings.Contains(view, want) {
			t.Fatalf("pipeline view missing %q:\n%s", want, view)
		}
	}
}

func TestPipelineTotalSaved(t *testing.T) {
	layers := []tracking.DashboardLayerSummary{{TotalSaved: 10}, {TotalSaved: 20}, {TotalSaved: 7}}
	if got := totalLayerSaved(layers); got != 37 {
		t.Fatalf("totalLayerSaved = %d, want 37", got)
	}
}
