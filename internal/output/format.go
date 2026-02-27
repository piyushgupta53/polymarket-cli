package output

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Package-level output options set by cmd/root.go PersistentPreRun.
var (
	optNoHeaders   bool
	optTSV         bool
	optCompactJSON bool
)

// SetOutputOptions configures package-level output behavior.
func SetOutputOptions(_, noHeaders, _, tsv, compactJSON bool) {
	optNoHeaders = noHeaders
	optTSV = tsv
	optCompactJSON = compactJSON
}

var (
	// Colors
	Green  = lipgloss.Color("#00CC88")
	Red    = lipgloss.Color("#FF4444")
	Yellow = lipgloss.Color("#FFAA00")
	Purple = lipgloss.Color("#AA55FF")
	Cyan   = lipgloss.Color("#00CCFF")
	Dim    = lipgloss.Color("#666666")
	White  = lipgloss.Color("#FFFFFF")

	// Styles
	GreenStyle  = lipgloss.NewStyle().Foreground(Green)
	RedStyle    = lipgloss.NewStyle().Foreground(Red)
	YellowStyle = lipgloss.NewStyle().Foreground(Yellow)
	PurpleStyle = lipgloss.NewStyle().Foreground(Purple)
	CyanStyle   = lipgloss.NewStyle().Foreground(Cyan)
	DimStyle    = lipgloss.NewStyle().Foreground(Dim)
	BoldStyle   = lipgloss.NewStyle().Bold(true)

	// Badge styles
	ActiveBadge = lipgloss.NewStyle().
			Background(Green).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	ClosedBadge = lipgloss.NewStyle().
			Background(Red).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)
)

// FormatPrice formats a price value (0-1) as a percentage-style display.
func FormatPrice(price float64) string {
	cents := price * 100
	s := fmt.Sprintf("%.1f¢", cents)
	if price >= 0.5 {
		return GreenStyle.Render(s)
	}
	return RedStyle.Render(s)
}

// FormatPricePtr formats a *float64 price.
func FormatPricePtr(price *float64) string {
	if price == nil {
		return DimStyle.Render("—")
	}
	return FormatPrice(*price)
}

// FormatPercent formats a 0-1 value as a percentage.
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

// FormatVolume formats a volume number with K/M/B suffixes.
func FormatVolume(volume string) string {
	v, err := strconv.ParseFloat(volume, 64)
	if err != nil {
		return volume
	}
	return FormatVolumeFloat(v)
}

// FormatVolumeFloat formats a float64 volume with K/M/B suffixes.
func FormatVolumeFloat(v float64) string {
	if v == 0 {
		return DimStyle.Render("$0")
	}

	abs := math.Abs(v)
	sign := ""
	if v < 0 {
		sign = "-"
	}

	switch {
	case abs >= 1_000_000_000:
		return CyanStyle.Render(fmt.Sprintf("%s$%.1fB", sign, abs/1_000_000_000))
	case abs >= 1_000_000:
		return CyanStyle.Render(fmt.Sprintf("%s$%.1fM", sign, abs/1_000_000))
	case abs >= 1_000:
		return CyanStyle.Render(fmt.Sprintf("%s$%.1fK", sign, abs/1_000))
	default:
		return CyanStyle.Render(fmt.Sprintf("%s$%.0f", sign, abs))
	}
}

// FormatTimestamp formats a timestamp string for display.
func FormatTimestamp(ts string) string {
	if ts == "" {
		return DimStyle.Render("—")
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, ts); err == nil {
			return t.Format("Jan 02, 2006 15:04")
		}
	}
	return ts
}

// FormatStatus returns a styled active/closed badge.
func FormatStatus(active, closed bool) string {
	if closed {
		return ClosedBadge.Render("CLOSED")
	}
	if active {
		return ActiveBadge.Render("ACTIVE")
	}
	return DimStyle.Render("INACTIVE")
}

// Truncate truncates a string to maxLen, adding "…" if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}

// FormatOutcomePrices returns a formatted string of outcomes and their prices.
func FormatOutcomePrices(outcomes, prices string) string {
	outcomeList := parseJSONStringArray(outcomes)
	priceList := parseJSONStringArray(prices)

	if len(outcomeList) == 0 {
		return DimStyle.Render("—")
	}

	var parts []string
	for i, outcome := range outcomeList {
		price := "?"
		if i < len(priceList) {
			if p, err := strconv.ParseFloat(priceList[i], 64); err == nil {
				price = FormatPrice(p)
			} else {
				price = priceList[i]
			}
		}
		parts = append(parts, fmt.Sprintf("%s: %s", outcome, price))
	}

	return strings.Join(parts, "  ")
}

func parseJSONStringArray(raw string) []string {
	if raw == "" {
		return nil
	}
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "[") {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil
	}
	return result
}
