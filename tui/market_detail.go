package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/api"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
)

// MarketDetailModel is the market detail screen.
type MarketDetailModel struct {
	market       *api.Market
	marketID     string
	orderBook    *clob.OrderBook
	viewport     viewport.Model
	progress     progress.Model
	spinner      spinner.Model
	spring       harmonica.Spring
	loading      bool
	err          error
	probTarget   float64
	probCurrent  float64
	springVel    float64
	descRendered string
	gammaClient  *api.GammaClient
	clobClient   *clob.Client
	helpModel    help.Model
	showFullHelp bool
	width        int
	height       int
}

// NewMarketDetailModel creates a new market detail screen.
func NewMarketDetailModel(marketID string, gamma *api.GammaClient, clobCl *clob.Client, w, h int) *MarketDetailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	// h - 5 accounts for: AppStyle top/bottom padding (2), help bar (1),
	// newline before help (1), status bar from App (1).
	vpH := h - 5
	if vpH < 4 {
		vpH = 4
	}
	vp := viewport.New(w-4, vpH)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	return &MarketDetailModel{
		marketID:    marketID,
		spinner:     s,
		progress:    p,
		viewport:    vp,
		spring:      harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.0),
		loading:     true,
		gammaClient: gamma,
		clobClient:  clobCl,
		helpModel:   newHelpModel(),
		width:       w,
		height:      h,
	}
}

func (m *MarketDetailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchMarketDetail())
}

func (m *MarketDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case marketDetailLoadedMsg:
		m.loading = false
		m.market = msg.market

		// Parse YES probability
		prices, _ := api.ParseOutcomePrices(m.market.OutcomePrices)
		if len(prices) > 0 {
			if p, err := strconv.ParseFloat(prices[0], 64); err == nil {
				m.probTarget = p
			}
		}

		// Render description with glamour (cached)
		m.renderDescription()

		// Fetch order book if token ID available
		var bookCmd tea.Cmd
		if tokID := m.market.FirstTokenID(); tokID != "" {
			bookCmd = m.fetchOrderBook(tokID)
		}

		m.rebuildContent()
		return m, tea.Batch(bookCmd, animateTick())

	case marketDetailErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case orderBookLoadedMsg:
		m.orderBook = msg.book
		m.rebuildContent()
		return m, nil

	case orderBookErrorMsg:
		// Non-fatal: just don't show order book
		return m, nil

	case animateTickMsg:
		newPos, newVel := m.spring.Update(m.probCurrent, m.springVel, m.probTarget)
		m.probCurrent = newPos
		m.springVel = newVel
		if math.Abs(m.probCurrent-m.probTarget) < 0.001 {
			m.probCurrent = m.probTarget
			m.rebuildContent()
			return m, nil
		}
		m.rebuildContent()
		return m, animateTick()

	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			m.showFullHelp = !m.showFullHelp
			return m, nil
		case "b", "esc":
			return m, func() tea.Msg { return popScreenMsg{} }
		case "q":
			return m, tea.Quit
		case "enter":
			// Open order book for first token
			if m.market != nil {
				if tokID := m.market.FirstTokenID(); tokID != "" {
					return m, func() tea.Msg {
						return pushScreenMsg{
							screen:   ScreenOrderBook,
							tokenID:  tokID,
							question: m.market.Question,
						}
					}
				}
			}
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		vpH := msg.Height - 5
		if vpH < 4 {
			vpH = 4
		}
		m.viewport.Height = vpH
		if m.market != nil {
			m.renderDescription()
			m.rebuildContent()
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			// If spinner.Update returned nil (e.g. stale tick with wrong ID/tag),
			// re-kick the spinner to prevent the animation from dying.
			if cmd == nil {
				cmd = m.spinner.Tick
			}
			return m, cmd
		}
		return m, nil

	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model.(progress.Model)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *MarketDetailModel) View() string {
	if m.loading {
		return AppStyle.Render(m.spinner.View() + " Loading market details…")
	}
	if m.err != nil {
		return AppStyle.Render(ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n" + HelpStyle.Render("b back"))
	}
	if m.market == nil {
		return AppStyle.Render(DimStyle.Render("No market data"))
	}

	// Viewport wraps all content; help bar pinned at bottom
	m.helpModel.Width = m.width - 4
	m.helpModel.ShowAll = m.showFullHelp
	helpView := m.helpModel.View(detailKeys)

	// Scroll indicator when content overflows viewport
	scrollInfo := ""
	if !(m.viewport.AtTop() && m.viewport.AtBottom()) {
		pct := int(m.viewport.ScrollPercent() * 100)
		scrollInfo = DimStyle.Render(fmt.Sprintf("  %d%%", pct))
	}

	return AppStyle.Render(m.viewport.View() + "\n" + helpView + scrollInfo)
}

// rebuildContent assembles all sections into the viewport.
func (m *MarketDetailModel) rebuildContent() {
	if m.market == nil {
		return
	}

	var sections []string

	// Header: Question + status + end date
	header := HeaderStyle.Render(m.market.Question)
	badge := StatusBadge(m.market.Active, m.market.Closed)
	endDate := ""
	if m.market.EndDateISO != "" {
		endDate = DimStyle.Render("Ends: " + formatEndDate(m.market.EndDateISO))
	}
	sections = append(sections, header+"\n"+badge+"  "+endDate)

	// Probability bar
	pct := fmt.Sprintf("YES %.1f%%", m.probCurrent*100)
	bar := m.progress.ViewAs(m.probCurrent) + "  " + GreenStyle.Render(pct)
	sections = append(sections, bar)

	// Stats panel (bordered card)
	stats := m.renderStats()
	sections = append(sections, stats)

	// Description
	sections = append(sections, SectionStyle.Render(LabelStyle.Render("Description")))
	sections = append(sections, m.descRendered)

	// Order book preview (bordered bids/asks)
	if m.orderBook != nil {
		sections = append(sections, m.renderOrderBook())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	m.viewport.SetContent(content)
}

func (m *MarketDetailModel) renderStats() string {
	// Left column: Volume, Liquidity
	vol := DimStyle.Render("—")
	if m.market.Volume != "" {
		vol = FormatVolume(m.market.Volume)
	}
	liq := DimStyle.Render("—")
	if m.market.Liquidity != "" {
		liq = FormatVolume(m.market.Liquidity)
	}

	labelW := 12
	labelStyle := LabelStyle.Width(labelW)

	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("Volume")+" "+vol,
		labelStyle.Render("Liquidity")+" "+liq,
	)

	// Right column: Best Bid, Best Ask, Last Trade
	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("Best Bid")+" "+FormatPricePtr(m.market.BestBid),
		labelStyle.Render("Best Ask")+" "+FormatPricePtr(m.market.BestAsk),
		labelStyle.Render("Last Trade")+" "+FormatPricePtr(m.market.LastTradePrice),
	)

	inner := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  │  ", rightCol)
	return CardStyle.Render(inner)
}

func (m *MarketDetailModel) renderOrderBook() string {
	colW := (m.width - 20) / 4
	if colW < 10 {
		colW = 10
	}
	cellStyle := lipgloss.NewStyle().Width(colW)

	renderSide := func(entries []clob.OrderBookEntry, title string, color lipgloss.Color) string {
		headerRow := lipgloss.NewStyle().Foreground(color).Bold(true).Width(colW)
		var sb strings.Builder
		sb.WriteString(headerRow.Render("Price"))
		sb.WriteString(headerRow.Render("Size"))
		sb.WriteByte('\n')

		count := min(5, len(entries))
		for i := 0; i < count; i++ {
			e := entries[i]
			cs := cellStyle.Foreground(color)
			sb.WriteString(cs.Render(e.Price))
			sb.WriteString(cs.Render(e.Size))
			sb.WriteByte('\n')
		}
		if count == 0 {
			sb.WriteString(DimStyle.Render("  No orders") + "\n")
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(color).
			Padding(0, 1).
			Render(lipgloss.NewStyle().Foreground(color).Bold(true).Render(title) + "\n" + sb.String())
		return box
	}

	bidBox := renderSide(m.orderBook.Bids, "BIDS", ColorGreen)
	askBox := renderSide(m.orderBook.Asks, "ASKS", ColorRed)

	heading := SectionStyle.Render(LabelStyle.Render("Order Book"))
	columns := lipgloss.JoinHorizontal(lipgloss.Top, bidBox, "  ", askBox)
	return heading + "\n" + columns
}

func (m *MarketDetailModel) renderDescription() {
	desc := m.market.Description
	if desc == "" {
		desc = "_No description available._"
	}
	width := m.width - 6
	if width < 20 {
		width = 20
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		m.descRendered = desc
		return
	}
	rendered, err := renderer.Render(desc)
	if err != nil {
		m.descRendered = desc
		return
	}
	m.descRendered = rendered
}

func (m *MarketDetailModel) fetchMarketDetail() tea.Cmd {
	return func() tea.Msg {
		if m.gammaClient == nil {
			return marketDetailErrorMsg{err: fmt.Errorf("no API client configured")}
		}
		market, err := m.gammaClient.GetMarket(m.marketID)
		if err != nil {
			return marketDetailErrorMsg{err: err}
		}
		return marketDetailLoadedMsg{market: market}
	}
}

func (m *MarketDetailModel) fetchOrderBook(tokenID string) tea.Cmd {
	return func() tea.Msg {
		if m.clobClient == nil {
			return orderBookErrorMsg{err: fmt.Errorf("no CLOB client")}
		}
		book, err := m.clobClient.GetBook(tokenID)
		if err != nil {
			return orderBookErrorMsg{err: err}
		}
		return orderBookLoadedMsg{book: book}
	}
}

func animateTick() tea.Cmd {
	return tea.Tick(time.Millisecond*16, func(t time.Time) tea.Msg {
		return animateTickMsg(t)
	})
}

func formatEndDate(iso string) string {
	t, err := api.ParseTime(iso)
	if err != nil {
		return iso
	}
	return t.Format("Jan 02, 2006")
}
