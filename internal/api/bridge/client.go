package bridge

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

const (
	DefaultBridgeBaseURL = "https://bridge.polymarket.com"
	defaultTimeout       = 30 * time.Second
)

// BridgeClient interacts with the Polymarket Bridge API.
type BridgeClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBridgeClient creates a new Bridge API client.
func NewBridgeClient(baseURL string) *BridgeClient {
	if baseURL == "" {
		baseURL = DefaultBridgeBaseURL
	}
	return &BridgeClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// get performs a GET request and returns the response body.
func (c *BridgeClient) get(path string) ([]byte, error) {
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

// GetDepositAddresses retrieves deposit addresses for supported chains.
func (c *BridgeClient) GetDepositAddresses() (*DepositAddresses, error) {
	data, err := c.get("/deposit-addresses")
	if err != nil {
		return nil, err
	}

	var addrs DepositAddresses
	if err := json.Unmarshal(data, &addrs); err != nil {
		return nil, fmt.Errorf("parsing deposit addresses: %w", err)
	}
	return &addrs, nil
}

// GetSupportedAssets retrieves the list of supported assets for bridging.
func (c *BridgeClient) GetSupportedAssets() ([]SupportedAsset, error) {
	data, err := c.get("/supported-assets")
	if err != nil {
		return nil, err
	}

	var assets []SupportedAsset
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, fmt.Errorf("parsing supported assets: %w", err)
	}
	return assets, nil
}

// GetDepositStatus retrieves the status of a deposit by transaction hash.
func (c *BridgeClient) GetDepositStatus(txHash string) (*DepositStatus, error) {
	data, err := c.get("/deposit-status?txHash=" + txHash)
	if err != nil {
		return nil, err
	}

	var status DepositStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("parsing deposit status: %w", err)
	}
	return &status, nil
}
