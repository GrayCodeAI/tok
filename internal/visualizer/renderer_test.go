package visualizer

import (
	"context"
	"testing"
)

func TestNewCanvas(t *testing.T) {
	canvas := NewCanvas(80, 24)
	if canvas == nil {
		t.Error("Expected non-nil canvas")
	}
	if canvas.width != 80 {
		t.Errorf("Expected width 80, got %d", canvas.width)
	}
	if canvas.height != 24 {
		t.Errorf("Expected height 24, got %d", canvas.height)
	}
}

func TestCanvasDrawChar(t *testing.T) {
	canvas := NewCanvas(10, 10)
	canvas.DrawChar(5, 5, 'X')

	if canvas.buffer[5][5] != 'X' {
		t.Errorf("Expected 'X' at position 5,5, got %c", canvas.buffer[5][5])
	}
}

func TestCanvasDrawText(t *testing.T) {
	canvas := NewCanvas(20, 10)
	canvas.DrawText(0, 0, "Hello")

	for i, ch := range "Hello" {
		if canvas.buffer[0][i] != ch {
			t.Errorf("Expected '%c' at position %d, got '%c'", ch, i, canvas.buffer[0][i])
		}
	}
}

func TestCanvasDrawLine(t *testing.T) {
	canvas := NewCanvas(20, 20)
	canvas.DrawLine(0, 0, 10, 10, '*')

	output := canvas.String()
	if len(output) == 0 {
		t.Error("Expected non-empty canvas output")
	}
}

func TestCanvasDrawBox(t *testing.T) {
	canvas := NewCanvas(20, 20)
	canvas.DrawBox(5, 5, 5, 5, '█')

	if canvas.buffer[5][5] != '█' {
		t.Error("Expected box corner character")
	}
}

func TestCanvasString(t *testing.T) {
	canvas := NewCanvas(10, 2)
	canvas.DrawText(0, 0, "Test")

	output := canvas.String()
	if len(output) == 0 {
		t.Error("Expected non-empty string output")
	}
}

func TestNewVisualRenderer(t *testing.T) {
	config := RenderConfig{Width: 80, Height: 24}
	renderer := NewVisualRenderer(config)

	if renderer == nil {
		t.Error("Expected non-nil renderer")
	}
}

func TestVisualRendererAddElement(t *testing.T) {
	renderer := NewVisualRenderer(RenderConfig{})

	elem := VisualElement{
		Type:     "text",
		Text:     "Hello",
		Position: Point{X: 0, Y: 0},
	}

	renderer.AddElement(context.Background(), elem)

	if len(renderer.elements) != 1 {
		t.Errorf("Expected 1 element, got %d", len(renderer.elements))
	}
}

func TestVisualRendererRender(t *testing.T) {
	renderer := NewVisualRenderer(RenderConfig{Width: 20, Height: 5})

	renderer.AddElement(context.Background(), VisualElement{
		Type:     "text",
		Text:     "Test",
		Position: Point{X: 0, Y: 0},
	})

	output := renderer.Render(context.Background())
	if len(output) == 0 {
		t.Error("Expected non-empty render output")
	}
}

func TestVisualRendererClear(t *testing.T) {
	renderer := NewVisualRenderer(RenderConfig{})

	renderer.AddElement(context.Background(), VisualElement{
		Type: "text",
		Text: "Test",
	})

	renderer.Clear()

	if len(renderer.elements) != 0 {
		t.Errorf("Expected 0 elements after clear, got %d", len(renderer.elements))
	}
}

func TestRenderTokenChart(t *testing.T) {
	tokens := map[string]int{
		"git":    100,
		"cargo":  80,
		"npm":    60,
		"docker": 40,
	}

	output := RenderTokenChart(tokens, 40, 10)
	if len(output) == 0 {
		t.Error("Expected non-empty chart output")
	}
}

func TestRenderTokenChartEmpty(t *testing.T) {
	tokens := map[string]int{}
	output := RenderTokenChart(tokens, 40, 10)

	if output != "No data to display\n" {
		t.Errorf("Expected 'No data to display', got %s", output)
	}
}

func TestRenderTokenChartZeroValues(t *testing.T) {
	tokens := map[string]int{
		"test": 0,
	}
	output := RenderTokenChart(tokens, 40, 10)

	if output != "All values are zero\n" {
		t.Errorf("Expected 'All values are zero', got %s", output)
	}
}

func TestCompactJSON(t *testing.T) {
	input := `{
  "name": "test",
  "value": 123
}`

	output := CompactJSON(input)
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestColorizeTokens(t *testing.T) {
	result := ColorizeTokens("keyword", "test")
	if len(result) == 0 {
		t.Error("Expected non-empty colorized output")
	}
}

func TestNewBarChart(t *testing.T) {
	chart := NewBarChart(20)
	if chart == nil {
		t.Error("Expected non-nil bar chart")
	}
}

func TestBarChartAddBar(t *testing.T) {
	chart := NewBarChart(20)
	chart.AddBar("test", 100)

	if len(chart.labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(chart.labels))
	}
}

func TestBarChartString(t *testing.T) {
	chart := NewBarChart(20)
	chart.AddBar("test1", 100)
	chart.AddBar("test2", 50)

	output := chart.String()
	if len(output) == 0 {
		t.Error("Expected non-empty bar chart output")
	}
}

func TestBarChartEmpty(t *testing.T) {
	chart := NewBarChart(20)
	output := chart.String()

	if output != "No data\n" {
		t.Errorf("Expected 'No data', got %s", output)
	}
}

func TestNewProgressBar(t *testing.T) {
	bar := NewProgressBar(100, 20)
	if bar == nil {
		t.Error("Expected non-nil progress bar")
	}
}

func TestProgressBarSetProgress(t *testing.T) {
	bar := NewProgressBar(100, 20)
	bar.SetProgress(50)

	if bar.completed != 50 {
		t.Errorf("Expected 50, got %d", bar.completed)
	}
}

func TestProgressBarString(t *testing.T) {
	bar := NewProgressBar(100, 20)
	bar.SetProgress(50)

	output := bar.String()
	if len(output) == 0 {
		t.Error("Expected non-empty progress bar output")
	}
}

func TestRenderHeatmap(t *testing.T) {
	data := [][]int{
		{10, 20, 30},
		{40, 50, 60},
	}

	labels := []string{"a", "b"}
	output := RenderHeatmap(data, labels)

	if len(output) == 0 {
		t.Error("Expected non-empty heatmap output")
	}
}

func TestRenderHeatmapEmpty(t *testing.T) {
	data := [][]int{}
	output := RenderHeatmap(data, nil)

	if output != "No data\n" {
		t.Errorf("Expected 'No data', got %s", output)
	}
}

func TestTruncateMiddle(t *testing.T) {
	result := TruncateMiddle("1234567890", 8)
	t.Logf("Result: %s", result)
}

func TestTruncateMiddleShort(t *testing.T) {
	result := TruncateMiddle("123", 10)
	if result != "123" {
		t.Errorf("Expected '123', got %s", result)
	}
}

func TestAbs(t *testing.T) {
	if abs(-5) != 5 {
		t.Error("Expected 5")
	}
	if abs(5) != 5 {
		t.Error("Expected 5")
	}
}
