package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/api"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Browse prediction market events",
	Long:  "List and view details of Polymarket events (groups of related markets).",
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	Long:  "List events with optional filters.",
	RunE:  runEventsList,
}

var eventsGetCmd = &cobra.Command{
	Use:   "get <id-or-slug>",
	Short: "Get event details",
	Long:  "Get detailed information about a specific event by ID or slug.",
	Args:  cobra.ExactArgs(1),
	RunE:  runEventsGet,
}

var (
	eventsLimit  int
	eventsOffset int
	eventsActive bool
	eventsClosed bool
)

func init() {
	rootCmd.AddCommand(eventsCmd)
	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsGetCmd)

	eventsListCmd.Flags().IntVarP(&eventsLimit, "limit", "l", 25, "Maximum number of events to return")
	eventsListCmd.Flags().IntVar(&eventsOffset, "offset", 0, "Offset for pagination")
	eventsListCmd.Flags().BoolVar(&eventsActive, "active", false, "Show only active events")
	eventsListCmd.Flags().BoolVar(&eventsClosed, "closed", false, "Show only closed events")
}

func runEventsList(cmd *cobra.Command, args []string) error {
	client := newGammaClient()

	params := api.EventListParams{
		Limit:  eventsLimit,
		Offset: eventsOffset,
	}

	if cmd.Flags().Changed("active") {
		params.Active = api.BoolPtr(eventsActive)
	}
	if cmd.Flags().Changed("closed") {
		params.Closed = api.BoolPtr(eventsClosed)
	}

	log.Debug("Listing events", "limit", params.Limit, "offset", params.Offset)

	events, err := client.ListEvents(params)
	if err != nil {
		return fmt.Errorf("listing events: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(events)
	}

	output.PrintEventsTable(events)
	return nil
}

func runEventsGet(cmd *cobra.Command, args []string) error {
	client := newGammaClient()
	identifier := args[0]

	var event *api.Event
	var err error

	if api.IsNumericID(identifier) {
		log.Debug("Getting event by ID", "id", identifier)
		event, err = client.GetEvent(identifier)
	} else {
		log.Debug("Getting event by slug", "slug", identifier)
		event, err = client.GetEventBySlug(identifier)
	}

	if err != nil {
		return fmt.Errorf("getting event: %w", err)
	}

	if getOutputFormat() == "json" {
		return output.PrintJSON(event)
	}

	output.PrintEventDetail(event)
	return nil
}
