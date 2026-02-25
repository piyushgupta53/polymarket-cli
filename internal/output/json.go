package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintJSON prints any value as JSON to stdout.
// In compact mode (quiet), uses single-line JSON; otherwise pretty-printed.
func PrintJSON(v any) error {
	var data []byte
	var err error
	if optCompactJSON {
		data, err = json.Marshal(v)
	} else {
		data, err = json.MarshalIndent(v, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintln(os.Stdout, string(data))
	return nil
}

// PrintRawJSON prints pre-encoded JSON bytes with indentation.
func PrintRawJSON(data json.RawMessage) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		// If we can't parse it, print raw
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	return PrintJSON(v)
}
