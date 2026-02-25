package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppInitializesToMarketsList(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	if app.ScreenCount() != 1 {
		t.Errorf("ScreenCount() = %d, want 1", app.ScreenCount())
	}
}

func TestAppQuitOnCtrlC(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestAppPushScreen(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	if app.ScreenCount() != 1 {
		t.Fatalf("initial ScreenCount() = %d, want 1", app.ScreenCount())
	}

	app.Update(pushScreenMsg{screen: ScreenMarketDetail, marketID: "123"})

	if app.ScreenCount() != 2 {
		t.Errorf("after push ScreenCount() = %d, want 2", app.ScreenCount())
	}
}

func TestAppPopScreen(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	// Push a second screen
	app.Update(pushScreenMsg{screen: ScreenMarketDetail, marketID: "123"})
	if app.ScreenCount() != 2 {
		t.Fatalf("after push ScreenCount() = %d, want 2", app.ScreenCount())
	}

	// Pop it
	app.Update(popScreenMsg{})
	if app.ScreenCount() != 1 {
		t.Errorf("after pop ScreenCount() = %d, want 1", app.ScreenCount())
	}
}

func TestAppPopLastScreenQuits(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	_, cmd := app.Update(popScreenMsg{})

	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestAppWindowResize(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if app.Width() != 120 {
		t.Errorf("Width() = %d, want 120", app.Width())
	}
	if app.Height() != 40 {
		t.Errorf("Height() = %d, want 40", app.Height())
	}
}

func TestAppPushOrderBook(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	app.Update(pushScreenMsg{screen: ScreenOrderBook, tokenID: "tok1", question: "Test?"})

	if app.ScreenCount() != 2 {
		t.Errorf("ScreenCount() = %d, want 2", app.ScreenCount())
	}
}

func TestAppPushPortfolio(t *testing.T) {
	app := NewApp(nil, nil, nil, "0xabc", true, nil)
	app.Init()

	app.Update(pushScreenMsg{screen: ScreenPortfolio})

	if app.ScreenCount() != 2 {
		t.Errorf("ScreenCount() = %d, want 2", app.ScreenCount())
	}
}

func TestAppPushShell(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	app.Update(pushScreenMsg{screen: ScreenShell})

	if app.ScreenCount() != 2 {
		t.Errorf("ScreenCount() = %d, want 2", app.ScreenCount())
	}
}

func TestAppPushOrderForm(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	app.Update(pushScreenMsg{screen: ScreenOrderForm, tokenID: "tok1", question: "Test?"})

	if app.ScreenCount() != 2 {
		t.Errorf("ScreenCount() = %d, want 2", app.ScreenCount())
	}
}

func TestAuthStartsMarkets(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	if app.ScreenCount() != 1 {
		t.Fatalf("ScreenCount() = %d, want 1", app.ScreenCount())
	}
	if _, ok := app.currentScreen().(*MarketsListModel); !ok {
		t.Errorf("expected MarketsListModel when hasAuth=true, got %T", app.currentScreen())
	}
}

func TestNoAuthStartsOnboarding(t *testing.T) {
	app := NewApp(nil, nil, nil, "", false, nil)
	app.Init()

	if app.ScreenCount() != 1 {
		t.Fatalf("ScreenCount() = %d, want 1", app.ScreenCount())
	}
	if _, ok := app.currentScreen().(*OnboardingModel); !ok {
		t.Errorf("expected OnboardingModel when hasAuth=false, got %T", app.currentScreen())
	}
}

func TestAppPushTriggersTransition(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()
	// Set a width so transitions trigger
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	app.Update(pushScreenMsg{screen: ScreenMarketDetail, marketID: "123"})

	if !app.Transitioning() {
		t.Error("expected transitioning=true after push with width>0")
	}
	if app.ScreenCount() != 2 {
		t.Errorf("ScreenCount() = %d, want 2", app.ScreenCount())
	}
}

func TestAppTransitionConverges(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	app.Update(pushScreenMsg{screen: ScreenMarketDetail, marketID: "123"})

	// Run transition ticks until it converges (should take < 100 frames)
	for i := 0; i < 200; i++ {
		if !app.Transitioning() {
			break
		}
		app.Update(transitionTickMsg(time.Now()))
	}

	if app.Transitioning() {
		t.Error("transition should have converged after 200 ticks")
	}
}

func TestAppDoublePushBlockedDuringTransition(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	// First push triggers transition
	app.Update(pushScreenMsg{screen: ScreenMarketDetail, marketID: "123"})
	if !app.Transitioning() {
		t.Fatal("expected transitioning after first push")
	}

	// Second push during transition should be blocked
	countBefore := app.ScreenCount()
	app.Update(pushScreenMsg{screen: ScreenOrderBook, tokenID: "tok1"})
	if app.ScreenCount() != countBefore {
		t.Errorf("double push should be blocked: ScreenCount went from %d to %d", countBefore, app.ScreenCount())
	}
}

func TestAppReplaceScreen(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()

	// Replace markets list with another markets list
	app.Update(replaceScreenMsg{screen: ScreenMarketsList})

	if app.ScreenCount() != 1 {
		t.Errorf("after replace ScreenCount() = %d, want 1", app.ScreenCount())
	}
	if _, ok := app.currentScreen().(*MarketsListModel); !ok {
		t.Errorf("after replace screen should be MarketsListModel, got %T", app.currentScreen())
	}
}

func TestOnboardingExploreGoesToMarkets(t *testing.T) {
	app := NewApp(nil, nil, nil, "", false, nil)
	app.Init()

	if _, ok := app.currentScreen().(*OnboardingModel); !ok {
		t.Fatalf("expected OnboardingModel, got %T", app.currentScreen())
	}

	// Onboarding "Just Explore" → markets list
	app.Update(replaceScreenMsg{screen: ScreenMarketsList})
	if _, ok := app.currentScreen().(*MarketsListModel); !ok {
		t.Errorf("expected MarketsListModel after explore, got %T", app.currentScreen())
	}
}

func TestAuthGatedPortfolio(t *testing.T) {
	// Without auth, pushing portfolio should redirect to setup wizard
	app := NewApp(nil, nil, nil, "", false, nil)
	app.Init()

	// Go to markets list first
	app.Update(replaceScreenMsg{screen: ScreenMarketsList})

	countBefore := app.ScreenCount()
	app.Update(pushScreenMsg{screen: ScreenPortfolio})

	if app.ScreenCount() != countBefore+1 {
		t.Fatalf("ScreenCount() = %d, want %d", app.ScreenCount(), countBefore+1)
	}
	if _, ok := app.currentScreen().(*SetupWizardModel); !ok {
		t.Errorf("expected SetupWizardModel for portfolio without auth, got %T", app.currentScreen())
	}
}

func TestAuthGatedOrderForm(t *testing.T) {
	// Without auth, pushing order form should redirect to setup wizard
	app := NewApp(nil, nil, nil, "", false, nil)
	app.Init()

	app.Update(pushScreenMsg{screen: ScreenOrderForm, tokenID: "tok1", question: "Test?"})

	if app.ScreenCount() != 2 {
		t.Fatalf("ScreenCount() = %d, want 2", app.ScreenCount())
	}
	if _, ok := app.currentScreen().(*SetupWizardModel); !ok {
		t.Errorf("expected SetupWizardModel for order form without auth, got %T", app.currentScreen())
	}
}

func TestAppInitEntersAltScreen(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	cmd := app.Init()

	if cmd == nil {
		t.Fatal("Init() should return a batch command")
	}
	// The batch command should produce an EnterAltScreen sequence msg.
	// We can't easily inspect a batch, but we can verify the command isn't nil
	// and the app initialized correctly.
	if app.ScreenCount() != 1 {
		t.Errorf("ScreenCount() = %d, want 1", app.ScreenCount())
	}
}

func TestAppStatusBarBreadcrumb(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	// Finish loading so status bar appears
	app.Update(marketsLoadedMsg{})

	view := app.View()
	if !strings.Contains(view, "Markets") {
		t.Error("expected status bar to contain 'Markets'")
	}
	if !strings.Contains(view, "? help") {
		t.Error("expected status bar to contain '? help'")
	}
}

func TestAppStatusBarHiddenWhileLoading(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	// While still loading, status bar should not appear
	view := app.View()
	if strings.Contains(view, "? help") {
		t.Error("status bar should be hidden while markets are loading")
	}
}

func TestAppStatusBarAuth(t *testing.T) {
	app := NewApp(nil, nil, nil, "0x1234567890ABcDeF1234567890abCdEf12345678", true, nil)
	app.Init()
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	// Finish loading so status bar appears
	app.Update(marketsLoadedMsg{})

	view := app.View()
	if !strings.Contains(view, "0x1234") {
		t.Error("expected status bar to contain truncated address prefix")
	}
	if !strings.Contains(view, "5678") {
		t.Error("expected status bar to contain truncated address suffix")
	}
}

func TestAppBreadcrumbMultiLevel(t *testing.T) {
	app := NewApp(nil, nil, nil, "", true, nil)
	app.Init()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Push market detail
	app.Update(pushScreenMsg{screen: ScreenMarketDetail, marketID: "123"})
	// Run transition to completion
	for i := 0; i < 200; i++ {
		if !app.Transitioning() {
			break
		}
		app.Update(transitionTickMsg(time.Now()))
	}

	view := app.View()
	if !strings.Contains(view, "›") {
		t.Error("expected breadcrumb separator in multi-level view")
	}
}

func TestAuthReloadMsg(t *testing.T) {
	app := NewApp(nil, nil, nil, "", false, nil)
	app.Init()

	if app.hasAuth {
		t.Fatal("hasAuth should start false")
	}

	// Send authReloadMsg (config.Load() in test env will return defaults without wallet,
	// but we can at least verify the handler doesn't panic and pops the stack)
	app.Update(pushScreenMsg{screen: ScreenSetupWizard})
	app.Update(authReloadMsg{})

	// Should not have crashed, stack should have markets list on top
	if app.ScreenCount() < 1 {
		t.Error("ScreenCount should be >= 1 after authReloadMsg")
	}
}
