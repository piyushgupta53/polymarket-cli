package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/api/bridge"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Bridge deposits and supported assets",
	Long:  "View deposit addresses, supported assets, and deposit status for the Polymarket bridge.",
}

var bridgeDepositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Show deposit addresses",
	Long:  "Show deposit addresses for supported chains (EVM, Solana, Bitcoin).",
	RunE:  runBridgeDeposit,
}

var bridgeSupportedAssetsCmd = &cobra.Command{
	Use:   "supported-assets",
	Short: "List supported bridge assets",
	Long:  "List all supported chains, tokens, and contract addresses for bridging.",
	RunE:  runBridgeSupportedAssets,
}

var bridgeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check deposit status",
	Long:  "Check the status of a deposit by transaction hash.",
	RunE:  runBridgeStatus,
}

var bridgeStatusTxHash string

func init() {
	rootCmd.AddCommand(bridgeCmd)
	bridgeCmd.AddCommand(bridgeDepositCmd)
	bridgeCmd.AddCommand(bridgeSupportedAssetsCmd)
	bridgeCmd.AddCommand(bridgeStatusCmd)

	bridgeStatusCmd.Flags().StringVar(&bridgeStatusTxHash, "tx", "", "Transaction hash to check")
	bridgeStatusCmd.MarkFlagRequired("tx")
}

func newBridgeClient() *bridge.BridgeClient {
	return bridge.NewBridgeClient(getBridgeAPIURL())
}

func runBridgeDeposit(cmd *cobra.Command, args []string) error {
	client := newBridgeClient()

	log.Debug("Fetching deposit addresses")

	addrs, err := client.GetDepositAddresses()
	if err != nil {
		return fmt.Errorf("fetching deposit addresses: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(addrs)
	}

	output.PrintDepositAddresses(addrs)
	return nil
}

func runBridgeSupportedAssets(cmd *cobra.Command, args []string) error {
	client := newBridgeClient()

	log.Debug("Fetching supported assets")

	assets, err := client.GetSupportedAssets()
	if err != nil {
		return fmt.Errorf("fetching supported assets: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(assets)
	}

	output.PrintSupportedAssetsTable(assets)
	return nil
}

func runBridgeStatus(cmd *cobra.Command, args []string) error {
	client := newBridgeClient()

	log.Debug("Fetching deposit status", "tx", bridgeStatusTxHash)

	status, err := client.GetDepositStatus(bridgeStatusTxHash)
	if err != nil {
		return fmt.Errorf("fetching deposit status: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(status)
	}

	output.PrintDepositStatus(status)
	return nil
}
