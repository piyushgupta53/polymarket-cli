package cmd

import (
	"bytes"
	"io"
	"os"
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
// It redirects os.Stdout via os.Pipe to capture all output including fmt.Println
// calls that bypass cobra's output writer. This is safe because BubbleTea's
// renderer holds its own reference to the original stdout from program creation.
func shellExecFn(input string) (string, error) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return "", nil
	}

	// Create a pipe to capture all writes to os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	origStdout := os.Stdout
	os.Stdout = w

	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
	rootCmd.SetArgs(args)

	execErr := rootCmd.Execute()

	// Restore stdout before reading pipe
	os.Stdout = origStdout
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
	w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	r.Close()

	captured := buf.String()
	if execErr != nil && captured == "" {
		return "", execErr
	}
	return strings.TrimSuffix(captured, "\n"), execErr
}
