package tui

import (
	"time"

	"github.com/piyushgupta/polymarket-cli/internal/api"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/api/data"
)

// Screen identifies which screen to display.
type Screen int

const (
	ScreenMarketsList Screen = iota
	ScreenMarketDetail
	ScreenOrderBook
	ScreenOrderForm
	ScreenPortfolio
	ScreenShell
	ScreenSetupWizard
	ScreenOnboarding
)

// Navigation messages
type pushScreenMsg struct {
	screen   Screen
	marketID string
	tokenID  string
	question string
	deriveFn func(string, string) (string, error)
}

type popScreenMsg struct{}

// replaceScreenMsg replaces the top of the stack instead of pushing.
type replaceScreenMsg struct {
	screen   Screen
	marketID string
	tokenID  string
	question string
	deriveFn func(string, string) (string, error)
}

// API result messages — markets list
type marketsLoadedMsg struct {
	markets []api.Market
}

type marketsErrorMsg struct {
	err error
}

// API result messages — market detail
type marketDetailLoadedMsg struct {
	market *api.Market
}

type marketDetailErrorMsg struct {
	err error
}

// API result messages — order book
type orderBookLoadedMsg struct {
	book *clob.OrderBook
}

type orderBookErrorMsg struct {
	err error
}

// API result messages — order book live refresh
type orderBookRefreshMsg struct {
	book *clob.OrderBook
}

type orderBookRefreshErrorMsg struct {
	err error
}

// API result messages — order form
type orderSubmittedMsg struct {
	resp *clob.OrderResponse
}

type orderSubmitErrorMsg struct {
	err error
}

// API result messages — portfolio
type positionsLoadedMsg struct {
	positions []data.Position
}

type positionsErrorMsg struct {
	err error
}

type openOrdersLoadedMsg struct {
	orders []clob.Order
}

type openOrdersErrorMsg struct {
	err error
}

type tradesLoadedMsg struct {
	trades []clob.Trade
}

type tradesErrorMsg struct {
	err error
}

type balanceLoadedMsg struct {
	balance *clob.BalanceAllowance
}

type balanceErrorMsg struct {
	err error
}

type orderCancelledMsg struct {
	resp *clob.CancelResponse
}

type orderCancelErrorMsg struct {
	err error
}

// Shell messages
type shellOutputMsg struct {
	output string
}

type shellErrorMsg struct {
	err error
}

// Animation tick
type animateTickMsg time.Time

// Refresh ticks — distinct types to prevent cross-screen leaking
type orderBookRefreshTickMsg time.Time
type portfolioRefreshTickMsg time.Time

// Animation-specific ticks
type waveTickMsg time.Time
type flashTickMsg time.Time
type transitionTickMsg time.Time

// Midpoint for order book header
type midpointLoadedMsg struct {
	mid string
}

// API result messages — events (for category tabs)
type eventsLoadedMsg struct {
	events []api.Event
}

type eventsErrorMsg struct {
	err error
}

// authReloadMsg signals the App to reload credentials from config after inline setup.
type authReloadMsg struct{}

