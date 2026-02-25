package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/api/data"
)

const portfolioRefreshInterval = 30 * time.Second

// PortfolioTab identifies the active tab.
type PortfolioTab int

const (
	TabPositions PortfolioTab = iota
	TabOrders
	TabTrades
)

// PortfolioModel is the portfolio dashboard screen.
type PortfolioModel struct {
	activeTab     PortfolioTab
	positions     []data.Position
	orders        []clob.Order
	trades        []clob.Trade
	balance       *clob.BalanceAllowance
	posTable      table.Model
	orderTable    table.Model
	tradeTable    table.Model
	spinner       spinner.Model
	loading       bool
	fetchesDone   int
	err           error
	address       string
	clobClient    *clob.Client
	dataClient    *data.DataClient
	helpModel        help.Model
	showFullHelp     bool
	confirmingCancel bool
	confirmOrderRow  []string
	active           bool // false when screen is not visible (suppresses refresh loops)
	width            int
	height           int
}

// NewPortfolioModel creates a new portfolio dashboard.
func NewPortfolioModel(address string, clobCl *clob.Client, dataCl *data.DataClient, w, h int) *PortfolioModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	tableH := h - 14
	if tableH < 5 {
		tableH = 5
	}

	posCols := []table.Column{
		{Title: "Market", Width: 30},
		{Title: "Side", Width: 6},
		{Title: "Size", Width: 10},
		{Title: "Avg Price", Width: 10},
		{Title: "Cur Price", Width: 10},
		{Title: "P&L", Width: 12},
	}
	pt := table.New(
		table.WithColumns(posCols),
		table.WithRows([]table.Row{}),
		table.WithHeight(tableH),
		table.WithFocused(true),
	)
	pt.SetStyles(table.DefaultStyles())

	ordCols := []table.Column{
		{Title: "ID", Width: 12},
		{Title: "Market", Width: 20},
		{Title: "Side", Width: 6},
		{Title: "Price", Width: 10},
		{Title: "Size", Width: 10},
		{Title: "Matched", Width: 10},
		{Title: "Status", Width: 10},
	}
	ot := table.New(
		table.WithColumns(ordCols),
		table.WithRows([]table.Row{}),
		table.WithHeight(tableH),
		table.WithFocused(false),
	)
	ot.SetStyles(table.DefaultStyles())

	trdCols := []table.Column{
		{Title: "ID", Width: 12},
		{Title: "Side", Width: 6},
		{Title: "Price", Width: 10},
		{Title: "Size", Width: 10},
		{Title: "Fee", Width: 8},
		{Title: "Status", Width: 10},
		{Title: "Time", Width: 18},
	}
	tt := table.New(
		table.WithColumns(trdCols),
		table.WithRows([]table.Row{}),
		table.WithHeight(tableH),
		table.WithFocused(false),
	)
	tt.SetStyles(table.DefaultStyles())

	return &PortfolioModel{
		activeTab:  TabPositions,
		posTable:   pt,
		orderTable: ot,
		tradeTable: tt,
		spinner:    s,
		loading:    true,
		active:     true,
		address:    address,
		clobClient: clobCl,
		dataClient: dataCl,
		helpModel:  newHelpModel(),
		width:      w,
		height:     h,
	}
}

func (m *PortfolioModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchPositions(),
		m.fetchOrders(),
		m.fetchTrades(),
		m.fetchBalance(),
		schedulePortfolioRefresh(),
	)
}

// SetActive controls whether the model processes refresh tick messages.
func (m *PortfolioModel) SetActive(v bool) { m.active = v }

func (m *PortfolioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Suppress refresh ticks when not the visible screen.
	if !m.active {
		switch msg.(type) {
		case refreshTickMsg:
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case positionsLoadedMsg:
		m.positions = msg.positions
		m.updatePositionsTable()
		m.checkLoading()
		return m, nil

	case positionsErrorMsg:
		if m.err == nil {
			m.err = msg.err
		}
		m.checkLoading()
		return m, nil

	case openOrdersLoadedMsg:
		m.orders = msg.orders
		m.updateOrdersTable()
		m.checkLoading()
		return m, nil

	case openOrdersErrorMsg:
		m.checkLoading()
		return m, nil

	case tradesLoadedMsg:
		m.trades = msg.trades
		m.updateTradesTable()
		m.checkLoading()
		return m, nil

	case tradesErrorMsg:
		m.checkLoading()
		return m, nil

	case balanceLoadedMsg:
		m.balance = msg.balance
		return m, nil

	case balanceErrorMsg:
		return m, nil

	case orderCancelledMsg:
		// Refresh orders after cancel
		return m, m.fetchOrders()

	case orderCancelErrorMsg:
		return m, nil

	case refreshTickMsg:
		m.fetchesDone = 0
		return m, tea.Batch(
			m.fetchPositions(),
			m.fetchOrders(),
			m.fetchTrades(),
			m.fetchBalance(),
			schedulePortfolioRefresh(),
		)

	case tea.KeyMsg:
		// Confirmation dialog takes priority over all other keys
		if m.confirmingCancel {
			switch msg.String() {
			case "y":
				m.confirmingCancel = false
				orderID := m.confirmOrderRow[0]
				m.confirmOrderRow = nil
				return m, m.cancelOrderByID(orderID)
			case "n", "esc":
				m.confirmingCancel = false
				m.confirmOrderRow = nil
				return m, nil
			}
			return m, nil // swallow all other keys during confirmation
		}

		switch msg.String() {
		case "?":
			m.showFullHelp = !m.showFullHelp
			return m, nil
		case "tab":
			m.cycleTab()
			return m, nil
		case "q", "esc":
			return m, func() tea.Msg { return popScreenMsg{} }
		case "r":
			m.loading = true
			m.fetchesDone = 0
			return m, tea.Batch(
				m.spinner.Tick,
				m.fetchPositions(),
				m.fetchOrders(),
				m.fetchTrades(),
				m.fetchBalance(),
			)
		case "c":
			if m.activeTab == TabOrders {
				row := m.orderTable.SelectedRow()
				if len(row) > 0 {
					m.confirmingCancel = true
					m.confirmOrderRow = row
				}
			}
		}

		// Forward to active table
		var cmd tea.Cmd
		switch m.activeTab {
		case TabPositions:
			m.posTable, cmd = m.posTable.Update(msg)
		case TabOrders:
			m.orderTable, cmd = m.orderTable.Update(msg)
		case TabTrades:
			m.tradeTable, cmd = m.tradeTable.Update(msg)
		}
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m *PortfolioModel) View() string {
	if m.loading && m.positions == nil && m.orders == nil {
		return AppStyle.Render(m.spinner.View() + " Loading portfolio…")
	}
	if m.err != nil && m.positions == nil {
		return AppStyle.Render(ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n" + HelpStyle.Render("r refresh │ q quit"))
	}

	var sections []string

	// Header with balance summary
	header := HeaderStyle.Render("Portfolio")
	balStr := ""
	if m.balance != nil {
		balStr = LabelStyle.Render("Balance: ") + CyanStyle.Render(m.balance.Balance) +
			"  " + LabelStyle.Render("Allowance: ") + CyanStyle.Render(m.balance.Allowance)
	}

	// P&L summary from positions
	var totalPnl float64
	for _, p := range m.positions {
		totalPnl += p.CashPnl
	}
	pnlStr := LabelStyle.Render("Total P&L: ") + FormatPnL(totalPnl)
	posCount := LabelStyle.Render(fmt.Sprintf("Positions: %d", len(m.positions)))

	sections = append(sections, header+"\n"+balStr+"  "+pnlStr+"  "+posCount)

	// Tabs
	tabs := m.renderTabs()
	sections = append(sections, tabs)

	// Active table (with empty state or confirm overlay)
	var rowInfo string
	switch m.activeTab {
	case TabPositions:
		if len(m.positions) == 0 && !m.loading {
			sections = append(sections, m.emptyState("No open positions", "Browse markets and place a trade to see positions here."))
		} else {
			sections = append(sections, m.posTable.View())
			if len(m.positions) > 0 {
				rowInfo = fmt.Sprintf("row %d/%d", m.posTable.Cursor()+1, len(m.positions))
			}
		}
	case TabOrders:
		if m.confirmingCancel {
			sections = append(sections, m.renderConfirmOverlay())
		} else if len(m.orders) == 0 && !m.loading {
			sections = append(sections, m.emptyState("No open orders", "Place an order from the order book to see it here."))
		} else {
			sections = append(sections, m.orderTable.View())
			if len(m.orders) > 0 {
				rowInfo = fmt.Sprintf("row %d/%d", m.orderTable.Cursor()+1, len(m.orders))
			}
		}
	case TabTrades:
		if len(m.trades) == 0 && !m.loading {
			sections = append(sections, m.emptyState("No trade history", "Completed trades will appear here."))
		} else {
			sections = append(sections, m.tradeTable.View())
			if len(m.trades) > 0 {
				rowInfo = fmt.Sprintf("row %d/%d", m.tradeTable.Cursor()+1, len(m.trades))
			}
		}
	}

	// Row indicator
	if rowInfo != "" {
		sections = append(sections, DimStyle.Render(rowInfo))
	}

	// Help
	keys := portfolioKeys
	keys.Cancel.SetEnabled(m.activeTab == TabOrders)
	m.helpModel.Width = m.width - 4
	m.helpModel.ShowAll = m.showFullHelp
	sections = append(sections, m.helpModel.View(keys))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return AppStyle.Render(content)
}

func (m *PortfolioModel) renderTabs() string {
	tabs := []struct {
		label string
		tab   PortfolioTab
	}{
		{"Positions", TabPositions},
		{"Orders", TabOrders},
		{"Trades", TabTrades},
	}

	var rendered []string
	for _, t := range tabs {
		if t.tab == m.activeTab {
			rendered = append(rendered, ActiveTabStyle.Render(t.label))
		} else {
			rendered = append(rendered, InactiveTabStyle.Render(t.label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, rendered...)
}

func (m *PortfolioModel) cycleTab() {
	m.posTable.Blur()
	m.orderTable.Blur()
	m.tradeTable.Blur()

	m.activeTab = (m.activeTab + 1) % 3

	switch m.activeTab {
	case TabPositions:
		m.posTable.Focus()
	case TabOrders:
		m.orderTable.Focus()
	case TabTrades:
		m.tradeTable.Focus()
	}
}

func (m *PortfolioModel) checkLoading() {
	m.fetchesDone++
	// Stop loading once all 3 fetches (positions, orders, trades) have responded.
	if m.fetchesDone >= 3 {
		m.loading = false
	}
}

func schedulePortfolioRefresh() tea.Cmd {
	return tea.Tick(portfolioRefreshInterval, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func (m *PortfolioModel) updatePositionsTable() {
	rows := make([]table.Row, len(m.positions))
	for i, p := range m.positions {
		rows[i] = table.Row{
			Truncate(p.EventTitle, 28),
			p.Outcome,
			fmt.Sprintf("%.2f", p.Size),
			fmt.Sprintf("%.4f", p.AvgPrice),
			fmt.Sprintf("%.4f", p.CurPrice),
			fmt.Sprintf("%.2f", p.CashPnl),
		}
	}
	m.posTable.SetRows(rows)
}

func (m *PortfolioModel) updateOrdersTable() {
	rows := make([]table.Row, len(m.orders))
	for i, o := range m.orders {
		rows[i] = table.Row{
			Truncate(o.ID, 10),
			Truncate(o.Market, 18),
			o.Side,
			o.Price,
			o.OriginalSize,
			o.SizeMatched,
			o.Status,
		}
	}
	m.orderTable.SetRows(rows)
}

func (m *PortfolioModel) updateTradesTable() {
	rows := make([]table.Row, len(m.trades))
	for i, t := range m.trades {
		rows[i] = table.Row{
			Truncate(t.ID, 10),
			t.Side,
			t.Price,
			t.Size,
			t.Fee,
			t.Status,
			Truncate(t.CreatedAt, 16),
		}
	}
	m.tradeTable.SetRows(rows)
}

func (m *PortfolioModel) fetchPositions() tea.Cmd {
	return func() tea.Msg {
		if m.dataClient == nil {
			return positionsErrorMsg{err: fmt.Errorf("no data client configured")}
		}
		if m.address == "" {
			return positionsErrorMsg{err: fmt.Errorf("no wallet address configured")}
		}
		positions, err := m.dataClient.GetPositions(m.address)
		if err != nil {
			return positionsErrorMsg{err: err}
		}
		return positionsLoadedMsg{positions: positions}
	}
}

func (m *PortfolioModel) fetchOrders() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil || !m.clobClient.IsAuthenticated() {
			return openOrdersErrorMsg{err: fmt.Errorf("not authenticated")}
		}
		orders, err := m.clobClient.GetOpenOrders(nil)
		if err != nil {
			return openOrdersErrorMsg{err: err}
		}
		return openOrdersLoadedMsg{orders: orders}
	}
}

func (m *PortfolioModel) fetchTrades() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil || !m.clobClient.IsAuthenticated() {
			return tradesErrorMsg{err: fmt.Errorf("not authenticated")}
		}
		trades, err := m.clobClient.GetTrades(nil)
		if err != nil {
			return tradesErrorMsg{err: err}
		}
		return tradesLoadedMsg{trades: trades}
	}
}

func (m *PortfolioModel) fetchBalance() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil || !m.clobClient.IsAuthenticated() {
			return balanceErrorMsg{err: fmt.Errorf("not authenticated")}
		}
		bal, err := m.clobClient.GetBalanceAllowance(nil)
		if err != nil {
			return balanceErrorMsg{err: err}
		}
		return balanceLoadedMsg{balance: bal}
	}
}

func (m *PortfolioModel) emptyState(title, hint string) string {
	tableH := m.height - 14
	if tableH < 5 {
		tableH = 5
	}
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(DimStyle.Render("  " + title))
	sb.WriteString("\n\n")
	sb.WriteString(DimStyle.Render("  " + hint))
	// Pad to fill the table height so layout doesn't collapse
	lines := 4 // newlines in msg
	for lines < tableH {
		sb.WriteByte('\n')
		lines++
	}
	return sb.String()
}

func (m *PortfolioModel) cancelOrderByID(orderID string) tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil || !m.clobClient.IsAuthenticated() {
			return orderCancelErrorMsg{err: fmt.Errorf("not authenticated")}
		}
		resp, err := m.clobClient.CancelOrder(orderID)
		if err != nil {
			return orderCancelErrorMsg{err: err}
		}
		return orderCancelledMsg{resp: resp}
	}
}

func (m *PortfolioModel) renderConfirmOverlay() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorYellow).Render("Cancel Order?")

	details := ""
	if len(m.confirmOrderRow) >= 5 {
		lbl := LabelStyle.Width(10)
		details = lbl.Render("Market:") + " " + ValueStyle.Render(m.confirmOrderRow[1]) + "\n" +
			lbl.Render("Side:") + " " + ValueStyle.Render(m.confirmOrderRow[2]) + "\n" +
			lbl.Render("Price:") + " " + ValueStyle.Render(m.confirmOrderRow[3]) + "\n" +
			lbl.Render("Size:") + " " + ValueStyle.Render(m.confirmOrderRow[4])
	}

	actions := CyanStyle.Render("y") + DimStyle.Render(" confirm  ") +
		CyanStyle.Render("n") + DimStyle.Render(" cancel")

	content := title + "\n\n" + details + "\n\n" + actions

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorYellow).
		Padding(1, 3).
		Render(content)
}

// ConfirmingCancel returns the confirmation dialog state (for testing).
func (m *PortfolioModel) ConfirmingCancel() bool {
	return m.confirmingCancel
}
