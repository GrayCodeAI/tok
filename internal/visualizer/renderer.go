package visualizer

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
)

type VisualElement struct {
	Type     string
	Text     string
	Width    int
	Height   int
	Color    string
	Position Point
}

type Point struct {
	X, Y int
}

type Canvas struct {
	width  int
	height int
	buffer [][]rune
	mu     sync.RWMutex
}

func NewCanvas(width, height int) *Canvas {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}
	return &Canvas{width: width, height: height, buffer: buffer}
}

func (c *Canvas) DrawChar(x, y int, ch rune) {
	if x >= 0 && x < c.width && y >= 0 && y < c.height {
		c.buffer[y][x] = ch
	}
}

func (c *Canvas) DrawText(x, y int, text string) {
	for i, ch := range text {
		c.DrawChar(x+i, y, ch)
	}
}

func (c *Canvas) DrawLine(x1, y1, x2, y2 int, ch rune) {
	dx := x2 - x1
	dy := y2 - y1
	steps := max(abs(dx), abs(dy))

	if steps == 0 {
		c.DrawChar(x1, y1, ch)
		return
	}

	xStep := float64(dx) / float64(steps)
	yStep := float64(dy) / float64(steps)

	x, y := float64(x1), float64(y1)
	for i := 0; i <= steps; i++ {
		c.DrawChar(int(math.Round(x)), int(math.Round(y)), ch)
		x += xStep
		y += yStep
	}
}

func (c *Canvas) DrawBox(x, y, w, h int, ch rune) {
	for i := 0; i < w; i++ {
		c.DrawChar(x+i, y, ch)
		c.DrawChar(x+i, y+h-1, ch)
	}
	for i := 0; i < h; i++ {
		c.DrawChar(x, y+i, ch)
		c.DrawChar(x+w-1, y+i, ch)
	}
}

func (c *Canvas) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sb strings.Builder
	for _, row := range c.buffer {
		sb.WriteString(string(row))
		sb.WriteString("\n")
	}
	return sb.String()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type VisualRenderer struct {
	mu       sync.RWMutex
	elements []VisualElement
	canvas   *Canvas
	config   RenderConfig
}

type RenderConfig struct {
	Width          int
	Height         int
	ColorEnabled   bool
	AsciiOnly      bool
	CompactMode    bool
	UnicodeEnabled bool
}

func NewVisualRenderer(config RenderConfig) *VisualRenderer {
	if config.Width == 0 {
		config.Width = 80
	}
	if config.Height == 0 {
		config.Height = 24
	}

	return &VisualRenderer{
		elements: make([]VisualElement, 0),
		canvas:   NewCanvas(config.Width, config.Height),
		config:   config,
	}
}

func (r *VisualRenderer) AddElement(ctx context.Context, elem VisualElement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.elements = append(r.elements, elem)
}

func (r *VisualRenderer) Render(ctx context.Context) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, elem := range r.elements {
		switch elem.Type {
		case "text":
			r.canvas.DrawText(elem.Position.X, elem.Position.Y, elem.Text)
		case "line":
			r.canvas.DrawLine(elem.Position.X, elem.Position.Y,
				elem.Position.X+elem.Width, elem.Position.Y+elem.Height, '─')
		case "box":
			r.canvas.DrawBox(elem.Position.X, elem.Position.Y, elem.Width, elem.Height, '█')
		}
	}

	return r.canvas.String()
}

func (r *VisualRenderer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.elements = r.elements[:0]
	r.canvas = NewCanvas(r.config.Width, r.config.Height)
}

func RenderTokenChart(tokens map[string]int, width, height int) string {
	if len(tokens) == 0 {
		return "No data to display\n"
	}

	maxVal := 0
	for _, v := range tokens {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		return "All values are zero\n"
	}

	canvas := NewCanvas(width, height)

	labelWidth := 10
	chartWidth := width - labelWidth - 2

	keys := make([]string, 0, len(tokens))
	for k := range tokens {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys) && i < height-2; i++ {
		label := keys[i]
		value := tokens[label]

		barWidth := int(float64(value) / float64(maxVal) * float64(chartWidth))
		if barWidth > chartWidth {
			barWidth = chartWidth
		}

		labelLine := label
		if len(labelLine) > labelWidth {
			labelLine = labelLine[:labelWidth-2] + ".."
		}

		canvas.DrawText(0, i+1, fmt.Sprintf("%-10s", labelLine))

		for j := 0; j < barWidth; j++ {
			canvas.DrawChar(labelWidth+1+j, i+1, '█')
		}
	}

	return canvas.String()
}

func CompactJSON(input string) string {
	var result strings.Builder
	inString := false
	indent := 0

	re := regexp.MustCompile(`[\t\n\r ]+`)

	for i, ch := range input {
		if ch == '"' && (i == 0 || input[i-1] != '\\') {
			inString = !inString
			result.WriteRune(ch)
			continue
		}

		if inString {
			result.WriteRune(ch)
			continue
		}

		switch ch {
		case '{', '[':
			result.WriteRune(ch)
			indent++
		case '}', ']':
			indent--
			result.WriteRune(ch)
		case ',':
			result.WriteRune(ch)
		case ':':
			result.WriteString(": ")
		default:
			if !unicodeIsSpace(ch) {
				result.WriteRune(ch)
			}
		}
	}

	output := re.ReplaceAllString(result.String(), " ")
	output = strings.TrimSpace(output)

	return output
}

func unicodeIsSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func ColorizeTokens(tokenType string, text string) string {
	colors := map[string]string{
		"keyword":  "\033[36m",
		"string":   "\033[32m",
		"number":   "\033[33m",
		"comment":  "\033[90m",
		"function": "\033[35m",
		"variable": "\033[34m",
		"reset":    "\033[0m",
	}

	if color, ok := colors[tokenType]; ok {
		return color + text + colors["reset"]
	}
	return text
}

type BarChart struct {
	mu       sync.RWMutex
	labels   []string
	values   []int
	maxWidth int
}

func NewBarChart(maxWidth int) *BarChart {
	return &BarChart{
		labels:   make([]string, 0),
		values:   make([]int, 0),
		maxWidth: maxWidth,
	}
}

func (b *BarChart) AddBar(label string, value int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.labels = append(b.labels, label)
	b.values = append(b.values, value)
}

func (b *BarChart) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.labels) == 0 {
		return "No data\n"
	}

	maxVal := 0
	for _, v := range b.values {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		return "All values are zero\n"
	}

	var result strings.Builder

	for i, label := range b.labels {
		value := b.values[i]
		barLen := int(float64(value) / float64(maxVal) * float64(b.maxWidth))

		bar := strings.Repeat("█", barLen)
		result.WriteString(fmt.Sprintf("%-20s %s %d\n", label, bar, value))
	}

	return result.String()
}

type ProgressBar struct {
	mu          sync.RWMutex
	completed   int
	total       int
	width       int
	showPercent bool
}

func NewProgressBar(total, width int) *ProgressBar {
	return &ProgressBar{
		total:       total,
		width:       width,
		showPercent: true,
	}
}

func (p *ProgressBar) SetProgress(completed int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.completed = completed
}

func (p *ProgressBar) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.total == 0 {
		return "[....................] 0%"
	}

	filled := int(float64(p.completed) / float64(p.total) * float64(p.width))
	empty := p.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat(".", empty)
	percent := int(float64(p.completed) / float64(p.total) * 100)

	if p.showPercent {
		return fmt.Sprintf("[%s] %d%%", bar, percent)
	}
	return fmt.Sprintf("[%s]", bar)
}

func RenderHeatmap(data [][]int, labels []string) string {
	if len(data) == 0 || len(data[0]) == 0 {
		return "No data\n"
	}

	cols := len(data[0])

	var result strings.Builder

	result.WriteString("   ")
	for j := 0; j < cols && j < 20; j++ {
		result.WriteString(fmt.Sprintf("%2d ", j))
	}
	result.WriteString("\n")

	for i, row := range data {
		if i >= 30 {
			break
		}

		if i < len(labels) {
			label := labels[i]
			if len(label) > 4 {
				label = label[:4]
			}
			result.WriteString(fmt.Sprintf("%4s ", label))
		} else {
			result.WriteString(fmt.Sprintf("%4d ", i))
		}

		for j, val := range row {
			if j >= 20 {
				break
			}

			bg := getHeatmapColor(val)
			result.WriteString(bg + "  " + "\033[0m")
		}
		result.WriteString("\n")
	}

	return result.String()
}

func getHeatmapColor(value int) string {
	if value < 25 {
		return "\033[48;5;22m"
	} else if value < 50 {
		return "\033[48;5;28m"
	} else if value < 75 {
		return "\033[48;5;34m"
	} else {
		return "\033[48;5;40m"
	}
}

func TruncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	ellipsis := "..."
	available := maxLen - len(ellipsis)

	half := available / 2
	return s[:half] + ellipsis + s[len(s)-half:]
}
