package cmd

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/chain"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var ctfCmd = &cobra.Command{
	Use:   "ctf",
	Short: "Conditional Token Framework operations",
	Long:  "Split, merge, and redeem conditional tokens, plus compute CTF IDs.",
}

// --- Transaction commands ---

var ctfSplitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split USDC into YES and NO conditional tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		conditionHex, err := cmd.Flags().GetString("condition")
		if err != nil {
			return err
		}
		amountUSDC, err := cmd.Flags().GetFloat64("amount")
		if err != nil {
			return err
		}

		conditionID, err := parseBytes32(conditionHex)
		if err != nil {
			return fmt.Errorf("invalid condition ID: %w", err)
		}

		ecdsaKey, err := requirePrivateKey()
		if err != nil {
			return err
		}

		client, err := chain.NewChainClient(getChainRPCURL())
		if err != nil {
			return fmt.Errorf("connecting to RPC: %w", err)
		}
		defer client.Close()

		amount := chain.USDCToRaw(amountUSDC)
		result, err := client.SplitPosition(context.Background(), ecdsaKey, conditionID, amount)
		if err != nil {
			return output.ErrChain(fmt.Errorf("split: %w", err))
		}

		return printTxResult("Split Position", result)
	},
}

var ctfMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge YES and NO tokens back into USDC",
	RunE: func(cmd *cobra.Command, args []string) error {
		conditionHex, err := cmd.Flags().GetString("condition")
		if err != nil {
			return err
		}
		amountShares, err := cmd.Flags().GetFloat64("amount")
		if err != nil {
			return err
		}

		conditionID, err := parseBytes32(conditionHex)
		if err != nil {
			return fmt.Errorf("invalid condition ID: %w", err)
		}

		ecdsaKey, err := requirePrivateKey()
		if err != nil {
			return err
		}

		client, err := chain.NewChainClient(getChainRPCURL())
		if err != nil {
			return fmt.Errorf("connecting to RPC: %w", err)
		}
		defer client.Close()

		amount := chain.USDCToRaw(amountShares)
		result, err := client.MergePositions(context.Background(), ecdsaKey, conditionID, amount)
		if err != nil {
			return output.ErrChain(fmt.Errorf("merge: %w", err))
		}

		return printTxResult("Merge Positions", result)
	},
}

var ctfRedeemCmd = &cobra.Command{
	Use:   "redeem",
	Short: "Redeem winning conditional tokens for USDC",
	RunE: func(cmd *cobra.Command, args []string) error {
		conditionHex, err := cmd.Flags().GetString("condition")
		if err != nil {
			return err
		}

		conditionID, err := parseBytes32(conditionHex)
		if err != nil {
			return fmt.Errorf("invalid condition ID: %w", err)
		}

		ecdsaKey, err := requirePrivateKey()
		if err != nil {
			return err
		}

		client, err := chain.NewChainClient(getChainRPCURL())
		if err != nil {
			return fmt.Errorf("connecting to RPC: %w", err)
		}
		defer client.Close()

		result, err := client.RedeemPositions(context.Background(), ecdsaKey, conditionID)
		if err != nil {
			return output.ErrChain(fmt.Errorf("redeem: %w", err))
		}

		return printTxResult("Redeem Positions", result)
	},
}

var ctfRedeemNegRiskCmd = &cobra.Command{
	Use:   "redeem-neg-risk",
	Short: "Redeem neg-risk conditional tokens for USDC",
	RunE: func(cmd *cobra.Command, args []string) error {
		conditionHex, err := cmd.Flags().GetString("condition")
		if err != nil {
			return err
		}

		conditionID, err := parseBytes32(conditionHex)
		if err != nil {
			return fmt.Errorf("invalid condition ID: %w", err)
		}

		ecdsaKey, err := requirePrivateKey()
		if err != nil {
			return err
		}

		client, err := chain.NewChainClient(getChainRPCURL())
		if err != nil {
			return fmt.Errorf("connecting to RPC: %w", err)
		}
		defer client.Close()

		result, err := client.RedeemNegRiskPositions(context.Background(), ecdsaKey, conditionID)
		if err != nil {
			return output.ErrChain(fmt.Errorf("redeem neg-risk: %w", err))
		}

		return printTxResult("Redeem Neg-Risk Positions", result)
	},
}

// --- Pure computation commands ---

var ctfConditionIDCmd = &cobra.Command{
	Use:   "condition-id",
	Short: "Compute a condition ID from oracle, question ID, and outcome count",
	RunE: func(cmd *cobra.Command, args []string) error {
		oracleHex, err := cmd.Flags().GetString("oracle")
		if err != nil {
			return err
		}
		questionHex, err := cmd.Flags().GetString("question-id")
		if err != nil {
			return err
		}
		outcomes, err := cmd.Flags().GetUint("outcomes")
		if err != nil {
			return err
		}

		oracle := common.HexToAddress(oracleHex)
		questionID, err := parseBytes32(questionHex)
		if err != nil {
			return fmt.Errorf("invalid question ID: %w", err)
		}

		result := chain.ComputeConditionID(oracle, questionID, outcomes)

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{
				"condition_id": "0x" + hex.EncodeToString(result[:]),
			})
		}

		fmt.Println("0x" + hex.EncodeToString(result[:]))
		return nil
	},
}

var ctfCollectionIDCmd = &cobra.Command{
	Use:   "collection-id",
	Short: "Compute a collection ID from condition ID and index set",
	RunE: func(cmd *cobra.Command, args []string) error {
		conditionHex, err := cmd.Flags().GetString("condition")
		if err != nil {
			return err
		}
		indexSet, err := cmd.Flags().GetUint("index-set")
		if err != nil {
			return err
		}
		parentHex, err := cmd.Flags().GetString("parent")
		if err != nil {
			return err
		}

		conditionID, err := parseBytes32(conditionHex)
		if err != nil {
			return fmt.Errorf("invalid condition ID: %w", err)
		}

		var parentCollectionID [32]byte
		if parentHex != "" {
			parentCollectionID, err = parseBytes32(parentHex)
			if err != nil {
				return fmt.Errorf("invalid parent collection ID: %w", err)
			}
		}

		result := chain.ComputeCollectionID(parentCollectionID, conditionID, indexSet)

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{
				"collection_id": "0x" + hex.EncodeToString(result[:]),
			})
		}

		fmt.Println("0x" + hex.EncodeToString(result[:]))
		return nil
	},
}

var ctfPositionIDCmd = &cobra.Command{
	Use:   "position-id",
	Short: "Compute a position ID from collateral token and collection ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		collectionHex, err := cmd.Flags().GetString("collection")
		if err != nil {
			return err
		}
		collateralHex, err := cmd.Flags().GetString("collateral")
		if err != nil {
			return err
		}

		collectionID, err := parseBytes32(collectionHex)
		if err != nil {
			return fmt.Errorf("invalid collection ID: %w", err)
		}

		collateral := common.HexToAddress(collateralHex)

		result := chain.ComputePositionID(collateral, collectionID)

		if getOutputFormat() == "json" {
			return output.PrintJSON(map[string]string{
				"position_id": "0x" + hex.EncodeToString(result[:]),
			})
		}

		fmt.Println("0x" + hex.EncodeToString(result[:]))
		return nil
	},
}

// --- Helpers ---

func requirePrivateKey() (*ecdsa.PrivateKey, error) {
	pk := resolvePrivateKey()
	if pk == "" {
		return nil, output.ErrAuthRequired("Run: polymarket wallet import --private-key <hex>")
	}

	hexKey := pk
	if len(hexKey) >= 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}

	ecdsaKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, output.ErrAuthFailed(err)
	}
	return ecdsaKey, nil
}

func parseBytes32(hexStr string) ([32]byte, error) {
	var result [32]byte
	hexStr = strings.TrimPrefix(hexStr, "0x")
	if len(hexStr) != 64 {
		return result, fmt.Errorf("expected 64 hex chars, got %d", len(hexStr))
	}
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return result, err
	}
	copy(result[:], b)
	return result, nil
}

func init() {
	rootCmd.AddCommand(ctfCmd)

	// Transaction commands
	ctfSplitCmd.Flags().String("condition", "", "Condition ID (hex, required)")
	ctfSplitCmd.MarkFlagRequired("condition")
	ctfSplitCmd.Flags().Float64("amount", 0, "USDC amount to split (required)")
	ctfSplitCmd.MarkFlagRequired("amount")
	ctfCmd.AddCommand(ctfSplitCmd)

	ctfMergeCmd.Flags().String("condition", "", "Condition ID (hex, required)")
	ctfMergeCmd.MarkFlagRequired("condition")
	ctfMergeCmd.Flags().Float64("amount", 0, "Number of shares to merge (required)")
	ctfMergeCmd.MarkFlagRequired("amount")
	ctfCmd.AddCommand(ctfMergeCmd)

	ctfRedeemCmd.Flags().String("condition", "", "Condition ID (hex, required)")
	ctfRedeemCmd.MarkFlagRequired("condition")
	ctfCmd.AddCommand(ctfRedeemCmd)

	ctfRedeemNegRiskCmd.Flags().String("condition", "", "Condition ID (hex, required)")
	ctfRedeemNegRiskCmd.MarkFlagRequired("condition")
	ctfCmd.AddCommand(ctfRedeemNegRiskCmd)

	// Pure computation commands
	ctfConditionIDCmd.Flags().String("oracle", "", "Oracle address (hex, required)")
	ctfConditionIDCmd.MarkFlagRequired("oracle")
	ctfConditionIDCmd.Flags().String("question-id", "", "Question ID (32-byte hex, required)")
	ctfConditionIDCmd.MarkFlagRequired("question-id")
	ctfConditionIDCmd.Flags().Uint("outcomes", 2, "Number of outcome slots")
	ctfCmd.AddCommand(ctfConditionIDCmd)

	ctfCollectionIDCmd.Flags().String("condition", "", "Condition ID (hex, required)")
	ctfCollectionIDCmd.MarkFlagRequired("condition")
	ctfCollectionIDCmd.Flags().Uint("index-set", 1, "Index set")
	ctfCollectionIDCmd.Flags().String("parent", "", "Parent collection ID (hex, optional)")
	ctfCmd.AddCommand(ctfCollectionIDCmd)

	ctfPositionIDCmd.Flags().String("collection", "", "Collection ID (hex, required)")
	ctfPositionIDCmd.MarkFlagRequired("collection")
	ctfPositionIDCmd.Flags().String("collateral", chain.USDCAddress.Hex(), "Collateral token address")
	ctfCmd.AddCommand(ctfPositionIDCmd)
}
