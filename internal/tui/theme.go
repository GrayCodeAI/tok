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

func newTheme() theme {
	bg := lipgloss.Color("#111827")
	panelBg := lipgloss.Color("#1B1C1E")
	panelBgAlt := lipgloss.Color("#202226")
	slate := lipgloss.Color("#8A94A6")
	fg := lipgloss.Color("#E8ECF3")
	line := lipgloss.Color("#2B3442")
	green := lipgloss.Color("#53D18D")
	amber := lipgloss.Color("#E8BC62")
	red := lipgloss.Color("#FF7A7A")
	cyan := lipgloss.Color("#4FC3F7")
	blue := lipgloss.Color("#7AB8FF")
	gold := lipgloss.Color("#F2CC70")
	accentColors := []lipgloss.Color{
		lipgloss.Color("#E056FD"),
		lipgloss.Color("#4FC3F7"),
		lipgloss.Color("#D6D644"),
		lipgloss.Color("#6DDA4F"),
		lipgloss.Color("#FF8A65"),
		lipgloss.Color("#BA68C8"),
		lipgloss.Color("#64FFDA"),
		lipgloss.Color("#FFB74D"),
		lipgloss.Color("#81C784"),
		lipgloss.Color("#7986CB"),
		lipgloss.Color("#4DD0E1"),
		lipgloss.Color("#F06292"),
		lipgloss.Color("#AED581"),
		lipgloss.Color("#FFD54F"),
		lipgloss.Color("#90CAF9"),
		lipgloss.Color("#CE93D8"),
		lipgloss.Color("#A5D6A7"),
		lipgloss.Color("#80CBC4"),
		lipgloss.Color("#FFCC80"),
		lipgloss.Color("#B39DDB"),
		lipgloss.Color("#EF9A9A"),
		lipgloss.Color("#80DEEA"),
		lipgloss.Color("#C5E1A5"),
		lipgloss.Color("#FFF59D"),
	}

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
			Foreground(lipgloss.Color("#F6F2A2")).
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
		SidebarItem: lipgloss.NewStyle().
			Foreground(slate).
			Padding(0, 1),
		SidebarActive: lipgloss.NewStyle().
			Foreground(fg).
			BorderLeft(true).
			BorderForeground(cyan).
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
