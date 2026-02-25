package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestShellInit(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()

	if len(m.history) != 0 {
		t.Errorf("history should be empty, got %d items", len(m.history))
	}
}

func TestShellExecuteCommand(t *testing.T) {
	execFn := func(input string) (string, error) {
		return "output: " + input, nil
	}
	m := NewShellModel(execFn, 80, 40)
	m.Init()

	// Type a command
	m.input.SetValue("markets list")

	// Press enter
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ShellModel)

	if len(m.history) != 1 {
		t.Errorf("history count = %d, want 1", len(m.history))
	}
	if m.history[0] != "markets list" {
		t.Errorf("history[0] = %q, want %q", m.history[0], "markets list")
	}

	// Execute the command
	if cmd == nil {
		t.Fatal("expected command from enter")
	}
	msg := cmd()
	if outMsg, ok := msg.(shellOutputMsg); ok {
		if outMsg.output != "output: markets list" {
			t.Errorf("output = %q, want %q", outMsg.output, "output: markets list")
		}
	}
}

func TestShellHistory(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()

	m.history = []string{"first", "second", "third"}
	m.historyIdx = 3

	// Press up
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*ShellModel)
	if m.historyIdx != 2 {
		t.Errorf("historyIdx = %d, want 2", m.historyIdx)
	}

	// Press up again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*ShellModel)
	if m.historyIdx != 1 {
		t.Errorf("historyIdx = %d, want 1", m.historyIdx)
	}

	// Press down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*ShellModel)
	if m.historyIdx != 2 {
		t.Errorf("historyIdx = %d, want 2", m.historyIdx)
	}
}

func TestShellClear(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()
	m.output.WriteString("some output")

	m.input.SetValue("clear")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ShellModel)

	if m.output.Len() != 0 {
		t.Errorf("output should be cleared, got %d bytes", m.output.Len())
	}
}

func TestShellQuit(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()

	m.input.SetValue("quit")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command from 'quit'")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestShellCtrlC(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("expected command from ctrl+c")
	}
	msg := cmd()
	if _, ok := msg.(popScreenMsg); !ok {
		t.Errorf("expected popScreenMsg, got %T", msg)
	}
}

func TestShellError(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()

	updated, _ := m.Update(shellErrorMsg{err: fmt.Errorf("command not found")})
	m = updated.(*ShellModel)

	// Error should be in the output
	if m.output.Len() == 0 {
		t.Error("expected error output to be appended")
	}
}

func TestShellResize(t *testing.T) {
	m := NewShellModel(nil, 80, 40)
	m.Init()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = updated.(*ShellModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}
