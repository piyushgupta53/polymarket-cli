package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/piyushgupta/polymarket-cli/internal/api"
)

var (
	// Table styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Purple).
			Padding(0, 1)

	CellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			Foreground(Dim)
)

// PrintMarketsTable prints a table of markets.
func PrintMarketsTable(markets []api.Market) {
	if len(markets) == 0 {
		fmt.Println(DimStyle.Render("No markets found."))
		return
	}

	headers := []string{"#", "Question", "Outcomes", "Volume", "Status"}
	rows := make([][]string, 0, len(markets))

	for i, m := range markets {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(m.Question, 60),
			FormatOutcomePrices(m.Outcomes, m.OutcomePrices),
			FormatVolume(m.Volume),
			FormatStatus(m.Active, m.Closed),
		})
	}

	printTable(headers, rows)
}

// PrintMarketDetail prints detailed information about a single market.
func PrintMarketDetail(m *api.Market) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White).MarginBottom(1)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)
	valueStyle := lipgloss.NewStyle()

	fmt.Println(titleStyle.Render(m.Question))
	fmt.Println()

	fields := []struct {
		label string
		value string
	}{
		{"ID", m.ID},
		{"Slug", m.Slug},
		{"Condition ID", m.ConditionID},
		{"Status", FormatStatus(m.Active, m.Closed)},
		{"Outcomes", FormatOutcomePrices(m.Outcomes, m.OutcomePrices)},
		{"Best Bid", FormatPricePtr(m.BestBid)},
		{"Best Ask", FormatPricePtr(m.BestAsk)},
		{"Last Trade", FormatPricePtr(m.LastTradePrice)},
		{"Volume", FormatVolume(m.Volume)},
		{"Volume (24h)", FormatVolumeFloat(m.Volume24hr)},
		{"Liquidity", FormatVolume(m.Liquidity)},
		{"End Date", FormatTimestamp(m.EndDateISO)},
	}

	for _, f := range fields {
		if f.value == "" || f.value == DimStyle.Render("—") {
			continue
		}
		fmt.Println(labelStyle.Render(f.label) + valueStyle.Render(f.value))
	}

	if m.Description != "" {
		fmt.Println()
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(80)
		fmt.Println(descStyle.Render(m.Description))
	}

	if len(m.Tokens) > 0 {
		fmt.Println()
		fmt.Println(BoldStyle.Render("Tokens:"))
		for _, t := range m.Tokens {
			fmt.Printf("  %s: %s (ID: %s)\n",
				t.Outcome,
				FormatPrice(t.Price),
				DimStyle.Render(Truncate(t.TokenID, 20)),
			)
		}
	}
}

// PrintEventsTable prints a table of events.
func PrintEventsTable(events []api.Event) {
	if len(events) == 0 {
		fmt.Println(DimStyle.Render("No events found."))
		return
	}

	headers := []string{"#", "Title", "Markets", "Volume", "Status"}
	rows := make([][]string, 0, len(events))

	for i, e := range events {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(e.Title, 60),
			fmt.Sprintf("%d", len(e.Markets)),
			FormatVolumeFloat(e.Volume),
			FormatStatus(e.Active, e.Closed),
		})
	}

	printTable(headers, rows)
}

// PrintEventDetail prints detailed information about a single event.
func PrintEventDetail(e *api.Event) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White).MarginBottom(1)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	fmt.Println(titleStyle.Render(e.Title))
	fmt.Println()

	fields := []struct {
		label string
		value string
	}{
		{"ID", e.ID},
		{"Slug", e.Slug},
		{"Ticker", e.Ticker},
		{"Status", FormatStatus(e.Active, e.Closed)},
		{"Volume", FormatVolumeFloat(e.Volume)},
		{"Volume (24h)", FormatVolumeFloat(e.Volume24hr)},
		{"Liquidity", FormatVolumeFloat(e.Liquidity)},
		{"Start Date", FormatTimestamp(e.StartDate)},
		{"End Date", FormatTimestamp(e.EndDate)},
		{"Comments", fmt.Sprintf("%d", e.CommentCount)},
	}

	for _, f := range fields {
		if f.value == "" || f.value == DimStyle.Render("—") {
			continue
		}
		fmt.Println(labelStyle.Render(f.label) + f.value)
	}

	if e.Description != "" {
		fmt.Println()
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(80)
		fmt.Println(descStyle.Render(e.Description))
	}

	if len(e.Markets) > 0 {
		fmt.Println()
		fmt.Println(BoldStyle.Render(fmt.Sprintf("Markets (%d):", len(e.Markets))))
		PrintMarketsTable(e.Markets)
	}
}

// PrintTagsTable prints a table of tags.
func PrintTagsTable(tags []api.Tag) {
	if len(tags) == 0 {
		fmt.Println(DimStyle.Render("No tags found."))
		return
	}

	headers := []string{"#", "Label", "Slug"}
	rows := make([][]string, 0, len(tags))

	for i, t := range tags {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			t.Label,
			DimStyle.Render(t.Slug),
		})
	}

	printTable(headers, rows)
}

func printTable(headers []string, rows [][]string) {
	if optTSV {
		printTSV(headers, rows)
		return
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return HeaderStyle
			}
			return CellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	fmt.Fprintln(os.Stdout, t.Render())
}

// printTSV writes tab-separated values to stdout.
func printTSV(headers []string, rows [][]string) {
	if !optNoHeaders {
		fmt.Fprintln(os.Stdout, strings.Join(headers, "\t"))
	}
	for _, row := range rows {
		fmt.Fprintln(os.Stdout, strings.Join(row, "\t"))
	}
}
