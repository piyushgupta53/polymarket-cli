package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
)

const refreshInterval = 3 * time.Second

// flashEntry tracks the flash intensity for a single row.
type flashEntry struct {
	intensity float64
	vel       float64
}

// orderRow is a manually-rendered row in the order book.
type orderRow struct {
	price string
	size  string
	total string
}

// OrderBookModel is the live order book viewer screen.
type OrderBookModel struct {
	tokenID    string
	question   string
	orderBook  *clob.OrderBook
	midpoint   string
	bidRows    []orderRow
	askRows    []orderRow
	spinner    spinner.Model
	loading    bool
	err        error
	clobClient *clob.Client
	width      int
	height     int

	// Flash animation state
	prevBidPrices map[string]string // price -> size (for change detection)
	prevAskPrices map[string]string
	bidFlash      map[int]flashEntry // row index -> flash
	askFlash      map[int]flashEntry
	flashSpring   harmonica.Spring
	flashing      bool
	firstLoad     bool // suppress flash on initial load
	helpModel     help.Model
	showFullHelp  bool
}

// NewOrderBookModel creates a new order book viewer.
func NewOrderBookModel(tokenID, question string, clobCl *clob.Client, w, h int) *OrderBookModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	return &OrderBookModel{
		tokenID:       tokenID,
		question:      question,
		spinner:       s,
		loading:       true,
		clobClient:    clobCl,
		width:         w,
		height:        h,
		prevBidPrices: make(map[string]string),
		prevAskPrices: make(map[string]string),
		bidFlash:      make(map[int]flashEntry),
		askFlash:      make(map[int]flashEntry),
		flashSpring:   harmonica.NewSpring(harmonica.FPS(60), 5.0, 0.6),
		firstLoad:     true,
		helpModel:     newHelpModel(),
	}
}

func (m *OrderBookModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchBook(),
		m.fetchMidpoint(),
	)
}

func (m *OrderBookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case orderBookLoadedMsg:
		m.loading = false
		m.orderBook = msg.book
		m.updateTables()
		m.firstLoad = false
		return m, scheduleRefresh()

	case orderBookErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, scheduleRefresh()

	case orderBookRefreshMsg:
		m.orderBook = msg.book
		m.err = nil
		m.updateTables()
		cmds := []tea.Cmd{scheduleRefresh()}
		if m.flashing {
			cmds = append(cmds, flashTick())
		}
		return m, tea.Batch(cmds...)

	case orderBookRefreshErrorMsg:
		// Keep stale data, just schedule next refresh
		return m, scheduleRefresh()

	case midpointLoadedMsg:
		m.midpoint = msg.mid
		return m, nil

	case refreshTickMsg:
		return m, tea.Batch(m.refreshBook(), m.fetchMidpoint())

	case flashTickMsg:
		if !m.flashing {
			return m, nil
		}
		anyActive := false
		for i, f := range m.bidFlash {
			f.intensity, f.vel = m.flashSpring.Update(f.intensity, f.vel, 0)
			if math.Abs(f.intensity) < 0.01 {
				delete(m.bidFlash, i)
			} else {
				m.bidFlash[i] = f
				anyActive = true
			}
		}
		for i, f := range m.askFlash {
			f.intensity, f.vel = m.flashSpring.Update(f.intensity, f.vel, 0)
			if math.Abs(f.intensity) < 0.01 {
				delete(m.askFlash, i)
			} else {
				m.askFlash[i] = f
				anyActive = true
			}
		}
		if !anyActive {
			m.flashing = false
			return m, nil
		}
		return m, flashTick()

	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			m.showFullHelp = !m.showFullHelp
			return m, nil
		case "b", "esc":
			return m, func() tea.Msg { return popScreenMsg{} }
		case "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, tea.Batch(m.fetchBook(), m.fetchMidpoint())
		case "o":
			if m.orderBook != nil {
				return m, func() tea.Msg {
					return pushScreenMsg{
						screen:   ScreenOrderForm,
						tokenID:  m.tokenID,
						question: m.question,
					}
				}
			}
		}

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

func (m *OrderBookModel) View() string {
	if m.loading && m.orderBook == nil {
		return AppStyle.Render(m.spinner.View() + " Loading order book…")
	}
	if m.err != nil && m.orderBook == nil {
		return AppStyle.Render(ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n" + HelpStyle.Render("r refresh │ b back"))
	}

	var sections []string

	// Header
	header := HeaderStyle.Render(Truncate(m.question, 70))
	midStr := ""
	if m.midpoint != "" {
		midStr = LabelStyle.Render("Mid: ") + CyanStyle.Render(m.midpoint)
	}
	lastTrade := ""
	if m.orderBook != nil && m.orderBook.LastTradePrice != "" {
		lastTrade = LabelStyle.Render("Last: ") + ValueStyle.Render(m.orderBook.LastTradePrice)
	}
	sections = append(sections, header+"\n"+midStr+"  "+lastTrade)

	// Spread
	if m.orderBook != nil && len(m.orderBook.Bids) > 0 && len(m.orderBook.Asks) > 0 {
		spread := SpreadBadge.Render(fmt.Sprintf("Spread: %s - %s", m.orderBook.Bids[0].Price, m.orderBook.Asks[0].Price))
		sections = append(sections, spread)
	}

	// Two-column layout: Bids | Asks with manual row rendering for flash
	bidView := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorGreen).
		Padding(0, 1).
		Render(BidStyle.Render("BIDS") + "\n" + m.renderRows(m.bidRows, m.bidFlash, ColorGreen))

	askView := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorRed).
		Padding(0, 1).
		Render(AskStyle.Render("ASKS") + "\n" + m.renderRows(m.askRows, m.askFlash, ColorRed))

	columns := lipgloss.JoinHorizontal(lipgloss.Top, bidView, "  ", askView)
	sections = append(sections, columns)

	// Level count
	if m.orderBook != nil {
		sections = append(sections, DimStyle.Render(fmt.Sprintf("  %d bids  %d asks", len(m.bidRows), len(m.askRows))))
	}

	// Help
	m.helpModel.Width = m.width - 4
	m.helpModel.ShowAll = m.showFullHelp
	sections = append(sections, m.helpModel.View(orderBookKeys))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return AppStyle.Render(content)
}

// renderRows renders order rows with optional flash coloring.
func (m *OrderBookModel) renderRows(rows []orderRow, flash map[int]flashEntry, baseColor lipgloss.Color) string {
	colW := (m.width - 16) / 6
	if colW < 8 {
		colW = 8
	}

	headerStyle := lipgloss.NewStyle().Foreground(baseColor).Bold(true).Width(colW)
	var sb strings.Builder
	sb.WriteString(headerStyle.Render("Price"))
	sb.WriteString(headerStyle.Render("Size"))
	sb.WriteString(headerStyle.Render("Total"))
	sb.WriteByte('\n')

	for i, row := range rows {
		color := baseColor
		if f, ok := flash[i]; ok && f.intensity > 0.01 {
			color = blendColor(ColorWhite, baseColor, f.intensity)
		}
		cellStyle := lipgloss.NewStyle().Foreground(color).Width(colW)
		sb.WriteString(cellStyle.Render(row.price))
		sb.WriteString(cellStyle.Render(row.size))
		sb.WriteString(cellStyle.Render(row.total))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func (m *OrderBookModel) updateTables() {
	if m.orderBook == nil {
		return
	}

	// Build new bid rows
	newBidPrices := make(map[string]string)
	m.bidRows = make([]orderRow, 0, min(20, len(m.orderBook.Bids)))
	cumBid := 0.0
	for i := 0; i < min(20, len(m.orderBook.Bids)); i++ {
		b := m.orderBook.Bids[i]
		size := parseFloat(b.Size)
		cumBid += size
		m.bidRows = append(m.bidRows, orderRow{b.Price, b.Size, fmt.Sprintf("%.2f", cumBid)})
		newBidPrices[b.Price] = b.Size
	}

	// Build new ask rows
	newAskPrices := make(map[string]string)
	m.askRows = make([]orderRow, 0, min(20, len(m.orderBook.Asks)))
	cumAsk := 0.0
	for i := 0; i < min(20, len(m.orderBook.Asks)); i++ {
		a := m.orderBook.Asks[i]
		size := parseFloat(a.Size)
		cumAsk += size
		m.askRows = append(m.askRows, orderRow{a.Price, a.Size, fmt.Sprintf("%.2f", cumAsk)})
		newAskPrices[a.Price] = a.Size
	}

	// Detect changes and trigger flash (skip on first load)
	if !m.firstLoad {
		for i, row := range m.bidRows {
			prevSize, existed := m.prevBidPrices[row.price]
			if !existed || prevSize != row.size {
				m.bidFlash[i] = flashEntry{intensity: 1.0, vel: 0}
				m.flashing = true
			}
		}
		for i, row := range m.askRows {
			prevSize, existed := m.prevAskPrices[row.price]
			if !existed || prevSize != row.size {
				m.askFlash[i] = flashEntry{intensity: 1.0, vel: 0}
				m.flashing = true
			}
		}
	}

	m.prevBidPrices = newBidPrices
	m.prevAskPrices = newAskPrices
}

func (m *OrderBookModel) fetchBook() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil {
			return orderBookErrorMsg{err: fmt.Errorf("no CLOB client")}
		}
		book, err := m.clobClient.GetBook(m.tokenID)
		if err != nil {
			return orderBookErrorMsg{err: err}
		}
		return orderBookLoadedMsg{book: book}
	}
}

func (m *OrderBookModel) refreshBook() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil {
			return orderBookRefreshErrorMsg{err: fmt.Errorf("no CLOB client")}
		}
		book, err := m.clobClient.GetBook(m.tokenID)
		if err != nil {
			return orderBookRefreshErrorMsg{err: err}
		}
		return orderBookRefreshMsg{book: book}
	}
}

func (m *OrderBookModel) fetchMidpoint() tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil {
			return midpointLoadedMsg{}
		}
		resp, err := m.clobClient.GetMidpoint(m.tokenID)
		if err != nil {
			return midpointLoadedMsg{}
		}
		return midpointLoadedMsg{mid: resp.Mid}
	}
}

func scheduleRefresh() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func parseFloat(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f
}

func flashTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return flashTickMsg(t)
	})
}

// Flashing returns the current flash state (for testing).
func (m *OrderBookModel) Flashing() bool {
	return m.flashing
}

// BidFlashCount returns the number of active bid flashes (for testing).
func (m *OrderBookModel) BidFlashCount() int {
	return len(m.bidFlash)
}

// AskFlashCount returns the number of active ask flashes (for testing).
func (m *OrderBookModel) AskFlashCount() int {
	return len(m.askFlash)
}
