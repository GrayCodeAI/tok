package tui

import "github.com/charmbracelet/lipgloss"

// World-Class TUI Color System - Purpose-Driven Colors
// Based on research of k9s, lazygit, grafana, and top monitoring tools
// Each color has a SPECIFIC PURPOSE - no random usage

const (
	// PRIMARY INTERACTION - Cyan/Teal
	// Used for: Active elements, selected items, borders of focused components
	ColorPrimary      = "#00D4AA" // Bright teal/cyan
	ColorPrimaryDim   = "#00A884" // Dimmer for secondary emphasis
	ColorPrimaryBright = "#00FFD4" // Bright for highlights

	// SUCCESS/HEALTHY - Green
	// Used for: Success messages, healthy status, positive trends, savings
	ColorSuccess      = "#4ADE80" // Soft green
	ColorSuccessDim   = "#22C55E" // Dimmer green
	ColorSuccessBright = "#86EFAC" // Bright green

	// WARNING/ATTENTION - Yellow/Orange
	// Used for: Warnings, medium priority, attention needed, cache misses
	ColorWarning      = "#FBBF24" // Amber
	ColorWarningDim   = "#F59E0B" // Darker amber
	ColorWarningBright = "#FCD34D" // Light amber

	// ERROR/CRITICAL - Red
	// Used for: Errors, critical issues, failed status, high memory
	ColorError        = "#F87171" // Soft red
	ColorErrorDim     = "#EF4444" // Standard red
	ColorErrorBright  = "#FCA5A5" // Light red

	// INFO/SPECIAL - Purple
	// Used for: Informational highlights, special features, research references
	ColorInfo         = "#A78BFA" // Soft purple
	ColorInfoDim      = "#8B5CF6" // Standard purple
	ColorInfoBright   = "#C4B5FD" // Light purple

	// DATA VISUALIZATION - Blue Scale
	// Used for: Charts, graphs, bar visualizations, cool metrics
	ColorData1        = "#60A5FA" // Blue
	ColorData2        = "#3B82F6" // Darker blue
	ColorData3        = "#93C5FD" // Lighter blue

	// NEUTRAL TEXT - Grayscale
	// Used for: Text content based on importance
	ColorTextPrimary  = "#F9FAFB" // Almost white - primary text
	ColorTextSecondary = "#D1D5DB" // Light gray - secondary text
	ColorTextMuted    = "#9CA3AF" // Medium gray - labels, hints
	ColorTextDim      = "#6B7280" // Dark gray - disabled, very muted

	// BACKGROUNDS - Dark theme
	ColorBg           = "#0B0B0B" // Main background (near black)
	ColorBgSurface    = "#141414" // Cards, panels
	ColorBgElevated   = "#1F1F1F" // Elevated elements, hover states
	ColorBgBorder     = "#262626" // Subtle borders

	// SPECIAL UI ELEMENTS
	ColorBorderActive = ColorPrimary  // Active component borders
	ColorBorderNormal = "#333333"     // Inactive borders
	ColorHeaderBg     = "#1A1A1A"     // Header background
)

// World-Class TUI Styles - Purpose-Driven Design
var (
	// TITLE - Application header
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimaryBright)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		MarginBottom(1)

	// HEADERS - Section headers with specific purposes
	HeaderPrimary = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 2)

	HeaderSuccess = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorSuccess)).
		Padding(0, 2)

	HeaderWarning = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorWarning)).
		Padding(0, 2)

	HeaderError = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorTextPrimary)).
		Background(lipgloss.Color(ColorError)).
		Padding(0, 2)

	HeaderInfo = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorInfo)).
		Padding(0, 2)

	HeaderSuccessDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSuccess)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2)

	HeaderWarningDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorWarning)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2)

	HeaderErrorDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorError)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2)

	HeaderInfoDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorInfo)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2)

	// BOXES - Container styles
	BoxPrimary = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2)

	BoxSuccess = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorSuccess)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2)

	BoxWarning = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorWarning)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2)

	BoxError = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorError)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2)

	BoxInfo = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorInfo)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2)

	BoxDim = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBgBorder)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(1, 2)

	BoxActive = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimaryBright)).
		Background(lipgloss.Color(ColorBgElevated)).
		Padding(1, 2)

	BoxStyle = BoxPrimary // Default

	// TABS - Navigation with purpose colors
	TabActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 2).
		MarginRight(1)

	TabSuccess = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorSuccess)).
		Padding(0, 2).
		MarginRight(1)

	TabWarning = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorWarning)).
		Padding(0, 2).
		MarginRight(1)

	TabError = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorTextPrimary)).
		Background(lipgloss.Color(ColorError)).
		Padding(0, 2).
		MarginRight(1)

	TabInfo = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorInfo)).
		Padding(0, 2).
		MarginRight(1)

	TabSuccessDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSuccess)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2).
		MarginRight(1)

	TabWarningDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorWarning)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2).
		MarginRight(1)

	TabErrorDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorError)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2).
		MarginRight(1)

	TabInfoDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorInfo)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2).
		MarginRight(1)

	TabInactive = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 2).
		MarginRight(1)

	TabActiveStyle = TabActive
	TabInactiveStyle = TabInactive

	// TEXT - Content colors by purpose
	TextPrimary = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextPrimary))
	TextSecondary = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextSecondary))
	TextMuted = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextMuted))
	TextDim = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim))

	TextPrimaryStyle = TextPrimary
	TextSecondaryStyle = TextSecondary
	TextMutedStyle = TextMuted
	TextDimStyle = TextDim

	// STATUS INDICATORS - Semantic colors
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning))
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError))
	InfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorInfo))

	// ACCENT - Primary highlight
	AccentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorPrimary))

	// STATS - Data display
	StatValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimaryBright)).
		Background(lipgloss.Color(ColorBg)).
		Padding(0, 1)

	StatLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextMuted))

	// FOOTER - Status bar
	FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Background(lipgloss.Color(ColorBgSurface)).
		Padding(0, 1).
		MarginTop(1)

	// KEYBOARD SHORTCUTS
	KeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 1)

	// HELP
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextMuted))

	// DATA VISUALIZATION - Charts and bars
	BarPrimary = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary))
	BarSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
	BarWarning = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning))
	BarError = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError))
	BarInfo = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorInfo))
	BarData1 = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorData1))
	BarData2 = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorData2))
	BarData3 = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorData3))

	// SPARKLINE - Mini charts
	SparklineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
)
