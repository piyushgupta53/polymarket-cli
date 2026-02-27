package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show CLI and API status",
	Long:  "Display version information and check API connectivity.",
	Example: `  # Check CLI and API status
  polymarket status

  # Agent: get status as JSON
  polymarket status -o json -q`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if getOutputFormat() == "json" {
		return runStatusJSON()
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(output.Purple)
	labelStyle := lipgloss.NewStyle().Foreground(output.Purple).Bold(true).Width(18)
	okStyle := lipgloss.NewStyle().Foreground(output.Green).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(output.Red).Bold(true)

	fmt.Println(titleStyle.Render("Polymarket CLI Status"))
	fmt.Println()

	// Version info
	fmt.Println(labelStyle.Render("Version") + appVersion)
	fmt.Println(labelStyle.Render("Go Version") + runtime.Version())
	fmt.Println(labelStyle.Render("Platform") + runtime.GOOS + "/" + runtime.GOARCH)
	fmt.Println()

	// API connectivity
	fmt.Println(titleStyle.Render("API Connectivity"))
	fmt.Println()

	client := newGammaClient()
	start := time.Now()
	err := client.Ping()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Println(labelStyle.Render("Gamma API") + errStyle.Render("✗ unreachable"))
		fmt.Println(labelStyle.Render("Error") + err.Error())
	} else {
		fmt.Println(labelStyle.Render("Gamma API") + okStyle.Render("✓ connected"))
		fmt.Println(labelStyle.Render("Latency") + elapsed.Round(time.Millisecond).String())
	}

	return nil
}

func runStatusJSON() error {
	client := newGammaClient()
	start := time.Now()
	err := client.Ping()
	elapsed := time.Since(start)

	result := map[string]any{
		"version":    appVersion,
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
		"gamma_api": map[string]any{
			"connected":  err == nil,
			"latency_ms": elapsed.Milliseconds(),
		},
	}

	if err != nil {
		result["gamma_api"].(map[string]any)["error"] = err.Error()
	}

	return output.PrintJSON(result)
}
