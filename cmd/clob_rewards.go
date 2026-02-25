package cmd

import (
	"fmt"

	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	rewardConditionFlag string
	rewardOrderIDFlag   string
	rewardOrderIDsFlag  string
)

var clobRewardsCmd = &cobra.Command{
	Use:   "rewards",
	Short: "Get current reward rates",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetCurrentRewards()
		if err != nil {
			return fmt.Errorf("getting rewards: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobEarningsCmd = &cobra.Command{
	Use:   "earnings",
	Short: "Get epoch earnings",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetEpochEarnings()
		if err != nil {
			return fmt.Errorf("getting earnings: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobEarningsTotalCmd = &cobra.Command{
	Use:   "earnings-total",
	Short: "Get total epoch earnings",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetEpochTotal()
		if err != nil {
			return fmt.Errorf("getting earnings total: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobRewardPercentagesCmd = &cobra.Command{
	Use:   "reward-percentages",
	Short: "Get reward percentage configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetRewardPercentages()
		if err != nil {
			return fmt.Errorf("getting reward percentages: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobMarketRewardCmd = &cobra.Command{
	Use:   "market-reward",
	Short: "Get reward info for a market",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetMarketRewards(rewardConditionFlag)
		if err != nil {
			return fmt.Errorf("getting market reward: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobOrderScoringCmd = &cobra.Command{
	Use:   "order-scoring",
	Short: "Get scoring info for an order",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		data, err := c.GetOrderScoring(rewardOrderIDFlag)
		if err != nil {
			return fmt.Errorf("getting order scoring: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

var clobOrdersScoringCmd = &cobra.Command{
	Use:   "orders-scoring",
	Short: "Get scoring info for multiple orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := requireAuth()
		if err != nil {
			return err
		}
		ids := splitTokens(rewardOrderIDsFlag)
		data, err := c.GetOrdersScoring(ids)
		if err != nil {
			return fmt.Errorf("getting orders scoring: %w", err)
		}
		return output.PrintRawJSON(data)
	},
}

func init() {
	clobCmd.AddCommand(clobRewardsCmd)
	clobCmd.AddCommand(clobEarningsCmd)
	clobCmd.AddCommand(clobEarningsTotalCmd)
	clobCmd.AddCommand(clobRewardPercentagesCmd)

	clobMarketRewardCmd.Flags().StringVar(&rewardConditionFlag, "condition", "", "Condition ID (required)")
	_ = clobMarketRewardCmd.MarkFlagRequired("condition")
	clobCmd.AddCommand(clobMarketRewardCmd)

	clobOrderScoringCmd.Flags().StringVar(&rewardOrderIDFlag, "order-id", "", "Order ID (required)")
	_ = clobOrderScoringCmd.MarkFlagRequired("order-id")
	clobCmd.AddCommand(clobOrderScoringCmd)

	clobOrdersScoringCmd.Flags().StringVar(&rewardOrderIDsFlag, "order-ids", "", "Comma-separated order IDs (required)")
	_ = clobOrdersScoringCmd.MarkFlagRequired("order-ids")
	clobCmd.AddCommand(clobOrdersScoringCmd)
}
