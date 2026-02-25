package cmd

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/api/data"
	"github.com/piyushgupta/polymarket-cli/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  "Launch an interactive terminal UI for browsing Polymarket markets.",
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	gammaClient := newGammaClient()
	clobClient := clob.NewClient("")
	dataClient := data.NewDataClient(getDataAPIURL())

	// Try to load auth for portfolio/trading features
	var address string
	var hasAuth bool
	authClient, _, err := requireAuth()
	if err == nil && authClient != nil {
		clobClient = authClient
		hasAuth = true
	}

	// Derive address from private key if available
	pk := resolvePrivateKey()
	if pk != "" {
		hexKey := strings.TrimPrefix(pk, "0x")
		if ecdsaKey, err := crypto.HexToECDSA(hexKey); err == nil {
			address = crypto.PubkeyToAddress(ecdsaKey.PublicKey).Hex()
		}
	}

	app := tui.NewApp(gammaClient, clobClient, dataClient, address, hasAuth, setupDeriveFn)
	app.SetExecFn(shellExecFn)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}
