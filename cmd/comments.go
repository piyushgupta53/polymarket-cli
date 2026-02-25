package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Browse and view comments",
	Long:  "List, view, and search comments on Polymarket entities.",
}

var commentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List comments for an entity",
	Long:  "List comments for a specific entity type and ID.",
	RunE:  runCommentsList,
}

var commentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get comment details",
	Long:  "Get detailed information about a specific comment by ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommentsGet,
}

var commentsByUserCmd = &cobra.Command{
	Use:   "by-user",
	Short: "List comments by a user",
	Long:  "List all comments made by a specific user address. Defaults to the configured wallet address.",
	RunE:  runCommentsByUser,
}

var (
	commentsEntityType string
	commentsEntityID   string
	commentsAddress    string
)

func init() {
	rootCmd.AddCommand(commentsCmd)
	commentsCmd.AddCommand(commentsListCmd)
	commentsCmd.AddCommand(commentsGetCmd)
	commentsCmd.AddCommand(commentsByUserCmd)

	commentsListCmd.Flags().StringVar(&commentsEntityType, "entity-type", "", "Entity type (e.g. event, market)")
	commentsListCmd.Flags().StringVar(&commentsEntityID, "entity-id", "", "Entity ID")

	commentsByUserCmd.Flags().StringVar(&commentsAddress, "address", "", "User wallet address (defaults to configured wallet)")
}

// resolveAddressForComments returns the address from flag or derives it from the wallet private key.
func resolveAddressForComments(flagAddr string) string {
	if flagAddr != "" {
		return flagAddr
	}
	// Derive address from private key
	pk := resolvePrivateKey()
	if pk == "" {
		return ""
	}
	hexKey := pk
	if len(hexKey) >= 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}
	ecdsaKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return ""
	}
	return crypto.PubkeyToAddress(ecdsaKey.PublicKey).Hex()
}

func runCommentsList(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	log.Debug("Listing comments", "entity_type", commentsEntityType, "entity_id", commentsEntityID)

	comments, err := client.ListComments(commentsEntityType, commentsEntityID)
	if err != nil {
		return fmt.Errorf("listing comments: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(comments)
	}

	output.PrintCommentsTable(comments)
	return nil
}

func runCommentsGet(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	id := args[0]

	log.Debug("Getting comment", "id", id)

	comment, err := client.GetComment(id)
	if err != nil {
		return fmt.Errorf("getting comment: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(comment)
	}

	output.PrintCommentDetail(comment)
	return nil
}

func runCommentsByUser(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	address := resolveAddressForComments(commentsAddress)
	if address == "" {
		return fmt.Errorf("no address specified. Use --address flag or configure a wallet")
	}

	log.Debug("Getting user comments", "address", address)

	comments, err := client.GetUserComments(address)
	if err != nil {
		return fmt.Errorf("getting user comments: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(comments)
	}

	fmt.Printf("Comments by %s (%d found):\n\n", address, len(comments))
	output.PrintCommentsTable(comments)
	return nil
}
