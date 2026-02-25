package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
)

func TestOrderBookInit(t *testing.T) {
	m := NewOrderBookModel("tok1", "Will BTC hit 100k?", nil, 80, 40)
	m.Init()

	if !m.loading {
		t.Error("expected loading=true after Init")
	}
}

func TestOrderBookLoaded(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	book := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{
			{Price: "0.65", Size: "100"},
			{Price: "0.64", Size: "200"},
		},
		Asks: []clob.OrderBookEntry{
			{Price: "0.66", Size: "150"},
			{Price: "0.67", Size: "50"},
		},
	}

	updated, _ := m.Update(orderBookLoadedMsg{book: book})
	m = updated.(*OrderBookModel)

	if m.loading {
		t.Error("expected loading=false after orderBookLoadedMsg")
	}
	if m.orderBook == nil {
		t.Fatal("expected orderBook to be set")
	}
	if len(m.orderBook.Bids) != 2 {
		t.Errorf("bids count = %d, want 2", len(m.orderBook.Bids))
	}
}

func TestOrderBookError(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	updated, _ := m.Update(orderBookErrorMsg{err: fmt.Errorf("connection failed")})
	m = updated.(*OrderBookModel)

	if m.loading {
		t.Error("expected loading=false after error")
	}
	if m.err == nil {
		t.Fatal("expected err to be set")
	}
}

func TestOrderBookRefresh(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()
	m.loading = false

	newBook := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.70", Size: "500"}},
		Asks: []clob.OrderBookEntry{},
	}

	updated, _ := m.Update(orderBookRefreshMsg{book: newBook})
	m = updated.(*OrderBookModel)

	if m.orderBook == nil || len(m.orderBook.Bids) != 1 {
		t.Error("expected refreshed order book")
	}
	if m.orderBook.Bids[0].Price != "0.70" {
		t.Errorf("bid price = %q, want 0.70", m.orderBook.Bids[0].Price)
	}
}

func TestOrderBookBack(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()
	m.loading = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if cmd == nil {
		t.Fatal("expected command from 'b' key")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestOrderBookOpenOrderForm(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()
	m.loading = false
	m.orderBook = &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "100"}},
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	if cmd == nil {
		t.Fatal("expected command from 'o' key")
	}
	msg := cmd()
	pushMsg, ok := msg.(pushScreenMsg)
	if !ok {
		t.Fatalf("expected pushScreenMsg, got %T", msg)
	}
	if pushMsg.screen != ScreenOrderForm {
		t.Errorf("screen = %d, want ScreenOrderForm", pushMsg.screen)
	}
}

func TestOrderBookResize(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = updated.(*OrderBookModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
}

func TestOrderBookHelpToggle(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()
	m.loading = false

	if m.showFullHelp {
		t.Error("showFullHelp should start false")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(*OrderBookModel)

	if !m.showFullHelp {
		t.Error("showFullHelp should be true after ?")
	}
}

func TestOrderBookNoFlashOnFirstLoad(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	book := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "100"}},
		Asks: []clob.OrderBookEntry{{Price: "0.66", Size: "150"}},
	}

	updated, _ := m.Update(orderBookLoadedMsg{book: book})
	m = updated.(*OrderBookModel)

	if m.Flashing() {
		t.Error("should not flash on first load")
	}
	if m.BidFlashCount() != 0 || m.AskFlashCount() != 0 {
		t.Error("no flash entries expected on first load")
	}
}

func TestOrderBookFlashOnPriceChange(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	// First load
	book1 := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "100"}},
		Asks: []clob.OrderBookEntry{{Price: "0.66", Size: "150"}},
	}
	updated, _ := m.Update(orderBookLoadedMsg{book: book1})
	m = updated.(*OrderBookModel)

	// Refresh with changed prices
	book2 := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "200"}}, // size changed
		Asks: []clob.OrderBookEntry{{Price: "0.67", Size: "150"}}, // price changed
	}
	updated, _ = m.Update(orderBookRefreshMsg{book: book2})
	m = updated.(*OrderBookModel)

	if !m.Flashing() {
		t.Error("should be flashing after price change")
	}
	if m.BidFlashCount() == 0 {
		t.Error("expected bid flash entries for changed size")
	}
	if m.AskFlashCount() == 0 {
		t.Error("expected ask flash entries for changed price")
	}
}

func TestOrderBookNoFlashOnSamePrices(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	book := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "100"}},
		Asks: []clob.OrderBookEntry{{Price: "0.66", Size: "150"}},
	}

	// First load
	updated, _ := m.Update(orderBookLoadedMsg{book: book})
	m = updated.(*OrderBookModel)

	// Refresh with same data
	updated, _ = m.Update(orderBookRefreshMsg{book: book})
	m = updated.(*OrderBookModel)

	if m.Flashing() {
		t.Error("should not flash when prices unchanged")
	}
}

func TestOrderBookFlashDecay(t *testing.T) {
	m := NewOrderBookModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	// First load
	book1 := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "100"}},
		Asks: []clob.OrderBookEntry{},
	}
	updated, _ := m.Update(orderBookLoadedMsg{book: book1})
	m = updated.(*OrderBookModel)

	// Refresh with change
	book2 := &clob.OrderBook{
		Bids: []clob.OrderBookEntry{{Price: "0.65", Size: "200"}},
		Asks: []clob.OrderBookEntry{},
	}
	updated, _ = m.Update(orderBookRefreshMsg{book: book2})
	m = updated.(*OrderBookModel)

	if !m.Flashing() {
		t.Fatal("should be flashing")
	}

	// Run flash ticks until converged
	for i := 0; i < 300; i++ {
		if !m.Flashing() {
			break
		}
		updated, _ = m.Update(flashTickMsg(time.Now()))
		m = updated.(*OrderBookModel)
	}

	if m.Flashing() {
		t.Error("flash should have converged")
	}
	if m.BidFlashCount() != 0 {
		t.Errorf("bid flash entries should be empty after convergence, got %d", m.BidFlashCount())
	}
}
