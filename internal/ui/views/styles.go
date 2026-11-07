package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Core Theme Palette: Red -> Purple -> Matrix Green, with Bright White text
	ColorPrimary     = lipgloss.Color("#FF1744") // Primary: Vibrant Crimson Red
	ColorSecondary   = lipgloss.Color("#AA00FF") // Secondary: Electric Purple
	ColorMatrixGreen = lipgloss.Color("#00FF41") // Accent: Matrix Green
	ColorSuccess     = lipgloss.Color("#00FF41") // Matrix Green
	ColorWarning     = lipgloss.Color("#FFD600") // Bright Gold
	ColorDanger      = lipgloss.Color("#FF1744") // Bright Red
	ColorMuted       = lipgloss.Color("#B0BEC5") // Light Silver-Grey for secondary hints
	ColorText        = lipgloss.Color("#FFFFFF") // Bright Pure White

	// Distinct Color Palette for each Stage (starting with Red -> Purple -> Matrix Green)
	StageColors = []lipgloss.Color{
		lipgloss.Color("#FF1744"), // Stage 1: Crimson Red
		lipgloss.Color("#AA00FF"), // Stage 2: Electric Purple
		lipgloss.Color("#00FF41"), // Stage 3: Matrix Green
		lipgloss.Color("#FF4081"), // Stage 4: Neon Magenta
		lipgloss.Color("#00E5FF"), // Stage 5: Cyber Cyan
		lipgloss.Color("#FFD600"), // Stage 6+: Cyber Gold
		lipgloss.Color("#FF6E40"), // Stage 7+: Neon Orange
	}

	// Bright White Default Text Style
	TextStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(false)

	BoldWhiteStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)

	// App Header Banner
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Box / Panel Styles
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Foreground(ColorText).
			Width(92).
			Padding(1, 2).
			MarginBottom(1)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Foreground(ColorText).
				Width(92).
				Padding(1, 2).
				MarginBottom(1)

	SuccessBannerStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorMatrixGreen).
				Foreground(ColorText).
				Width(92).
				Padding(1, 2).
				MarginBottom(1)

	WarningBannerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSecondary).
				Foreground(ColorText).
				Width(92).
				Padding(1, 2).
				MarginBottom(1)

	// Badges
	BadgeChainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#607D8B")).
			Padding(0, 1)

	BadgeSelectedChainStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(ColorMatrixGreen).
				Padding(0, 2)

	InputBoxBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorMatrixGreen).
				Padding(0, 1).
				Width(84)

	BadgeExchangeHitStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(ColorMatrixGreen).
				Padding(0, 1)

	BadgeSkippedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary)

	// Text helpers
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	AddressStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMatrixGreen)

	AmountStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText)

	ArrowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger)
)

// RenderBigBanner renders a compact block-art "CRYPTO TRACKER" banner in solid crimson red,
// sized to match the exact 94-column outer width of the UI panels below it.
func RenderBigBanner() string {
	lines := []string{
		`▄█████ █████▄ ██  ██ █████▄ ██████ ▄████▄     ██████ █████▄ ▄████▄ ▄█████ ██  ▄█ ██████ █████▄`,
		`██     ██  ██  ████  ██  ██   ██   ██  ██       ██   ██  ██ ██  ██ ██     ██▄█▀  ██▄▄   ██  ██`,
		`██     █████▀   ██   █████▀   ██   ██  ██       ██   █████▀ ██████ ██     ████▄  ██▀▀   █████▀`,
		`▀█████ ██  ██   ██   ██       ██   ▀████▀       ██   ██  ██ ██  ██ ▀█████ ██  ██ ██████ ██  ██`,
	}

	redStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)

	var out strings.Builder
	out.WriteString("\n")
	for _, line := range lines {
		out.WriteString(redStyle.Render(line) + "\n")
	}
	return out.String()
}

// StageColor returns a distinct color for each stage number (1-indexed).
func StageColor(stage int) lipgloss.Color {
	if stage <= 0 {
		return ColorMatrixGreen
	}
	idx := (stage - 1) % len(StageColors)
	return StageColors[idx]
}

// StageBadgeStyle returns a bold foreground style colored specifically for the given stage.
func StageBadgeStyle(stage int) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(StageColor(stage))
}

// StagePillStyle returns a high-contrast pill badge colored specifically for the given stage.
func StagePillStyle(stage int) lipgloss.Style {
	fg := ColorText
	if StageColor(stage) == ColorMatrixGreen || StageColor(stage) == ColorWarning {
		fg = lipgloss.Color("#000000")
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(fg).
		Background(StageColor(stage)).
		Padding(0, 1)
}

// FormatShortAddr returns a readable shortened address while keeping prefix and suffix clear.
func FormatShortAddr(addr string) string {
	if len(addr) <= 18 {
		return addr
	}
	return fmt.Sprintf("%s...%s", addr[:8], addr[len(addr)-6:])
}
