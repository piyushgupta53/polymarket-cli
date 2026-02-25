package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var seriesCmd = &cobra.Command{
	Use:   "series",
	Short: "Browse event series",
	Long:  "List and view details of Polymarket event series.",
}

var seriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List series",
	Long:  "List event series with optional pagination.",
	RunE:  runSeriesList,
}

var seriesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get series details",
	Long:  "Get detailed information about a specific series by ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runSeriesGet,
}

var (
	seriesLimit  int
	seriesOffset int
)

func init() {
	rootCmd.AddCommand(seriesCmd)
	seriesCmd.AddCommand(seriesListCmd)
	seriesCmd.AddCommand(seriesGetCmd)

	seriesListCmd.Flags().IntVarP(&seriesLimit, "limit", "l", 25, "Maximum number of series to return")
	seriesListCmd.Flags().IntVar(&seriesOffset, "offset", 0, "Offset for pagination")
}

func runSeriesList(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	log.Debug("Listing series", "limit", seriesLimit, "offset", seriesOffset)

	series, err := client.ListSeries(seriesLimit, seriesOffset)
	if err != nil {
		return fmt.Errorf("listing series: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(series)
	}

	output.PrintSeriesTable(series)
	return nil
}

func runSeriesGet(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	id := args[0]

	log.Debug("Getting series", "id", id)

	series, err := client.GetSeries(id)
	if err != nil {
		return fmt.Errorf("getting series: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(series)
	}

	output.PrintSeriesDetail(series)
	return nil
}
