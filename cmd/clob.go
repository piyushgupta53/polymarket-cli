package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var clobCmd = &cobra.Command{
	Use:   "clob",
	Short: "CLOB API commands (order book, prices, markets)",
	Long:  "Interact with the Polymarket CLOB (Central Limit Order Book) API.",
}

// --- Health & Status ---

var clobOkCmd = &cobra.Command{
	Use:   "ok",
	Short: "Check CLOB API health",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		status, err := c.HealthCheck()
		if err != nil {
			return fmt.Errorf("health check: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"status": status})
		}
		fmt.Println(output.GreenStyle.Bold(true).Render("CLOB API: " + status))
		return nil
	},
}

var clobTimeCmd = &cobra.Command{
	Use:   "time",
	Short: "Get CLOB server time",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		t, err := c.ServerTime()
		if err != nil {
			return fmt.Errorf("server time: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"time": t})
		}
		fmt.Println("Server time:", output.CyanStyle.Render(t))
		return nil
	},
}

// --- Single Token Commands ---

var (
	clobTokenFlag    string
	clobSideFlag     string
	clobTokensFlag   string
	clobConditionFlag string
	clobIntervalFlag string
	clobFidelityFlag int
	clobCursorFlag   string
)

var clobPriceCmd = &cobra.Command{
	Use:   "price",
	Short: "Get token price",
	Example: `  # Get buy price for a token
  polymarket clob price --token <TOKEN_ID> --side buy

  # Agent: get price as JSON
  polymarket clob price --token <TOKEN_ID> -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetPrice(clobTokenFlag, clobSideFlag)
		if err != nil {
			return fmt.Errorf("getting price: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		fmt.Println(label.Render("Price") + " (" + clobSideFlag + "): " + output.CyanStyle.Render(resp.Price))
		return nil
	},
}

var clobMidpointCmd = &cobra.Command{
	Use:   "midpoint",
	Short: "Get token midpoint price",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetMidpoint(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting midpoint: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		fmt.Println(label.Render("Midpoint: ") + output.CyanStyle.Render(resp.Mid))
		return nil
	},
}

var clobSpreadCmd = &cobra.Command{
	Use:   "spread",
	Short: "Get token bid-ask spread",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetSpread(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting spread: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		fmt.Println(label.Render("Spread: ") + output.YellowStyle.Render(resp.Spread))
		return nil
	},
}

var clobBookCmd = &cobra.Command{
	Use:   "book",
	Short: "Get token order book",
	Example: `  # View order book for a token
  polymarket clob book --token <TOKEN_ID>

  # Agent: get order book as JSON
  polymarket clob book --token <TOKEN_ID> -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		book, err := c.GetBook(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting order book: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(book)
		}
		output.PrintOrderBook(book)
		return nil
	},
}

var clobLastTradeCmd = &cobra.Command{
	Use:   "last-trade",
	Short: "Get last trade price for token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetLastTradePrice(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting last trade: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		fmt.Println(label.Render("Last Trade: ") + output.CyanStyle.Render(resp.Price) + " (" + resp.Side + ")")
		return nil
	},
}

var clobTickSizeCmd = &cobra.Command{
	Use:   "tick-size",
	Short: "Get tick size for token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetTickSize(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting tick size: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		fmt.Println(label.Render("Tick Size: ") + fmt.Sprintf("%.4f", resp.MinimumTickSize))
		return nil
	},
}

var clobFeeRateCmd = &cobra.Command{
	Use:   "fee-rate",
	Short: "Get fee rate for token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetFeeRate(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting fee rate: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		fmt.Println(label.Render("Base Fee: ") + fmt.Sprintf("%.4f", resp.BaseFee))
		return nil
	},
}

var clobNegRiskCmd = &cobra.Command{
	Use:   "neg-risk",
	Short: "Get neg-risk flag for token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetNegRisk(clobTokenFlag)
		if err != nil {
			return fmt.Errorf("getting neg-risk: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		label := lipgloss.NewStyle().Foreground(output.Purple).Bold(true)
		val := output.RedStyle.Render("false")
		if resp.NegRisk {
			val = output.GreenStyle.Render("true")
		}
		fmt.Println(label.Render("Neg Risk: ") + val)
		return nil
	},
}

// --- Batch Commands ---

var clobBatchPricesCmd = &cobra.Command{
	Use:   "batch-prices",
	Short: "Get prices for multiple tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		tokens := splitTokens(clobTokensFlag)
		resp, err := c.GetBatchPrices(tokens, clobSideFlag)
		if err != nil {
			return fmt.Errorf("getting batch prices: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintBatchPrices(resp)
		return nil
	},
}

var clobMidpointsCmd = &cobra.Command{
	Use:   "midpoints",
	Short: "Get midpoints for multiple tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		tokens := splitTokens(clobTokensFlag)
		resp, err := c.GetBatchMidpoints(tokens)
		if err != nil {
			return fmt.Errorf("getting batch midpoints: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintBatchValues("Midpoint", resp)
		return nil
	},
}

var clobSpreadsCmd = &cobra.Command{
	Use:   "spreads",
	Short: "Get spreads for multiple tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		tokens := splitTokens(clobTokensFlag)
		resp, err := c.GetBatchSpreads(tokens)
		if err != nil {
			return fmt.Errorf("getting batch spreads: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintBatchValues("Spread", resp)
		return nil
	},
}

var clobBooksCmd = &cobra.Command{
	Use:   "books",
	Short: "Get order books for multiple tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		tokens := splitTokens(clobTokensFlag)
		resp, err := c.GetBatchBooks(tokens)
		if err != nil {
			return fmt.Errorf("getting batch books: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		for id, book := range resp {
			fmt.Println(output.BoldStyle.Render("Token: " + output.Truncate(id, 30)))
			bookCopy := book
			output.PrintOrderBook(&bookCopy)
			fmt.Println()
		}
		return nil
	},
}

var clobLastTradesCmd = &cobra.Command{
	Use:   "last-trades",
	Short: "Get last trade prices for multiple tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		tokens := splitTokens(clobTokensFlag)
		resp, err := c.GetBatchLastTrades(tokens)
		if err != nil {
			return fmt.Errorf("getting batch last trades: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		for id, lt := range resp {
			fmt.Printf("%s  Price: %s  Side: %s\n",
				output.DimStyle.Render(output.Truncate(id, 30)),
				output.CyanStyle.Render(lt.Price),
				lt.Side)
		}
		return nil
	},
}

// --- Market Commands ---

var clobMarketCmd = &cobra.Command{
	Use:   "market",
	Short: "Get CLOB market by condition ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		m, err := c.GetMarket(clobConditionFlag)
		if err != nil {
			return fmt.Errorf("getting CLOB market: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(m)
		}
		output.PrintClobMarketDetail(m)
		return nil
	},
}

var clobMarketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "List all CLOB markets (paginated)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.ListMarkets(clobCursorFlag)
		if err != nil {
			return fmt.Errorf("listing CLOB markets: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintClobMarketsTable(resp.Data)
		if resp.NextCursor != "" {
			fmt.Println()
			fmt.Println(output.DimStyle.Render(fmt.Sprintf("Next cursor: %s  (use --cursor to paginate)", resp.NextCursor)))
		}
		return nil
	},
}

var clobSamplingMarketsCmd = &cobra.Command{
	Use:   "sampling-markets",
	Short: "List reward-eligible markets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		markets, err := c.GetSamplingMarkets()
		if err != nil {
			return fmt.Errorf("getting sampling markets: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(markets)
		}
		output.PrintClobMarketsTable(markets)
		return nil
	},
}

var clobSimplifiedMarketsCmd = &cobra.Command{
	Use:   "simplified-markets",
	Short: "List simplified markets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		markets, err := c.GetSimplifiedMarkets()
		if err != nil {
			return fmt.Errorf("getting simplified markets: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(markets)
		}
		output.PrintClobMarketsTable(markets)
		return nil
	},
}

// --- Price History ---

var clobPriceHistoryCmd = &cobra.Command{
	Use:   "price-history",
	Short: "Get price history for token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClobClient()
		resp, err := c.GetPriceHistory(clobTokenFlag, clobIntervalFlag, clobFidelityFlag)
		if err != nil {
			return fmt.Errorf("getting price history: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintPriceHistory(resp.History)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clobCmd)

	// Health
	clobCmd.AddCommand(clobOkCmd)
	clobCmd.AddCommand(clobTimeCmd)

	// Single token commands
	for _, cmd := range []*cobra.Command{clobPriceCmd, clobMidpointCmd, clobSpreadCmd, clobBookCmd, clobLastTradeCmd, clobTickSizeCmd, clobFeeRateCmd, clobNegRiskCmd, clobPriceHistoryCmd} {
		cmd.Flags().StringVar(&clobTokenFlag, "token", "", "Token ID (required)")
		_ = cmd.MarkFlagRequired("token")
		clobCmd.AddCommand(cmd)
	}

	// Side flag for price
	clobPriceCmd.Flags().StringVar(&clobSideFlag, "side", "buy", "Side (buy or sell)")

	// Price history flags
	clobPriceHistoryCmd.Flags().StringVar(&clobIntervalFlag, "interval", "1d", "Interval (1m, 1h, 6h, 1d, 1w, max)")
	clobPriceHistoryCmd.Flags().IntVar(&clobFidelityFlag, "fidelity", 0, "Number of data points")

	// Batch commands
	for _, cmd := range []*cobra.Command{clobBatchPricesCmd, clobMidpointsCmd, clobSpreadsCmd, clobBooksCmd, clobLastTradesCmd} {
		cmd.Flags().StringVar(&clobTokensFlag, "tokens", "", "Comma-separated token IDs (required)")
		_ = cmd.MarkFlagRequired("tokens")
		clobCmd.AddCommand(cmd)
	}
	clobBatchPricesCmd.Flags().StringVar(&clobSideFlag, "side", "buy", "Side (buy or sell)")

	// Market commands
	clobMarketCmd.Flags().StringVar(&clobConditionFlag, "condition", "", "Condition ID (required)")
	_ = clobMarketCmd.MarkFlagRequired("condition")
	clobCmd.AddCommand(clobMarketCmd)

	clobMarketsCmd.Flags().StringVar(&clobCursorFlag, "cursor", "", "Pagination cursor")
	clobCmd.AddCommand(clobMarketsCmd)

	clobCmd.AddCommand(clobSamplingMarketsCmd)
	clobCmd.AddCommand(clobSimplifiedMarketsCmd)
}

func newClobClient() *clob.Client {
	return clob.NewClient("")
}

func splitTokens(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
