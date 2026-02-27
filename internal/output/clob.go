package output

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
)

// PrintOrderBook prints a styled order book with bids and asks.
func PrintOrderBook(book *clob.OrderBook) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	spreadLabel := lipgloss.NewStyle().Foreground(Yellow).Bold(true)

	fmt.Println(titleStyle.Render("Order Book"))
	fmt.Println(DimStyle.Render(fmt.Sprintf("Market: %s  Token: %s", Truncate(book.Market, 20), Truncate(book.AssetID, 20))))
	if book.LastTradePrice != "" {
		fmt.Println(DimStyle.Render(fmt.Sprintf("Last trade: %s  Tick: %s  Min size: %s", book.LastTradePrice, book.TickSize, book.MinOrderSize)))
	}
	fmt.Println()

	// Asks (sorted ascending by price — show in reverse so highest ask at top)
	if len(book.Asks) > 0 {
		fmt.Println(RedStyle.Bold(true).Render("Asks"))
		printOrderBookSide(book.Asks, false)
	}

	// Spread
	if len(book.Bids) > 0 && len(book.Asks) > 0 {
		bestBid := book.Bids[len(book.Bids)-1].Price // bids sorted ascending, last = highest = best
		bestAsk := book.Asks[0].Price                 // asks sorted ascending, first = lowest = best
		bidF, errB := strconv.ParseFloat(bestBid, 64)
		askF, errA := strconv.ParseFloat(bestAsk, 64)
		if errB == nil && errA == nil {
			spread := askF - bidF
			fmt.Println(spreadLabel.Render(fmt.Sprintf("  ── Spread: %.4f ──", spread)))
		}
	}
	fmt.Println()

	// Bids (sorted descending by price)
	if len(book.Bids) > 0 {
		fmt.Println(GreenStyle.Bold(true).Render("Bids"))
		printOrderBookSide(book.Bids, true)
	}
}

func printOrderBookSide(entries []clob.OrderBookEntry, isBid bool) {
	colorStyle := RedStyle
	if isBid {
		colorStyle = GreenStyle
	}

	headers := []string{"Price", "Size", "Total"}

	// Calculate cumulative totals
	type row struct {
		price, size, total string
	}
	rows := make([]row, 0, len(entries))
	var cumulative float64

	if isBid {
		// Bids: show from highest to lowest, cumulate from top
		for i := len(entries) - 1; i >= 0; i-- {
			e := entries[i]
			sz, _ := strconv.ParseFloat(e.Size, 64)
			cumulative += sz
			rows = append(rows, row{e.Price, e.Size, strconv.FormatFloat(cumulative, 'f', 2, 64)})
		}
	} else {
		// Asks: show from lowest to highest, cumulate from bottom
		totals := make([]float64, len(entries))
		var cum float64
		for i := len(entries) - 1; i >= 0; i-- {
			sz, _ := strconv.ParseFloat(entries[i].Size, 64)
			cum += sz
			totals[i] = cum
		}
		for i, e := range entries {
			rows = append(rows, row{e.Price, e.Size, strconv.FormatFloat(totals[i], 'f', 2, 64)})
		}
	}

	// Limit display to top 20 entries
	maxRows := 20
	if len(rows) > maxRows {
		trimmed := make([]row, maxRows)
		copy(trimmed, rows[:maxRows])
		rows = trimmed
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(r, col int) lipgloss.Style {
			if r == table.HeaderRow {
				return HeaderStyle
			}
			if col == 0 {
				return CellStyle.Foreground(colorStyle.GetForeground())
			}
			return CellStyle
		})

	for _, r := range rows {
		t.Row(r.price, r.size, r.total)
	}

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintClobMarketsTable prints a table of CLOB markets.
func PrintClobMarketsTable(markets []clob.ClobMarket) {
	if len(markets) == 0 {
		fmt.Println(DimStyle.Render("No CLOB markets found."))
		return
	}

	headers := []string{"#", "Question", "Tokens", "Min Size", "Tick", "Status"}
	rows := make([][]string, 0, len(markets))

	for i, m := range markets {
		status := FormatStatus(m.Active, m.Closed)
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(m.Question, 55),
			fmt.Sprintf("%d", len(m.Tokens)),
			fmt.Sprintf("%.0f", m.MinimumOrderSize),
			fmt.Sprintf("%.3f", m.MinimumTickSize),
			status,
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(r, col int) lipgloss.Style {
			if r == table.HeaderRow {
				return HeaderStyle
			}
			return CellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintClobMarketDetail prints detailed info about a single CLOB market.
func PrintClobMarketDetail(m *clob.ClobMarket) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(22)

	fmt.Println(titleStyle.Render(m.Question))
	fmt.Println()

	fmt.Println(labelStyle.Render("Condition ID") + DimStyle.Render(m.ConditionID))
	fmt.Println(labelStyle.Render("Slug") + m.MarketSlug)
	fmt.Println(labelStyle.Render("Status") + FormatStatus(m.Active, m.Closed))
	fmt.Println(labelStyle.Render("Order Book") + fmt.Sprintf("%v", m.EnableOrderBook))
	fmt.Println(labelStyle.Render("Accepting Orders") + fmt.Sprintf("%v", m.AcceptingOrders))
	fmt.Println(labelStyle.Render("Min Order Size") + fmt.Sprintf("%.0f", m.MinimumOrderSize))
	fmt.Println(labelStyle.Render("Min Tick Size") + fmt.Sprintf("%.3f", m.MinimumTickSize))
	fmt.Println(labelStyle.Render("Neg Risk") + fmt.Sprintf("%v", m.NegRisk))
	fmt.Println(labelStyle.Render("Maker Fee") + fmt.Sprintf("%.4f", m.MakerBaseFee))
	fmt.Println(labelStyle.Render("Taker Fee") + fmt.Sprintf("%.4f", m.TakerBaseFee))

	if m.EndDateISO != "" {
		fmt.Println(labelStyle.Render("End Date") + FormatTimestamp(m.EndDateISO))
	}

	if len(m.Tokens) > 0 {
		fmt.Println()
		fmt.Println(BoldStyle.Render("Tokens:"))
		for _, t := range m.Tokens {
			winner := ""
			if t.Winner {
				winner = GreenStyle.Render(" (winner)")
			}
			fmt.Printf("  %s: %s%s\n", t.Outcome, FormatPrice(t.Price), winner)
			fmt.Printf("    ID: %s\n", DimStyle.Render(t.TokenID))
		}
	}
}

// PrintPriceHistory prints a table of price history points.
func PrintPriceHistory(history []clob.PriceHistoryPoint) {
	if len(history) == 0 {
		fmt.Println(DimStyle.Render("No price history found."))
		return
	}

	headers := []string{"Time", "Price"}
	rows := make([][]string, 0, len(history))

	for _, h := range history {
		t := time.Unix(h.Timestamp, 0)
		rows = append(rows, []string{
			t.Format("Jan 02, 15:04"),
			fmt.Sprintf("%.4f", h.Price),
		})
	}

	// Limit to last 50 rows
	if len(rows) > 50 {
		trimmed := make([][]string, 50)
		copy(trimmed, rows[len(rows)-50:])
		rows = trimmed
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(r, col int) lipgloss.Style {
			if r == table.HeaderRow {
				return HeaderStyle
			}
			return CellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintBatchPrices prints batch price results.
func PrintBatchPrices(prices map[string]clob.BatchPriceEntry) {
	headers := []string{"Token ID", "Buy", "Sell"}
	var rows [][]string

	for id, p := range prices {
		buy := p.Buy
		if buy == "" {
			buy = "—"
		}
		sell := p.Sell
		if sell == "" {
			sell = "—"
		}
		rows = append(rows, []string{Truncate(id, 30), buy, sell})
	}

	printClobTable(headers, rows)
}

// PrintBatchValues prints a map of token ID to string value.
func PrintBatchValues(label string, values map[string]string) {
	headers := []string{"Token ID", label}
	var rows [][]string

	for id, v := range values {
		rows = append(rows, []string{Truncate(id, 30), v})
	}

	printClobTable(headers, rows)
}

func printClobTable(headers []string, rows [][]string) {
	if optTSV {
		printTSV(headers, rows)
		return
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(r, col int) lipgloss.Style {
			if r == table.HeaderRow {
				return HeaderStyle
			}
			return CellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintOrder prints a single order's details.
func PrintOrder(o *clob.Order) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(22)

	fmt.Println(titleStyle.Render("Order " + o.ID))
	fmt.Println()

	sideStyle := GreenStyle
	if o.Side == "SELL" {
		sideStyle = RedStyle
	}

	fmt.Println(labelStyle.Render("Status") + o.Status)
	fmt.Println(labelStyle.Render("Side") + sideStyle.Render(o.Side))
	fmt.Println(labelStyle.Render("Price") + o.Price)
	fmt.Println(labelStyle.Render("Original Size") + o.OriginalSize)
	fmt.Println(labelStyle.Render("Size Matched") + o.SizeMatched)
	fmt.Println(labelStyle.Render("Type") + o.OrderType)
	fmt.Println(labelStyle.Render("Market") + DimStyle.Render(Truncate(o.Market, 30)))
	fmt.Println(labelStyle.Render("Asset") + DimStyle.Render(Truncate(o.AssetID, 30)))
	if o.Outcome != "" {
		fmt.Println(labelStyle.Render("Outcome") + o.Outcome)
	}
	fmt.Println(labelStyle.Render("Created") + FormatTimestamp(o.CreatedAt))
	if o.ExpiresAt != "" && o.ExpiresAt != "0" {
		fmt.Println(labelStyle.Render("Expires") + FormatTimestamp(o.ExpiresAt))
	}
}

// PrintOrdersTable prints a table of orders.
func PrintOrdersTable(orders []clob.Order) {
	if len(orders) == 0 {
		fmt.Println(DimStyle.Render("No orders found."))
		return
	}

	headers := []string{"#", "ID", "Side", "Price", "Size", "Matched", "Status", "Created"}
	rows := make([][]string, 0, len(orders))

	for i, o := range orders {
		side := o.Side
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(o.ID, 12),
			side,
			o.Price,
			o.OriginalSize,
			o.SizeMatched,
			o.Status,
			FormatTimestamp(o.CreatedAt),
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(r, col int) lipgloss.Style {
			if r == table.HeaderRow {
				return HeaderStyle
			}
			if col == 2 { // Side column
				if r > 0 && r-1 < len(orders) && orders[r-1].Side == "BUY" {
					return CellStyle.Foreground(Green)
				}
				return CellStyle.Foreground(Red)
			}
			return CellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintTradesTable prints a table of trades.
func PrintTradesTable(trades []clob.Trade) {
	if len(trades) == 0 {
		fmt.Println(DimStyle.Render("No trades found."))
		return
	}

	headers := []string{"#", "ID", "Side", "Price", "Size", "Fee", "Status", "Time"}
	rows := make([][]string, 0, len(trades))

	for i, tr := range trades {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			Truncate(tr.ID, 12),
			tr.Side,
			tr.Price,
			tr.Size,
			tr.Fee,
			tr.Status,
			FormatTimestamp(tr.MatchTime),
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		Headers(headers...).
		StyleFunc(func(r, col int) lipgloss.Style {
			if r == table.HeaderRow {
				return HeaderStyle
			}
			if col == 2 { // Side column
				if r > 0 && r-1 < len(trades) && trades[r-1].Side == "BUY" {
					return CellStyle.Foreground(Green)
				}
				return CellStyle.Foreground(Red)
			}
			return CellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintBalance prints balance and allowance info.
func PrintBalance(bal *clob.BalanceAllowance) {
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)
	fmt.Println(BoldStyle.Render("Balance & Allowance"))
	fmt.Println(labelStyle.Render("Balance") + CyanStyle.Render(bal.Balance))
	fmt.Println(labelStyle.Render("Allowance") + CyanStyle.Render(bal.Allowance))
}

// PrintCancelResult prints the result of a cancel operation.
func PrintCancelResult(resp *clob.CancelResponse) {
	if len(resp.Canceled) > 0 {
		fmt.Println(GreenStyle.Bold(true).Render(fmt.Sprintf("Canceled %d order(s)", len(resp.Canceled))))
		for _, id := range resp.Canceled {
			fmt.Println("  " + DimStyle.Render(id))
		}
	}
	if len(resp.NotCanceled) > 0 {
		fmt.Println(RedStyle.Bold(true).Render(fmt.Sprintf("Failed to cancel %d order(s)", len(resp.NotCanceled))))
		for _, f := range resp.NotCanceled {
			fmt.Printf("  %s: %s\n", DimStyle.Render(f.ID), f.Reason)
		}
	}
	if len(resp.Canceled) == 0 && len(resp.NotCanceled) == 0 {
		fmt.Println(DimStyle.Render("No orders to cancel."))
	}
}
