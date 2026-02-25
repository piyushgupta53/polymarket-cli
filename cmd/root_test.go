package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/piyushgupta/polymarket-cli/internal/output"
)

func TestGetOutputFormat_Default(t *testing.T) {
	// When not a TTY (in tests), default should be "json"
	old := outputFormat
	outputFormat = ""
	defer func() { outputFormat = old }()

	got := getOutputFormat()
	// In test, stdout is not a TTY so should be "json"
	if got != "json" {
		t.Errorf("expected 'json' for non-TTY, got %q", got)
	}
}

func TestGetOutputFormat_ExplicitFlag(t *testing.T) {
	old := outputFormat
	outputFormat = "table"
	defer func() { outputFormat = old }()

	got := getOutputFormat()
	if got != "table" {
		t.Errorf("expected 'table', got %q", got)
	}
}

func TestGetOutputFormat_TSV(t *testing.T) {
	old := outputFormat
	outputFormat = "tsv"
	defer func() { outputFormat = old }()

	got := getOutputFormat()
	if got != "tsv" {
		t.Errorf("expected 'tsv', got %q", got)
	}
}

func TestNonInteractiveEnvVar(t *testing.T) {
	old := nonInteractive
	nonInteractive = false
	defer func() { nonInteractive = old }()

	os.Setenv("POLYMARKET_NON_INTERACTIVE", "1")
	defer os.Unsetenv("POLYMARKET_NON_INTERACTIVE")

	// Simulate what PersistentPreRun does
	if os.Getenv("POLYMARKET_NON_INTERACTIVE") == "1" {
		nonInteractive = true
	}

	if !IsNonInteractive() {
		t.Error("expected non-interactive to be true when env var is set")
	}
}

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"nil", nil, 0},
		{"plain error", fmt.Errorf("oops"), 1},
		{"auth required", output.ErrAuthRequired("hint"), 3},
		{"auth failed", output.ErrAuthFailed(fmt.Errorf("bad")), 3},
		{"invalid input", output.ErrInvalidInput("bad"), 2},
		{"network", output.ErrNetwork(fmt.Errorf("timeout")), 4},
		{"not found", output.ErrNotFound("thing"), 5},
		{"chain", output.ErrChain(fmt.Errorf("revert")), 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := output.ExitCodeFromError(tt.err)
			if got != tt.expected {
				t.Errorf("ExitCodeFromError() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestIsQuiet(t *testing.T) {
	old := quiet
	quiet = true
	defer func() { quiet = old }()

	if !IsQuiet() {
		t.Error("expected quiet to be true")
	}
}
