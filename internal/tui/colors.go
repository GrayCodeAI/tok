package tui

import "github.com/charmbracelet/lipgloss"

// Single solid color scheme - ONE accent only, NO gradients
const (
	// ONE primary accent color - Cyan (used everywhere)
	ColorAccent      = "#00D4AA" // Bright cyan/teal - ONLY accent color
	ColorAccentDim   = "#008F70" // Dimmer cyan for subtle elements

	// Status colors (minimal, not competing)
	ColorSuccess     = "#22C55E" // Green - success only
	ColorWarning     = "#EAB308" // Yellow - warnings only
	ColorError       = "#EF4444" // Red - errors only

	// Background - Dark solid
	ColorBg          = "#0A0A0A" // Pure dark background
	ColorBgPanel     = "#111111" // Panel background
	ColorBgBorder    = "#00D4AA" // Accent border (same as accent)

	// Text - Solid grayscale
	ColorText        = "#FFFFFF" // White text
	ColorTextMuted   = "#888888" // Gray text
	ColorTextDim     = "#555555" // Dim text
)

// Single accent color styles - ONE color only, NO gradients
var (
	// Title - Single cyan accent
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent)).
		Background(lipgloss.Color(ColorBgPanel)).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		MarginBottom(1)

	// Header - Solid accent
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorAccent)).
		Padding(0, 2).
		MarginBottom(1)

	// ONE unified box style - single accent border
	BoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		Background(lipgloss.Color(ColorBgPanel)).
		Padding(1, 2).
		Margin(0, 1)

	// Tab styles
	TabActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorAccent)).
		Padding(0, 2).
		MarginRight(1)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorBgPanel)).
		Padding(0, 2).
		MarginRight(1)

	// Text styles
	TextPrimaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText))

	TextSecondaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted))

	TextMutedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim))

	// Status styles - only for indicators
	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess))

	WarningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorWarning))

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError))

	// Accent style - single color
	AccentStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent))

	// Stat styles - accent color
	StatValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent)).
		Background(lipgloss.Color(ColorBg)).
		Padding(0, 1)

	StatLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted))

	// Footer
	FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Background(lipgloss.Color(ColorBgPanel)).
		Padding(0, 1).
		MarginTop(1)

	// Key style - accent
	KeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorAccent)).
		Padding(0, 1)

	// Help style
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted))

	// Bar style - single accent color (NO gradients)
	BarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccent))
)
