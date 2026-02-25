package output

import (
	"strings"
	"testing"
)

func TestFormatVolumeFloat(t *testing.T) {
	tests := []struct {
		name string
		v    float64
		want string // We check the raw content (contains)
	}{
		{"zero", 0, "$0"},
		{"hundreds", 500, "$500"},
		{"thousands", 1500, "$1.5K"},
		{"millions", 2500000, "$2.5M"},
		{"billions", 1200000000, "$1.2B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatVolumeFloat(tt.v)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatVolumeFloat(%f) = %q, want to contain %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestFormatVolume(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1500", "$1.5K"},
		{"0", "$0"},
		{"invalid", "invalid"},
		{"2500000", "$2.5M"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatVolume(tt.input)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatVolume(%q) = %q, want to contain %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.65, "65.0%"},
		{0.0, "0.0%"},
		{1.0, "100.0%"},
		{0.123, "12.3%"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatPercent(tt.input)
			if got != tt.want {
				t.Errorf("FormatPercent(%f) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 2, "hi"},
		{"abc", 1, "…"},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := Truncate(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		input string
		want  string // just check it's non-empty or matches expected
	}{
		{"", "—"},
		{"2024-01-15T10:30:00Z", "Jan 15, 2024 10:30"},
		{"2024-01-15", "Jan 15, 2024 00:00"},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatTimestamp(tt.input)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatTimestamp(%q) = %q, want to contain %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatPricePtr(t *testing.T) {
	val := 0.65

	// nil price should return "—" (with styling)
	got := FormatPricePtr(nil)
	if !strings.Contains(got, "—") {
		t.Errorf("FormatPricePtr(nil) = %q, want to contain —", got)
	}

	// non-nil price should contain the formatted value
	got = FormatPricePtr(&val)
	if !strings.Contains(got, "65.0") {
		t.Errorf("FormatPricePtr(&0.65) = %q, want to contain 65.0", got)
	}
}
