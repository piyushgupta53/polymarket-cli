package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Exit codes for structured error reporting.
const (
	ExitSuccess           = 0
	ExitGeneralError      = 1
	ExitInvalidInput      = 2
	ExitAuthError         = 3
	ExitNetworkError      = 4
	ExitNotFound          = 5
	ExitInsufficientFunds = 6
	ExitChainError        = 7
)

// Error code strings for machine-readable error identification.
const (
	CodeAuthRequired      = "AUTH_REQUIRED"
	CodeAuthFailed        = "AUTH_FAILED"
	CodeNetworkError      = "NETWORK_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	CodeRateLimited       = "RATE_LIMITED"
	CodeChainError        = "CHAIN_ERROR"
	CodeInternal          = "INTERNAL"
)

// CLIError is a structured error with machine-readable code, human hint, and exit code.
type CLIError struct {
	Code     string // e.g. "AUTH_REQUIRED"
	Message  string // human-readable message
	Hint     string // actionable suggestion, e.g. "polymarket setup --key <hex>"
	ExitCode int    // process exit code (0-7)
	Wrapped  error  // underlying error
}

func (e *CLIError) Error() string { return e.Message }

func (e *CLIError) Unwrap() error { return e.Wrapped }

// PrintError writes a structured error to stderr.
// In JSON mode it emits {"error":{"code":"...","message":"...","hint":"..."}}.
// In table mode it prints styled text.
func PrintError(err error, jsonMode bool) {
	if err == nil {
		return
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		cliErr = &CLIError{
			Code:     CodeInternal,
			Message:  err.Error(),
			ExitCode: ExitGeneralError,
		}
	}

	if jsonMode {
		errObj := map[string]any{
			"code":    cliErr.Code,
			"message": cliErr.Message,
		}
		if cliErr.Hint != "" {
			errObj["hint"] = cliErr.Hint
		}
		envelope := map[string]any{"error": errObj}
		data, merr := json.Marshal(envelope)
		if merr != nil {
			_, _ = fmt.Fprintf(os.Stderr, `{"error":{"code":%q,"message":%q}}`+"\n", cliErr.Code, cliErr.Message)
			return
		}
		_, _ = fmt.Fprintln(os.Stderr, string(data))
	} else {
		if cliErr.Hint != "" {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\nHint: %s\n", cliErr.Message, cliErr.Hint)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", cliErr.Message)
		}
	}
}

// ExitCodeFromError returns the appropriate exit code for an error.
// Returns 0 for nil, the CLIError's ExitCode if available, or 1 for plain errors.
func ExitCodeFromError(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr.ExitCode
	}
	return ExitGeneralError
}

// --- Constructors ---

// ErrAuthRequired returns a CLIError for missing authentication.
func ErrAuthRequired(hint string) *CLIError {
	return &CLIError{
		Code:     CodeAuthRequired,
		Message:  "authentication required",
		Hint:     hint,
		ExitCode: ExitAuthError,
	}
}

// ErrAuthFailed returns a CLIError for authentication failure.
func ErrAuthFailed(err error) *CLIError {
	return &CLIError{
		Code:     CodeAuthFailed,
		Message:  fmt.Sprintf("authentication failed: %v", err),
		ExitCode: ExitAuthError,
		Wrapped:  err,
	}
}

// ErrInvalidInput returns a CLIError for invalid user input.
func ErrInvalidInput(msg string) *CLIError {
	return &CLIError{
		Code:     CodeInvalidInput,
		Message:  msg,
		ExitCode: ExitInvalidInput,
	}
}

// ErrNetwork returns a CLIError for network failures.
func ErrNetwork(err error) *CLIError {
	return &CLIError{
		Code:     CodeNetworkError,
		Message:  fmt.Sprintf("network error: %v", err),
		ExitCode: ExitNetworkError,
		Wrapped:  err,
	}
}

// ErrNotFound returns a CLIError for missing resources.
func ErrNotFound(resource string) *CLIError {
	return &CLIError{
		Code:     CodeNotFound,
		Message:  fmt.Sprintf("not found: %s", resource),
		ExitCode: ExitNotFound,
	}
}

// ErrInsufficientFunds returns a CLIError for insufficient balance.
func ErrInsufficientFunds(msg string) *CLIError {
	return &CLIError{
		Code:     CodeInsufficientFunds,
		Message:  msg,
		ExitCode: ExitInsufficientFunds,
	}
}

// ErrChain returns a CLIError for on-chain transaction failures.
func ErrChain(err error) *CLIError {
	return &CLIError{
		Code:     CodeChainError,
		Message:  fmt.Sprintf("chain error: %v", err),
		ExitCode: ExitChainError,
		Wrapped:  err,
	}
}
