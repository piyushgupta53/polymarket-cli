package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var sportsCmd = &cobra.Command{
	Use:   "sports",
	Short: "Browse sports markets",
	Long:  "List sports, market types, and teams for Polymarket sports betting.",
}

var sportsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sports",
	Long:  "List all available sports categories.",
	RunE:  runSportsList,
}

var sportsMarketTypesCmd = &cobra.Command{
	Use:   "market-types",
	Short: "List sport market types",
	Long:  "List all available sport market types (e.g. moneyline, spread).",
	RunE:  runSportsMarketTypes,
}

var sportsTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "List teams",
	Long:  "List sports teams with optional league filter.",
	RunE:  runSportsTeams,
}

var (
	teamsLeague string
	teamsLimit  int
)

func init() {
	rootCmd.AddCommand(sportsCmd)
	sportsCmd.AddCommand(sportsListCmd)
	sportsCmd.AddCommand(sportsMarketTypesCmd)
	sportsCmd.AddCommand(sportsTeamsCmd)

	sportsTeamsCmd.Flags().StringVar(&teamsLeague, "league", "", "Filter by league (e.g. NBA, NFL)")
	sportsTeamsCmd.Flags().IntVarP(&teamsLimit, "limit", "l", 25, "Maximum number of teams to return")
}

func runSportsList(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	log.Debug("Listing sports")

	sports, err := client.ListSports()
	if err != nil {
		return fmt.Errorf("listing sports: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(sports)
	}

	output.PrintSportsTable(sports)
	return nil
}

func runSportsMarketTypes(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	log.Debug("Getting market types")

	types, err := client.GetMarketTypes()
	if err != nil {
		return fmt.Errorf("getting market types: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(types)
	}

	output.PrintMarketTypesTable(types)
	return nil
}

func runSportsTeams(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	log.Debug("Listing teams", "league", teamsLeague, "limit", teamsLimit)

	teams, err := client.ListTeams(teamsLeague, teamsLimit)
	if err != nil {
		return fmt.Errorf("listing teams: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(teams)
	}

	output.PrintTeamsTable(teams)
	return nil
}
