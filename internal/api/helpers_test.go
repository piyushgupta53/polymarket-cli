package api

import (
	"testing"
)

func TestBuildQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params QueryParams
		want   string
	}{
		{
			name:   "empty params",
			params: QueryParams{},
			want:   "",
		},
		{
			name:   "single param",
			params: QueryParams{"limit": "10"},
			want:   "?limit=10",
		},
		{
			name:   "skip empty values",
			params: QueryParams{"limit": "10", "offset": ""},
			want:   "?limit=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildQueryString(tt.params)
			if tt.want == "" && got != "" {
				t.Errorf("BuildQueryString() = %q, want empty", got)
			}
			if tt.want != "" && got == "" {
				t.Errorf("BuildQueryString() = empty, want %q", tt.want)
			}
		})
	}
}

func TestIsNumericID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"0", true},
		{"abc", false},
		{"will-trump-win", false},
		{"12abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsNumericID(tt.input); got != tt.want {
				t.Errorf("IsNumericID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		want     bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.haystack+"_"+tt.needle, func(t *testing.T) {
			if got := ContainsIgnoreCase(tt.haystack, tt.needle); got != tt.want {
				t.Errorf("ContainsIgnoreCase(%q, %q) = %v, want %v", tt.haystack, tt.needle, got, tt.want)
			}
		})
	}
}

func TestFilterMarkets(t *testing.T) {
	markets := []Market{
		{Question: "Will Trump win?", Slug: "trump-win"},
		{Question: "Will Biden win?", Slug: "biden-win"},
		{Question: "BTC price above 100k?", Slug: "btc-100k"},
	}

	filtered := FilterMarkets(markets, "trump")
	if len(filtered) != 1 {
		t.Errorf("FilterMarkets() returned %d results, want 1", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Slug != "trump-win" {
		t.Errorf("FilterMarkets() got slug %q, want %q", filtered[0].Slug, "trump-win")
	}

	all := FilterMarkets(markets, "")
	if len(all) != 3 {
		t.Errorf("FilterMarkets('') returned %d results, want 3", len(all))
	}

	// Search in slug
	btc := FilterMarkets(markets, "btc")
	if len(btc) != 1 {
		t.Errorf("FilterMarkets('btc') returned %d results, want 1", len(btc))
	}
}

func TestMarketListParamsToQueryParams(t *testing.T) {
	active := true
	params := MarketListParams{
		Limit:  10,
		Offset: 5,
		Active: &active,
		Order:  "liquidity",
	}

	qp := params.ToQueryParams()
	if qp["limit"] != "10" {
		t.Errorf("limit = %q, want %q", qp["limit"], "10")
	}
	if qp["offset"] != "5" {
		t.Errorf("offset = %q, want %q", qp["offset"], "5")
	}
	if qp["active"] != "true" {
		t.Errorf("active = %q, want %q", qp["active"], "true")
	}
	if qp["order"] != "liquidity" {
		t.Errorf("order = %q, want %q", qp["order"], "liquidity")
	}
}

func TestFormatFloat64Ptr(t *testing.T) {
	val := 0.6523
	tests := []struct {
		name  string
		input *float64
		want  string
	}{
		{"nil", nil, "—"},
		{"value", &val, "0.6523"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFloat64Ptr(tt.input)
			if got != tt.want {
				t.Errorf("FormatFloat64Ptr() = %q, want %q", got, tt.want)
			}
		})
	}
}
