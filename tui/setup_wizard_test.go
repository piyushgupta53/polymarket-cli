package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSetupWizardInit(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	if m.GetStep() != stepWelcome {
		t.Errorf("expected stepWelcome, got %d", m.GetStep())
	}
}

func TestSetupWizardWelcomeView(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	view := m.View()
	if view == "" {
		t.Error("expected non-empty welcome view")
	}
}

func TestSetupWizardEnterGoesToForm(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	wiz := updated.(*SetupWizardModel)
	if wiz.GetStep() != stepForm {
		t.Errorf("expected stepForm after enter, got %d", wiz.GetStep())
	}
}

func TestSetupWizardEscPopsScreen(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command on esc")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestSetupWizardDerivedSuccess(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	// Simulate deriving step
	m.step = stepDeriving
	updated, _ := m.Update(derivedMsg{address: "0x1234"})
	wiz := updated.(*SetupWizardModel)
	if wiz.GetStep() != stepDone {
		t.Errorf("expected stepDone, got %d", wiz.GetStep())
	}
	result := wiz.GetResult()
	if result.Address != "0x1234" {
		t.Errorf("expected address 0x1234, got %s", result.Address)
	}
	if !result.Saved {
		t.Error("expected Saved=true")
	}
}

func TestSetupWizardDerivedError(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	m.step = stepDeriving
	updated, _ := m.Update(derivedMsg{err: fmt.Errorf("bad key")})
	wiz := updated.(*SetupWizardModel)
	if wiz.GetStep() != stepDone {
		t.Errorf("expected stepDone, got %d", wiz.GetStep())
	}
	if wiz.err == nil {
		t.Error("expected error to be set")
	}
}

func TestSetupWizardDoneEnterPops(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	m.step = stepDone
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter at done step")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestSetupWizardSpinnerTick(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	m.step = stepDeriving
	// Should not panic on spinner tick
	_, _ = m.Update(spinner.TickMsg{})
}

func TestSetupWizardWindowResize(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	wiz := updated.(*SetupWizardModel)
	if wiz.width != 120 || wiz.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", wiz.width, wiz.height)
	}
}

func TestSetupWizardDoneView(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	m.step = stepDone
	m.result.Address = "0xABC"
	m.result.SignatureType = "EOA"
	view := m.View()
	if view == "" {
		t.Error("expected non-empty done view")
	}
}

func TestSetupWizardDoneErrorView(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	m.step = stepDone
	m.err = fmt.Errorf("test error")
	view := m.View()
	if view == "" {
		t.Error("expected non-empty error view")
	}
}

func TestSetupWizardDeriveNilFn(t *testing.T) {
	m := NewSetupWizardModel(nil, 80, 24)
	cmd := m.derive()
	msg := cmd()
	if d, ok := msg.(derivedMsg); ok {
		if d.err == nil {
			t.Error("expected error with nil deriveFn")
		}
	} else {
		t.Errorf("expected derivedMsg, got %T", msg)
	}
}
