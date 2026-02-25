package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "View user profiles",
	Long:  "View Polymarket user profiles.",
}

var profilesGetCmd = &cobra.Command{
	Use:   "get <address>",
	Short: "Get profile details",
	Long:  "Get detailed information about a user profile by wallet address.",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfilesGet,
}

func init() {
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.AddCommand(profilesGetCmd)
}

func runProfilesGet(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	address := args[0]

	log.Debug("Getting profile", "address", address)

	profile, err := client.GetProfile(address)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(profile)
	}

	output.PrintProfileDetail(profile)
	return nil
}
