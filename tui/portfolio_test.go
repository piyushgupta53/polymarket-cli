package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/api/data"
)

func TestPortfolioInit(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()

	if !m.loading {
		t.Error("expected loading=true after Init")
	}
}

func TestPortfolioPositionsLoaded(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()

	positions := []data.Position{
		{EventTitle: "BTC 100k?", Outcome: "Yes", Size: 10, AvgPrice: 0.65, CurPrice: 0.70, CashPnl: 0.50},
		{EventTitle: "ETH 10k?", Outcome: "No", Size: 5, AvgPrice: 0.30, CurPrice: 0.25, CashPnl: -0.25},
	}

	updated, _ := m.Update(positionsLoadedMsg{positions: positions})
	m = updated.(*PortfolioModel)

	if m.positions == nil {
		t.Fatal("expected positions to be set")
	}
	if len(m.positions) != 2 {
		t.Errorf("positions count = %d, want 2", len(m.positions))
	}
}

func TestPortfolioOrdersLoaded(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()

	orders := []clob.Order{
		{ID: "ord-1", Side: "BUY", Price: "0.65", OriginalSize: "10", Status: "LIVE"},
	}

	updated, _ := m.Update(openOrdersLoadedMsg{orders: orders})
	m = updated.(*PortfolioModel)

	if len(m.orders) != 1 {
		t.Errorf("orders count = %d, want 1", len(m.orders))
	}
}

func TestPortfolioTabCycle(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	if m.activeTab != TabPositions {
		t.Errorf("initial tab = %d, want TabPositions", m.activeTab)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*PortfolioModel)
	if m.activeTab != TabOrders {
		t.Errorf("after first tab = %d, want TabOrders", m.activeTab)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*PortfolioModel)
	if m.activeTab != TabTrades {
		t.Errorf("after second tab = %d, want TabTrades", m.activeTab)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*PortfolioModel)
	if m.activeTab != TabPositions {
		t.Errorf("after third tab = %d, want TabPositions (wrap)", m.activeTab)
	}
}

func TestPortfolioBack(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("expected command from 'q' key")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestPortfolioError(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()

	updated, _ := m.Update(positionsErrorMsg{err: fmt.Errorf("no data client")})
	m = updated.(*PortfolioModel)

	if m.err == nil {
		t.Fatal("expected err to be set")
	}
}

func TestPortfolioResize(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = updated.(*PortfolioModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
}

func TestPortfolioHelpToggle(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	if m.showFullHelp {
		t.Error("showFullHelp should start false")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(*PortfolioModel)

	if !m.showFullHelp {
		t.Error("showFullHelp should be true after ?")
	}
}

func TestPortfolioEmptyPositions(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.positions = []data.Position{}
	m.orders = []clob.Order{}
	m.trades = []clob.Trade{}

	view := m.View()
	if !strings.Contains(view, "No open positions") {
		t.Error("expected 'No open positions' in empty positions view")
	}
}

func TestPortfolioEmptyOrders(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.positions = []data.Position{}
	m.orders = []clob.Order{}
	m.trades = []clob.Trade{}

	// Switch to orders tab
	m.Update(tea.KeyMsg{Type: tea.KeyTab})

	view := m.View()
	if !strings.Contains(view, "No open orders") {
		t.Error("expected 'No open orders' in empty orders view")
	}
}

func TestPortfolioEmptyTrades(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.positions = []data.Position{}
	m.orders = []clob.Order{}
	m.trades = []clob.Trade{}

	// Switch to trades tab (two tabs)
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m.Update(tea.KeyMsg{Type: tea.KeyTab})

	view := m.View()
	if !strings.Contains(view, "No trade history") {
		t.Error("expected 'No trade history' in empty trades view")
	}
}

func TestPortfolioConfirmCancel(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.orders = []clob.Order{
		{ID: "ord-1", Market: "Test", Side: "BUY", Price: "0.65", OriginalSize: "10", Status: "LIVE"},
	}
	m.updateOrdersTable()

	// Switch to orders tab
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*PortfolioModel)

	if m.activeTab != TabOrders {
		t.Fatalf("expected TabOrders, got %d", m.activeTab)
	}

	// Press c to start confirmation
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(*PortfolioModel)

	if !m.ConfirmingCancel() {
		t.Error("expected confirmingCancel=true after c")
	}
}

func TestPortfolioConfirmCancelY(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.orders = []clob.Order{
		{ID: "ord-1", Market: "Test", Side: "BUY", Price: "0.65", OriginalSize: "10", Status: "LIVE"},
	}
	m.updateOrdersTable()

	// Switch to orders tab
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*PortfolioModel)

	// Press c then y
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(*PortfolioModel)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(*PortfolioModel)

	if m.ConfirmingCancel() {
		t.Error("expected confirmingCancel=false after y")
	}
	if cmd == nil {
		t.Error("expected cancel command after y")
	}
}

func TestPortfolioConfirmCancelN(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.orders = []clob.Order{
		{ID: "ord-1", Market: "Test", Side: "BUY", Price: "0.65", OriginalSize: "10", Status: "LIVE"},
	}
	m.updateOrdersTable()

	// Switch to orders tab
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*PortfolioModel)

	// Press c then n
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(*PortfolioModel)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(*PortfolioModel)

	if m.ConfirmingCancel() {
		t.Error("expected confirmingCancel=false after n")
	}
	if cmd != nil {
		t.Error("expected nil command after n (no cancel)")
	}
}

func TestPortfolioConfirmOverlayInView(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.orders = []clob.Order{
		{ID: "ord-1", Market: "Test", Side: "BUY", Price: "0.65", OriginalSize: "10", Status: "LIVE"},
	}
	m.updateOrdersTable()

	// Switch to orders tab and start confirmation
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	view := m.View()
	if !strings.Contains(view, "Cancel Order?") {
		t.Error("expected 'Cancel Order?' in confirm overlay view")
	}
	if !strings.Contains(view, "confirm") {
		t.Error("expected 'confirm' action in overlay")
	}
}

func TestPortfolioRowIndicator(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.loading = false
	m.positions = []data.Position{
		{EventTitle: "BTC 100k?", Outcome: "Yes", Size: 10, AvgPrice: 0.65, CurPrice: 0.70, CashPnl: 0.50},
		{EventTitle: "ETH 10k?", Outcome: "No", Size: 5, AvgPrice: 0.30, CurPrice: 0.25, CashPnl: -0.25},
	}
	m.updatePositionsTable()

	view := m.View()
	if !strings.Contains(view, "row 1/2") {
		t.Error("expected 'row 1/2' indicator in view")
	}
}

func TestPortfolioBalanceLoaded(t *testing.T) {
	m := NewPortfolioModel("0xabc", nil, nil, 80, 40)
	m.Init()

	bal := &clob.BalanceAllowance{Balance: "1000.00", Allowance: "500.00"}
	updated, _ := m.Update(balanceLoadedMsg{balance: bal})
	m = updated.(*PortfolioModel)

	if m.balance == nil {
		t.Fatal("expected balance to be set")
	}
	if m.balance.Balance != "1000.00" {
		t.Errorf("balance = %q, want %q", m.balance.Balance, "1000.00")
	}
}
