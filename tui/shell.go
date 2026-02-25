package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxShellHistory    = 1000
	maxOutputLines     = 5000
)

// ShellModel is the interactive REPL screen.
type ShellModel struct {
	input      textinput.Model
	viewport   viewport.Model
	history    []string
	historyIdx int
	outputLines []string
	width      int
	height     int

	// execFn is the function that executes a command string.
	// It's injected so the shell can dispatch to Cobra commands.
	execFn func(input string) (string, error)
}

// NewShellModel creates a new shell REPL screen.
func NewShellModel(execFn func(string) (string, error), w, h int) *ShellModel {
	ti := textinput.New()
	ti.Prompt = "polymarket> "
	ti.PromptStyle = PromptStyle
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = w - 20

	vp := viewport.New(w-4, h-6)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	return &ShellModel{
		input:       ti,
		viewport:    vp,
		history:     make([]string, 0, 64),
		historyIdx:  -1,
		outputLines: make([]string, 0, 128),
		width:       w,
		height:      h,
		execFn:      execFn,
	}
}

func (m *ShellModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			input := strings.TrimSpace(m.input.Value())
			if input == "" {
				return m, nil
			}

			// Handle built-in commands
			if input == "quit" || input == "exit" {
				return m, func() tea.Msg { return popScreenMsg{} }
			}
			if input == "clear" {
				m.outputLines = m.outputLines[:0]
				m.viewport.SetContent("")
				m.input.SetValue("")
				return m, nil
			}
			if input == "help" {
				m.appendOutput(shellHelp())
				m.input.SetValue("")
				return m, nil
			}

			// Add to history (capped)
			m.history = append(m.history, input)
			if len(m.history) > maxShellHistory {
				copy(m.history, m.history[len(m.history)-maxShellHistory:])
				m.history = m.history[:maxShellHistory]
			}
			m.historyIdx = len(m.history)

			// Execute command
			m.appendOutput(PromptStyle.Render("polymarket> ") + input)
			return m, m.executeCommand(input)

		case "up":
			if len(m.history) > 0 && m.historyIdx > 0 {
				m.historyIdx--
				m.input.SetValue(m.history[m.historyIdx])
				m.input.CursorEnd()
			}
			return m, nil

		case "down":
			if m.historyIdx < len(m.history)-1 {
				m.historyIdx++
				m.input.SetValue(m.history[m.historyIdx])
				m.input.CursorEnd()
			} else {
				m.historyIdx = len(m.history)
				m.input.SetValue("")
			}
			return m, nil

		case "ctrl+l":
			m.outputLines = m.outputLines[:0]
			m.viewport.SetContent("")
			return m, nil

		case "ctrl+c":
			return m, func() tea.Msg { return popScreenMsg{} }
		}

	case shellOutputMsg:
		m.appendOutput(msg.output)
		m.input.SetValue("")
		return m, nil

	case shellErrorMsg:
		m.appendOutput(ErrorStyle.Render("Error: " + msg.err.Error()))
		m.input.SetValue("")
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		m.input.Width = msg.Width - 20
		return m, nil
	}

	// Forward to text input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *ShellModel) View() string {
	header := HeaderStyle.Render("Polymarket Shell")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewport.View(),
		m.input.View(),
		HelpStyle.Render("enter run │ up/down history │ ctrl+l clear │ ctrl+c exit"),
	)
	return AppStyle.Render(content)
}

func (m *ShellModel) appendOutput(s string) {
	m.outputLines = append(m.outputLines, s)
	if len(m.outputLines) > maxOutputLines {
		m.outputLines = m.outputLines[len(m.outputLines)-maxOutputLines:]
	}
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
}

func (m *ShellModel) executeCommand(input string) tea.Cmd {
	return func() tea.Msg {
		if m.execFn == nil {
			return shellErrorMsg{err: fmt.Errorf("command execution not configured")}
		}
		out, err := m.execFn(input)
		if err != nil {
			return shellErrorMsg{err: err}
		}
		return shellOutputMsg{output: out}
	}
}

func shellHelp() string {
	return DimStyle.Render(`Available commands:
  Any polymarket CLI command (e.g. "markets list", "clob price ...")

Built-in:
  help     Show this help
  clear    Clear output
  quit     Exit shell`)
}
