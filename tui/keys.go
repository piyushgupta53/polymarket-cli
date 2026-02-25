package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// newHelpModel creates a help.Model styled to match the TUI theme.
func newHelpModel() help.Model {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(ColorCyan)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ColorDim)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ColorDim)
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(ColorCyan)
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(ColorDim)
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(ColorDim)
	return h
}

// --- Markets List additional keys ---

type marketsExtraKeys struct {
	Select    key.Binding
	Portfolio key.Binding
	PrevTab   key.Binding
	NextTab   key.Binding
	Quit      key.Binding
}

var marketsExtra = marketsExtraKeys{
	Select:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open market")),
	Portfolio: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "portfolio")),
	PrevTab:   key.NewBinding(key.WithKeys("["), key.WithHelp("[", "prev category")),
	NextTab:   key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next category")),
	Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

// --- Market Detail keys ---

type detailKeyMap struct {
	OpenOrderBook key.Binding
	ScrollUp      key.Binding
	ScrollDown    key.Binding
	Portfolio     key.Binding
	Back          key.Binding
	Quit          key.Binding
	Help          key.Binding
}

func (k detailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.OpenOrderBook, k.ScrollDown, k.Back, k.Help}
}

func (k detailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.OpenOrderBook, k.Portfolio},
		{k.ScrollUp, k.ScrollDown},
		{k.Back, k.Quit, k.Help},
	}
}

var detailKeys = detailKeyMap{
	OpenOrderBook: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "order book")),
	ScrollUp:      key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "scroll up")),
	ScrollDown:    key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "scroll down")),
	Portfolio:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "portfolio")),
	Back:          key.NewBinding(key.WithKeys("b", "esc"), key.WithHelp("b/esc", "back")),
	Quit:          key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Help:          key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

// --- Order Book keys ---

type orderBookKeyMap struct {
	Refresh key.Binding
	Order   key.Binding
	Back    key.Binding
	Quit    key.Binding
	Help    key.Binding
}

func (k orderBookKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.Order, k.Back, k.Help}
}

func (k orderBookKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Refresh, k.Order},
		{k.Back, k.Quit, k.Help},
	}
}

var orderBookKeys = orderBookKeyMap{
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Order:   key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "place order")),
	Back:    key.NewBinding(key.WithKeys("b", "esc"), key.WithHelp("b/esc", "back")),
	Quit:    key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

// --- Portfolio keys ---

type portfolioKeyMap struct {
	Tab        key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Cancel     key.Binding
	Refresh    key.Binding
	Back       key.Binding
	Help       key.Binding
}

func (k portfolioKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ScrollDown, k.Refresh, k.Back, k.Help}
}

func (k portfolioKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Cancel},
		{k.ScrollUp, k.ScrollDown},
		{k.Refresh, k.Back, k.Help},
	}
}

var portfolioKeys = portfolioKeyMap{
	Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch tab")),
	ScrollUp:   key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "scroll up")),
	ScrollDown: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "scroll down")),
	Cancel:     key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cancel order")),
	Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Back:       key.NewBinding(key.WithKeys("q", "esc"), key.WithHelp("q/esc", "back")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}
