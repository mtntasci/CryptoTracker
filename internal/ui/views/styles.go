package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color Palette
	ColorPrimary   = lipgloss.Color("#00E5FF") // Cyan
	ColorSecondary = lipgloss.Color("#7C4DFF") // Deep Purple
	ColorSuccess   = lipgloss.Color("#00E676") // Neon Green
	ColorWarning   = lipgloss.Color("#FFB300") // Amber
	ColorDanger    = lipgloss.Color("#FF5252") // Coral Red
	ColorMuted     = lipgloss.Color("#78909C") // Blue Grey
	ColorText      = lipgloss.Color("#ECEFF1") // Off-White

	// App Header Banner
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0D1117")).
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
			Padding(1, 2).
			MarginBottom(1)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2).
				MarginBottom(1)

	SuccessBannerStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorSuccess).
				Padding(1, 2).
				MarginBottom(1)

	WarningBannerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorWarning).
				Padding(1, 2).
				MarginBottom(1)

	// Badges
	BadgeChainStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSecondary).
			Padding(0, 1)

	BadgeSelectedChainStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0D1117")).
				Background(ColorPrimary).
				Padding(0, 1)

	BadgeExchangeHitStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0D1117")).
				Background(ColorSuccess).
				Padding(0, 1)

	BadgeHopStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	BadgeSkippedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	// Text helpers
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	AddressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#80CBC4"))

	AmountStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD54F"))

	ArrowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger)
)

// FormatShortAddr returns a readable shortened address while keeping prefix and suffix clear.
func FormatShortAddr(addr string) string {
	if len(addr) <= 18 {
		return addr
	}
	return fmt.Sprintf("%s...%s", addr[:8], addr[len(addr)-6:])
}

