package tui

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OnboardingModel presents a welcome screen with two choices:
// "Just Explore" (default) and "Set Up Wallet".
// Includes an animated gradient sweep (cyan ↔ purple).
type OnboardingModel struct {
	choice   int // 0 = Just Explore, 1 = Set Up Wallet
	deriveFn func(string, string) (string, error)
	width    int
	height   int
	phase    float64
	active   bool // false when screen is not visible (suppresses wave ticks)
}

// NewOnboardingModel creates a new onboarding screen.
func NewOnboardingModel(deriveFn func(string, string) (string, error), w, h int) *OnboardingModel {
	return &OnboardingModel{
		choice:   0,
		deriveFn: deriveFn,
		width:    w,
		height:   h,
		active:   true,
	}
}

func (m *OnboardingModel) Init() tea.Cmd {
	return waveTick()
}

// SetActive controls whether the model processes wave tick messages.
func (m *OnboardingModel) SetActive(v bool) { m.active = v }

func (m *OnboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case waveTickMsg:
		if !m.active {
			return m, nil
		}
		m.phase += 0.10
		return m, waveTick()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.choice > 0 {
				m.choice--
			}
		case "down", "j":
			if m.choice < 1 {
				m.choice++
			}
		case "enter":
			if m.choice == 0 {
				return m, func() tea.Msg {
					return replaceScreenMsg{screen: ScreenMarketsList}
				}
			}
			return m, func() tea.Msg {
				return pushScreenMsg{screen: ScreenSetupWizard, deriveFn: m.deriveFn}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *OnboardingModel) View() string {
	cardWidth := min(60, m.width-8)
	if cardWidth < 20 {
		cardWidth = 20
	}

	// Gradient bar inside the card
	gradientWidth := cardWidth - 8 // padding (3 each side) + border (1 each side)
	gradient := renderGradient(gradientWidth, m.phase)

	title := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Bold(true).
		MarginBottom(1).
		Render("Welcome to Polymarket")

	body := ValueStyle.Render("Browse live prediction markets, view order books,") + "\n" +
		ValueStyle.Render("and track prices — all from your terminal.") + "\n\n" +
		DimStyle.Render("To trade or view your portfolio, you'll need to") + "\n" +
		DimStyle.Render("connect your wallet.") + "\n\n"

	// Render choices
	choices := [2]string{"Just Explore", "Set Up Wallet"}
	for i, label := range choices {
		cursor := "  "
		style := DimStyle
		if i == m.choice {
			cursor = "▸ "
			style = lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
		}
		body += style.Render(cursor+label) + "\n"
	}

	help := "\n" + HelpStyle.Render("↑/↓ navigate  enter select")

	content := gradient + "\n\n" + title + "\n\n" + body + help

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(1, 3).
		Width(cardWidth).
		Render(content)

	// Center vertically and horizontally
	cardHeight := strings.Count(card, "\n") + 1
	hPad := (m.width - lipgloss.Width(card)) / 2
	vPad := (m.height - cardHeight) / 2
	if hPad < 0 {
		hPad = 0
	}
	if vPad < 0 {
		vPad = 0
	}

	return strings.Repeat("\n", vPad) +
		strings.Repeat(" ", hPad) +
		strings.ReplaceAll(card, "\n", "\n"+strings.Repeat(" ", hPad))
}

// renderGradient renders a 3-row animated gradient bar that sweeps
// cyan ↔ purple. The top and bottom rows use half-blocks (▄/▀) for
// soft edges; the center row is a solid bright band.
func renderGradient(width int, phase float64) string {
	if width < 4 {
		return ""
	}

	// Row specs: character and intensity multiplier.
	// ▄ = bottom-half lit (soft top edge), █ = full (bright center), ▀ = top-half lit (soft bottom edge)
	type rowSpec struct {
		ch        string
		intensity float64
	}
	rows := []rowSpec{
		{"▄", 0.35},
		{"█", 1.00},
		{"▀", 0.35},
	}

	var lines []string
	for _, rs := range rows {
		var sb strings.Builder
		for col := 0; col < width; col++ {
			// Scroll position: one full color cycle across the bar width
			t := math.Mod(float64(col)/float64(width)-phase*0.15, 1.0)
			if t < 0 {
				t += 1.0
			}
			base := sweepColor(t)
			color := blendColor(base, lipgloss.Color("#000000"), rs.intensity)
			sb.WriteString(lipgloss.NewStyle().Foreground(color).Render(rs.ch))
		}
		lines = append(lines, sb.String())
	}

	return strings.Join(lines, "\n")
}

// sweepColor returns a color for position t (0–1) in the gradient cycle.
// Smoothly interpolates: cyan → purple → cyan using cosine easing.
func sweepColor(t float64) lipgloss.Color {
	blend := (1.0 + math.Cos(t*2*math.Pi)) / 2.0
	return blendColor(ColorCyan, ColorPurple, blend)
}

func waveTick() tea.Cmd {
	return tea.Tick(time.Second/20, func(t time.Time) tea.Msg {
		return waveTickMsg(t)
	})
}

// Choice returns the current choice index (for testing).
func (m *OnboardingModel) Choice() int {
	return m.choice
}

// Phase returns the current wave phase (for testing).
func (m *OnboardingModel) Phase() float64 {
	return m.phase
}
