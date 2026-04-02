package tokscalenavigation

import (
	"strings"
)

type TUIEngine struct {
	currentView int
	theme       int
}

func NewTUIEngine() *TUIEngine {
	return &TUIEngine{
		currentView: 0,
		theme:       0,
	}
}

func (e *TUIEngine) SwitchView(view int) {
	if view >= 0 && view < 4 {
		e.currentView = view
	}
}

func (e *TUIEngine) GetCurrentView() int {
	return e.currentView
}

func (e *TUIEngine) SwitchTheme(theme int) {
	if theme >= 0 && theme < 9 {
		e.theme = theme
	}
}

func (e *TUIEngine) GetTheme() int {
	return e.theme
}

func (e *TUIEngine) RenderOverview(stats map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("Overview\n")
	sb.WriteString("Commands: " + toString(stats["commands"]) + "\n")
	sb.WriteString("Tokens: " + toString(stats["tokens"]) + "\n")
	sb.WriteString("Saved: " + toString(stats["saved"]) + "\n")
	return sb.String()
}

func (e *TUIEngine) RenderModels(models []map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("Models\n")
	for _, m := range models {
		sb.WriteString(toString(m["name"]) + ": " + toString(m["tokens"]) + " tokens\n")
	}
	return sb.String()
}

func (e *TUIEngine) RenderDaily(daily []map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("Daily\n")
	for _, d := range daily {
		sb.WriteString(toString(d["date"]) + ": " + toString(d["tokens"]) + "\n")
	}
	return sb.String()
}

func (e *TUIEngine) RenderStats() string {
	return "Stats\n"
}

func toString(v interface{}) string {
	if v == nil {
		return "0"
	}
	return strings.TrimSpace(strings.Replace(strings.Trim(strings.TrimPrefix(strings.TrimPrefix(v.(string), " "), " "), " "), "\n", " ", -1))
}
