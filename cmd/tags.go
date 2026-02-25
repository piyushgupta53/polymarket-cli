package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Browse market tags",
	Long:  "List and view details of Polymarket topic tags.",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	Long:  "List all available topic tags.",
	RunE:  runTagsList,
}

var tagsGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get tag details",
	Long:  "Get details of a specific tag by slug.",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagsGet,
}

func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsGetCmd)
}

func runTagsList(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	log.Debug("Listing tags")

	tags, err := client.ListTags()
	if err != nil {
		return fmt.Errorf("listing tags: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(tags)
	}

	output.PrintTagsTable(tags)
	return nil
}

func runTagsGet(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	slug := args[0]

	log.Debug("Getting tag", "slug", slug)

	tag, err := client.GetTag(slug)
	if err != nil {
		return fmt.Errorf("getting tag: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(tag)
	}

	fmt.Printf("Tag: %s\n", output.BoldStyle.Render(tag.Label))
	fmt.Printf("Slug: %s\n", output.DimStyle.Render(tag.Slug))
	fmt.Printf("ID: %s\n", tag.ID)
	return nil
}
