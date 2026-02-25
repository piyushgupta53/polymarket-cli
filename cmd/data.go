package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/api/data"
	"github.com/piyushgupta/polymarket-cli/internal/auth"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Query portfolio, trades, and market data",
	Long:  "Access the Polymarket Data API for positions, trades, activity, holders, leaderboards, and more.",
}

// Flags
var (
	dataAddress string
	dataLimit   int
	dataOffset  int
)

func init() {
	rootCmd.AddCommand(dataCmd)

	// Subcommands
	dataCmd.AddCommand(dataPositionsCmd)
	dataCmd.AddCommand(dataClosedCmd)
	dataCmd.AddCommand(dataValueCmd)
	dataCmd.AddCommand(dataTradedCmd)
	dataCmd.AddCommand(dataTradesCmd)
	dataCmd.AddCommand(dataActivityCmd)
	dataCmd.AddCommand(dataHoldersCmd)
	dataCmd.AddCommand(dataOpenInterestCmd)
	dataCmd.AddCommand(dataVolumeCmd)
	dataCmd.AddCommand(dataLeaderboardCmd)
	dataCmd.AddCommand(dataBuilderLeaderboardCmd)
	dataCmd.AddCommand(dataBuilderVolumeCmd)

	// Address flag for address-based commands
	for _, cmd := range []*cobra.Command{
		dataPositionsCmd, dataClosedCmd, dataValueCmd, dataTradedCmd,
		dataTradesCmd, dataActivityCmd,
	} {
		cmd.Flags().StringVar(&dataAddress, "address", "", "Wallet address (defaults to configured wallet)")
	}

	// Limit/offset for paginated commands
	for _, cmd := range []*cobra.Command{
		dataTradesCmd, dataActivityCmd, dataLeaderboardCmd,
	} {
		cmd.Flags().IntVarP(&dataLimit, "limit", "l", 25, "Maximum number of results")
		cmd.Flags().IntVar(&dataOffset, "offset", 0, "Offset for pagination")
	}
}

func newDataClient() *data.DataClient {
	baseURL := os.Getenv("POLYMARKET_DATA_API_URL")
	if baseURL == "" {
		baseURL = getDataAPIURL()
	}
	return data.NewDataClient(baseURL)
}

func resolveAddress(flagAddr string) string {
	if flagAddr != "" {
		return flagAddr
	}
	pk := resolvePrivateKey()
	if pk == "" {
		return ""
	}
	addr, err := auth.PrivateKeyToAddress(pk)
	if err != nil {
		return ""
	}
	return addr.Hex()
}

// --- Positions ---

var dataPositionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "List open positions",
	Long:  "List open positions for a wallet address.",
	Example: `  # List your open positions
  polymarket data positions

  # Agent: get positions as JSON for a specific address
  polymarket data positions --address 0xABC... -o json -q`,
	RunE: runDataPositions,
}

func runDataPositions(cmd *cobra.Command, args []string) error {
	addr := resolveAddress(dataAddress)
	if addr == "" {
		return fmt.Errorf("no address specified. Use --address or configure a wallet")
	}

	client := newDataClient()
	log.Debug("Getting positions", "address", addr)

	positions, err := client.GetPositions(addr)
	if err != nil {
		return fmt.Errorf("getting positions: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(positions)
	}
	output.PrintPositionsTable(positions)
	return nil
}

// --- Closed Positions ---

var dataClosedCmd = &cobra.Command{
	Use:   "closed",
	Short: "List closed positions",
	Long:  "List closed positions for a wallet address.",
	RunE:  runDataClosed,
}

func runDataClosed(cmd *cobra.Command, args []string) error {
	addr := resolveAddress(dataAddress)
	if addr == "" {
		return fmt.Errorf("no address specified. Use --address or configure a wallet")
	}

	client := newDataClient()
	log.Debug("Getting closed positions", "address", addr)

	positions, err := client.GetClosedPositions(addr)
	if err != nil {
		return fmt.Errorf("getting closed positions: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(positions)
	}
	output.PrintPositionsTable(positions)
	return nil
}

// --- Portfolio Value ---

var dataValueCmd = &cobra.Command{
	Use:   "value",
	Short: "Show portfolio value history",
	Long:  "Show portfolio value history for a wallet address.",
	RunE:  runDataValue,
}

func runDataValue(cmd *cobra.Command, args []string) error {
	addr := resolveAddress(dataAddress)
	if addr == "" {
		return fmt.Errorf("no address specified. Use --address or configure a wallet")
	}

	client := newDataClient()
	log.Debug("Getting portfolio value", "address", addr)

	values, err := client.GetPortfolioValue(addr)
	if err != nil {
		return fmt.Errorf("getting portfolio value: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(values)
	}

	// Print as simple value list
	if len(values) == 0 {
		fmt.Println(output.DimStyle.Render("No portfolio value data found."))
		return nil
	}
	for _, v := range values {
		fmt.Printf("%s  %s\n", output.FormatTimestamp(v.Timestamp), output.FormatVolumeFloat(v.Value))
	}
	return nil
}

// --- Portfolio Traded ---

var dataTradedCmd = &cobra.Command{
	Use:   "traded",
	Short: "Show total volume traded",
	Long:  "Show total volume traded by a wallet address.",
	RunE:  runDataTraded,
}

func runDataTraded(cmd *cobra.Command, args []string) error {
	addr := resolveAddress(dataAddress)
	if addr == "" {
		return fmt.Errorf("no address specified. Use --address or configure a wallet")
	}

	client := newDataClient()
	log.Debug("Getting portfolio traded", "address", addr)

	raw, err := client.GetPortfolioTraded(addr)
	if err != nil {
		return fmt.Errorf("getting portfolio traded: %w", err)
	}

	return output.PrintRawJSON(raw)
}

// --- Trades ---

var dataTradesCmd = &cobra.Command{
	Use:   "trades",
	Short: "List trades",
	Long:  "List trades for a wallet address.",
	Example: `  # List recent trades
  polymarket data trades --limit 10

  # Agent: get trades as JSON
  polymarket data trades -o json -q`,
	RunE: runDataTrades,
}

func runDataTrades(cmd *cobra.Command, args []string) error {
	addr := resolveAddress(dataAddress)
	if addr == "" {
		return fmt.Errorf("no address specified. Use --address or configure a wallet")
	}

	client := newDataClient()
	log.Debug("Getting trades", "address", addr, "limit", dataLimit, "offset", dataOffset)

	trades, err := client.GetTrades(addr, dataLimit, dataOffset)
	if err != nil {
		return fmt.Errorf("getting trades: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(trades)
	}
	output.PrintTradesRecordTable(trades)
	return nil
}

// --- Activity ---

var dataActivityCmd = &cobra.Command{
	Use:   "activity",
	Short: "List account activity",
	Long:  "List account activity for a wallet address.",
	RunE:  runDataActivity,
}

func runDataActivity(cmd *cobra.Command, args []string) error {
	addr := resolveAddress(dataAddress)
	if addr == "" {
		return fmt.Errorf("no address specified. Use --address or configure a wallet")
	}

	client := newDataClient()
	log.Debug("Getting activity", "address", addr, "limit", dataLimit, "offset", dataOffset)

	activities, err := client.GetActivity(addr, dataLimit, dataOffset)
	if err != nil {
		return fmt.Errorf("getting activity: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(activities)
	}
	output.PrintActivityTable(activities)
	return nil
}

// --- Holders ---

var dataHoldersCmd = &cobra.Command{
	Use:   "holders <condition-id>",
	Short: "List top holders for a market",
	Long:  "List top holders for a market condition.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataHolders,
}

func runDataHolders(cmd *cobra.Command, args []string) error {
	client := newDataClient()
	conditionID := args[0]

	log.Debug("Getting holders", "conditionId", conditionID)

	holders, err := client.GetHolders(conditionID)
	if err != nil {
		return fmt.Errorf("getting holders: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(holders)
	}
	output.PrintHoldersTable(holders)
	return nil
}

// --- Open Interest ---

var dataOpenInterestCmd = &cobra.Command{
	Use:   "open-interest <condition-id>",
	Short: "Show open interest for a market",
	Long:  "Show open interest for a market condition.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataOpenInterest,
}

func runDataOpenInterest(cmd *cobra.Command, args []string) error {
	client := newDataClient()
	conditionID := args[0]

	log.Debug("Getting open interest", "conditionId", conditionID)

	oi, err := client.GetOpenInterest(conditionID)
	if err != nil {
		return fmt.Errorf("getting open interest: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(oi)
	}

	labelStyle := output.PurpleStyle.Bold(true).Width(18)
	fmt.Println(output.BoldStyle.Render("Open Interest"))
	fmt.Println(labelStyle.Render("Condition ID") + oi.ConditionID)
	fmt.Println(labelStyle.Render("Open Interest") + output.FormatVolumeFloat(oi.OpenInterest))
	return nil
}

// --- Event Volume ---

var dataVolumeCmd = &cobra.Command{
	Use:   "volume <event-id>",
	Short: "Show volume for an event",
	Long:  "Show volume data for an event.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataVolume,
}

func runDataVolume(cmd *cobra.Command, args []string) error {
	client := newDataClient()
	eventID := args[0]

	log.Debug("Getting event volume", "eventId", eventID)

	vol, err := client.GetEventVolume(eventID)
	if err != nil {
		return fmt.Errorf("getting event volume: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(vol)
	}

	labelStyle := output.PurpleStyle.Bold(true).Width(18)
	fmt.Println(output.BoldStyle.Render("Event Volume"))
	if vol.EventTitle != "" {
		fmt.Println(labelStyle.Render("Event") + vol.EventTitle)
	}
	fmt.Println(labelStyle.Render("Volume") + output.FormatVolumeFloat(vol.Volume))
	fmt.Println(labelStyle.Render("Volume (24h)") + output.FormatVolumeFloat(vol.Volume24hr))
	return nil
}

// --- Leaderboard ---

var dataLeaderboardCmd = &cobra.Command{
	Use:   "leaderboard",
	Short: "Show the trading leaderboard",
	Long:  "Show the Polymarket trading leaderboard.",
	RunE:  runDataLeaderboard,
}

func runDataLeaderboard(cmd *cobra.Command, args []string) error {
	client := newDataClient()

	log.Debug("Getting leaderboard", "limit", dataLimit, "offset", dataOffset)

	entries, err := client.GetLeaderboard(dataLimit, dataOffset)
	if err != nil {
		return fmt.Errorf("getting leaderboard: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(entries)
	}
	output.PrintLeaderboardTable(entries)
	return nil
}

// --- Builder Leaderboard ---

var dataBuilderLeaderboardCmd = &cobra.Command{
	Use:   "builder-leaderboard",
	Short: "Show the builder leaderboard",
	Long:  "Show the Polymarket builder leaderboard.",
	RunE:  runDataBuilderLeaderboard,
}

func runDataBuilderLeaderboard(cmd *cobra.Command, args []string) error {
	client := newDataClient()

	log.Debug("Getting builder leaderboard")

	raw, err := client.GetBuilderLeaderboard()
	if err != nil {
		return fmt.Errorf("getting builder leaderboard: %w", err)
	}

	return output.PrintRawJSON(raw)
}

// --- Builder Volume ---

var dataBuilderVolumeCmd = &cobra.Command{
	Use:   "builder-volume",
	Short: "Show builder volume data",
	Long:  "Show builder volume data from the Data API.",
	RunE:  runDataBuilderVolume,
}

func runDataBuilderVolume(cmd *cobra.Command, args []string) error {
	client := newDataClient()

	log.Debug("Getting builder volume")

	raw, err := client.GetBuilderVolume()
	if err != nil {
		return fmt.Errorf("getting builder volume: %w", err)
	}

	return output.PrintRawJSON(raw)
}
