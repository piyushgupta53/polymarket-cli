package api

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// QueryParams holds query parameters for API requests.
type QueryParams map[string]string

// BuildQueryString constructs a URL query string from parameters.
func BuildQueryString(params QueryParams) string {
	if len(params) == 0 {
		return ""
	}

	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Set(k, v)
		}
	}

	encoded := values.Encode()
	if encoded == "" {
		return ""
	}
	return "?" + encoded
}

// MarketListParams holds parameters for listing markets.
type MarketListParams struct {
	Limit     int
	Offset    int
	Active    *bool
	Closed    *bool
	Slug      string
	Order     string // e.g. "volume_24hr", "liquidity"
	Ascending bool
}

// ToQueryParams converts MarketListParams to QueryParams.
func (p MarketListParams) ToQueryParams() QueryParams {
	params := QueryParams{}
	if p.Limit > 0 {
		params["limit"] = strconv.Itoa(p.Limit)
	}
	if p.Offset > 0 {
		params["offset"] = strconv.Itoa(p.Offset)
	}
	if p.Active != nil {
		params["active"] = strconv.FormatBool(*p.Active)
	}
	if p.Closed != nil {
		params["closed"] = strconv.FormatBool(*p.Closed)
	}
	if p.Slug != "" {
		params["slug"] = p.Slug
	}
	if p.Order != "" {
		params["order"] = p.Order
		if p.Ascending {
			params["ascending"] = "true"
		}
	}
	return params
}

// EventListParams holds parameters for listing events.
type EventListParams struct {
	Limit   int
	Offset  int
	Active  *bool
	Closed  *bool
	Slug    string
	Order   string
	TagSlug string
}

// ToQueryParams converts EventListParams to QueryParams.
func (p EventListParams) ToQueryParams() QueryParams {
	params := QueryParams{}
	if p.Limit > 0 {
		params["limit"] = strconv.Itoa(p.Limit)
	}
	if p.Offset > 0 {
		params["offset"] = strconv.Itoa(p.Offset)
	}
	if p.Active != nil {
		params["active"] = strconv.FormatBool(*p.Active)
	}
	if p.Closed != nil {
		params["closed"] = strconv.FormatBool(*p.Closed)
	}
	if p.Slug != "" {
		params["slug"] = p.Slug
	}
	if p.Order != "" {
		params["order"] = p.Order
	}
	if p.TagSlug != "" {
		params["tag_slug"] = p.TagSlug
	}
	return params
}

// IsNumericID checks if a string is a numeric ID (vs a slug).
func IsNumericID(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// ContainsIgnoreCase checks if haystack contains needle (case-insensitive).
func ContainsIgnoreCase(haystack, needle string) bool {
	return strings.Contains(
		strings.ToLower(haystack),
		strings.ToLower(needle),
	)
}

// FilterMarkets filters markets by a search query (client-side).
func FilterMarkets(markets []Market, query string) []Market {
	if query == "" {
		return markets
	}

	var filtered []Market
	for _, m := range markets {
		if ContainsIgnoreCase(m.Question, query) ||
			ContainsIgnoreCase(m.Description, query) ||
			ContainsIgnoreCase(m.Slug, query) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// BoolPtr returns a pointer to a bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// FormatFloat64Ptr formats a *float64 for display, returning "—" for nil.
func FormatFloat64Ptr(f *float64) string {
	if f == nil {
		return "—"
	}
	return fmt.Sprintf("%.4f", *f)
}
