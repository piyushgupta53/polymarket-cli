package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/config"
)

// wizardStep tracks the setup flow.
type wizardStep int

const (
	stepWelcome wizardStep = iota
	stepForm
	stepDeriving
	stepDone
)

// SetupResult holds the result of the wizard for the caller.
type SetupResult struct {
	PrivateKey    string
	SignatureType string
	Address       string
	Saved         bool
}

// derivedMsg is sent when API key derivation completes.
type derivedMsg struct {
	address string
	err     error
}

// SetupWizardModel is the interactive setup wizard screen.
type SetupWizardModel struct {
	step    wizardStep
	form    *huh.Form
	spinner spinner.Model
	result  SetupResult
	err     error
	width   int
	height  int

	// Form field values
	privateKey    string
	signatureType string

	// deriveFn is called after form completion to derive keys and save config.
	// Injected by the caller so the wizard doesn't import auth/crypto directly.
	deriveFn func(privateKey, sigType string) (address string, err error)

	// inline means the wizard was launched inside the TUI (not standalone).
	// On success, it sends authReloadMsg instead of popScreenMsg.
	inline bool
}

// NewSetupWizardModel creates a new setup wizard.
func NewSetupWizardModel(deriveFn func(string, string) (string, error), w, h int) *SetupWizardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	return &SetupWizardModel{
		step:          stepWelcome,
		spinner:       s,
		signatureType: "EOA",
		width:         w,
		height:        h,
		deriveFn:      deriveFn,
	}
}

func (m *SetupWizardModel) Init() tea.Cmd {
	return nil
}

func (m *SetupWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case stepWelcome:
			if msg.String() == "enter" {
				m.buildForm()
				m.step = stepForm
				return m, m.form.Init()
			}
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg { return popScreenMsg{} }
			}
		case stepForm:
			if msg.String() == "esc" {
				return m, func() tea.Msg { return popScreenMsg{} }
			}
		case stepDone:
			if msg.String() == "enter" || msg.String() == "q" || msg.String() == "esc" {
				if m.inline && m.result.Saved {
					return m, func() tea.Msg { return authReloadMsg{} }
				}
				return m, func() tea.Msg { return popScreenMsg{} }
			}
		}

	case derivedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepDone
			return m, nil
		}
		m.result.Address = msg.address
		m.result.Saved = true
		m.step = stepDone
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.step == stepDeriving {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.step == stepForm {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form.State == huh.StateCompleted {
			m.result.PrivateKey = m.privateKey
			m.result.SignatureType = m.signatureType
			m.step = stepDeriving
			return m, tea.Batch(m.spinner.Tick, m.derive())
		}
		return m, cmd
	}

	return m, nil
}

func (m *SetupWizardModel) View() string {
	switch m.step {
	case stepWelcome:
		return m.viewWelcome()
	case stepForm:
		return m.viewForm()
	case stepDeriving:
		return AppStyle.Render(m.spinner.View() + " Validating key and deriving API credentials…")
	case stepDone:
		return m.viewDone()
	}
	return ""
}

func (m *SetupWizardModel) viewWelcome() string {
	title := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Bold(true).
		MarginBottom(1).
		Render("Welcome to Polymarket CLI")

	body := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyan).
		Padding(1, 3).
		Width(min(60, m.width-8)).
		Render(
			title + "\n\n" +
				ValueStyle.Render("This wizard will help you set up your wallet for trading.") + "\n\n" +
				LabelStyle.Render("You'll need:") + "\n" +
				ValueStyle.Render("  • Your Ethereum private key (hex format)") + "\n" +
				ValueStyle.Render("  • Your signature type (EOA for most users)") + "\n\n" +
				DimStyle.Render("Your key is stored locally at:") + "\n" +
				DimStyle.Render("  "+config.DefaultConfigPath()) + "\n\n" +
				HelpStyle.Render("Press Enter to continue, esc to go back"),
		)

	return AppStyle.Render(body)
}

func (m *SetupWizardModel) viewForm() string {
	header := HeaderStyle.Render("Setup — Import Wallet")
	return AppStyle.Render(header + "\n\n" + m.form.View() + "\n" + HelpStyle.Render("esc to cancel"))
}

func (m *SetupWizardModel) viewDone() string {
	if m.err != nil {
		return AppStyle.Render(
			ErrorBoxStyle.Render(
				ErrorStyle.Render("Setup Failed")+"\n\n"+m.err.Error(),
			) + "\n\n" + HelpStyle.Render("Press Enter to exit"),
		)
	}

	content := SuccessBoxStyle.Render(
		GreenStyle.Render("Setup Complete!") + "\n\n" +
			LabelStyle.Render("Address: ") + ValueStyle.Render(m.result.Address) + "\n" +
			LabelStyle.Render("Signature Type: ") + ValueStyle.Render(m.result.SignatureType) + "\n" +
			LabelStyle.Render("Config: ") + DimStyle.Render(config.DefaultConfigPath()) + "\n\n" +
			DimStyle.Render("Next steps:") + "\n" +
			ValueStyle.Render("  polymarket tui        — Browse markets") + "\n" +
			ValueStyle.Render("  polymarket markets list — List markets") + "\n" +
			ValueStyle.Render("  polymarket wallet show  — View wallet"),
	)

	return AppStyle.Render(content + "\n\n" + HelpStyle.Render("Press Enter to exit"))
}

func (m *SetupWizardModel) buildForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Private Key").
				Description("Hex-encoded Ethereum private key (with or without 0x prefix)").
				EchoMode(huh.EchoModePassword).
				Value(&m.privateKey).
				Validate(func(s string) error {
					s = strings.TrimPrefix(s, "0x")
					if len(s) != 64 {
						return fmt.Errorf("private key must be 64 hex characters (got %d)", len(s))
					}
					for _, c := range s {
						if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
							return fmt.Errorf("invalid hex character: %c", c)
						}
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Signature Type").
				Description("How your wallet signs transactions").
				Options(
					huh.NewOption("EOA (Standard wallet)", "EOA"),
					huh.NewOption("Polymarket Proxy", "POLY_PROXY"),
					huh.NewOption("Gnosis Safe", "POLY_GNOSIS_SAFE"),
				).
				Value(&m.signatureType),
		),
	).WithWidth(min(60, m.width-8)).WithShowHelp(true)
}

func (m *SetupWizardModel) derive() tea.Cmd {
	return func() tea.Msg {
		if m.deriveFn == nil {
			return derivedMsg{err: fmt.Errorf("derive function not configured")}
		}
		address, err := m.deriveFn(m.privateKey, m.signatureType)
		if err != nil {
			return derivedMsg{err: err}
		}
		return derivedMsg{address: address}
	}
}

// GetResult returns the setup result (for testing).
func (m *SetupWizardModel) GetResult() SetupResult {
	return m.result
}

// GetStep returns the current step (for testing).
func (m *SetupWizardModel) GetStep() wizardStep {
	return m.step
}
