package cmd

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/chain"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	approveCheckSpenderFlag string
	approveSetSpenderFlag   string
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Manage USDC approvals for Polymarket contracts",
	Long:  "Check and set ERC-20 USDC approval for the CTF Exchange or Neg Risk Exchange.",
}

var approveCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Show USDC balance and allowance",
	RunE: func(cmd *cobra.Command, args []string) error {
		pk := resolvePrivateKey()
		if pk == "" {
			return output.ErrAuthRequired("Run: polymarket wallet import --private-key <hex>")
		}

		hexKey := pk
		if len(hexKey) >= 2 && hexKey[:2] == "0x" {
			hexKey = hexKey[2:]
		}
		ecdsaKey, err := crypto.HexToECDSA(hexKey)
		if err != nil {
			return output.ErrAuthFailed(err)
		}
		owner := chain.AddressFromKey(ecdsaKey)

		spender, err := chain.SpenderFromName(approveCheckSpenderFlag)
		if err != nil {
			return err
		}

		client, err := chain.NewChainClient(getChainRPCURL())
		if err != nil {
			return output.ErrChain(fmt.Errorf("connecting to RPC: %w", err))
		}
		defer client.Close()

		ctx := context.Background()

		balance, err := client.GetUSDCBalance(ctx, owner)
		if err != nil {
			return output.ErrChain(fmt.Errorf("fetching USDC balance: %w", err))
		}

		allowance, err := client.GetUSDCAllowance(ctx, owner, spender)
		if err != nil {
			return output.ErrChain(fmt.Errorf("fetching USDC allowance: %w", err))
		}

		balStr := chain.RawToUSDC(balance)
		allowStr := chain.RawToUSDC(allowance)
		unlimited := allowance.Cmp(chain.MaxUint256) == 0

		if getOutputFormat() == "json" {
			data := map[string]string{
				"address":   owner.Hex(),
				"spender":   approveCheckSpenderFlag,
				"balance":   balStr,
				"allowance": allowStr,
			}
			if unlimited {
				data["allowance"] = "unlimited"
			}
			return output.PrintJSON(data)
		}

		if unlimited {
			allowStr = "unlimited"
		}

		output.PrintBalanceTable([][]string{
			{"USDC Balance", output.FormatUSDC(balStr)},
			{"USDC Allowance (" + approveCheckSpenderFlag + ")", allowStr},
			{"Wallet", owner.Hex()},
		})
		return nil
	},
}

var approveSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Approve USDC spending for a Polymarket contract",
	Long:  "Send an ERC-20 approve transaction. Use --amount 0 to revoke, or omit --amount for unlimited approval.",
	RunE: func(cmd *cobra.Command, args []string) error {
		pk := resolvePrivateKey()
		if pk == "" {
			return output.ErrAuthRequired("Run: polymarket wallet import --private-key <hex>")
		}

		hexKey := pk
		if len(hexKey) >= 2 && hexKey[:2] == "0x" {
			hexKey = hexKey[2:]
		}
		ecdsaKey, err := crypto.HexToECDSA(hexKey)
		if err != nil {
			return output.ErrAuthFailed(err)
		}

		spender, err := chain.SpenderFromName(approveSetSpenderFlag)
		if err != nil {
			return err
		}

		amountFlag, err := cmd.Flags().GetFloat64("amount")
		if err != nil {
			return err
		}
		amountSet := cmd.Flags().Changed("amount")

		var amount *big.Int
		if amountSet {
			amount = chain.USDCToRaw(amountFlag)
		}
		// nil amount = MaxUint256 (unlimited)

		client, err := chain.NewChainClient(getChainRPCURL())
		if err != nil {
			return output.ErrChain(fmt.Errorf("connecting to RPC: %w", err))
		}
		defer client.Close()

		result, err := client.ApproveUSDC(context.Background(), ecdsaKey, spender, amount)
		if err != nil {
			return output.ErrChain(fmt.Errorf("approve: %w", err))
		}

		return printTxResult("USDC Approval", result)
	},
}

func printTxResult(title string, result *chain.TxResult) error {
	data := output.TxReceiptData{
		Title:       title,
		TxHash:      result.TxHash.Hex(),
		BlockNumber: fmt.Sprintf("%d", result.BlockNumber),
		GasUsed:     fmt.Sprintf("%d", result.GasUsed),
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(output.TxReceiptJSON(data))
	}

	output.PrintTxReceipt(data)
	return nil
}

func init() {
	rootCmd.AddCommand(approveCmd)

	approveCheckCmd.Flags().StringVar(&approveCheckSpenderFlag, "spender", "exchange", "Spender contract (exchange, neg-risk-exchange)")
	approveCmd.AddCommand(approveCheckCmd)

	approveSetCmd.Flags().StringVar(&approveSetSpenderFlag, "spender", "exchange", "Spender contract (exchange, neg-risk-exchange)")
	approveSetCmd.Flags().Float64("amount", 0, "USDC amount to approve (omit for unlimited)")
	approveCmd.AddCommand(approveSetCmd)
}
