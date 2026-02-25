package tui

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/api"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/api/data"
	"github.com/piyushgupta/polymarket-cli/internal/config"
)

// App is the root BubbleTea model with stack-based screen navigation.
type App struct {
	gammaClient *api.GammaClient
	clobClient  *clob.Client
	dataClient  *data.DataClient
	address     string
	screenStack []tea.Model
	width       int
	height      int

	// execFn dispatches shell commands to Cobra.
	execFn func(string) (string, error)

	// hasAuth is true when the user has configured credentials.
	hasAuth bool
	// deriveFn derives API keys from a private key and signature type.
	deriveFn func(string, string) (string, error)

	// Screen transition animation state.
	transitioning    bool
	transitionOffset float64
	transitionVel    float64
	transitionSpring harmonica.Spring
	pendingPop       bool
}

// NewApp creates a new TUI application.
func NewApp(gamma *api.GammaClient, clobCl *clob.Client, dataCl *data.DataClient, address string, hasAuth bool, deriveFn func(string, string) (string, error)) *App {
	return &App{
		gammaClient:      gamma,
		clobClient:       clobCl,
		dataClient:       dataCl,
		address:          address,
		hasAuth:          hasAuth,
		deriveFn:         deriveFn,
		transitionSpring: harmonica.NewSpring(harmonica.FPS(60), 8.0, 0.8),
	}
}

// SetExecFn sets the shell command executor function.
func (a *App) SetExecFn(fn func(string) (string, error)) {
	a.execFn = fn
}

func (a *App) Init() tea.Cmd {
	var screen tea.Model
	if a.hasAuth {
		screen = NewMarketsListModel(a.gammaClient, a.width, a.height)
	} else {
		screen = NewOnboardingModel(a.deriveFn, a.width, a.height)
	}
	a.screenStack = []tea.Model{screen}
	return tea.Batch(tea.EnterAltScreen, screen.Init())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to current screen
		if screen := a.currentScreen(); screen != nil {
			updated, cmd := screen.Update(msg)
			a.screenStack[len(a.screenStack)-1] = updated
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "p":
			// Global 'p' for portfolio — only if not in shell/form/splash and not transitioning
			if !a.isInputScreen() && !a.transitioning {
				return a, func() tea.Msg {
					return pushScreenMsg{screen: ScreenPortfolio}
				}
			}
		}

	case pushScreenMsg:
		// Block pushes during transition
		if a.transitioning {
			return a, nil
		}
		// Auth-gate: redirect portfolio and order form to setup wizard when no auth
		if !a.hasAuth && (msg.screen == ScreenPortfolio || msg.screen == ScreenOrderForm) {
			msg = pushScreenMsg{screen: ScreenSetupWizard, deriveFn: a.deriveFn}
		}
		// Deactivate the current top screen so its tick loops stop.
		a.deactivateTop()
		screen := a.constructScreen(pushScreenMsg(msg))
		if screen == nil {
			return a, nil
		}
		a.screenStack = append(a.screenStack, screen)
		initCmd := screen.Init()

		// Start slide-in transition if we have width
		if a.width > 0 {
			a.transitioning = true
			a.transitionOffset = float64(a.width)
			a.transitionVel = 0
			a.pendingPop = false
			return a, tea.Batch(initCmd, transitionTick())
		}
		return a, initCmd

	case popScreenMsg:
		// Block pops during transition
		if a.transitioning {
			return a, nil
		}
		if len(a.screenStack) <= 1 {
			return a, tea.Quit
		}
		// Deactivate the screen being popped.
		a.deactivateTop()
		// Start slide-out transition
		if a.width > 0 {
			a.transitioning = true
			a.transitionOffset = 0
			a.transitionVel = 0
			a.pendingPop = true
			return a, transitionTick()
		}
		a.screenStack = a.screenStack[:len(a.screenStack)-1]
		// Re-activate the new top screen.
		a.activateTop()
		return a, nil

	case replaceScreenMsg:
		a.deactivateTop()
		screen := a.constructScreen(pushScreenMsg(msg))
		if screen == nil {
			return a, nil
		}
		// Replace top of stack
		if len(a.screenStack) > 0 {
			a.screenStack[len(a.screenStack)-1] = screen
		} else {
			a.screenStack = []tea.Model{screen}
		}
		initCmd := screen.Init()

		// Start push-style transition
		if a.width > 0 {
			a.transitioning = true
			a.transitionOffset = float64(a.width)
			a.transitionVel = 0
			a.pendingPop = false
			return a, tea.Batch(initCmd, transitionTick())
		}
		return a, initCmd

	case authReloadMsg:
		// Reload credentials from config after successful inline setup
		cfg, err := config.Load()
		if err == nil && cfg.HasWallet() && cfg.HasAPIKeys() {
			a.hasAuth = true
			a.clobClient = clob.NewAuthenticatedClient(cfg.CLOBAPIURL, &clob.AuthCredentials{
				Key:        cfg.APIKey,
				Secret:     cfg.APISecret,
				Passphrase: cfg.Passphrase,
				Address:    deriveAddress(cfg.PrivateKey),
			})
			a.address = deriveAddress(cfg.PrivateKey)
		}
		// Pop wizard, replace with markets list
		if len(a.screenStack) > 1 {
			a.screenStack = a.screenStack[:len(a.screenStack)-1]
		}
		screen := a.constructScreen(pushScreenMsg{screen: ScreenMarketsList})
		if screen != nil {
			if len(a.screenStack) > 0 {
				a.screenStack[len(a.screenStack)-1] = screen
			} else {
				a.screenStack = []tea.Model{screen}
			}
			return a, screen.Init()
		}
		return a, nil

	case transitionTickMsg:
		if !a.transitioning {
			return a, nil
		}
		target := 0.0
		if a.pendingPop {
			target = float64(a.width)
		}
		a.transitionOffset, a.transitionVel = a.transitionSpring.Update(
			a.transitionOffset, a.transitionVel, target,
		)
		// Converged?
		if math.Abs(a.transitionOffset-target) < 0.5 {
			a.transitionOffset = target
			a.transitioning = false
			if a.pendingPop {
				a.pendingPop = false
				if len(a.screenStack) > 1 {
					a.screenStack = a.screenStack[:len(a.screenStack)-1]
				}
				// Re-activate the new top screen after pop animation.
				a.activateTop()
			}
			// Force a full repaint so the un-offset view is flushed to the terminal.
			// Without this, BubbleTea's diff renderer may hold a stale padded frame.
			return a, tea.ClearScreen
		}
		return a, transitionTick()
	}

	// Delegate to current screen
	if screen := a.currentScreen(); screen != nil {
		updated, cmd := screen.Update(msg)
		a.screenStack[len(a.screenStack)-1] = updated
		return a, cmd
	}
	return a, nil
}

func (a *App) View() string {
	screen := a.currentScreen()
	if screen == nil {
		return ""
	}

	view := screen.View()
	if a.transitioning {
		offset := int(a.transitionOffset)
		if offset > 0 && offset < a.width {
			// Pad each line with spaces from the left
			var sb strings.Builder
			pad := strings.Repeat(" ", offset)
			for i, line := range strings.Split(view, "\n") {
				if i > 0 {
					sb.WriteByte('\n')
				}
				sb.WriteString(pad)
				sb.WriteString(line)
			}
			view = sb.String()
		}
	}

	// Append status bar — skip for onboarding and screens still loading
	showStatusBar := a.width > 0
	switch s := screen.(type) {
	case *OnboardingModel:
		showStatusBar = false
	case *MarketsListModel:
		if s.loading {
			showStatusBar = false
		}
	}
	if showStatusBar {
		var sb strings.Builder
		sb.WriteString(view)
		sb.WriteByte('\n')
		sb.WriteString(a.renderStatusBar())
		view = sb.String()
	}

	return view
}

func (a *App) currentScreen() tea.Model {
	if len(a.screenStack) == 0 {
		return nil
	}
	return a.screenStack[len(a.screenStack)-1]
}

func (a *App) constructScreen(msg pushScreenMsg) tea.Model {
	switch msg.screen {
	case ScreenMarketDetail:
		return NewMarketDetailModel(msg.marketID, a.gammaClient, a.clobClient, a.width, a.height)
	case ScreenMarketsList:
		return NewMarketsListModel(a.gammaClient, a.width, a.height)
	case ScreenOrderBook:
		return NewOrderBookModel(msg.tokenID, msg.question, a.clobClient, a.width, a.height)
	case ScreenOrderForm:
		return NewOrderFormModel(msg.tokenID, msg.question, a.clobClient, a.width, a.height)
	case ScreenPortfolio:
		return NewPortfolioModel(a.address, a.clobClient, a.dataClient, a.width, a.height)
	case ScreenShell:
		return NewShellModel(a.execFn, a.width, a.height)
	case ScreenSetupWizard:
		wiz := NewSetupWizardModel(msg.deriveFn, a.width, a.height)
		wiz.inline = true
		return wiz
	case ScreenOnboarding:
		return NewOnboardingModel(a.deriveFn, a.width, a.height)
	}
	return nil
}

// isInputScreen returns true if the current screen captures text input or is the splash.
func (a *App) isInputScreen() bool {
	switch s := a.currentScreen().(type) {
	case *OrderFormModel, *ShellModel, *SetupWizardModel, *OnboardingModel:
		return true
	case *MarketsListModel:
		return s.Filtering()
	}
	return false
}

// screenTitle returns a display title for a screen (used in breadcrumb).
func (a *App) screenTitle(screen tea.Model) string {
	switch s := screen.(type) {
	case *MarketsListModel:
		return "Markets"
	case *MarketDetailModel:
		if s.market != nil {
			return Truncate(s.market.Question, 30)
		}
		return "Market"
	case *OrderBookModel:
		return "Order Book"
	case *OrderFormModel:
		return "Order"
	case *PortfolioModel:
		return "Portfolio"
	case *ShellModel:
		return "Shell"
	case *SetupWizardModel:
		return "Setup"
	case *OnboardingModel:
		return "Welcome"
	}
	return ""
}

// renderStatusBar renders the bottom chrome with breadcrumbs and auth info.
func (a *App) renderStatusBar() string {
	var crumbs []string
	for _, screen := range a.screenStack {
		if title := a.screenTitle(screen); title != "" {
			crumbs = append(crumbs, title)
		}
	}
	left := strings.Join(crumbs, " › ")

	right := "? help"
	if a.hasAuth && a.address != "" {
		addr := a.address
		if len(addr) > 10 {
			addr = addr[:6] + "…" + addr[len(addr)-4:]
		}
		right = addr + "  " + right
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := a.width - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	return StatusBarStyle.Width(a.width).Render(left + strings.Repeat(" ", gap) + right)
}

// ScreenCount returns the number of screens on the stack (for testing).
func (a *App) ScreenCount() int {
	return len(a.screenStack)
}

// Width returns the stored width (for testing).
func (a *App) Width() int {
	return a.width
}

// Height returns the stored height (for testing).
func (a *App) Height() int {
	return a.height
}

// Transitioning returns the current transition state (for testing).
func (a *App) Transitioning() bool {
	return a.transitioning
}

// activatable is implemented by screens that run background tick loops.
type activatable interface {
	SetActive(bool)
}

// deactivateTop marks the current top screen as inactive (stops its tick loops).
func (a *App) deactivateTop() {
	if len(a.screenStack) == 0 {
		return
	}
	if s, ok := a.screenStack[len(a.screenStack)-1].(activatable); ok {
		s.SetActive(false)
	}
}

// activateTop marks the current top screen as active (resumes its tick loops).
func (a *App) activateTop() {
	if len(a.screenStack) == 0 {
		return
	}
	if s, ok := a.screenStack[len(a.screenStack)-1].(activatable); ok {
		s.SetActive(true)
	}
}

func transitionTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return transitionTickMsg(t)
	})
}

// deriveAddress converts a hex private key to an Ethereum address string.
func deriveAddress(hexKey string) string {
	hexKey = strings.TrimPrefix(hexKey, "0x")
	ecdsaKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return ""
	}
	return crypto.PubkeyToAddress(ecdsaKey.PublicKey).Hex()
}
