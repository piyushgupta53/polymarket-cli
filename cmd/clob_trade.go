package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	tradeTokenFlag     string
	tradeSideFlag      string
	tradePriceFlag     string
	tradeSizeFlag      string
	tradeTypeFlag      string
	tradeExpirationFlag string
	tradeOrderIDFlag   string
	tradeOrderIDsFlag  string
	tradeMarketFlag    string
	tradeAssetFlag     string
	tradeConfirmFlag   bool
	balanceTypeFlag    string
	balanceTokenFlag   string
)

// --- Order Commands ---

var clobCreateOrderCmd = &cobra.Command{
	Use:   "create-order",
	Short: "Create a new order",
	Long:  "Submit a limit order to the CLOB. Requires wallet configuration.",
	Example: `  # Place a limit buy order
  polymarket clob create-order --token <TOKEN_ID> --side BUY --price 0.55 --size 10

  # Agent: place order and get JSON response
  polymarket clob create-order --token <TOKEN_ID> --side BUY --price 0.55 --size 10 -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, cfg, err := requireAuth()
		if err != nil {
			return err
		}

		// Parse private key for order signing.
		hexKey := strings.TrimPrefix(resolvePrivateKey(), "0x")
		ecdsaKey, err := crypto.HexToECDSA(hexKey)
		if err != nil {
			return output.ErrAuthFailed(fmt.Errorf("parsing private key: %w", err))
		}

		price, err := strconv.ParseFloat(tradePriceFlag, 64)
		if err != nil {
			return fmt.Errorf("invalid price %q: %w", tradePriceFlag, err)
		}
		size, err := strconv.ParseFloat(tradeSizeFlag, 64)
		if err != nil {
			return fmt.Errorf("invalid size %q: %w", tradeSizeFlag, err)
		}

		// Auto-detect neg-risk for this token.
		negRisk := false
		if nr, err := c.GetNegRisk(tradeTokenFlag); err == nil {
			negRisk = nr.NegRisk
		}

		payload, err := clob.BuildSignedOrder(clob.OrderParams{
			TokenID:    tradeTokenFlag,
			Side:       tradeSideFlag,
			Price:      price,
			Size:       size,
			OrderType:  tradeTypeFlag,
			Expiration: tradeExpirationFlag,
			NegRisk:    negRisk,
		}, ecdsaKey, cfg.ChainID)
		if err != nil {
			return fmt.Errorf("building order: %w", err)
		}

		resp, err := c.PostOrder(payload)
		if err != nil {
			return fmt.Errorf("submitting order: %w", err)
		}

		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		if resp.Success {
			fmt.Println(output.GreenStyle.Render("Order placed successfully"))
			fmt.Printf("  Order ID: %s\n", resp.OrderID)
			if resp.Status != "" {
				fmt.Printf("  Status:   %s\n", resp.Status)
			}
		} else {
			fmt.Println(output.RedStyle.Render("Order rejected"))
			if resp.ErrorMsg != "" {
				fmt.Printf("  Error: %s\n", resp.ErrorMsg)
			}
		}
		return nil
	},
}

// --- Cancel Commands ---

var clobCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel an order by ID",
	Example: `  # Cancel a specific order
  polymarket clob cancel --order-id <ORDER_ID>

  # Agent: cancel and get JSON response
  polymarket clob cancel --order-id <ORDER_ID> -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		resp, err := c.CancelOrder(tradeOrderIDFlag)
		if err != nil {
			return fmt.Errorf("canceling order: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintCancelResult(resp)
		return nil
	},
}

var clobCancelOrdersCmd = &cobra.Command{
	Use:   "cancel-orders",
	Short: "Cancel multiple orders by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		ids := splitTokens(tradeOrderIDsFlag)
		resp, err := c.CancelOrders(ids)
		if err != nil {
			return fmt.Errorf("canceling orders: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintCancelResult(resp)
		return nil
	},
}

var clobCancelMarketCmd = &cobra.Command{
	Use:   "cancel-market",
	Short: "Cancel all orders for a market",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		resp, err := c.CancelMarketOrders(tradeMarketFlag)
		if err != nil {
			return fmt.Errorf("canceling market orders: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintCancelResult(resp)
		return nil
	},
}

var clobCancelAllCmd = &cobra.Command{
	Use:   "cancel-all",
	Short: "Cancel all open orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tradeConfirmFlag {
			return fmt.Errorf("this will cancel ALL open orders. Use --confirm to proceed")
		}
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		resp, err := c.CancelAll()
		if err != nil {
			return fmt.Errorf("canceling all orders: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(resp)
		}
		output.PrintCancelResult(resp)
		return nil
	},
}

// --- Order/Trade Query Commands ---

var clobOrdersCmd = &cobra.Command{
	Use:   "orders",
	Short: "List open orders",
	Example: `  # List all open orders
  polymarket clob orders

  # Agent: list orders as JSON
  polymarket clob orders -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		params := &clob.OpenOrdersParams{
			Market: tradeMarketFlag,
			Asset:  tradeAssetFlag,
		}
		orders, err := c.GetOpenOrders(params)
		if err != nil {
			return fmt.Errorf("getting orders: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(orders)
		}
		output.PrintOrdersTable(orders)
		return nil
	},
}

var clobOrderCmd = &cobra.Command{
	Use:   "order",
	Short: "Get a single order by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		order, err := c.GetOrder(tradeOrderIDFlag)
		if err != nil {
			return fmt.Errorf("getting order: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(order)
		}
		output.PrintOrder(order)
		return nil
	},
}

var clobTradesCmd = &cobra.Command{
	Use:   "trades",
	Short: "List trade history",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		params := &clob.TradesParams{
			Market: tradeMarketFlag,
		}
		trades, err := c.GetTrades(params)
		if err != nil {
			return fmt.Errorf("getting trades: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(trades)
		}
		output.PrintTradesTable(trades)
		return nil
	},
}

// --- Balance Commands ---

var clobBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get balance and allowance",
	Example: `  # Check your balance
  polymarket clob balance

  # Agent: get balance as JSON
  polymarket clob balance -o json -q`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		params := &clob.BalanceParams{
			AssetType: balanceTypeFlag,
			TokenID:   balanceTokenFlag,
		}
		bal, err := c.GetBalanceAllowance(params)
		if err != nil {
			return fmt.Errorf("getting balance: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(bal)
		}
		output.PrintBalance(bal)
		return nil
	},
}

var clobUpdateBalanceCmd = &cobra.Command{
	Use:   "update-balance",
	Short: "Refresh balance and allowance",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		if err := c.UpdateBalanceAllowance(); err != nil {
			return fmt.Errorf("updating balance: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{"status": "updated"})
		}
		fmt.Println(output.GreenStyle.Render("Balance and allowance refresh triggered."))
		return nil
	},
}

func init() {
	// Create order
	clobCreateOrderCmd.Flags().StringVar(&tradeTokenFlag, "token", "", "Token ID (required)")
	_ = clobCreateOrderCmd.MarkFlagRequired("token")
	clobCreateOrderCmd.Flags().StringVar(&tradeSideFlag, "side", "", "Side: BUY or SELL (required)")
	_ = clobCreateOrderCmd.MarkFlagRequired("side")
	clobCreateOrderCmd.Flags().StringVar(&tradePriceFlag, "price", "", "Limit price (required)")
	_ = clobCreateOrderCmd.MarkFlagRequired("price")
	clobCreateOrderCmd.Flags().StringVar(&tradeSizeFlag, "size", "", "Order size (required)")
	_ = clobCreateOrderCmd.MarkFlagRequired("size")
	clobCreateOrderCmd.Flags().StringVar(&tradeTypeFlag, "type", "GTC", "Order type (GTC, GTD, FOK)")
	clobCreateOrderCmd.Flags().StringVar(&tradeExpirationFlag, "expiration", "0", "Expiration timestamp")
	clobCmd.AddCommand(clobCreateOrderCmd)

	// Cancel single
	clobCancelCmd.Flags().StringVar(&tradeOrderIDFlag, "order-id", "", "Order ID (required)")
	_ = clobCancelCmd.MarkFlagRequired("order-id")
	clobCmd.AddCommand(clobCancelCmd)

	// Cancel multiple
	clobCancelOrdersCmd.Flags().StringVar(&tradeOrderIDsFlag, "order-ids", "", "Comma-separated order IDs (required)")
	_ = clobCancelOrdersCmd.MarkFlagRequired("order-ids")
	clobCmd.AddCommand(clobCancelOrdersCmd)

	// Cancel market
	clobCancelMarketCmd.Flags().StringVar(&tradeMarketFlag, "condition", "", "Condition ID (required)")
	_ = clobCancelMarketCmd.MarkFlagRequired("condition")
	clobCmd.AddCommand(clobCancelMarketCmd)

	// Cancel all
	clobCancelAllCmd.Flags().BoolVar(&tradeConfirmFlag, "confirm", false, "Confirm cancellation")
	clobCmd.AddCommand(clobCancelAllCmd)

	// Orders list
	clobOrdersCmd.Flags().StringVar(&tradeMarketFlag, "market", "", "Filter by market (condition ID)")
	clobOrdersCmd.Flags().StringVar(&tradeAssetFlag, "asset", "", "Filter by asset (token ID)")
	clobCmd.AddCommand(clobOrdersCmd)

	// Single order
	clobOrderCmd.Flags().StringVar(&tradeOrderIDFlag, "id", "", "Order ID (required)")
	_ = clobOrderCmd.MarkFlagRequired("id")
	clobCmd.AddCommand(clobOrderCmd)

	// Trades
	clobTradesCmd.Flags().StringVar(&tradeMarketFlag, "market", "", "Filter by market (condition ID)")
	clobCmd.AddCommand(clobTradesCmd)

	// Balance
	clobBalanceCmd.Flags().StringVar(&balanceTypeFlag, "type", "", "Asset type filter")
	clobBalanceCmd.Flags().StringVar(&balanceTokenFlag, "token", "", "Token ID filter")
	clobCmd.AddCommand(clobBalanceCmd)

	// Update balance
	clobCmd.AddCommand(clobUpdateBalanceCmd)
}

