package clob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

const (
	DefaultClobBaseURL = "https://clob.polymarket.com"
	defaultTimeout     = 30 * time.Second
)

// Client interacts with the Polymarket CLOB API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       *AuthCredentials
}

// NewClient creates a new unauthenticated CLOB API client.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultClobBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// NewAuthenticatedClient creates a CLOB API client with auth credentials.
func NewAuthenticatedClient(baseURL string, auth *AuthCredentials) *Client {
	if baseURL == "" {
		baseURL = DefaultClobBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		auth: auth,
	}
}

// IsAuthenticated returns true if the client has auth credentials.
func (c *Client) IsAuthenticated() bool {
	return c.auth != nil
}

// doGet performs a GET request and returns the response body.
func (c *Client) doGet(path string) ([]byte, error) {
	u := c.baseURL + path
	log.Debug("CLOB GET", "url", u)

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CLOB API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// doGetWithBody performs a GET request with a JSON body (used by batch endpoints).
func (c *Client) doGetWithBody(path string, body any) ([]byte, error) {
	u := c.baseURL + path
	log.Debug("CLOB GET+body", "url", u)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("GET", u, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CLOB API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// HealthCheck checks if the CLOB API is reachable.
func (c *Client) HealthCheck() (string, error) {
	data, err := c.doGet("/")
	if err != nil {
		return "", err
	}
	return strings.Trim(string(data), `"`), nil
}

// ServerTime returns the CLOB server timestamp.
func (c *Client) ServerTime() (string, error) {
	data, err := c.doGet("/time")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// GetPrice returns the price for a token on a given side.
func (c *Client) GetPrice(tokenID, side string) (*PriceResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	params.Set("side", side)
	data, err := c.doGet("/price?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp PriceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing price: %w", err)
	}
	return &resp, nil
}

// GetBatchPrices returns prices for multiple tokens.
func (c *Client) GetBatchPrices(tokenIDs []string, side string) (map[string]BatchPriceEntry, error) {
	body := make([]BatchTokenRequest, len(tokenIDs))
	for i, id := range tokenIDs {
		body[i] = BatchTokenRequest{TokenID: id, Side: side}
	}
	data, err := c.doGetWithBody("/prices", body)
	if err != nil {
		return nil, err
	}
	var resp map[string]BatchPriceEntry
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing batch prices: %w", err)
	}
	return resp, nil
}

// GetMidpoint returns the midpoint price for a token.
func (c *Client) GetMidpoint(tokenID string) (*MidpointResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/midpoint?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp MidpointResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing midpoint: %w", err)
	}
	return &resp, nil
}

// GetBatchMidpoints returns midpoints for multiple tokens.
// Response is map[tokenID]midpointString
func (c *Client) GetBatchMidpoints(tokenIDs []string) (map[string]string, error) {
	body := make([]BatchTokenRequest, len(tokenIDs))
	for i, id := range tokenIDs {
		body[i] = BatchTokenRequest{TokenID: id}
	}
	data, err := c.doGetWithBody("/midpoints", body)
	if err != nil {
		return nil, err
	}
	var resp map[string]string
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing batch midpoints: %w", err)
	}
	return resp, nil
}

// GetSpread returns the bid-ask spread for a token.
func (c *Client) GetSpread(tokenID string) (*SpreadResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/spread?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp SpreadResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing spread: %w", err)
	}
	return &resp, nil
}

// GetBatchSpreads returns spreads for multiple tokens.
func (c *Client) GetBatchSpreads(tokenIDs []string) (map[string]string, error) {
	body := make([]BatchTokenRequest, len(tokenIDs))
	for i, id := range tokenIDs {
		body[i] = BatchTokenRequest{TokenID: id}
	}
	data, err := c.doGetWithBody("/spreads", body)
	if err != nil {
		return nil, err
	}
	var resp map[string]string
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing batch spreads: %w", err)
	}
	return resp, nil
}

// GetBook returns the full order book for a token.
func (c *Client) GetBook(tokenID string) (*OrderBook, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/book?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp OrderBook
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing order book: %w", err)
	}
	return &resp, nil
}

// GetBatchBooks returns order books for multiple tokens.
func (c *Client) GetBatchBooks(tokenIDs []string) (map[string]OrderBook, error) {
	body := make([]BatchTokenRequest, len(tokenIDs))
	for i, id := range tokenIDs {
		body[i] = BatchTokenRequest{TokenID: id}
	}
	data, err := c.doGetWithBody("/books", body)
	if err != nil {
		return nil, err
	}
	var resp map[string]OrderBook
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing batch books: %w", err)
	}
	return resp, nil
}

// GetLastTradePrice returns the last trade price for a token.
func (c *Client) GetLastTradePrice(tokenID string) (*LastTradeResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/last-trade-price?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp LastTradeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing last trade: %w", err)
	}
	return &resp, nil
}

// GetBatchLastTrades returns last trade prices for multiple tokens.
func (c *Client) GetBatchLastTrades(tokenIDs []string) (map[string]LastTradeResponse, error) {
	body := make([]BatchTokenRequest, len(tokenIDs))
	for i, id := range tokenIDs {
		body[i] = BatchTokenRequest{TokenID: id}
	}
	data, err := c.doGetWithBody("/last-trades-prices", body)
	if err != nil {
		return nil, err
	}
	var resp map[string]LastTradeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing batch last trades: %w", err)
	}
	return resp, nil
}

// GetMarket returns a single CLOB market by condition ID.
func (c *Client) GetMarket(conditionID string) (*ClobMarket, error) {
	data, err := c.doGet("/markets/" + conditionID)
	if err != nil {
		return nil, err
	}
	var resp ClobMarket
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing CLOB market: %w", err)
	}
	return &resp, nil
}

// ListMarkets returns paginated CLOB markets.
func (c *Client) ListMarkets(nextCursor string) (*ClobMarketsResponse, error) {
	path := "/markets"
	if nextCursor != "" {
		params := url.Values{}
		params.Set("next_cursor", nextCursor)
		path += "?" + params.Encode()
	}
	data, err := c.doGet(path)
	if err != nil {
		return nil, err
	}
	var resp ClobMarketsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing CLOB markets: %w", err)
	}
	return &resp, nil
}

// GetSamplingMarkets returns reward-eligible markets.
func (c *Client) GetSamplingMarkets() ([]ClobMarket, error) {
	data, err := c.doGet("/sampling-markets")
	if err != nil {
		return nil, err
	}
	var resp ClobMarketsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing sampling markets: %w", err)
	}
	return resp.Data, nil
}

// GetSimplifiedMarkets returns reduced-detail markets.
func (c *Client) GetSimplifiedMarkets() ([]ClobMarket, error) {
	data, err := c.doGet("/simplified-markets")
	if err != nil {
		return nil, err
	}
	var resp ClobMarketsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing simplified markets: %w", err)
	}
	return resp.Data, nil
}

// GetTickSize returns the tick size for a token.
func (c *Client) GetTickSize(tokenID string) (*TickSizeResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/tick-size?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp TickSizeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing tick size: %w", err)
	}
	return &resp, nil
}

// GetFeeRate returns the fee rate for a token.
func (c *Client) GetFeeRate(tokenID string) (*FeeRateResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/fee-rate?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp FeeRateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing fee rate: %w", err)
	}
	return &resp, nil
}

// GetNegRisk returns the neg-risk flag for a token.
func (c *Client) GetNegRisk(tokenID string) (*NegRiskResponse, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	data, err := c.doGet("/neg-risk?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp NegRiskResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing neg risk: %w", err)
	}
	return &resp, nil
}

// GetPriceHistory returns price history for a token.
// interval: 1m, 1h, 6h, 1d, 1w, max
// fidelity: number of data points
func (c *Client) GetPriceHistory(tokenID, interval string, fidelity int) (*PriceHistoryResponse, error) {
	params := url.Values{}
	params.Set("market", tokenID)
	params.Set("interval", interval)
	if fidelity > 0 {
		params.Set("fidelity", fmt.Sprintf("%d", fidelity))
	}
	data, err := c.doGet("/prices-history?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var resp PriceHistoryResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing price history: %w", err)
	}
	return &resp, nil
}
