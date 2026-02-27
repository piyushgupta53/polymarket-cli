package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/api"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var marketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "Browse and search prediction markets",
	Long:  "List, search, and view details of Polymarket prediction markets.",
}

var marketsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List markets",
	Long:  "List prediction markets with optional filters.",
	Example: `  # List active markets
  polymarket markets list --active --limit 10

  # Agent: list markets as JSON
  polymarket markets list --active -o json -q`,
	RunE: runMarketsList,
}

var marketsGetCmd = &cobra.Command{
	Use:   "get <id-or-slug>",
	Short: "Get market details",
	Long:  "Get detailed information about a specific market by ID or slug.",
	Example: `  # Get market by slug
  polymarket markets get will-trump-win-2024

  # Agent: get market JSON by ID
  polymarket markets get 12345 -o json -q`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketsGet,
}

var marketsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search markets",
	Long:  "Search for markets by keyword (client-side filtering).",
	Example: `  # Search for election markets
  polymarket markets search "election"

  # Agent: search and get JSON
  polymarket markets search "bitcoin" -o json -q`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketsSearch,
}

var (
	marketsLimit  int
	marketsOffset int
	marketsActive bool
	marketsClosed bool
	marketsOrder  string
)

func init() {
	rootCmd.AddCommand(marketsCmd)
	marketsCmd.AddCommand(marketsListCmd)
	marketsCmd.AddCommand(marketsGetCmd)
	marketsCmd.AddCommand(marketsSearchCmd)

	marketsListCmd.Flags().IntVarP(&marketsLimit, "limit", "l", 25, "Maximum number of markets to return")
	marketsListCmd.Flags().IntVar(&marketsOffset, "offset", 0, "Offset for pagination")
	marketsListCmd.Flags().BoolVar(&marketsActive, "active", false, "Show only active markets")
	marketsListCmd.Flags().BoolVar(&marketsClosed, "closed", false, "Show only closed markets")
	marketsListCmd.Flags().StringVar(&marketsOrder, "order", "", "Order by field (e.g. volume_24hr, liquidity)")

	marketsSearchCmd.Flags().IntVarP(&marketsLimit, "limit", "l", 25, "Maximum number of results")
}

func newGammaClient() *api.GammaClient {
	baseURL := getAPIURL()
	return api.NewGammaClient(baseURL)
}

func runMarketsList(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	params := api.MarketListParams{
		Limit:  marketsLimit,
		Offset: marketsOffset,
		Order:  marketsOrder,
	}

	if cmd.Flags().Changed("active") {
		params.Active = api.BoolPtr(marketsActive)
	}
	if cmd.Flags().Changed("closed") {
		params.Closed = api.BoolPtr(marketsClosed)
	}

	log.Debug("Listing markets", "limit", params.Limit, "offset", params.Offset)

	markets, err := client.ListMarkets(params)
	if err != nil {
		return output.ErrNetwork(err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(markets)
	}

	output.PrintMarketsTable(markets)
	return nil
}

func runMarketsGet(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	identifier := args[0]

	var market *api.Market
	var err error

	if api.IsNumericID(identifier) {
		log.Debug("Getting market by ID", "id", identifier)
		market, err = client.GetMarket(identifier)
	} else {
		log.Debug("Getting market by slug", "slug", identifier)
		market, err = client.GetMarketBySlug(identifier)
	}

	if err != nil {
		return fmt.Errorf("getting market %q: %w", identifier, err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(market)
	}

	output.PrintMarketDetail(market)
	return nil
}

func runMarketsSearch(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	query := args[0]

	log.Debug("Searching markets", "query", query, "limit", marketsLimit)

	markets, err := client.SearchMarkets(query, marketsLimit)
	if err != nil {
		return output.ErrNetwork(err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(markets)
	}

	fmt.Printf("Search results for %q (%d found):\n\n", query, len(markets))
	output.PrintMarketsTable(markets)
	return nil
}
