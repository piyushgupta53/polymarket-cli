package api

import (
	"encoding/json"
	"time"
)

// Market represents a Polymarket market from the Gamma API.
type Market struct {
	ID                 string   `json:"id"`
	Question           string   `json:"question"`
	ConditionID        string   `json:"conditionId"`
	Slug               string   `json:"slug"`
	EndDateISO         string   `json:"endDateIso"`
	GameStartTime      string   `json:"gameStartTime,omitempty"`
	Description        string   `json:"description"`
	Outcomes           string   `json:"outcomes"` // JSON-stringified array e.g. `["Yes","No"]`
	OutcomePrices      string   `json:"outcomePrices"` // JSON-stringified array e.g. `["0.65","0.35"]`
	Active             bool     `json:"active"`
	Closed             bool     `json:"closed"`
	MarketMakerAddress string   `json:"marketMakerAddress,omitempty"`
	Volume             string   `json:"volume,omitempty"`
	Volume24hr         float64  `json:"volume24hr,omitempty"`
	Liquidity          string   `json:"liquidity,omitempty"`
	BestBid            *float64 `json:"bestBid"`
	BestAsk            *float64 `json:"bestAsk"`
	LastTradePrice     *float64 `json:"lastTradePrice"`
	Image              string   `json:"image,omitempty"`
	Icon               string   `json:"icon,omitempty"`
	TokenID            string   `json:"tokenId,omitempty"`
	ClobTokenIds       string   `json:"clobTokenIds,omitempty"` // JSON-stringified array of token IDs

	// Tokens associated with the market
	Tokens []Token `json:"tokens,omitempty"`

	// Timestamps
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// FirstTokenID returns the first token ID from Tokens, ClobTokenIds, or TokenID.
func (m *Market) FirstTokenID() string {
	if len(m.Tokens) > 0 {
		return m.Tokens[0].TokenID
	}
	if ids, _ := ParseClobTokenIds(m.ClobTokenIds); len(ids) > 0 {
		return ids[0]
	}
	return m.TokenID
}

// Token represents a conditional token for a market outcome.
type Token struct {
	TokenID  string  `json:"token_id"`
	Outcome  string  `json:"outcome"`
	Price    float64 `json:"price"`
	Winner   bool    `json:"winner"`
}

// Event represents a Polymarket event from the Gamma API.
type Event struct {
	ID              string   `json:"id"`
	Ticker          string   `json:"ticker"`
	Slug            string   `json:"slug"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	StartDate       string   `json:"startDate,omitempty"`
	EndDate         string   `json:"endDate,omitempty"`
	CreationDate    string   `json:"creationDate,omitempty"`
	Active          bool     `json:"active"`
	Closed          bool     `json:"closed"`
	Archived        bool     `json:"archived"`
	New             bool     `json:"new"`
	Featured        bool     `json:"featured"`
	Restricted      bool     `json:"restricted"`
	Volume          float64  `json:"volume"`
	Liquidity       float64  `json:"liquidity"`
	Volume24hr      float64  `json:"volume24hr"`
	Markets         []Market `json:"markets,omitempty"`
	CommentCount    int      `json:"commentCount"`
	Image           string   `json:"image,omitempty"`
	Icon            string   `json:"icon,omitempty"`

	// Tags
	Tags []Tag `json:"tags,omitempty"`
}

// Tag represents a topic tag.
type Tag struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Slug      string `json:"slug"`
	ForceShow bool   `json:"forceShow,omitempty"`
}

// PaginatedResponse wraps a paginated API response.
type PaginatedResponse[T any] struct {
	Data   []T `json:"data"`
	Count  int `json:"count,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ParseOutcomePrices parses the JSON-stringified outcome_prices field.
func ParseOutcomePrices(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var prices []string
	if err := json.Unmarshal([]byte(raw), &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

// ParseClobTokenIds parses the JSON-stringified clobTokenIds field.
func ParseClobTokenIds(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// ParseOutcomes parses the JSON-stringified outcomes field.
func ParseOutcomes(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var outcomes []string
	if err := json.Unmarshal([]byte(raw), &outcomes); err != nil {
		return nil, err
	}
	return outcomes, nil
}

// Series represents a series of related events.
type Series struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description string  `json:"description,omitempty"`
	Events      []Event `json:"events,omitempty"`
}

// Comment represents a user comment on an entity.
type Comment struct {
	ID         string `json:"id"`
	Author     string `json:"author"`
	Body       string `json:"body"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	CreatedAt  string `json:"createdAt"`
}

// Profile represents a Polymarket user profile.
type Profile struct {
	Address      string  `json:"address"`
	Username     string  `json:"username,omitempty"`
	Bio          string  `json:"bio,omitempty"`
	ProfileImage string  `json:"profileImage,omitempty"`
	Volume       float64 `json:"volume,omitempty"`
}

// Sport represents a sport category.
type Sport struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// SportMarketType represents a type of sports market.
type SportMarketType struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Sport string `json:"sport"`
}

// Team represents a sports team.
type Team struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	League string `json:"league,omitempty"`
	Slug   string `json:"slug,omitempty"`
}

// ParseTime parses a timestamp string in common formats.
func ParseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Value: s}
}
