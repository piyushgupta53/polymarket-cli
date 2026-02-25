package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOnboardingDefaultChoice(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)

	if m.Choice() != 0 {
		t.Errorf("Choice() = %d, want 0 (Just Explore)", m.Choice())
	}
}

func TestOnboardingExplore(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)

	// Default choice is 0 (Just Explore), press Enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}
	msg := cmd()
	rmsg, ok := msg.(replaceScreenMsg)
	if !ok {
		t.Fatalf("expected replaceScreenMsg, got %T", msg)
	}
	if rmsg.screen != ScreenMarketsList {
		t.Errorf("screen = %d, want ScreenMarketsList", rmsg.screen)
	}
}

func TestOnboardingSetup(t *testing.T) {
	dummyDerive := func(pk, sig string) (string, error) { return "0xabc", nil }
	m := NewOnboardingModel(dummyDerive, 80, 24)

	// Move down to "Set Up Wallet"
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Choice() != 1 {
		t.Fatalf("Choice() = %d, want 1 after down", m.Choice())
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}
	msg := cmd()
	pmsg, ok := msg.(pushScreenMsg)
	if !ok {
		t.Fatalf("expected pushScreenMsg, got %T", msg)
	}
	if pmsg.screen != ScreenSetupWizard {
		t.Errorf("screen = %d, want ScreenSetupWizard", pmsg.screen)
	}
	if pmsg.deriveFn == nil {
		t.Error("deriveFn should be passed through")
	}
}

func TestOnboardingNavigationWrap(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)

	// Already at 0, pressing up should stay at 0
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Choice() != 0 {
		t.Errorf("Choice() = %d, want 0 (should not wrap)", m.Choice())
	}

	// Go down twice — should stop at 1
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Choice() != 1 {
		t.Errorf("Choice() = %d, want 1 (should not wrap)", m.Choice())
	}
}

func TestOnboardingJKNavigation(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.Choice() != 1 {
		t.Errorf("Choice() = %d, want 1 after j", m.Choice())
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.Choice() != 0 {
		t.Errorf("Choice() = %d, want 0 after k", m.Choice())
	}
}

func TestOnboardingView(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)
	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestOnboardingInitReturnsTick(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)
	cmd := m.Init()

	if cmd == nil {
		t.Fatal("Init() should return a tick command, got nil")
	}
}

func TestOnboardingPhaseAdvances(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)

	if m.Phase() != 0 {
		t.Errorf("initial Phase() = %f, want 0", m.Phase())
	}

	updated, _ := m.Update(waveTickMsg(time.Now()))
	m = updated.(*OnboardingModel)

	if m.Phase() <= 0 {
		t.Errorf("Phase() = %f after tick, want > 0", m.Phase())
	}
}

func TestOnboardingViewContainsGradient(t *testing.T) {
	m := NewOnboardingModel(nil, 80, 24)
	view := m.View()

	if !strings.Contains(view, "█") {
		t.Error("View() should contain gradient block characters (█)")
	}
}

func TestRenderGradient(t *testing.T) {
	result := renderGradient(40, 0)

	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("renderGradient returned %d lines, want 3", len(lines))
	}

	if !strings.Contains(result, "█") {
		t.Error("renderGradient should contain full block characters (█)")
	}
	if !strings.Contains(result, "▄") {
		t.Error("renderGradient should contain half-block characters (▄)")
	}
	if !strings.Contains(result, "▀") {
		t.Error("renderGradient should contain half-block characters (▀)")
	}
}

func TestRenderGradientTooNarrow(t *testing.T) {
	result := renderGradient(3, 0)

	if result != "" {
		t.Errorf("renderGradient(3, 0) = %q, want empty string", result)
	}
}
