package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors mirrored from internal/output/format.go.
// TUI re-declares them to avoid importing output (which writes to stdout).
var (
	ColorGreen  = lipgloss.Color("#00CC88")
	ColorRed    = lipgloss.Color("#FF4444")
	ColorYellow = lipgloss.Color("#FFAA00")
	ColorPurple = lipgloss.Color("#AA55FF")
	ColorCyan   = lipgloss.Color("#00CCFF")
	ColorDim    = lipgloss.Color("#666666")
	ColorWhite  = lipgloss.Color("#FFFFFF")
)

// Styles
var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true).
			MarginBottom(1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	SectionStyle = lipgloss.NewStyle().
			MarginTop(1)

	ActiveBadge = lipgloss.NewStyle().
			Background(ColorGreen).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	ClosedBadge = lipgloss.NewStyle().
			Background(ColorRed).
			Foreground(ColorWhite).
			Padding(0, 1).
			Bold(true)

	BidStyle = lipgloss.NewStyle().Foreground(ColorGreen)
	AskStyle = lipgloss.NewStyle().Foreground(ColorRed)

	GreenStyle  = lipgloss.NewStyle().Foreground(ColorGreen)
	RedStyle    = lipgloss.NewStyle().Foreground(ColorRed)
	YellowStyle = lipgloss.NewStyle().Foreground(ColorYellow)
	CyanStyle   = lipgloss.NewStyle().Foreground(ColorCyan)
	DimStyle    = lipgloss.NewStyle().Foreground(ColorDim)

	// Tab styles for portfolio
	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorCyan)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Padding(0, 2)

	// Success/error boxes
	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGreen).
			Padding(1, 2)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorRed).
			Padding(1, 2)

	// Spread badge
	SpreadBadge = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true).
			Padding(0, 1)

	// Shell prompt
	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	// Bordered content card
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim).
			Padding(0, 1)

	// Markets list accent styles
	SelectedAccent   = lipgloss.NewStyle().Background(ColorCyan).Foreground(ColorCyan)
	UnselectedAccent = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#333333"))
	SelectedItemBg   = lipgloss.NewStyle().Background(lipgloss.Color("#111122"))
	ProbBarEmpty     = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ProbPctStyle     = lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
	MetaSepStyle     = lipgloss.NewStyle().Foreground(ColorDim)

	// Category tab styles (markets list)
	CategoryActiveStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true).
				Underline(true).
				Padding(0, 1)

	CategoryInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Padding(0, 1)

	CategorySepStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333"))

	// Status bar (bottom chrome)
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#1A1A2E")).
			PaddingTop(1)
)

// StatusBadge returns a styled status badge string.
func StatusBadge(active, closed bool) string {
	if closed {
		return ClosedBadge.Render("CLOSED")
	}
	if active {
		return ActiveBadge.Render("ACTIVE")
	}
	return DimStyle.Render("INACTIVE")
}

// FormatPrice formats a 0-1 price as colored cents (e.g. "65.0¢").
func FormatPrice(price float64) string {
	s := fmt.Sprintf("%.1f¢", price*100)
	if price >= 0.5 {
		return GreenStyle.Render(s)
	}
	return RedStyle.Render(s)
}

// FormatPricePtr formats a nullable price pointer.
func FormatPricePtr(price *float64) string {
	if price == nil {
		return DimStyle.Render("—")
	}
	return FormatPrice(*price)
}

// FormatVolume formats a volume string with K/M/B suffixes.
func FormatVolume(vol string) string {
	v, err := strconv.ParseFloat(vol, 64)
	if err != nil {
		return vol
	}
	abs := math.Abs(v)
	switch {
	case abs >= 1_000_000_000:
		return CyanStyle.Render(fmt.Sprintf("$%.1fB", abs/1_000_000_000))
	case abs >= 1_000_000:
		return CyanStyle.Render(fmt.Sprintf("$%.1fM", abs/1_000_000))
	case abs >= 1_000:
		return CyanStyle.Render(fmt.Sprintf("$%.1fK", abs/1_000))
	default:
		return CyanStyle.Render(fmt.Sprintf("$%.0f", abs))
	}
}

// FormatPnL formats a P&L value with color (green positive, red negative).
func FormatPnL(pnl float64) string {
	s := fmt.Sprintf("$%.2f", math.Abs(pnl))
	if pnl >= 0 {
		return GreenStyle.Render("+" + s)
	}
	return RedStyle.Render("-" + s)
}

// FormatPercent formats a 0-1 value as percentage.
func FormatPercent(v float64) string {
	return fmt.Sprintf("%.1f%%", v*100)
}

// RenderProbBar renders a probability bar of the given width.
// Uses green fill for prob >= 0.5, red otherwise.
func RenderProbBar(prob float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(prob * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	empty := width - filled

	color := ColorGreen
	if prob < 0.5 {
		color = ColorRed
	}
	fillStyle := lipgloss.NewStyle().Foreground(color)

	bar := fillStyle.Render(strings.Repeat("█", filled)) +
		ProbBarEmpty.Render(strings.Repeat("░", empty))
	pct := ProbPctStyle.Render(fmt.Sprintf("%3.0f%%", prob*100))
	return bar + " " + pct
}

// FormatUSD formats a dollar value.
func FormatUSD(v float64) string {
	if math.Abs(v) >= 1_000_000 {
		return fmt.Sprintf("$%.1fM", v/1_000_000)
	}
	if math.Abs(v) >= 1_000 {
		return fmt.Sprintf("$%.1fK", v/1_000)
	}
	return fmt.Sprintf("$%.2f", v)
}

// Truncate truncates a string to max characters, appending "…" if needed.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return s[:max-1] + "…"
}

// blendColor linearly interpolates between two hex colors.
// t=1.0 returns 'from', t=0.0 returns 'to'. Clamped to [0,1].
func blendColor(from, to lipgloss.Color, t float64) lipgloss.Color {
	if t > 1.0 {
		t = 1.0
	}
	if t < 0.0 {
		t = 0.0
	}

	r1, g1, b1 := parseHexColor(string(from))
	r2, g2, b2 := parseHexColor(string(to))

	r := uint8(float64(r1)*t + float64(r2)*(1-t))
	g := uint8(float64(g1)*t + float64(g2)*(1-t))
	b := uint8(float64(b1)*t + float64(b2)*(1-t))

	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// parseHexColor parses a "#RRGGBB" hex string to RGB components.
func parseHexColor(hex string) (uint8, uint8, uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return uint8(r), uint8(g), uint8(b)
}
