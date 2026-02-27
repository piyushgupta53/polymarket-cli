package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/muesli/termenv"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/auth"
	"github.com/piyushgupta/polymarket-cli/internal/config"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	outputFormat   string
	verbose        bool
	apiURL         string
	privateKeyFlag string
	rpcURLFlag     string

	nonInteractive bool
	quiet          bool
	noHeaders      bool
	noColor        bool
)

// Version info set at build time.
var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

// SetVersionInfo is called from main.go with ldflags values.
func SetVersionInfo(version, commit, date string) {
	appVersion = version
	appCommit = commit
	appDate = date
	rootCmd.Version = version
}

var rootCmd = &cobra.Command{
	Use:     "polymarket",
	Short:   "Polymarket CLI — trade prediction markets from the terminal",
	Long:    "A full-featured CLI for interacting with Polymarket prediction markets.\nBrowse markets, place trades, and manage your portfolio.",
	Version: appVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Reset state that may have been set by a previous command in the shell REPL.
		log.SetLevel(log.InfoLevel)
		log.SetOutput(os.Stderr)

		if verbose {
			log.SetLevel(log.DebugLevel)
		}

		// Auto-detect non-interactive
		if os.Getenv("POLYMARKET_NON_INTERACTIVE") == "1" {
			nonInteractive = true
		}
		if !nonInteractive && !term.IsTerminal(int(os.Stdin.Fd())) {
			nonInteractive = true
		}

		// NO_COLOR env
		if os.Getenv("NO_COLOR") != "" {
			noColor = true
		}
		if noColor {
			lipgloss.SetColorProfile(termenv.Ascii)
		} else {
			lipgloss.SetColorProfile(termenv.ColorProfile())
		}

		// Quiet mode: suppress log output
		if quiet {
			log.SetOutput(io.Discard)
		}

		// Determine if TSV and compact JSON
		isTSV := outputFormat == "tsv"
		compactJSON := quiet

		output.SetOutputOptions(quiet, noHeaders, noColor, isTSV, compactJSON)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (table, json, tsv)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Override Gamma API base URL")
	rootCmd.PersistentFlags().StringVar(&privateKeyFlag, "private-key", "", "Ethereum private key (hex)")
	rootCmd.PersistentFlags().StringVar(&rpcURLFlag, "rpc-url", "", "Polygon RPC URL")
	rootCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "Disable interactive prompts")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-data output")
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Omit table headers")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	rootCmd.SetVersionTemplate("polymarket-cli {{.Version}}\n")

	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

// getOutputFormat returns the output format, auto-detecting from TTY if not set.
func getOutputFormat() string {
	if outputFormat != "" {
		return outputFormat
	}
	// Auto-detect: TTY → table, non-TTY (piped) → json
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "table"
	}
	return "json"
}

func getAPIURL() string {
	if apiURL != "" {
		return apiURL
	}
	if env := os.Getenv("POLYMARKET_API_URL"); env != "" {
		return env
	}
	return ""
}

func getChainRPCURL() string {
	if rpcURLFlag != "" {
		return rpcURLFlag
	}
	if env := os.Getenv("POLYMARKET_RPC_URL"); env != "" {
		return env
	}
	cfg, err := config.Load()
	if err != nil {
		return config.DefaultRPCURL
	}
	return cfg.RPCURL
}

func getDataAPIURL() string {
	if env := os.Getenv("POLYMARKET_DATA_API_URL"); env != "" {
		return env
	}
	cfg, err := config.Load()
	if err != nil {
		return config.DefaultDataAPIURL
	}
	return cfg.DataAPIURL
}

func getBridgeAPIURL() string {
	if env := os.Getenv("POLYMARKET_BRIDGE_API_URL"); env != "" {
		return env
	}
	cfg, err := config.Load()
	if err != nil {
		return config.DefaultBridgeAPIURL
	}
	return cfg.BridgeAPIURL
}

// IsNonInteractive returns true if interactive prompts should be suppressed.
func IsNonInteractive() bool {
	return nonInteractive
}

// IsQuiet returns true if non-data output should be suppressed.
func IsQuiet() bool {
	return quiet
}

// NoHeaders returns true if table headers should be omitted.
func NoHeaders() bool {
	return noHeaders
}

// GetResolvedOutputFormat returns the resolved output format string for use by main.go.
func GetResolvedOutputFormat() string {
	return getOutputFormat()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Example: `  # Print version
  polymarket version

  # Agent: get version as JSON
  polymarket version -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{
				"version": appVersion,
				"commit":  appCommit,
				"date":    appDate,
				"go":      runtime.Version(),
				"os":      runtime.GOOS + "/" + runtime.GOARCH,
			})
		}
		fmt.Printf("polymarket-cli %s (commit: %s, built: %s, %s)\n",
			appVersion, appCommit, appDate, runtime.Version())
		return nil
	},
}

// resolvePrivateKey returns the private key from flag, env, or config.
func resolvePrivateKey() string {
	if privateKeyFlag != "" {
		return privateKeyFlag
	}
	if env := os.Getenv("POLYMARKET_PRIVATE_KEY"); env != "" {
		return env
	}
	cfg, err := config.Load()
	if err != nil {
		return ""
	}
	return cfg.PrivateKey
}

// requireAuth loads config, resolves the private key, derives API keys if needed,
// and returns an authenticated CLOB client.
func requireAuth() (*clob.Client, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	pk := resolvePrivateKey()
	if pk == "" {
		return nil, nil, output.ErrAuthRequired("Run: polymarket setup --key <hex>")
	}

	// Strip 0x prefix for go-ethereum
	hexKey := pk
	if len(hexKey) >= 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}

	ecdsaKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, nil, output.ErrAuthFailed(err)
	}

	address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)
	log.Debug("wallet address", "address", address.Hex())

	// Derive API keys if not cached
	if !cfg.HasAPIKeys() {
		log.Debug("deriving API keys...")
		creds, err := auth.DeriveAPIKey(cfg.CLOBAPIURL, ecdsaKey, cfg.ChainID, cfg.GetSignatureTypeInt())
		if err != nil {
			return nil, nil, output.ErrAuthFailed(fmt.Errorf("deriving API keys: %w", err))
		}
		cfg.APIKey = creds.APIKey
		cfg.APISecret = creds.Secret
		cfg.Passphrase = creds.Passphrase
		if err := cfg.Save(); err != nil {
			log.Warn("could not cache API keys", "err", err)
		}
	}

	authCreds := &clob.AuthCredentials{
		Key:        cfg.APIKey,
		Secret:     cfg.APISecret,
		Passphrase: cfg.Passphrase,
		Address:    address.Hex(),
	}

	client := clob.NewAuthenticatedClient(cfg.CLOBAPIURL, authCreds)
	return client, cfg, nil
}

// GetVersion returns the current version string.
func GetVersion() string {
	return appVersion
}
