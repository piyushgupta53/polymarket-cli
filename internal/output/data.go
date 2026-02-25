package output

import (
	"fmt"

	"github.com/piyushgupta/polymarket-cli/internal/api/data"
)

// PrintPositionsTable prints a table of positions.
func PrintPositionsTable(positions []data.Position) {
	if len(positions) == 0 {
		fmt.Println(DimStyle.Render("No positions found."))
		return
	}

	headers := []string{"#", "Event", "Outcome", "Size", "Avg Price", "Cur Price", "Value", "P&L"}
	rows := make([][]string, 0, len(positions))

	for i, p := range positions {
		pnl := fmt.Sprintf("%.2f", p.CashPnl)
		if p.CashPnl > 0 {
			pnl = GreenStyle.Render("+" + pnl)
		} else if p.CashPnl < 0 {
			pnl = RedStyle.Render(pnl)
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(p.EventTitle, 40),
			p.Outcome,
			fmt.Sprintf("%.2f", p.Size),
			fmt.Sprintf("%.4f", p.AvgPrice),
			fmt.Sprintf("%.4f", p.CurPrice),
			FormatVolumeFloat(p.CurrentValue),
			pnl,
		})
	}

	printTable(headers, rows)
}

// PrintTradesRecordTable prints a table of trade records from the Data API.
func PrintTradesRecordTable(trades []data.TradeRecord) {
	if len(trades) == 0 {
		fmt.Println(DimStyle.Render("No trades found."))
		return
	}

	headers := []string{"#", "ID", "Side", "Price", "Size", "USDC", "Outcome", "Time"}
	rows := make([][]string, 0, len(trades))

	for i, tr := range trades {
		side := tr.Side
		if side == "BUY" {
			side = GreenStyle.Render(side)
		} else {
			side = RedStyle.Render(side)
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(tr.ID, 12),
			side,
			fmt.Sprintf("%.4f", tr.Price),
			fmt.Sprintf("%.2f", tr.Size),
			fmt.Sprintf("$%.2f", tr.UsdcSize),
			tr.Outcome,
			FormatTimestamp(tr.Timestamp),
		})
	}

	printTable(headers, rows)
}

// PrintActivityTable prints a table of account activities.
func PrintActivityTable(activities []data.Activity) {
	if len(activities) == 0 {
		fmt.Println(DimStyle.Render("No activity found."))
		return
	}

	headers := []string{"#", "Type", "Side", "Outcome", "Size", "Price", "USDC", "Time"}
	rows := make([][]string, 0, len(activities))

	for i, a := range activities {
		side := a.Side
		if side == "BUY" {
			side = GreenStyle.Render(side)
		} else if side == "SELL" {
			side = RedStyle.Render(side)
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			a.Type,
			side,
			a.Outcome,
			fmt.Sprintf("%.2f", a.Size),
			fmt.Sprintf("%.4f", a.Price),
			fmt.Sprintf("$%.2f", a.UsdcSize),
			FormatTimestamp(a.Timestamp),
		})
	}

	printTable(headers, rows)
}

// PrintHoldersTable prints a table of top holders.
func PrintHoldersTable(holders []data.Holder) {
	if len(holders) == 0 {
		fmt.Println(DimStyle.Render("No holders found."))
		return
	}

	headers := []string{"Rank", "Address", "Position", "Value"}
	rows := make([][]string, 0, len(holders))

	for _, h := range holders {
		rows = append(rows, []string{
			fmt.Sprintf("%d", h.Rank),
			Truncate(h.Address, 20),
			fmt.Sprintf("%.2f", h.Position),
			FormatVolumeFloat(h.Value),
		})
	}

	printTable(headers, rows)
}

// PrintLeaderboardTable prints a table of leaderboard entries.
func PrintLeaderboardTable(entries []data.LeaderboardEntry) {
	if len(entries) == 0 {
		fmt.Println(DimStyle.Render("No leaderboard entries found."))
		return
	}

	headers := []string{"Rank", "Address", "Volume", "P&L", "Markets"}
	rows := make([][]string, 0, len(entries))

	for _, e := range entries {
		addr := e.Address
		if e.DisplayName != "" {
			addr = e.DisplayName
		}

		pnl := fmt.Sprintf("%.2f", e.ProfitLoss)
		if e.ProfitLoss > 0 {
			pnl = GreenStyle.Render("+" + pnl)
		} else if e.ProfitLoss < 0 {
			pnl = RedStyle.Render(pnl)
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", e.Rank),
			Truncate(addr, 20),
			FormatVolumeFloat(e.Volume),
			pnl,
			fmt.Sprintf("%d", e.MarketsTraded),
		})
	}

	printTable(headers, rows)
}
