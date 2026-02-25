package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piyushgupta/polymarket-cli/internal/api"
)

func TestMarketsListInit(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	if !m.loading {
		t.Error("expected loading=true after Init")
	}
}

func TestMarketsListLoaded(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	markets := []api.Market{
		{ID: "1", Question: "Will BTC hit 100k?", Active: true, Volume: "1000000"},
		{ID: "2", Question: "Will ETH hit 10k?", Active: true, Volume: "500000"},
	}

	updated, _ := m.Update(marketsLoadedMsg{markets: markets})
	m = updated.(*MarketsListModel)

	if m.loading {
		t.Error("expected loading=false after marketsLoadedMsg")
	}
	if m.err != nil {
		t.Errorf("expected err=nil, got %v", m.err)
	}
}

func TestMarketsListError(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	testErr := fmt.Errorf("API connection failed")
	updated, _ := m.Update(marketsErrorMsg{err: testErr})
	m = updated.(*MarketsListModel)

	if m.loading {
		t.Error("expected loading=false after marketsErrorMsg")
	}
	if m.err == nil {
		t.Fatal("expected err to be set")
	}
	if m.err.Error() != "API connection failed" {
		t.Errorf("err = %q, want %q", m.err.Error(), "API connection failed")
	}
}

func TestMarketsListSelectItem(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	// Load markets
	markets := []api.Market{
		{ID: "42", Question: "Test market?", Active: true},
	}
	updated, _ := m.Update(marketsLoadedMsg{markets: markets})
	m = updated.(*MarketsListModel)

	// Press enter
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = updated

	if cmd == nil {
		t.Fatal("expected command from enter key, got nil")
	}
	msg := cmd()
	pushMsg, ok := msg.(pushScreenMsg)
	if !ok {
		t.Fatalf("expected pushScreenMsg, got %T", msg)
	}
	if pushMsg.marketID != "42" {
		t.Errorf("marketID = %q, want %q", pushMsg.marketID, "42")
	}
	if pushMsg.screen != ScreenMarketDetail {
		t.Errorf("screen = %d, want ScreenMarketDetail", pushMsg.screen)
	}
}

func TestMarketsListHelpToggle(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()
	m.loading = false

	// Load some markets so list is active
	markets := []api.Market{{ID: "1", Question: "Test?", Active: true}}
	m.Update(marketsLoadedMsg{markets: markets})

	if m.list.Help.ShowAll {
		t.Error("Help.ShowAll should start false")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(*MarketsListModel)

	if !m.list.Help.ShowAll {
		t.Error("Help.ShowAll should be true after ?")
	}
}

func TestMarketsListResize(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = updated.(*MarketsListModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestMarketItemProbability(t *testing.T) {
	tests := []struct {
		name string
		item marketItem
		want float64
	}{
		{
			name: "from OutcomePrices",
			item: marketItem{market: api.Market{OutcomePrices: `["0.65","0.35"]`}},
			want: 0.65,
		},
		{
			name: "from BestBid fallback",
			item: marketItem{market: api.Market{BestBid: float64Ptr(0.42)}},
			want: 0.42,
		},
		{
			name: "no data returns 0",
			item: marketItem{market: api.Market{}},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.probability()
			if got != tt.want {
				t.Errorf("probability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderProbBar(t *testing.T) {
	bar := RenderProbBar(0.75, 20)
	if bar == "" {
		t.Fatal("expected non-empty prob bar")
	}
	if !strings.Contains(bar, "75%") {
		t.Errorf("expected bar to contain '75%%', got %q", bar)
	}

	bar50 := RenderProbBar(0.5, 20)
	if !strings.Contains(bar50, "50%") {
		t.Errorf("expected bar to contain '50%%', got %q", bar50)
	}

	barZero := RenderProbBar(0, 10)
	if !strings.Contains(barZero, " 0%") {
		t.Errorf("expected bar to contain ' 0%%', got %q", barZero)
	}
}

func TestMarketDelegateHeight(t *testing.T) {
	d := marketDelegate{}
	if d.Height() != 3 {
		t.Errorf("Height() = %d, want 3", d.Height())
	}
	if d.Spacing() != 1 {
		t.Errorf("Spacing() = %d, want 1", d.Spacing())
	}
}

func TestEffectiveStatus(t *testing.T) {
	// Mock time to 2026-02-25
	origNow := timeNow
	timeNow = func() time.Time { return time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origNow }()

	tests := []struct {
		name       string
		item       marketItem
		wantActive bool
		wantClosed bool
	}{
		{
			name:       "active market with future end date",
			item:       marketItem{market: api.Market{Active: true, EndDateISO: "2026-06-01T00:00:00Z"}},
			wantActive: true,
			wantClosed: false,
		},
		{
			name:       "active market with past end date shown as closed",
			item:       marketItem{market: api.Market{Active: true, EndDateISO: "2025-12-31T00:00:00Z"}},
			wantActive: false,
			wantClosed: true,
		},
		{
			name:       "already closed market stays closed",
			item:       marketItem{market: api.Market{Active: true, Closed: true, EndDateISO: "2026-06-01T00:00:00Z"}},
			wantActive: true,
			wantClosed: true,
		},
		{
			name:       "no end date preserves API status",
			item:       marketItem{market: api.Market{Active: true}},
			wantActive: true,
			wantClosed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			active, closed := tt.item.effectiveStatus()
			if active != tt.wantActive {
				t.Errorf("active = %v, want %v", active, tt.wantActive)
			}
			if closed != tt.wantClosed {
				t.Errorf("closed = %v, want %v", closed, tt.wantClosed)
			}
		})
	}
}

func TestEndDateString(t *testing.T) {
	tests := []struct {
		name string
		iso  string
		want string
	}{
		{"RFC3339", "2026-03-15T00:00:00Z", "Mar 15, 2026"},
		{"date only", "2026-03-15", "Mar 15, 2026"},
		{"empty", "", ""},
		{"invalid", "not-a-date", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := marketItem{market: api.Market{EndDateISO: tt.iso}}
			got := mi.endDateString()
			if got != tt.want {
				t.Errorf("endDateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func float64Ptr(v float64) *float64 { return &v }

// --- Category tabs tests ---

func TestMarketsListEventsLoaded(t *testing.T) {
	origNow := timeNow
	timeNow = func() time.Time { return time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origNow }()

	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	events := []api.Event{
		{
			ID:    "e1",
			Title: "BTC Price Event",
			Tags:  []api.Tag{{ID: "1", Label: "Crypto", Slug: "crypto"}},
			Markets: []api.Market{
				{ID: "m1", Question: "Will BTC hit 100k?", Active: true, EndDateISO: "2026-12-31T00:00:00Z"},
				{ID: "m2", Question: "Will BTC hit 150k?", Active: true, EndDateISO: "2026-12-31T00:00:00Z"},
			},
		},
		{
			ID:    "e2",
			Title: "Election Event",
			Tags:  []api.Tag{{ID: "2", Label: "Politics", Slug: "politics"}},
			Markets: []api.Market{
				{ID: "m3", Question: "Who wins 2028?", Active: true, EndDateISO: "2028-11-05T00:00:00Z"},
			},
		},
		{
			ID:    "e3",
			Title: "Multi-tag Event",
			Tags: []api.Tag{
				{ID: "1", Label: "Crypto", Slug: "crypto"},
				{ID: "3", Label: "Finance", Slug: "finance"},
			},
			Markets: []api.Market{
				{ID: "m1", Question: "Will BTC hit 100k?", Active: true, EndDateISO: "2026-12-31T00:00:00Z"}, // duplicate
				{ID: "m4", Question: "DeFi market cap?", Active: true, EndDateISO: "2027-01-01T00:00:00Z"},
			},
		},
	}

	updated, _ := m.Update(eventsLoadedMsg{events: events})
	m = updated.(*MarketsListModel)

	if m.loading {
		t.Error("expected loading=false after eventsLoadedMsg")
	}

	// All markets should be deduped: m1, m2, m3, m4
	if len(m.allMarkets) != 4 {
		t.Errorf("expected 4 allMarkets, got %d", len(m.allMarkets))
	}

	// Crypto tag should have m1, m2, m4
	cryptoMarkets := m.tagIndex["crypto"]
	if len(cryptoMarkets) != 3 {
		t.Errorf("expected 3 crypto markets, got %d", len(cryptoMarkets))
	}

	// Politics tag should have m3
	politicsMarkets := m.tagIndex["politics"]
	if len(politicsMarkets) != 1 {
		t.Errorf("expected 1 politics market, got %d", len(politicsMarkets))
	}

	// Finance tag should have m1, m4
	financeMarkets := m.tagIndex["finance"]
	if len(financeMarkets) != 2 {
		t.Errorf("expected 2 finance markets, got %d", len(financeMarkets))
	}

	// Categories should be derived from topLevelCategories that have markets
	// Order is: Politics, Crypto, Finance (matching topLevelCategories order)
	if len(m.categories) != 3 {
		t.Fatalf("expected 3 categories, got %d: %+v", len(m.categories), m.categories)
	}
	if m.categories[0].Slug != "politics" {
		t.Errorf("categories[0].Slug = %q, want %q", m.categories[0].Slug, "politics")
	}
	if m.categories[1].Slug != "crypto" {
		t.Errorf("categories[1].Slug = %q, want %q", m.categories[1].Slug, "crypto")
	}
	if m.categories[2].Slug != "finance" {
		t.Errorf("categories[2].Slug = %q, want %q", m.categories[2].Slug, "finance")
	}
}

func TestMarketsListCategorySwitch(t *testing.T) {
	origNow := timeNow
	timeNow = func() time.Time { return time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origNow }()

	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	// Load events — categories derived automatically
	events := []api.Event{
		{
			ID:   "e1",
			Tags: []api.Tag{{ID: "1", Label: "Crypto", Slug: "crypto"}},
			Markets: []api.Market{
				{ID: "m1", Question: "BTC 100k?", Active: true, EndDateISO: "2027-01-01T00:00:00Z"},
			},
		},
		{
			ID:   "e2",
			Tags: []api.Tag{{ID: "2", Label: "Politics", Slug: "politics"}},
			Markets: []api.Market{
				{ID: "m2", Question: "Election?", Active: true, EndDateISO: "2028-11-05T00:00:00Z"},
				{ID: "m3", Question: "Senate?", Active: true, EndDateISO: "2028-11-05T00:00:00Z"},
			},
		},
	}
	updated, _ := m.Update(eventsLoadedMsg{events: events})
	m = updated.(*MarketsListModel)

	// Should have 2 categories: Politics, Crypto (in topLevelCategories order)
	if len(m.categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(m.categories))
	}

	// Start on "All" tab
	if m.activeTab != 0 {
		t.Errorf("expected activeTab=0, got %d", m.activeTab)
	}

	// Press ] to move to first category (Politics)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updated.(*MarketsListModel)
	if m.activeTab != 1 {
		t.Errorf("expected activeTab=1 after ], got %d", m.activeTab)
	}

	// Press ] again to move to second category (Crypto)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updated.(*MarketsListModel)
	if m.activeTab != 2 {
		t.Errorf("expected activeTab=2 after ]], got %d", m.activeTab)
	}

	// Press ] at the end — should not go past last tab
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updated.(*MarketsListModel)
	if m.activeTab != 2 {
		t.Errorf("expected activeTab=2 (clamped), got %d", m.activeTab)
	}

	// Press [ to go back
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updated.(*MarketsListModel)
	if m.activeTab != 1 {
		t.Errorf("expected activeTab=1 after [, got %d", m.activeTab)
	}

	// Press [ to go back to "All"
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updated.(*MarketsListModel)
	if m.activeTab != 0 {
		t.Errorf("expected activeTab=0 after [[, got %d", m.activeTab)
	}

	// Press [ at the start — should not go below 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updated.(*MarketsListModel)
	if m.activeTab != 0 {
		t.Errorf("expected activeTab=0 (clamped), got %d", m.activeTab)
	}
}

func TestMarketsListCategoryTabsInView(t *testing.T) {
	origNow := timeNow
	timeNow = func() time.Time { return time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origNow }()

	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	// Categories are derived from events, so load events with known tags
	events := []api.Event{
		{
			ID:   "e1",
			Tags: []api.Tag{{ID: "1", Label: "Politics", Slug: "politics"}},
			Markets: []api.Market{
				{ID: "m1", Question: "Test?", Active: true, EndDateISO: "2027-01-01T00:00:00Z"},
			},
		},
		{
			ID:   "e2",
			Tags: []api.Tag{{ID: "2", Label: "Sports", Slug: "sports"}},
			Markets: []api.Market{
				{ID: "m2", Question: "Game?", Active: true, EndDateISO: "2027-01-01T00:00:00Z"},
			},
		},
	}
	updated, _ := m.Update(eventsLoadedMsg{events: events})
	m = updated.(*MarketsListModel)

	view := m.View()

	if !strings.Contains(view, "All") {
		t.Error("expected View to contain 'All' tab")
	}
	if !strings.Contains(view, "Politics") {
		t.Error("expected View to contain 'Politics' tab")
	}
	if !strings.Contains(view, "Sports") {
		t.Error("expected View to contain 'Sports' tab")
	}
	if !strings.Contains(view, "─") {
		t.Error("expected View to contain tab separator line")
	}
}

func TestMarketsListOnlyTopLevelCategories(t *testing.T) {
	origNow := timeNow
	timeNow = func() time.Time { return time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origNow }()

	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	// Events with both top-level ("crypto") and niche ("detroit-pistons") tags
	events := []api.Event{
		{
			ID: "e1",
			Tags: []api.Tag{
				{ID: "1", Label: "Crypto", Slug: "crypto"},
				{ID: "99", Label: "detroit pistons", Slug: "detroit-pistons"},
			},
			Markets: []api.Market{
				{ID: "m1", Question: "BTC?", Active: true, EndDateISO: "2027-01-01T00:00:00Z"},
			},
		},
	}
	updated, _ := m.Update(eventsLoadedMsg{events: events})
	m = updated.(*MarketsListModel)

	// Only "Crypto" should appear as a category tab (not "detroit-pistons")
	if len(m.categories) != 1 {
		t.Fatalf("expected 1 category, got %d: %+v", len(m.categories), m.categories)
	}
	if m.categories[0].Slug != "crypto" {
		t.Errorf("categories[0].Slug = %q, want %q", m.categories[0].Slug, "crypto")
	}

	// But the niche tag should still be in the tagIndex (just not shown as a tab)
	if len(m.tagIndex["detroit-pistons"]) != 1 {
		t.Errorf("expected detroit-pistons in tagIndex, got %d", len(m.tagIndex["detroit-pistons"]))
	}
}

func TestMarketsListEventsFallback(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	// Send events with no embedded markets → should trigger fallback
	events := []api.Event{
		{
			ID:    "e1",
			Title: "Event with no markets",
			Tags:  []api.Tag{{ID: "1", Label: "Crypto", Slug: "crypto"}},
		},
	}
	updated, cmd := m.Update(eventsLoadedMsg{events: events})
	m = updated.(*MarketsListModel)

	if !m.loading {
		t.Error("expected loading=true after empty events (fallback)")
	}
	if cmd == nil {
		t.Fatal("expected fallback command, got nil")
	}
}

func TestMarketsListEventsError(t *testing.T) {
	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	updated, cmd := m.Update(eventsErrorMsg{err: fmt.Errorf("events API failed")})
	m = updated.(*MarketsListModel)

	if m.err != nil {
		t.Errorf("expected err=nil on events error (fallback), got %v", m.err)
	}
	if cmd == nil {
		t.Fatal("expected fallback command on events error, got nil")
	}
}

func TestMarketsListEventsFilterExpired(t *testing.T) {
	origNow := timeNow
	timeNow = func() time.Time { return time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origNow }()

	m := NewMarketsListModel(nil, 80, 40)
	m.Init()

	events := []api.Event{
		{
			ID:   "e1",
			Tags: []api.Tag{{ID: "1", Label: "Crypto", Slug: "crypto"}},
			Markets: []api.Market{
				{ID: "m1", Question: "Future market", Active: true, EndDateISO: "2027-01-01T00:00:00Z"},
				{ID: "m2", Question: "Expired market", Active: true, EndDateISO: "2025-01-01T00:00:00Z"},
			},
		},
	}

	updated, _ := m.Update(eventsLoadedMsg{events: events})
	m = updated.(*MarketsListModel)

	if len(m.allMarkets) != 1 {
		t.Errorf("expected 1 market (expired filtered), got %d", len(m.allMarkets))
	}
	if m.allMarkets[0].ID != "m1" {
		t.Errorf("expected market m1, got %s", m.allMarkets[0].ID)
	}
}
