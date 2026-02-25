package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCLIError_ImplementsError(t *testing.T) {
	var err error = &CLIError{Code: CodeInternal, Message: "test"}
	if err.Error() != "test" {
		t.Errorf("expected 'test', got %q", err.Error())
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("root cause")
	cliErr := &CLIError{Code: CodeNetworkError, Message: "net fail", Wrapped: inner}
	if !errors.Is(cliErr, inner) {
		t.Error("Unwrap should expose wrapped error")
	}
}

func TestPrintError_JSON(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cliErr := &CLIError{
		Code:    CodeAuthRequired,
		Message: "authentication required",
		Hint:    "polymarket setup --key <hex>",
	}
	PrintError(cliErr, true)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	var parsed map[string]map[string]string
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	errObj := parsed["error"]
	if errObj["code"] != CodeAuthRequired {
		t.Errorf("expected code %q, got %q", CodeAuthRequired, errObj["code"])
	}
	if errObj["message"] != "authentication required" {
		t.Errorf("unexpected message: %q", errObj["message"])
	}
	if errObj["hint"] != "polymarket setup --key <hex>" {
		t.Errorf("unexpected hint: %q", errObj["hint"])
	}
}

func TestPrintError_Table(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cliErr := &CLIError{
		Code:    CodeNotFound,
		Message: "not found: market xyz",
		Hint:    "check the market ID",
	}
	PrintError(cliErr, false)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "not found: market xyz") {
		t.Errorf("expected message in output: %q", output)
	}
	if !strings.Contains(output, "check the market ID") {
		t.Errorf("expected hint in output: %q", output)
	}
}

func TestPrintError_PlainError_JSON(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintError(fmt.Errorf("something broke"), true)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	var parsed map[string]map[string]string
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["error"]["code"] != CodeInternal {
		t.Errorf("plain error should map to INTERNAL, got %q", parsed["error"]["code"])
	}
}

func TestPrintError_Nil(t *testing.T) {
	// Should not panic or print anything
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintError(nil, true)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if buf.Len() > 0 {
		t.Error("nil error should produce no output")
	}
}

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"nil", nil, ExitSuccess},
		{"plain error", fmt.Errorf("oops"), ExitGeneralError},
		{"auth required", ErrAuthRequired("hint"), ExitAuthError},
		{"auth failed", ErrAuthFailed(fmt.Errorf("bad")), ExitAuthError},
		{"invalid input", ErrInvalidInput("bad arg"), ExitInvalidInput},
		{"network", ErrNetwork(fmt.Errorf("timeout")), ExitNetworkError},
		{"not found", ErrNotFound("market"), ExitNotFound},
		{"chain", ErrChain(fmt.Errorf("revert")), ExitChainError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCodeFromError(tt.err)
			if got != tt.expected {
				t.Errorf("ExitCodeFromError(%v) = %d, want %d", tt.err, got, tt.expected)
			}
		})
	}
}

func TestConstructors(t *testing.T) {
	t.Run("ErrAuthRequired", func(t *testing.T) {
		e := ErrAuthRequired("run setup")
		if e.Code != CodeAuthRequired || e.Hint != "run setup" || e.ExitCode != ExitAuthError {
			t.Errorf("unexpected: %+v", e)
		}
	})
	t.Run("ErrAuthFailed", func(t *testing.T) {
		inner := fmt.Errorf("bad key")
		e := ErrAuthFailed(inner)
		if e.Code != CodeAuthFailed || e.Wrapped != inner {
			t.Errorf("unexpected: %+v", e)
		}
	})
	t.Run("ErrInvalidInput", func(t *testing.T) {
		e := ErrInvalidInput("missing flag")
		if e.Code != CodeInvalidInput || e.ExitCode != ExitInvalidInput {
			t.Errorf("unexpected: %+v", e)
		}
	})
	t.Run("ErrNetwork", func(t *testing.T) {
		inner := fmt.Errorf("timeout")
		e := ErrNetwork(inner)
		if e.Code != CodeNetworkError || e.Wrapped != inner {
			t.Errorf("unexpected: %+v", e)
		}
	})
	t.Run("ErrNotFound", func(t *testing.T) {
		e := ErrNotFound("market abc")
		if e.Code != CodeNotFound || e.ExitCode != ExitNotFound {
			t.Errorf("unexpected: %+v", e)
		}
	})
	t.Run("ErrChain", func(t *testing.T) {
		inner := fmt.Errorf("revert")
		e := ErrChain(inner)
		if e.Code != CodeChainError || e.Wrapped != inner {
			t.Errorf("unexpected: %+v", e)
		}
	})
}
