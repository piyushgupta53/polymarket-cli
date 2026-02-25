package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/api"
)

// timeNow is a function variable for mocking in tests.
var timeNow = time.Now

// marketItem wraps a Market for the list.Model.
type marketItem struct {
	market api.Market
}

func (i marketItem) Title() string       { return i.market.Question }
func (i marketItem) FilterValue() string { return i.market.Question }
func (i marketItem) Description() string { return "" }

// probability returns the YES probability (0-1) parsed from OutcomePrices,
// falling back to BestBid. Returns 0 if unavailable.
func (i marketItem) probability() float64 {
	prices, _ := api.ParseOutcomePrices(i.market.OutcomePrices)
	if len(prices) > 0 {
		if p, err := strconv.ParseFloat(prices[0], 64); err == nil {
			return p
		}
	}
	if i.market.BestBid != nil {
		return *i.market.BestBid
	}
	return 0
}

// priceString returns the formatted YES price string, or a dim placeholder.
func (i marketItem) priceString() string {
	prices, _ := api.ParseOutcomePrices(i.market.OutcomePrices)
	if len(prices) > 0 {
		if p, err := strconv.ParseFloat(prices[0], 64); err == nil {
			return "YES " + FormatPrice(p)
		}
	}
	if i.market.BestBid != nil {
		return "YES " + FormatPrice(*i.market.BestBid)
	}
	return DimStyle.Render("YES --")
}

// effectiveStatus returns the display status, treating markets past their
// end date as closed even if the API still reports active=true.
func (i marketItem) effectiveStatus() (active, closed bool) {
	active, closed = i.market.Active, i.market.Closed
	if !closed && i.market.EndDateISO != "" {
		if t, err := api.ParseTime(i.market.EndDateISO); err == nil {
			if timeNow().After(t) {
				return false, true
			}
		}
	}
	return active, closed
}

// endDateString returns a formatted end date, or empty if unavailable.
func (i marketItem) endDateString() string {
	if i.market.EndDateISO == "" {
		return ""
	}
	t, err := api.ParseTime(i.market.EndDateISO)
	if err != nil {
		return ""
	}
	return t.Format("Jan 02, 2006")
}

// marketDelegate renders list items with accent bar + prob bar + 3-line layout.
type marketDelegate struct{}

func (d marketDelegate) Height() int                             { return 3 }
func (d marketDelegate) Spacing() int                            { return 1 }
func (d marketDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d marketDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	mi, ok := item.(marketItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	listWidth := m.Width()

	// Accent bar: 1-char-wide colored block
	accent := UnselectedAccent.Render(" ")
	if isSelected {
		accent = SelectedAccent.Render(" ")
	}
	accentW := 1

	// --- Line 1: accent + title (full width, no badge) ---
	titleMaxW := listWidth - accentW - 1
	if titleMaxW < 10 {
		titleMaxW = 10
	}
	title := Truncate(mi.market.Question, titleMaxW)
	if isSelected {
		title = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render(title)
	} else {
		title = lipgloss.NewStyle().Foreground(ColorWhite).Render(title)
	}
	line1 := accent + " " + title

	// --- Line 2: accent + price + probability bar ---
	price := mi.priceString()
	probBar := RenderProbBar(mi.probability(), 20)
	line2 := accent + "  " + price + "  " + probBar

	// --- Line 3: accent + Vol + Liq + End date ---
	var metaParts []string
	if mi.market.Volume != "" {
		metaParts = append(metaParts, "Vol "+FormatVolume(mi.market.Volume))
	}
	if mi.market.Liquidity != "" {
		metaParts = append(metaParts, "Liq "+FormatVolume(mi.market.Liquidity))
	}
	if endStr := mi.endDateString(); endStr != "" {
		metaParts = append(metaParts, "Ends "+endStr)
	}
	sep := MetaSepStyle.Render(" │ ")
	meta := DimStyle.Render("  ") + strings.Join(metaParts, sep)
	line3 := accent + meta

	// Apply background tint to selected item
	if isSelected {
		bgStyle := SelectedItemBg.Width(listWidth)
		line1 = bgStyle.Render(line1)
		line2 = bgStyle.Render(line2)
		line3 = bgStyle.Render(line3)
	}

	fmt.Fprintf(w, "%s\n%s\n%s", line1, line2, line3)
}

// tabBarHeight is the number of lines occupied by the category tab bar + separator.
const tabBarHeight = 2

// topLevelCategories defines the curated set of top-level category slugs and
// their display labels, in display order. The /tags API returns hundreds of
// niche tags ("detroit pistons", "spider-man"); the real top-level categories
// only appear as tags embedded on events. We match against these slugs when
// building the tag index from events.
var topLevelCategories = []struct {
	Slug  string
	Label string
}{
	{"politics", "Politics"},
	{"sports", "Sports"},
	{"crypto", "Crypto"},
	{"pop-culture", "Culture"},
	{"finance", "Finance"},
	{"tech", "Tech"},
	{"world", "World"},
}

// categoryTab holds a resolved category tab (slug + label) that has markets.
type categoryTab struct {
	Slug  string
	Label string
}

// MarketsListModel is the markets list screen.
type MarketsListModel struct {
	list        list.Model
	spinner     spinner.Model
	loading     bool
	err         error
	gammaClient *api.GammaClient
	width       int
	height      int

	// Category tabs — derived from event tags, not /tags API
	categories []categoryTab            // resolved tabs (only those with markets)
	activeTab  int                      // 0 = All, 1+ maps to categories[i-1]
	allMarkets []api.Market             // all markets (deduped from events)
	tagIndex   map[string][]api.Market  // tag.Slug → markets with that tag
}

// NewMarketsListModel creates a new markets list screen.
func NewMarketsListModel(gamma *api.GammaClient, w, h int) *MarketsListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	delegate := marketDelegate{}
	l := list.New([]list.Item{}, delegate, w, h-4-tabBarHeight)
	l.Title = "Polymarket — Markets"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(ColorCyan)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(ColorCyan)

	// Style the list's built-in help to match our theme
	l.Help.Styles.ShortKey = lipgloss.NewStyle().Foreground(ColorCyan)
	l.Help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ColorDim)
	l.Help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ColorDim)
	l.Help.Styles.FullKey = lipgloss.NewStyle().Foreground(ColorCyan)
	l.Help.Styles.FullDesc = lipgloss.NewStyle().Foreground(ColorDim)
	l.Help.Styles.FullSeparator = lipgloss.NewStyle().Foreground(ColorDim)

	// Add extra keys to the list's help
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{marketsExtra.Select, marketsExtra.PrevTab, marketsExtra.NextTab, marketsExtra.Portfolio}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{marketsExtra.Select, marketsExtra.PrevTab, marketsExtra.NextTab, marketsExtra.Portfolio, marketsExtra.Quit}
	}

	return &MarketsListModel{
		list:        l,
		spinner:     s,
		loading:     true,
		gammaClient: gamma,
		width:       w,
		height:      h,
		tagIndex:    make(map[string][]api.Market),
	}
}

func (m *MarketsListModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchEvents())
}

func (m *MarketsListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case eventsLoadedMsg:
		m.loading = false
		m.buildFromEvents(msg.events)
		if len(m.allMarkets) == 0 {
			// Fallback: events didn't embed markets, fetch directly
			m.loading = true
			return m, m.fetchMarketsDirect()
		}
		cmd := m.list.SetItems(marketsToItems(m.allMarkets))
		return m, cmd

	case eventsErrorMsg:
		// Fall back to direct market fetch
		return m, m.fetchMarketsDirect()

	case marketsLoadedMsg:
		// Fallback path (direct market fetch)
		m.loading = false
		m.allMarkets = msg.markets
		cmd := m.list.SetItems(marketsToItems(m.allMarkets))
		return m, cmd

	case marketsErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4-tabBarHeight)
		return m, nil

	case tea.KeyMsg:
		// Don't handle custom keys when filtering
		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "?":
			m.list.Help.ShowAll = !m.list.Help.ShowAll
			return m, nil
		case "[":
			if m.activeTab > 0 {
				m.activeTab--
				m.updateListForCategory()
			}
			return m, nil
		case "]":
			if m.activeTab < len(m.categories) {
				m.activeTab++
				m.updateListForCategory()
			}
			return m, nil
		}
		if msg.String() == "enter" {
			if id := m.SelectedMarketID(); id != "" {
				return m, func() tea.Msg {
					return pushScreenMsg{screen: ScreenMarketDetail, marketID: id}
				}
			}
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MarketsListModel) View() string {
	if m.loading {
		return AppStyle.Render(m.spinner.View() + " Loading markets…")
	}
	if m.err != nil {
		return AppStyle.Render(ErrorStyle.Render("Error: " + m.err.Error()))
	}
	tabs := m.renderCategoryTabs()
	sepLine := CategorySepStyle.Render(strings.Repeat("─", m.width))
	return tabs + "\n" + sepLine + "\n" + m.list.View()
}

// Filtering returns true if the list filter input is active.
func (m *MarketsListModel) Filtering() bool {
	return m.list.FilterState() == list.Filtering
}

// SelectedMarketID returns the ID of the currently selected market.
func (m *MarketsListModel) SelectedMarketID() string {
	item := m.list.SelectedItem()
	if item == nil {
		return ""
	}
	if mi, ok := item.(marketItem); ok {
		return mi.market.ID
	}
	return ""
}

// buildFromEvents extracts markets from events, deduplicates, builds the tag
// index, and derives which top-level categories have markets.
func (m *MarketsListModel) buildFromEvents(events []api.Event) {
	now := timeNow()
	seen := make(map[string]bool)
	tagSeen := make(map[string]map[string]bool) // slug → set of market IDs
	m.allMarkets = nil
	m.tagIndex = make(map[string][]api.Market)

	for _, ev := range events {
		for _, mkt := range ev.Markets {
			// Filter out expired markets
			if mkt.EndDateISO != "" {
				if t, err := api.ParseTime(mkt.EndDateISO); err == nil && now.After(t) {
					continue
				}
			}
			if !seen[mkt.ID] {
				seen[mkt.ID] = true
				m.allMarkets = append(m.allMarkets, mkt)
			}
			// Index by each of the event's tags
			for _, tag := range ev.Tags {
				if tagSeen[tag.Slug] == nil {
					tagSeen[tag.Slug] = make(map[string]bool)
				}
				if !tagSeen[tag.Slug][mkt.ID] {
					tagSeen[tag.Slug][mkt.ID] = true
					m.tagIndex[tag.Slug] = append(m.tagIndex[tag.Slug], mkt)
				}
			}
		}
	}

	// Derive category tabs from topLevelCategories — only include those
	// that actually have markets in the tagIndex.
	m.categories = nil
	for _, cat := range topLevelCategories {
		if len(m.tagIndex[cat.Slug]) > 0 {
			m.categories = append(m.categories, categoryTab{Slug: cat.Slug, Label: cat.Label})
		}
	}
}

// updateListForCategory sets the list items based on the active category tab.
func (m *MarketsListModel) updateListForCategory() {
	var markets []api.Market
	if m.activeTab == 0 {
		markets = m.allMarkets
	} else if m.activeTab-1 < len(m.categories) {
		slug := m.categories[m.activeTab-1].Slug
		markets = m.tagIndex[slug]
	}
	m.list.SetItems(marketsToItems(markets))
	m.list.ResetSelected()
}

// renderCategoryTabs renders the horizontal category tab bar.
func (m *MarketsListModel) renderCategoryTabs() string {
	labels := make([]string, 0, 1+len(m.categories))

	// "All" tab
	if m.activeTab == 0 {
		labels = append(labels, CategoryActiveStyle.Render("All"))
	} else {
		labels = append(labels, CategoryInactiveStyle.Render("All"))
	}

	for i, cat := range m.categories {
		if m.activeTab == i+1 {
			labels = append(labels, CategoryActiveStyle.Render(cat.Label))
		} else {
			labels = append(labels, CategoryInactiveStyle.Render(cat.Label))
		}
	}

	tabSep := CategorySepStyle.Render("│")
	full := strings.Join(labels, tabSep)

	// If the rendered tab bar fits, return it directly
	if m.width == 0 || lipgloss.Width(full) <= m.width {
		return " " + full
	}

	// Overflow: show a visible window around the active tab with ◂/▸ indicators
	totalTabs := 1 + len(m.categories)
	start := m.activeTab - 2
	if start < 0 {
		start = 0
	}
	end := start + 5
	if end > totalTabs {
		end = totalTabs
		start = end - 5
		if start < 0 {
			start = 0
		}
	}

	var visible []string
	for idx := start; idx < end; idx++ {
		var label string
		if idx == 0 {
			label = "All"
		} else {
			label = m.categories[idx-1].Label
		}
		if idx == m.activeTab {
			visible = append(visible, CategoryActiveStyle.Render(label))
		} else {
			visible = append(visible, CategoryInactiveStyle.Render(label))
		}
	}

	result := strings.Join(visible, tabSep)
	if start > 0 {
		result = CategorySepStyle.Render("◂ ") + result
	}
	if end < totalTabs {
		result = result + CategorySepStyle.Render(" ▸")
	}
	return " " + result
}

// fetchEvents fetches events (with embedded markets and tags) from the API.
func (m *MarketsListModel) fetchEvents() tea.Cmd {
	return func() tea.Msg {
		if m.gammaClient == nil {
			return eventsErrorMsg{err: fmt.Errorf("no API client configured")}
		}
		events, err := m.gammaClient.ListEvents(api.EventListParams{
			Limit:  200,
			Active: api.BoolPtr(true),
			Closed: api.BoolPtr(false),
		})
		if err != nil {
			return eventsErrorMsg{err: err}
		}
		return eventsLoadedMsg{events: events}
	}
}

// fetchMarketsDirect fetches markets directly (fallback when events don't embed markets).
func (m *MarketsListModel) fetchMarketsDirect() tea.Cmd {
	return func() tea.Msg {
		if m.gammaClient == nil {
			return marketsErrorMsg{err: fmt.Errorf("no API client configured")}
		}
		markets, err := m.gammaClient.ListMarkets(api.MarketListParams{
			Limit:  200,
			Active: api.BoolPtr(true),
			Closed: api.BoolPtr(false),
		})
		if err != nil {
			return marketsErrorMsg{err: err}
		}
		// Filter out markets that are effectively closed (past end date)
		now := timeNow()
		filtered := make([]api.Market, 0, len(markets))
		for _, mkt := range markets {
			if mkt.EndDateISO != "" {
				if t, err := api.ParseTime(mkt.EndDateISO); err == nil && now.After(t) {
					continue
				}
			}
			filtered = append(filtered, mkt)
		}
		return marketsLoadedMsg{markets: filtered}
	}
}

// marketsToItems converts a slice of markets to list items.
func marketsToItems(markets []api.Market) []list.Item {
	items := make([]list.Item, len(markets))
	for i, mkt := range markets {
		items[i] = marketItem{market: mkt}
	}
	return items
}
