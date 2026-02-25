package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piyushgupta/polymarket-cli/internal/api"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
)

func TestMarketDetailInit(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()

	if !m.loading {
		t.Error("expected loading=true after Init")
	}
}

func TestMarketDetailLoaded(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()

	market := &api.Market{
		ID:            "123",
		Question:      "Will BTC hit 100k?",
		Description:   "A test market description.",
		OutcomePrices: `["0.65","0.35"]`,
		Active:        true,
		Tokens: []api.Token{
			{TokenID: "tok1", Outcome: "Yes", Price: 0.65},
		},
	}

	updated, _ := m.Update(marketDetailLoadedMsg{market: market})
	m = updated.(*MarketDetailModel)

	if m.loading {
		t.Error("expected loading=false after marketDetailLoadedMsg")
	}
	if m.market == nil {
		t.Fatal("expected market to be set")
	}
	if m.market.ID != "123" {
		t.Errorf("market.ID = %q, want %q", m.market.ID, "123")
	}
	if m.probTarget != 0.65 {
		t.Errorf("probTarget = %f, want 0.65", m.probTarget)
	}
}

func TestMarketDetailOrderBookLoaded(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	book := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{
			{Price: "0.65", Size: "100"},
			{Price: "0.64", Size: "200"},
		},
		Asks: []clob.OrderBookEntry{
			{Price: "0.66", Size: "150"},
		},
	}

	updated, _ := m.Update(orderBookLoadedMsg{book: book})
	m = updated.(*MarketDetailModel)

	if m.orderBook == nil {
		t.Fatal("expected orderBook to be set")
	}
	if len(m.orderBook.Bids) != 2 {
		t.Errorf("bids count = %d, want 2", len(m.orderBook.Bids))
	}
	if len(m.orderBook.Asks) != 1 {
		t.Errorf("asks count = %d, want 1", len(m.orderBook.Asks))
	}
}

func TestMarketDetailBack(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if cmd == nil {
		t.Fatal("expected command from 'b' key, got nil")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestMarketDetailError(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()

	testErr := fmt.Errorf("market not found")
	updated, _ := m.Update(marketDetailErrorMsg{err: testErr})
	m = updated.(*MarketDetailModel)

	if m.loading {
		t.Error("expected loading=false after marketDetailErrorMsg")
	}
	if m.err == nil {
		t.Fatal("expected err to be set")
	}
	if m.err.Error() != "market not found" {
		t.Errorf("err = %q, want %q", m.err.Error(), "market not found")
	}
}

func TestMarketDetailHelpToggle(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	if m.showFullHelp {
		t.Error("showFullHelp should start false")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(*MarketDetailModel)

	if !m.showFullHelp {
		t.Error("showFullHelp should be true after ?")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(*MarketDetailModel)

	if m.showFullHelp {
		t.Error("showFullHelp should be false after second ?")
	}
}

func TestMarketDetailResize(t *testing.T) {
	m := NewMarketDetailModel("123", nil, nil, 80, 40)
	m.Init()
	m.loading = false

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = updated.(*MarketDetailModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}
