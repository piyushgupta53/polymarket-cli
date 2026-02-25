package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/api"
)

const (
	DefaultDataBaseURL = "https://data-api.polymarket.com"
	defaultTimeout     = 30 * time.Second
)

// DataClient interacts with the Polymarket Data API.
type DataClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewDataClient creates a new Data API client.
func NewDataClient(baseURL string) *DataClient {
	if baseURL == "" {
		baseURL = DefaultDataBaseURL
	}
	return &DataClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// get performs a GET request and returns the response body.
func (c *DataClient) get(path string) ([]byte, error) {
	url := c.baseURL + path
	log.Debug("GET", "url", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetPositions retrieves open positions for an address.
func (c *DataClient) GetPositions(address string) ([]Position, error) {
	data, err := c.get("/positions" + api.BuildQueryString(api.QueryParams{"address": address}))
	if err != nil {
		return nil, err
	}

	var positions []Position
	if err := json.Unmarshal(data, &positions); err != nil {
		return nil, fmt.Errorf("parsing positions: %w", err)
	}
	return positions, nil
}

// GetClosedPositions retrieves closed positions for an address.
func (c *DataClient) GetClosedPositions(address string) ([]Position, error) {
	data, err := c.get("/positions/closed" + api.BuildQueryString(api.QueryParams{"address": address}))
	if err != nil {
		return nil, err
	}

	var positions []Position
	if err := json.Unmarshal(data, &positions); err != nil {
		return nil, fmt.Errorf("parsing closed positions: %w", err)
	}
	return positions, nil
}

// GetPortfolioValue retrieves portfolio value history for an address.
func (c *DataClient) GetPortfolioValue(address string) ([]PortfolioValue, error) {
	data, err := c.get("/value" + api.BuildQueryString(api.QueryParams{"address": address}))
	if err != nil {
		return nil, err
	}

	var values []PortfolioValue
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("parsing portfolio value: %w", err)
	}
	return values, nil
}

// GetPortfolioTraded retrieves total volume traded by an address.
func (c *DataClient) GetPortfolioTraded(address string) (json.RawMessage, error) {
	data, err := c.get("/traded" + api.BuildQueryString(api.QueryParams{"address": address}))
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetActivity retrieves account activity for an address.
func (c *DataClient) GetActivity(address string, limit, offset int) ([]Activity, error) {
	params := api.QueryParams{"address": address}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["offset"] = fmt.Sprintf("%d", offset)
	}
	data, err := c.get("/activity" + api.BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var activities []Activity
	if err := json.Unmarshal(data, &activities); err != nil {
		return nil, fmt.Errorf("parsing activity: %w", err)
	}
	return activities, nil
}

// GetTrades retrieves trades for an address.
func (c *DataClient) GetTrades(address string, limit, offset int) ([]TradeRecord, error) {
	params := api.QueryParams{"address": address}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["offset"] = fmt.Sprintf("%d", offset)
	}
	data, err := c.get("/trades" + api.BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var trades []TradeRecord
	if err := json.Unmarshal(data, &trades); err != nil {
		return nil, fmt.Errorf("parsing trades: %w", err)
	}
	return trades, nil
}

// GetHolders retrieves top holders for a market condition.
func (c *DataClient) GetHolders(conditionID string) ([]Holder, error) {
	data, err := c.get("/holders" + api.BuildQueryString(api.QueryParams{"conditionId": conditionID}))
	if err != nil {
		return nil, err
	}

	var holders []Holder
	if err := json.Unmarshal(data, &holders); err != nil {
		return nil, fmt.Errorf("parsing holders: %w", err)
	}
	return holders, nil
}

// GetOpenInterest retrieves open interest for a market condition.
func (c *DataClient) GetOpenInterest(conditionID string) (*OpenInterest, error) {
	data, err := c.get("/open-interest" + api.BuildQueryString(api.QueryParams{"conditionId": conditionID}))
	if err != nil {
		return nil, err
	}

	var oi OpenInterest
	if err := json.Unmarshal(data, &oi); err != nil {
		return nil, fmt.Errorf("parsing open interest: %w", err)
	}
	return &oi, nil
}

// GetEventVolume retrieves volume data for an event.
func (c *DataClient) GetEventVolume(eventID string) (*EventVolume, error) {
	data, err := c.get("/volume" + api.BuildQueryString(api.QueryParams{"eventId": eventID}))
	if err != nil {
		return nil, err
	}

	var vol EventVolume
	if err := json.Unmarshal(data, &vol); err != nil {
		return nil, fmt.Errorf("parsing event volume: %w", err)
	}
	return &vol, nil
}

// GetLeaderboard retrieves the trading leaderboard.
func (c *DataClient) GetLeaderboard(limit, offset int) ([]LeaderboardEntry, error) {
	params := api.QueryParams{}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["offset"] = fmt.Sprintf("%d", offset)
	}
	data, err := c.get("/leaderboard" + api.BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var entries []LeaderboardEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing leaderboard: %w", err)
	}
	return entries, nil
}

// GetBuilderLeaderboard retrieves the builder leaderboard.
func (c *DataClient) GetBuilderLeaderboard() (json.RawMessage, error) {
	data, err := c.get("/builder-leaderboard")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetBuilderVolume retrieves builder volume data.
func (c *DataClient) GetBuilderVolume() (json.RawMessage, error) {
	data, err := c.get("/builder-volume")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
