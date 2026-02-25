package cmd

import (
	"bytes"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/tui"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Launch interactive REPL",
	Long:  "Launch an interactive shell for running Polymarket CLI commands.",
	RunE:  runShell,
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func runShell(cmd *cobra.Command, args []string) error {
	gammaClient := newGammaClient()
	clobClient := clob.NewClient("")

	app := tui.NewApp(gammaClient, clobClient, nil, "", false, nil)
	app.SetExecFn(shellExecFn)

	// Start directly on the shell screen
	p := tea.NewProgram(
		&shellLauncher{app: app},
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}

// shellLauncher wraps App to start on the shell screen.
type shellLauncher struct {
	app *tui.App
}

func (s *shellLauncher) Init() tea.Cmd {
	return s.app.Init()
}

func (s *shellLauncher) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s.app.Update(msg)
}

func (s *shellLauncher) View() string {
	return s.app.View()
}

// shellExecFn dispatches a shell input line to the Cobra command tree.
func shellExecFn(input string) (string, error) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return "", nil
	}

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	output := buf.String()
	if err != nil && output == "" {
		return "", err
	}
	return strings.TrimSuffix(output, "\n"), err
}
