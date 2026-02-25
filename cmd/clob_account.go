package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/auth"
	"github.com/piyushgupta/polymarket-cli/internal/config"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteAPIKeyFlag string

var clobAPIKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetAPIKeys()
		if err != nil {
			return fmt.Errorf("getting API keys: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobCreateAPIKeyCmd = &cobra.Command{
	Use:   "create-api-key",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		pk := resolvePrivateKey()
		if pk == "" {
			return fmt.Errorf("no wallet configured. Run 'polymarket wallet import' or set POLYMARKET_PRIVATE_KEY")
		}

		hexKey := pk
		if len(hexKey) >= 2 && hexKey[:2] == "0x" {
			hexKey = hexKey[2:]
		}

		ecdsaKey, err := crypto.HexToECDSA(hexKey)
		if err != nil {
			return fmt.Errorf("invalid private key: %w", err)
		}

		creds, err := auth.CreateAPIKey(cfg.CLOBAPIURL, ecdsaKey, cfg.ChainID)
		if err != nil {
			return fmt.Errorf("creating API key: %w", err)
		}

		if getOutputFormat() == "json" {
			return output.PrintJSON(creds)
		}

		fmt.Println(output.GreenStyle.Bold(true).Render("API key created"))
		fmt.Println("Key: " + creds.APIKey)
		fmt.Println(output.DimStyle.Render("Secret and passphrase not shown. Use JSON output to see full details."))
		return nil
	},
}

var clobDeleteAPIKeyCmd = &cobra.Command{
	Use:   "delete-api-key",
	Short: "Delete an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		if err := c.DeleteAPIKey(deleteAPIKeyFlag); err != nil {
			return fmt.Errorf("deleting API key: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"status": "deleted"})
		}
		fmt.Println(output.GreenStyle.Render("API key deleted."))
		return nil
	},
}

var clobNotificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetNotifications()
		if err != nil {
			return fmt.Errorf("getting notifications: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobDeleteNotificationsCmd = &cobra.Command{
	Use:   "delete-notifications",
	Short: "Delete all notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		if err := c.DeleteNotifications(); err != nil {
			return fmt.Errorf("deleting notifications: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"status": "deleted"})
		}
		fmt.Println(output.GreenStyle.Render("Notifications deleted."))
		return nil
	},
}

func init() {
	clobCmd.AddCommand(clobAPIKeysCmd)
	clobCmd.AddCommand(clobCreateAPIKeyCmd)

	clobDeleteAPIKeyCmd.Flags().StringVar(&deleteAPIKeyFlag, "key", "", "API key to delete (required)")
	clobDeleteAPIKeyCmd.MarkFlagRequired("key")
	clobCmd.AddCommand(clobDeleteAPIKeyCmd)

	clobCmd.AddCommand(clobNotificationsCmd)
	clobCmd.AddCommand(clobDeleteNotificationsCmd)
}
