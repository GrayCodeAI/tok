package tuiviews

import (
	"fmt"
	"strings"
	"time"
)

type TUIView int

const (
	ViewOverview TUIView = iota
	ViewModels
	ViewDaily
	ViewStats
)

type Theme struct {
	Name      string
	Primary   string
	Secondary string
	Accent    string
	Bg        string
}

var Themes = []Theme{
	{Name: "tokman", Primary: "#e94560", Secondary: "#16c79a", Accent: "#f8b500", Bg: "#1a1a2e"},
	{Name: "dark", Primary: "#00ff88", Secondary: "#0088ff", Accent: "#ff0088", Bg: "#000000"},
	{Name: "light", Primary: "#333333", Secondary: "#666666", Accent: "#0066cc", Bg: "#ffffff"},
	{Name: "nord", Primary: "#88c0d0", Secondary: "#81a1c1", Accent: "#ebcb8b", Bg: "#2e3440"},
	{Name: "gruvbox", Primary: "#fabd2f", Secondary: "#8ec07c", Accent: "#fb4934", Bg: "#282828"},
	{Name: "solarized", Primary: "#268bd2", Secondary: "#2aa198", Accent: "#b58900", Bg: "#002b36"},
	{Name: "monokai", Primary: "#f92672", Secondary: "#a6e22e", Accent: "#fd971f", Bg: "#272822"},
	{Name: "dracula", Primary: "#ff79c6", Secondary: "#50fa7b", Accent: "#f1fa8c", Bg: "#282a36"},
	{Name: "catppuccin", Primary: "#cba6f7", Secondary: "#a6e3a1", Accent: "#f9e2af", Bg: "#1e1e2e"},
}

type TUIViewEngine struct {
	currentView  TUIView
	currentTheme int
	theme        Theme
}

func NewTUIViewEngine() *TUIViewEngine {
	return &TUIViewEngine{
		currentView:  ViewOverview,
		currentTheme: 0,
		theme:        Themes[0],
	}
}

func (e *TUIViewEngine) SwitchView(view TUIView) {
	e.currentView = view
}

func (e *TUIViewEngine) GetCurrentView() TUIView {
	return e.currentView
}

func (e *TUIViewEngine) SwitchTheme(index int) {
	if index >= 0 && index < len(Themes) {
		e.currentTheme = index
		e.theme = Themes[index]
	}
}

func (e *TUIViewEngine) GetTheme() Theme {
	return e.theme
}

func (e *TUIViewEngine) GetThemes() []Theme {
	return Themes
}

func (e *TUIViewEngine) RenderOverview(stats map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─ TokMan Dashboard [%s] ─────────────────────┐\n", e.theme.Name))
	sb.WriteString("│                                                    │\n")
	sb.WriteString(fmt.Sprintf("│  Commands: %-8d  Tokens: %-12d  │\n", stats["commands"], stats["tokens"]))
	sb.WriteString(fmt.Sprintf("│  Saved:    %-8d  Cost:   $%-10.2f  │\n", stats["saved"], stats["cost"]))
	sb.WriteString(fmt.Sprintf("│  Savings:  %.1f%%                              │\n", stats["savings_pct"]))
	sb.WriteString("│                                                    │\n")
	sb.WriteString("│  [1] Overview  [2] Models  [3] Daily  [4] Stats   │\n")
	sb.WriteString("└────────────────────────────────────────────────────┘\n")
	return sb.String()
}

func (e *TUIViewEngine) RenderModels(models []map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("┌─ Models ───────────────────────────────────┐\n")
	sb.WriteString("│ Model            Tokens    Cost    Savings │\n")
	sb.WriteString("├─────────────────────────────────────────────┤\n")
	for _, m := range models {
		sb.WriteString(fmt.Sprintf("│ %-16s %8d $%6.2f %6.1f%% │\n",
			m["name"], m["tokens"], m["cost"], m["savings_pct"]))
	}
	sb.WriteString("└─────────────────────────────────────────────┘\n")
	return sb.String()
}

func (e *TUIViewEngine) RenderDaily(daily []map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("┌─ Daily ────────────────────────────────────┐\n")
	sb.WriteString("│ Date        Commands  Tokens   Cost Savings│\n")
	sb.WriteString("├─────────────────────────────────────────────┤\n")
	for _, d := range daily {
		date := d["date"].(time.Time).Format("2006-01-02")
		sb.WriteString(fmt.Sprintf("│ %s %8d %8d $%5.2f %5.1f%% │\n",
			date, d["commands"], d["tokens"], d["cost"], d["savings_pct"]))
	}
	sb.WriteString("└─────────────────────────────────────────────┘\n")
	return sb.String()
}

func (e *TUIViewEngine) RenderStats() string {
	var sb strings.Builder
	sb.WriteString("┌─ Stats ────────────────────────────────────┐\n")
	sb.WriteString("│  Avg Compression: 65.3%                    │\n")
	sb.WriteString("│  Best Savings:    89.2% (git log)          │\n")
	sb.WriteString("│  Top Model:       gpt-4o                   │\n")
	sb.WriteString("│  Top Command:     git status               │\n")
	sb.WriteString("└─────────────────────────────────────────────┘\n")
	return sb.String()
}
