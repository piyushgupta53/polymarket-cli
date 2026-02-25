package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/auth"
	"github.com/piyushgupta/polymarket-cli/internal/config"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	walletPKFlag      string
	walletSigTypeFlag string
	walletConfirmFlag bool
)

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Manage your Polymarket wallet",
	Long:  "Create, import, view, and reset your Ethereum wallet for Polymarket trading.",
}

var walletCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.HasWallet() {
			return fmt.Errorf("wallet already exists. Use 'wallet reset --confirm' first")
		}

		key, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("generating key: %w", err)
		}

		cfg.PrivateKey = fmt.Sprintf("%x", crypto.FromECDSA(key))
		cfg.SignatureType = "EOA"

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		address := crypto.PubkeyToAddress(key.PublicKey)

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"address": address.Hex()})
		}

		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true).Width(18)
		fmt.Println(output.GreenStyle.Bold(true).Render("Wallet created"))
		fmt.Println(label.Render("Address") + address.Hex())
		fmt.Println(output.DimStyle.Render("Private key saved to config. Never share your private key."))
		return nil
	},
}

var walletImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an existing wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		if walletPKFlag == "" {
			return fmt.Errorf("--private-key is required")
		}

		address, err := auth.PrivateKeyToAddress(walletPKFlag)
		if err != nil {
			return fmt.Errorf("invalid private key: %w", err)
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Strip 0x prefix for storage
		hexKey := walletPKFlag
		if len(hexKey) >= 2 && hexKey[:2] == "0x" {
			hexKey = hexKey[2:]
		}
		cfg.PrivateKey = hexKey

		if walletSigTypeFlag != "" {
			cfg.SignatureType = walletSigTypeFlag
		} else if cfg.SignatureType == "" {
			cfg.SignatureType = "EOA"
		}

		// Clear cached API keys since they're tied to the old wallet
		cfg.APIKey = ""
		cfg.APISecret = ""
		cfg.Passphrase = ""

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"address": address.Hex()})
		}

		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true).Width(18)
		fmt.Println(output.GreenStyle.Bold(true).Render("Wallet imported"))
		fmt.Println(label.Render("Address") + address.Hex())
		fmt.Println(label.Render("Signature Type") + cfg.SignatureType)
		return nil
	},
}

var walletAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Show wallet address",
	RunE: func(cmd *cobra.Command, args []string) error {
		pk := resolvePrivateKey()
		if pk == "" {
			return fmt.Errorf("no wallet configured. Run 'polymarket wallet import' or set POLYMARKET_PRIVATE_KEY")
		}

		address, err := auth.PrivateKeyToAddress(pk)
		if err != nil {
			return fmt.Errorf("invalid private key: %w", err)
		}

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"address": address.Hex()})
		}

		fmt.Println(address.Hex())
		return nil
	},
}

var walletShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show wallet details",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		pk := resolvePrivateKey()
		if pk == "" {
			return fmt.Errorf("no wallet configured. Run 'polymarket wallet import' or set POLYMARKET_PRIVATE_KEY")
		}

		address, err := auth.PrivateKeyToAddress(pk)
		if err != nil {
			return fmt.Errorf("invalid private key: %w", err)
		}

		sigType := cfg.SignatureType
		if sigType == "" {
			sigType = "EOA"
		}

		apiKeyCached := "no"
		if cfg.HasAPIKeys() {
			apiKeyCached = "yes"
		}

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{
				"address":        address.Hex(),
				"signature_type": sigType,
				"api_key_cached": apiKeyCached,
			})
		}

		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true).Width(22)
		fmt.Println(output.BoldStyle.Render("Wallet"))
		fmt.Println(label.Render("Address") + address.Hex())
		fmt.Println(label.Render("Signature Type") + sigType)
		fmt.Println(label.Render("API Keys Cached") + apiKeyCached)
		fmt.Println(label.Render("Chain ID") + fmt.Sprintf("%d", cfg.ChainID))
		return nil
	},
}

var walletResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear all wallet data from config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !walletConfirmFlag {
			return fmt.Errorf("this will delete your wallet from the config. Use --confirm to proceed")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfg.ClearWallet()

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"status": "wallet cleared"})
		}

		fmt.Println(output.YellowStyle.Render("Wallet data cleared from config."))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)

	walletCmd.AddCommand(walletCreateCmd)

	walletImportCmd.Flags().StringVar(&walletPKFlag, "private-key", "", "Private key in hex (required)")
	_ = walletImportCmd.MarkFlagRequired("private-key")
	walletImportCmd.Flags().StringVar(&walletSigTypeFlag, "signature-type", "", "Signature type (EOA, POLY_PROXY, POLY_GNOSIS_SAFE)")
	walletCmd.AddCommand(walletImportCmd)

	walletCmd.AddCommand(walletAddressCmd)
	walletCmd.AddCommand(walletShowCmd)

	walletResetCmd.Flags().BoolVar(&walletConfirmFlag, "confirm", false, "Confirm wallet reset")
	walletCmd.AddCommand(walletResetCmd)
}
