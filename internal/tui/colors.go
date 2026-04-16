package tui

import "github.com/charmbracelet/lipgloss"

// Vibrant color palette for high contrast
const (
	// Primary colors
	ColorPrimary    = "#FF6B6B" // Coral red
	ColorSecondary  = "#4ECDC4" // Turquoise
	ColorAccent     = "#FFE66D" // Yellow
	ColorSuccess    = "#95E1D3" // Mint green
	ColorWarning    = "#FFA07A" // Light salmon
	ColorError      = "#FF4757" // Bright red
	ColorInfo       = "#70A1FF" // Sky blue

	// Background colors
	ColorBgDark     = "#1E1E2E" // Dark purple-gray
	ColorBgDarker   = "#16161E" // Almost black
	ColorBgLight    = "#2D2D44" // Lighter purple-gray
	ColorBgLighter  = "#3D3D5C" // Even lighter

	// Text colors
	ColorTextPrimary   = "#FFFFFF" // White
	ColorTextSecondary = "#A0A0B0" // Gray
	ColorTextMuted     = "#6E6E8A" // Muted gray

	// Gradient colors
	ColorGradientStart = "#FF6B6B" // Red
	ColorGradientMid   = "#4ECDC4" // Cyan
	ColorGradientEnd   = "#95E1D3" // Green
)

// Styles with vibrant colors
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgDark)).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		MarginBottom(1)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorTextPrimary)).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 1).
		MarginBottom(1)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorSecondary)).
		Background(lipgloss.Color(ColorBgDark)).
		Padding(1, 2).
		Margin(0, 1)

	BoxActiveStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgLight)).
		Padding(1, 2).
		Margin(0, 1)

	// Tab styles
	TabActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBgDark)).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 3).
		MarginRight(1)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextSecondary)).
		Background(lipgloss.Color(ColorBgLight)).
		Padding(0, 3).
		MarginRight(1)

	// Text styles
	TextPrimaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextPrimary))

	TextSecondaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextSecondary))

	TextMutedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted))

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSuccess))

	WarningStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorWarning))

	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorError))

	InfoStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorInfo))

	// Stat value styles
	StatValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent)).
		Background(lipgloss.Color(ColorBgDarker)).
		Padding(0, 1)

	StatLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextSecondary))

	// Gradient bar style
	GradientBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgLight))

	// Footer style
	FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorBgDarker)).
		Padding(0, 1).
		MarginTop(1)

	// Key style (for keyboard shortcuts)
	KeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBgDark)).
		Background(lipgloss.Color(ColorAccent)).
		Padding(0, 1)

	// Help style
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextSecondary))
)

// GetGradientColors returns colors for progress bars
func GetGradientColors() []string {
	return []string{
		ColorGradientStart,
		"#FF8E53",
		"#FF6B9D",
		"#C44569",
		ColorGradientMid,
		"#38B2AC",
		ColorGradientEnd,
	}
}
