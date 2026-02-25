package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/config"
)

// orderFormState tracks the form workflow.
type orderFormState int

const (
	stateForm orderFormState = iota
	stateConfirm
	stateSubmitting
	stateResult
)

// OrderFormModel is the order placement form screen.
type OrderFormModel struct {
	tokenID    string
	question   string
	form       *huh.Form
	confirmForm *huh.Form
	spinner    spinner.Model
	state      orderFormState
	result     *clob.OrderResponse
	err        error
	clobClient *clob.Client
	width      int
	height     int

	// Form field values
	side      string
	orderType string
	price     string
	size      string
	confirmed bool
}

// NewOrderFormModel creates a new order form screen.
func NewOrderFormModel(tokenID, question string, clobCl *clob.Client, w, h int) *OrderFormModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	m := &OrderFormModel{
		tokenID:    tokenID,
		question:   question,
		spinner:    s,
		state:      stateForm,
		clobClient: clobCl,
		width:      w,
		height:     h,
		side:       "BUY",
		orderType:  "GTC",
		price:      "0.50",
		size:       "10",
	}
	m.buildForm()
	return m
}

func (m *OrderFormModel) buildForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Side").
				Options(
					huh.NewOption("YES (Buy)", "BUY"),
					huh.NewOption("NO (Sell)", "SELL"),
				).
				Value(&m.side),

			huh.NewInput().
				Title("Price (0.01 - 0.99)").
				Value(&m.price).
				Validate(validatePrice),

			huh.NewInput().
				Title("Size (shares)").
				Value(&m.size).
				Validate(validateSize),

			huh.NewSelect[string]().
				Title("Order Type").
				Options(
					huh.NewOption("GTC (Good Til Cancelled)", "GTC"),
					huh.NewOption("FOK (Fill Or Kill)", "FOK"),
					huh.NewOption("GTD (Good Til Date)", "GTD"),
				).
				Value(&m.orderType),
		),
	).WithWidth(m.width - 8).WithShowHelp(true)
}

func (m *OrderFormModel) buildConfirmForm() {
	m.confirmForm = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Submit this order?").
				Affirmative("Yes, submit").
				Negative("Cancel").
				Value(&m.confirmed),
		),
	).WithWidth(m.width - 8)
}

func (m *OrderFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *OrderFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.state == stateResult || m.state == stateConfirm {
				return m, func() tea.Msg { return popScreenMsg{} }
			}
			if m.state == stateForm {
				return m, func() tea.Msg { return popScreenMsg{} }
			}
		}

	case orderSubmittedMsg:
		m.state = stateResult
		m.result = msg.resp
		return m, nil

	case orderSubmitErrorMsg:
		m.state = stateResult
		m.err = msg.err
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.state == stateSubmitting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	switch m.state {
	case stateForm:
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form.State == huh.StateCompleted {
			m.state = stateConfirm
			m.buildConfirmForm()
			return m, m.confirmForm.Init()
		}
		return m, cmd

	case stateConfirm:
		form, cmd := m.confirmForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.confirmForm = f
		}
		if m.confirmForm.State == huh.StateCompleted {
			if m.confirmed {
				m.state = stateSubmitting
				return m, tea.Batch(m.spinner.Tick, m.submitOrder())
			}
			// User cancelled
			return m, func() tea.Msg { return popScreenMsg{} }
		}
		return m, cmd
	}

	return m, nil
}

func (m *OrderFormModel) View() string {
	var sections []string

	// Header
	header := HeaderStyle.Render("Place Order")
	question := DimStyle.Render(Truncate(m.question, 60))
	sections = append(sections, header+"\n"+question)

	switch m.state {
	case stateForm:
		sections = append(sections, m.form.View())

	case stateConfirm:
		// Order summary
		summary := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2).
			Render(
				LabelStyle.Render("Side: ") + ValueStyle.Render(m.side) + "\n" +
					LabelStyle.Render("Price: ") + ValueStyle.Render(m.price) + "\n" +
					LabelStyle.Render("Size: ") + ValueStyle.Render(m.size) + " shares\n" +
					LabelStyle.Render("Type: ") + ValueStyle.Render(m.orderType) + "\n" +
					LabelStyle.Render("Cost: ") + CyanStyle.Render(m.estimateCost()),
			)
		sections = append(sections, summary)
		sections = append(sections, m.confirmForm.View())

	case stateSubmitting:
		sections = append(sections, m.spinner.View()+" Submitting order…")

	case stateResult:
		if m.err != nil {
			sections = append(sections, ErrorBoxStyle.Render(
				ErrorStyle.Render("Order Failed")+"\n"+m.err.Error(),
			))
		} else if m.result != nil {
			status := GreenStyle.Render("SUCCESS")
			if !m.result.Success {
				status = RedStyle.Render("FAILED")
			}
			sections = append(sections, SuccessBoxStyle.Render(
				status+"\n"+
					LabelStyle.Render("Order ID: ")+ValueStyle.Render(m.result.OrderID)+"\n"+
					LabelStyle.Render("Status: ")+ValueStyle.Render(m.result.Status),
			))
		}
		sections = append(sections, "\n"+HelpStyle.Render("esc back"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return AppStyle.Render(content)
}

func (m *OrderFormModel) estimateCost() string {
	p, err1 := strconv.ParseFloat(m.price, 64)
	s, err2 := strconv.ParseFloat(m.size, 64)
	if err1 != nil || err2 != nil {
		return "—"
	}
	return fmt.Sprintf("$%.2f", p*s)
}

func (m *OrderFormModel) submitOrder() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil || !m.clobClient.IsAuthenticated() {
			return orderSubmitErrorMsg{err: fmt.Errorf("not authenticated — configure wallet first")}
		}

		// Load private key from config for EIP-712 signing.
		cfg, err := config.Load()
		if err != nil {
			return orderSubmitErrorMsg{err: fmt.Errorf("loading config: %w", err)}
		}
		pk := cfg.PrivateKey
		if pk == "" {
			return orderSubmitErrorMsg{err: fmt.Errorf("no private key configured — run setup first")}
		}
		ecdsaKey, err := crypto.HexToECDSA(strings.TrimPrefix(pk, "0x"))
		if err != nil {
			return orderSubmitErrorMsg{err: fmt.Errorf("invalid private key: %w", err)}
		}

		price, err := strconv.ParseFloat(m.price, 64)
		if err != nil {
			return orderSubmitErrorMsg{err: fmt.Errorf("invalid price: %w", err)}
		}
		size, err := strconv.ParseFloat(m.size, 64)
		if err != nil {
			return orderSubmitErrorMsg{err: fmt.Errorf("invalid size: %w", err)}
		}

		// Auto-detect neg-risk.
		negRisk := false
		if nr, nrErr := m.clobClient.GetNegRisk(m.tokenID); nrErr == nil {
			negRisk = nr.NegRisk
		}

		payload, err := clob.BuildSignedOrder(clob.OrderParams{
			TokenID:   m.tokenID,
			Side:      m.side,
			Price:     price,
			Size:      size,
			OrderType: m.orderType,
			NegRisk:   negRisk,
		}, ecdsaKey, cfg.ChainID)
		if err != nil {
			return orderSubmitErrorMsg{err: fmt.Errorf("building order: %w", err)}
		}

		resp, err := m.clobClient.PostOrder(payload)
		if err != nil {
			return orderSubmitErrorMsg{err: fmt.Errorf("submitting order: %w", err)}
		}
		return orderSubmittedMsg{resp: resp}
	}
}

func validatePrice(s string) error {
	p, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid price: must be a number")
	}
	if p < 0.01 || p > 0.99 {
		return fmt.Errorf("price must be between 0.01 and 0.99")
	}
	return nil
}

func validateSize(s string) error {
	sz, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid size: must be a number")
	}
	if sz <= 0 {
		return fmt.Errorf("size must be greater than 0")
	}
	return nil
}
