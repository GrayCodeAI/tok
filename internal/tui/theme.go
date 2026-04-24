package tui

import "github.com/charmbracelet/lipgloss"

type theme struct {
	App            lipgloss.Style
	Header         lipgloss.Style
	HeaderMuted    lipgloss.Style
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	SectionTitle   lipgloss.Style
	PanelTitle     lipgloss.Style
	Panel          lipgloss.Style
	Card           lipgloss.Style
	CardAccent     lipgloss.Style
	CardPositive   lipgloss.Style
	CardWarning    lipgloss.Style
	CardInfo       lipgloss.Style
	CardLabel      lipgloss.Style
	CardMeta       lipgloss.Style
	Sidebar        lipgloss.Style
	SidebarItem    lipgloss.Style
	SidebarActive  lipgloss.Style
	SidebarKey     lipgloss.Style
	Main           lipgloss.Style
	RightPane      lipgloss.Style
	Insight        lipgloss.Style
	Footer         lipgloss.Style
	FooterKey      lipgloss.Style
	Muted          lipgloss.Style
	Positive       lipgloss.Style
	Warning        lipgloss.Style
	Danger         lipgloss.Style
	Focus          lipgloss.Style
	TableHeader    lipgloss.Style
	TableRow       lipgloss.Style
	TableRowAccent lipgloss.Style
	ValuePositive  lipgloss.Style
	ValueWarning   lipgloss.Style
	ValueFocus     lipgloss.Style
	ValueGold      lipgloss.Style
	BarEmpty       lipgloss.Style
	AccentColors   []lipgloss.Color
}

// ThemeName enumerates the bundled themes. Unknown names fall back to
// ThemeDark so typos don't crash the TUI.
type ThemeName string

const (
	ThemeDark         ThemeName = "dark"
	ThemeLight        ThemeName = "light"
	ThemeHighContrast ThemeName = "high-contrast"
	ThemeColorblind   ThemeName = "colorblind"
)

// AvailableThemes is the list users can cycle through via the palette.
// Kept in this order because it's the order the theme picker renders.
var AvailableThemes = []ThemeName{
	ThemeDark,
	ThemeLight,
	ThemeHighContrast,
	ThemeColorblind,
}

// themePalette carries just the color constants for one theme. Kept
// separate from the lipgloss.Style construction below so adding a new
// theme is a one-struct change, not a copy-and-mutate of newTheme.
type themePalette struct {
	bg, panelBg, panelBgAlt, fg, slate, line lipgloss.Color
	green, amber, red, cyan, blue, gold      lipgloss.Color
	accents                                  []lipgloss.Color
}

func paletteFor(name ThemeName) themePalette {
	switch name {
	case ThemeLight:
		return themePalette{
			bg:         lipgloss.Color("#F5F7FA"),
			panelBg:    lipgloss.Color("#FFFFFF"),
			panelBgAlt: lipgloss.Color("#EFF1F5"),
			fg:         lipgloss.Color("#1F2328"),
			slate:      lipgloss.Color("#5A6371"),
			line:       lipgloss.Color("#D0D7DE"),
			green:      lipgloss.Color("#1A7F37"),
			amber:      lipgloss.Color("#9A6700"),
			red:        lipgloss.Color("#CF222E"),
			cyan:       lipgloss.Color("#0969DA"),
			blue:       lipgloss.Color("#0550AE"),
			gold:       lipgloss.Color("#7D4E00"),
			accents: []lipgloss.Color{
				"#8250DF", "#0969DA", "#BF8700", "#1A7F37",
				"#D4613A", "#6F42C1", "#116B79", "#A3531F",
				"#24754A", "#3B5CCC", "#0E4F5E", "#B4307B",
				"#5B7F2B", "#8B6400", "#2D6EAE", "#9C5A9F",
				"#3C7650", "#386564", "#A36125", "#755AB6",
				"#C73232", "#2B758A", "#7A8C3A", "#998300",
			},
		}
	case ThemeHighContrast:
		return themePalette{
			bg:         lipgloss.Color("#000000"),
			panelBg:    lipgloss.Color("#000000"),
			panelBgAlt: lipgloss.Color("#0A0A0A"),
			fg:         lipgloss.Color("#FFFFFF"),
			slate:      lipgloss.Color("#C0C0C0"),
			line:       lipgloss.Color("#808080"),
			green:      lipgloss.Color("#00FF00"),
			amber:      lipgloss.Color("#FFFF00"),
			red:        lipgloss.Color("#FF4040"),
			cyan:       lipgloss.Color("#00FFFF"),
			blue:       lipgloss.Color("#4040FF"),
			gold:       lipgloss.Color("#FFD700"),
			accents: []lipgloss.Color{
				"#FFFFFF", "#00FFFF", "#FFFF00", "#00FF00",
				"#FF8080", "#FF00FF", "#00FFC0", "#FFA500",
				"#80FF80", "#8080FF", "#00FFFF", "#FF80FF",
				"#C0FF00", "#FFE000", "#80C0FF", "#E080FF",
				"#C0FFC0", "#80FFFF", "#FFC080", "#C080FF",
				"#FFB0B0", "#80FFFF", "#E0FF80", "#FFFFA0",
			},
		}
	case ThemeColorblind:
		// Okabe-Ito palette: 8 colors distinguishable for most forms of
		// color-vision deficiency. Used as-is for accents; hue-coded
		// semantic colors (green/red) remain but chosen from the palette.
		return themePalette{
			bg:         lipgloss.Color("#111827"),
			panelBg:    lipgloss.Color("#1B1C1E"),
			panelBgAlt: lipgloss.Color("#202226"),
			fg:         lipgloss.Color("#E8ECF3"),
			slate:      lipgloss.Color("#8A94A6"),
			line:       lipgloss.Color("#2B3442"),
			green:      lipgloss.Color("#009E73"), // bluish-green
			amber:      lipgloss.Color("#E69F00"), // orange
			red:        lipgloss.Color("#D55E00"), // vermillion
			cyan:       lipgloss.Color("#56B4E9"), // sky blue
			blue:       lipgloss.Color("#0072B2"), // blue
			gold:       lipgloss.Color("#F0E442"), // yellow
			accents: []lipgloss.Color{
				"#E69F00", "#56B4E9", "#009E73", "#F0E442",
				"#0072B2", "#D55E00", "#CC79A7", "#999999",
				"#E69F00", "#56B4E9", "#009E73", "#F0E442",
				"#0072B2", "#D55E00", "#CC79A7", "#999999",
				"#E69F00", "#56B4E9", "#009E73", "#F0E442",
				"#0072B2", "#D55E00", "#CC79A7", "#999999",
			},
		}
	default: // ThemeDark
		return themePalette{
			bg:         lipgloss.Color("#111827"),
			panelBg:    lipgloss.Color("#1B1C1E"),
			panelBgAlt: lipgloss.Color("#202226"),
			fg:         lipgloss.Color("#E8ECF3"),
			slate:      lipgloss.Color("#8A94A6"),
			line:       lipgloss.Color("#2B3442"),
			green:      lipgloss.Color("#53D18D"),
			amber:      lipgloss.Color("#E8BC62"),
			red:        lipgloss.Color("#FF7A7A"),
			cyan:       lipgloss.Color("#4FC3F7"),
			blue:       lipgloss.Color("#7AB8FF"),
			gold:       lipgloss.Color("#F2CC70"),
			accents: []lipgloss.Color{
				"#E056FD", "#4FC3F7", "#D6D644", "#6DDA4F",
				"#FF8A65", "#BA68C8", "#64FFDA", "#FFB74D",
				"#81C784", "#7986CB", "#4DD0E1", "#F06292",
				"#AED581", "#FFD54F", "#90CAF9", "#CE93D8",
				"#A5D6A7", "#80CBC4", "#FFCC80", "#B39DDB",
				"#EF9A9A", "#80DEEA", "#C5E1A5", "#FFF59D",
			},
		}
	}
}

// newTheme returns the default (dark) theme. Prefer newThemeByName for
// new callers; this helper is kept because early tests invoke it
// directly and switching them all in one sweep isn't worth the churn.
func newTheme() theme {
	return newThemeByName(ThemeDark)
}

// newThemeByName returns a fully-populated theme for the given name.
// Unknown names fall through to the dark palette.
func newThemeByName(name ThemeName) theme {
	p := paletteFor(name)
	bg := p.bg
	panelBg := p.panelBg
	panelBgAlt := p.panelBgAlt
	slate := p.slate
	fg := p.fg
	line := p.line
	green := p.green
	amber := p.amber
	red := p.red
	cyan := p.cyan
	blue := p.blue
	gold := p.gold
	accentColors := p.accents

	return theme{
		App: lipgloss.NewStyle().
			Foreground(fg).
			Background(bg),
		Header: lipgloss.NewStyle().
			Foreground(fg).
			Background(bg).
			BorderBottom(true).
			BorderForeground(line).
			Padding(0, 1),
		HeaderMuted: lipgloss.NewStyle().
			Foreground(slate),
		Title: lipgloss.NewStyle().
			Foreground(fg).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(slate),
		SectionTitle: lipgloss.NewStyle().
			Foreground(blue).
			Bold(true),
		PanelTitle: lipgloss.NewStyle().
			Foreground(gold).
			Bold(true),
		Panel: lipgloss.NewStyle().
			Background(panelBg).
			Border(lipgloss.NormalBorder()).
			BorderForeground(line).
			Padding(1, 1),
		Card: lipgloss.NewStyle().
			Background(panelBgAlt).
			Border(lipgloss.NormalBorder()).
			BorderForeground(line).
			Padding(0, 1),
		CardAccent: lipgloss.NewStyle().
			Background(panelBgAlt).
			Border(lipgloss.NormalBorder()).
			BorderForeground(cyan).
			Padding(0, 1),
		CardPositive: lipgloss.NewStyle().
			Background(panelBgAlt).
			Border(lipgloss.NormalBorder()).
			BorderForeground(green).
			Padding(0, 1),
		CardWarning: lipgloss.NewStyle().
			Background(panelBgAlt).
			Border(lipgloss.NormalBorder()).
			BorderForeground(amber).
			Padding(0, 1),
		CardInfo: lipgloss.NewStyle().
			Background(panelBgAlt).
			Border(lipgloss.NormalBorder()).
			BorderForeground(blue).
			Padding(0, 1),
		CardLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A7A2")).
			Bold(true),
		CardMeta: lipgloss.NewStyle().
			Foreground(slate),
		Sidebar: lipgloss.NewStyle().
			Background(bg).
			BorderRight(true).
			BorderForeground(line).
			Padding(1, 1),
		// Sidebar items and the active marker share the same frame so
		// selecting a section doesn't shift the name column by one.
		// The active state is conveyed by a prefix glyph (▸) and a
		// brighter foreground, applied in renderSidebar.
		SidebarItem: lipgloss.NewStyle().
			Foreground(slate).
			Padding(0, 1),
		SidebarActive: lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true).
			Padding(0, 1),
		SidebarKey: lipgloss.NewStyle().
			Foreground(gold),
		Main: lipgloss.NewStyle().
			Background(bg).
			Padding(0, 1),
		RightPane: lipgloss.NewStyle().
			Background(bg).
			BorderLeft(true).
			BorderForeground(line).
			Padding(0, 1),
		Insight: lipgloss.NewStyle().
			Foreground(slate).
			Padding(0, 0, 1, 0),
		Footer: lipgloss.NewStyle().
			Foreground(slate).
			Background(bg).
			BorderTop(true).
			BorderForeground(line).
			Padding(0, 1),
		FooterKey: lipgloss.NewStyle().
			Foreground(fg).
			Underline(true),
		Muted: lipgloss.NewStyle().
			Foreground(slate),
		Positive: lipgloss.NewStyle().
			Foreground(green).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(amber).
			Bold(true),
		Danger: lipgloss.NewStyle().
			Foreground(red).
			Bold(true),
		Focus: lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true),
		TableHeader: lipgloss.NewStyle().
			Foreground(slate).
			Bold(true),
		TableRow: lipgloss.NewStyle().
			Foreground(fg),
		TableRowAccent: lipgloss.NewStyle().
			Foreground(cyan),
		ValuePositive: lipgloss.NewStyle().
			Foreground(green).
			Bold(true),
		ValueWarning: lipgloss.NewStyle().
			Foreground(amber).
			Bold(true),
		ValueFocus: lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true),
		ValueGold: lipgloss.NewStyle().
			Foreground(gold).
			Bold(true),
		BarEmpty: lipgloss.NewStyle().
			Foreground(line),
		AccentColors: accentColors,
	}
}
