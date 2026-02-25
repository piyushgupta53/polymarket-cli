package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/auth"
	"github.com/piyushgupta/polymarket-cli/internal/config"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/piyushgupta/polymarket-cli/tui"
	"github.com/spf13/cobra"
)

var (
	setupKeyFlag     string
	setupKeyEnvFlag  string
	setupSigTypeFlag string
	setupShowFlag    bool
	setupResetFlag   bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up your wallet for trading",
	Long:  "Interactive setup wizard to configure your Polymarket wallet.\nImports your private key, derives API credentials, and saves config.",
	Example: `  # Interactive setup wizard
  polymarket setup

  # Non-interactive setup with a key
  polymarket setup --key 0xYOUR_PRIVATE_KEY

  # Agent: setup from environment variable
  polymarket setup --key-env POLYMARKET_PRIVATE_KEY -o json -q`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().StringVar(&setupKeyFlag, "key", "", "Private key (hex) for non-interactive setup")
	setupCmd.Flags().StringVar(&setupKeyEnvFlag, "key-env", "", "Environment variable name containing the private key")
	setupCmd.Flags().StringVar(&setupSigTypeFlag, "sig-type", "EOA", "Signature type (EOA, POLY_PROXY, POLY_GNOSIS_SAFE)")
	setupCmd.Flags().BoolVar(&setupShowFlag, "show", false, "Print current config as JSON (key redacted)")
	setupCmd.Flags().BoolVar(&setupResetFlag, "reset", false, "Clear wallet configuration")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	// --show: print current config
	if setupShowFlag {
		return runSetupShow()
	}

	// --reset: clear wallet config
	if setupResetFlag {
		return runSetupReset()
	}

	// --key-env: read key from named env var
	if setupKeyEnvFlag != "" {
		key := os.Getenv(setupKeyEnvFlag)
		if key == "" {
			return output.ErrInvalidInput(fmt.Sprintf("environment variable %s is empty or not set", setupKeyEnvFlag))
		}
		return runSetupHeadless(key, setupSigTypeFlag)
	}

	// --key: direct key
	if setupKeyFlag != "" {
		return runSetupHeadless(setupKeyFlag, setupSigTypeFlag)
	}

	// No key + non-interactive → error
	if IsNonInteractive() {
		return output.ErrInvalidInput("setup requires --key or --key-env in non-interactive mode")
	}

	// Interactive TUI wizard
	wizard := tui.NewSetupWizardModel(setupDeriveFn, 80, 24)
	p := tea.NewProgram(wizard, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("setup wizard: %w", err)
	}

	// Check result
	if m, ok := finalModel.(*tui.SetupWizardModel); ok {
		result := m.GetResult()
		if result.Saved {
			fmt.Printf("Wallet configured: %s\n", result.Address)
		}
	}
	return nil
}

// runSetupShow prints the current config as JSON with a redacted private key.
func runSetupShow() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Redact private key
	redactedKey := ""
	if cfg.PrivateKey != "" {
		k := cfg.PrivateKey
		if len(k) > 10 {
			redactedKey = k[:6] + "..." + k[len(k)-4:]
		} else {
			redactedKey = "***"
		}
	}

	data := map[string]any{
		"private_key":    redactedKey,
		"signature_type": cfg.SignatureType,
		"chain_id":       cfg.ChainID,
		"gamma_api_url":  cfg.GammaAPIURL,
		"clob_api_url":   cfg.CLOBAPIURL,
		"rpc_url":        cfg.RPCURL,
		"data_api_url":   cfg.DataAPIURL,
		"bridge_api_url": cfg.BridgeAPIURL,
		"has_api_keys":   cfg.HasAPIKeys(),
	}

	return output.PrintJSON(data)
}

// runSetupReset clears wallet configuration.
func runSetupReset() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg.ClearWallet()
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(map[string]string{"status": "reset"})
	}

	fmt.Println("Wallet configuration cleared.")
	return nil
}

// runSetupHeadless performs non-interactive setup using flags.
func runSetupHeadless(privateKey, sigType string) error {
	address, err := setupDeriveFn(privateKey, sigType)
	if err != nil {
		return err
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(map[string]any{
			"address":          address,
			"signature_type":   sigType,
			"api_keys_derived": true,
		})
	}

	fmt.Printf("Wallet configured: %s\n", address)
	return nil
}

// setupDeriveFn validates the key, derives API credentials, and saves config.
func setupDeriveFn(privateKey, sigType string) (string, error) {
	hexKey := strings.TrimPrefix(privateKey, "0x")

	ecdsaKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)

	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}

	cfg.PrivateKey = hexKey
	cfg.SignatureType = sigType

	// Clear old cached API keys
	cfg.APIKey = ""
	cfg.APISecret = ""
	cfg.Passphrase = ""

	// Derive API keys
	sigTypeInt := 0
	switch sigType {
	case "POLY_PROXY":
		sigTypeInt = 1
	case "POLY_GNOSIS_SAFE":
		sigTypeInt = 2
	}

	creds, err := auth.DeriveAPIKey(cfg.CLOBAPIURL, ecdsaKey, cfg.ChainID, sigTypeInt)
	if err != nil {
		return "", fmt.Errorf("deriving API keys: %w", err)
	}

	cfg.APIKey = creds.APIKey
	cfg.APISecret = creds.Secret
	cfg.Passphrase = creds.Passphrase

	if err := cfg.Save(); err != nil {
		return "", fmt.Errorf("saving config: %w", err)
	}

	return address.Hex(), nil
}
