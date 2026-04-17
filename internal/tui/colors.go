package tui

import "github.com/charmbracelet/lipgloss"

// Professional Dashboard Color Palette - 20+ SOLID colors, HIGH CONTRAST
// No gradients, pure solid colors only
const (
	// Row 1: Blues & Cyans (cool)
	ColorBlue1   = "#00D4FF" // Cyan
	ColorBlue2   = "#0099CC" // Deep Blue
	ColorBlue3   = "#0066AA" // Navy
	ColorBlue4   = "#33CCFF" // Sky
	ColorTeal    = "#00D4AA" // Teal

	// Row 2: Greens (nature)
	ColorGreen1  = "#00FF66" // Neon Green
	ColorGreen2  = "#66CC00" // Lime
	ColorGreen3  = "#00AA44" // Forest
	ColorGreen4  = "#88FF00" // Chartreuse

	// Row 3: Yellows & Oranges (warm)
	ColorYellow1 = "#FFDD00" // Gold
	ColorYellow2 = "#FFAA00" // Amber
	ColorOrange1 = "#FF7700" // Orange
	ColorOrange2 = "#FF4400" // Deep Orange

	// Row 4: Reds & Pinks (hot)
	ColorRed1    = "#FF0044" // Red
	ColorRed2    = "#CC0066" // Crimson
	ColorPink1   = "#FF66AA" // Pink
	ColorPink2   = "#FF00CC" // Magenta

	// Row 5: Purples & Violets
	ColorPurple1 = "#9900FF" // Purple
	ColorPurple2 = "#6600CC" // Deep Purple
	ColorViolet  = "#CC66FF" // Violet
	ColorIndigo  = "#4444FF" // Indigo

	// Row 6: Neutrals & Grays (contrast)
	ColorWhite   = "#FFFFFF"
	ColorGray1   = "#CCCCCC"
	ColorGray2   = "#888888"
	ColorGray3   = "#444444"
	ColorBlack   = "#0A0A0A"

	// Backgrounds
	ColorBg      = "#0F0F0F" // Main background
	ColorBgPanel = "#1A1A1A" // Panel
	ColorBgDark  = "#050505" // Darker

	// Semantic shortcuts
	ColorSuccess = ColorGreen1
	ColorWarning = ColorYellow1
	ColorError   = ColorRed1
	ColorInfo    = ColorBlue1
)

// Professional Dashboard Styles - 20+ SOLID colors
var (
	// Title - Cyan
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBlue1)).
		Background(lipgloss.Color(ColorBgPanel)).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBlue1)).
		MarginBottom(1)

	// Headers - Different colors
	HeaderCyan   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorBlue1)).Padding(0, 2)
	HeaderGreen  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorGreen1)).Padding(0, 2)
	HeaderYellow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorYellow1)).Padding(0, 2)
	HeaderOrange = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorOrange1)).Padding(0, 2)
	HeaderRed    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorRed1)).Padding(0, 2)
	HeaderPurple = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorPurple1)).Padding(0, 2)

	// Box styles - 20 different colored borders
	BoxCyan   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorBlue1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxBlue   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorBlue2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxNavy   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorBlue3)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxSky    = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorBlue4)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxTeal   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorTeal)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxGreen  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorGreen1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxLime   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorGreen2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxForest = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorGreen3)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxChart  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorGreen4)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxGold   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorYellow1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxAmber  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorYellow2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxOrange = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorOrange1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxDeepO  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorOrange2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxRed    = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorRed1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxCrim   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorRed2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxPink   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPink1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxMagent = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPink2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxPurple = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPurple1)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxDeepP  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPurple2)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxViolet = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorViolet)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)
	BoxIndigo = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorIndigo)).Background(lipgloss.Color(ColorBgPanel)).Padding(1, 2).Margin(0, 1)

	// Default box (cyan)
	BoxStyle = BoxCyan

	// Tab styles - colorful
	TabCyan   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorBlue1)).Padding(0, 2).MarginRight(1)
	TabGreen  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorGreen1)).Padding(0, 2).MarginRight(1)
	TabYellow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorYellow1)).Padding(0, 2).MarginRight(1)
	TabOrange = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorOrange1)).Padding(0, 2).MarginRight(1)
	TabRed    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorRed1)).Padding(0, 2).MarginRight(1)
	TabPurple = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorPurple1)).Padding(0, 2).MarginRight(1)
	TabPink   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorPink1)).Padding(0, 2).MarginRight(1)
	TabBlue   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorBlue2)).Padding(0, 2).MarginRight(1)
	TabTeal   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorTeal)).Padding(0, 2).MarginRight(1)
	TabIndigo = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorIndigo)).Padding(0, 2).MarginRight(1)

	TabActiveStyle   = TabCyan
	TabInactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray2)).Background(lipgloss.Color(ColorBgPanel)).Padding(0, 2).MarginRight(1)

	// Text colors - all 20 colors available
	TextCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue1))
	TextBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue2))
	TextGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen1))
	TextLime   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen2))
	TextYellow = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow1))
	TextOrange = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange1))
	TextRed    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed1))
	TextPink   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPink1))
	TextPurple = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple1))
	TextViolet = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorViolet))
	TextWhite  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWhite))
	TextGray   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray1))
	TextMuted  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray2))
	TextDim    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray3))

	TextPrimaryStyle   = TextWhite
	TextSecondaryStyle = TextGray
	TextMutedStyle     = TextMuted

	// Status
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen1))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow1))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed1))
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue1))

	// Accent (cyan)
	AccentStyle = TextCyan

	// Stats
	StatValueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlue1)).Background(lipgloss.Color(ColorBlack)).Padding(0, 1)
	StatLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray2))

	// Footer
	FooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray3)).Background(lipgloss.Color(ColorBgPanel)).Padding(0, 1).MarginTop(1)

	// Key
	KeyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorBlack)).Background(lipgloss.Color(ColorBlue1)).Padding(0, 1)

	// Help
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray2))

	// Bar colors - all 20
	BarCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue1))
	BarBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue2))
	BarGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen1))
	BarLime   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen2))
	BarYellow = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow1))
	BarOrange = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange1))
	BarRed    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed1))
	BarPink   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPink1))
	BarPurple = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple1))
	BarViolet = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorViolet))
	BarTeal   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal))
)
