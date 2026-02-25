package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// TxReceiptData holds transaction receipt fields for display.
type TxReceiptData struct {
	Title       string
	TxHash      string
	BlockNumber string
	GasUsed     string
}

// PrintTxReceipt prints a styled transaction receipt.
func PrintTxReceipt(data TxReceiptData) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	fmt.Println(titleStyle.Render(data.Title))
	fmt.Println()
	fmt.Println(labelStyle.Render("Tx Hash") + CyanStyle.Render(data.TxHash))
	fmt.Println(labelStyle.Render("Block") + data.BlockNumber)
	fmt.Println(labelStyle.Render("Gas Used") + data.GasUsed)
}

// PrintTxReceiptJSON returns the receipt data as a map for JSON output.
func TxReceiptJSON(data TxReceiptData) map[string]string {
	return map[string]string{
		"tx_hash":      data.TxHash,
		"block_number": data.BlockNumber,
		"gas_used":     data.GasUsed,
	}
}

// FormatUSDC formats a USDC amount string with cyan styling.
func FormatUSDC(amount string) string {
	return CyanStyle.Render(amount + " USDC")
}

// PrintBalanceTable prints a table of on-chain balances.
func PrintBalanceTable(rows [][]string) {
	headers := []string{"Asset", "Balance"}

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

	fmt.Fprintln(os.Stdout, t.Render())
}
