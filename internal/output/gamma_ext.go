package output

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/api"
)

// PrintSeriesTable prints a table of series.
func PrintSeriesTable(series []api.Series) {
	if len(series) == 0 {
		fmt.Println(DimStyle.Render("No series found."))
		return
	}

	headers := []string{"#", "Title", "Slug"}
	rows := make([][]string, 0, len(series))

	for i, s := range series {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(s.Title, 60),
			DimStyle.Render(s.Slug),
		})
	}

	printTable(headers, rows)
}

// PrintSeriesDetail prints detailed information about a single series.
func PrintSeriesDetail(s *api.Series) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White).MarginBottom(1)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	fmt.Println(titleStyle.Render(s.Title))
	fmt.Println()

	fields := []struct {
		label string
		value string
	}{
		{"ID", s.ID},
		{"Slug", s.Slug},
	}

	for _, f := range fields {
		if f.value == "" {
			continue
		}
		fmt.Println(labelStyle.Render(f.label) + f.value)
	}

	if s.Description != "" {
		fmt.Println()
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(80)
		fmt.Println(descStyle.Render(s.Description))
	}

	if len(s.Events) > 0 {
		fmt.Println()
		fmt.Println(BoldStyle.Render(fmt.Sprintf("Events (%d):", len(s.Events))))
		PrintEventsTable(s.Events)
	}
}

// PrintCommentsTable prints a table of comments.
func PrintCommentsTable(comments []api.Comment) {
	if len(comments) == 0 {
		fmt.Println(DimStyle.Render("No comments found."))
		return
	}

	headers := []string{"#", "Author", "Body", "Created"}
	rows := make([][]string, 0, len(comments))

	for i, c := range comments {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(c.Author, 16),
			Truncate(c.Body, 60),
			FormatTimestamp(c.CreatedAt),
		})
	}

	printTable(headers, rows)
}

// PrintCommentDetail prints detailed information about a single comment.
func PrintCommentDetail(c *api.Comment) {
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	fields := []struct {
		label string
		value string
	}{
		{"ID", c.ID},
		{"Author", c.Author},
		{"Entity Type", c.EntityType},
		{"Entity ID", c.EntityID},
		{"Created", FormatTimestamp(c.CreatedAt)},
	}

	for _, f := range fields {
		if f.value == "" {
			continue
		}
		fmt.Println(labelStyle.Render(f.label) + f.value)
	}

	if c.Body != "" {
		fmt.Println()
		bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(80)
		fmt.Println(bodyStyle.Render(c.Body))
	}
}

// PrintProfileDetail prints detailed information about a user profile.
func PrintProfileDetail(p *api.Profile) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White).MarginBottom(1)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	displayName := p.Username
	if displayName == "" {
		displayName = p.Address
	}
	fmt.Println(titleStyle.Render(displayName))
	fmt.Println()

	fields := []struct {
		label string
		value string
	}{
		{"Address", p.Address},
		{"Username", p.Username},
		{"Bio", p.Bio},
		{"Volume", FormatVolumeFloat(p.Volume)},
	}

	for _, f := range fields {
		if f.value == "" {
			continue
		}
		fmt.Println(labelStyle.Render(f.label) + f.value)
	}
}

// PrintSportsTable prints a table of sports.
func PrintSportsTable(sports []api.Sport) {
	if len(sports) == 0 {
		fmt.Println(DimStyle.Render("No sports found."))
		return
	}

	headers := []string{"#", "Label", "Slug"}
	rows := make([][]string, 0, len(sports))

	for i, s := range sports {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			s.Label,
			DimStyle.Render(s.Slug),
		})
	}

	printTable(headers, rows)
}

// PrintMarketTypesTable prints a table of sport market types.
func PrintMarketTypesTable(types []api.SportMarketType) {
	if len(types) == 0 {
		fmt.Println(DimStyle.Render("No market types found."))
		return
	}

	headers := []string{"#", "Label", "Sport"}
	rows := make([][]string, 0, len(types))

	for i, mt := range types {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			mt.Label,
			mt.Sport,
		})
	}

	printTable(headers, rows)
}

// PrintTeamsTable prints a table of teams.
func PrintTeamsTable(teams []api.Team) {
	if len(teams) == 0 {
		fmt.Println(DimStyle.Render("No teams found."))
		return
	}

	headers := []string{"#", "Name", "League", "Slug"}
	rows := make([][]string, 0, len(teams))

	for i, t := range teams {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			t.Name,
			t.League,
			DimStyle.Render(t.Slug),
		})
	}

	printTable(headers, rows)
}
