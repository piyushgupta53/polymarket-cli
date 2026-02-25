package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/piyushgupta/polymarket-cli/internal/api/bridge"
)

// PrintDepositAddresses prints deposit addresses as styled key-value pairs.
func PrintDepositAddresses(addrs *bridge.DepositAddresses) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	fmt.Println(titleStyle.Render("Deposit Addresses"))
	fmt.Println()

	fields := []struct {
		label string
		value string
	}{
		{"EVM", addrs.EVMAddress},
		{"Solana", addrs.SolanaAddress},
		{"Bitcoin", addrs.BitcoinAddress},
	}

	for _, f := range fields {
		if f.value == "" {
			continue
		}
		fmt.Println(labelStyle.Render(f.label) + CyanStyle.Render(f.value))
	}
}

// PrintSupportedAssetsTable prints a table of supported bridge assets.
func PrintSupportedAssetsTable(assets []bridge.SupportedAsset) {
	if len(assets) == 0 {
		fmt.Println(DimStyle.Render("No supported assets found."))
		return
	}

	headers := []string{"#", "Chain", "Token", "Address"}
	var rows [][]string

	for i, a := range assets {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			a.Chain,
			a.Token,
			DimStyle.Render(a.Address),
		})
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

	_, _ = fmt.Fprintln(os.Stdout, t.Render())
}

// PrintDepositStatus prints deposit status as styled key-value pairs.
func PrintDepositStatus(status *bridge.DepositStatus) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	labelStyle := lipgloss.NewStyle().Foreground(Purple).Bold(true).Width(18)

	fmt.Println(titleStyle.Render("Deposit Status"))
	fmt.Println()

	fmt.Println(labelStyle.Render("Tx Hash") + CyanStyle.Render(status.TxHash))

	statusStyle := DimStyle
	switch status.Status {
	case "completed":
		statusStyle = GreenStyle
	case "pending":
		statusStyle = YellowStyle
	case "failed":
		statusStyle = RedStyle
	}
	fmt.Println(labelStyle.Render("Status") + statusStyle.Render(status.Status))

	if status.Amount != "" {
		fmt.Println(labelStyle.Render("Amount") + CyanStyle.Render(status.Amount))
	}
	if status.Chain != "" {
		fmt.Println(labelStyle.Render("Chain") + status.Chain)
	}
}
