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

// getJSON performs a GET request and streams the JSON response into dest.
func (c *BridgeClient) getJSON(path string, dest any) error {
	url := c.baseURL + path
	log.Debug("GET", "url", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return nil
}

// GetDepositAddresses retrieves deposit addresses for supported chains.
func (c *BridgeClient) GetDepositAddresses() (*DepositAddresses, error) {
	var addrs DepositAddresses
	if err := c.getJSON("/deposit-addresses", &addrs); err != nil {
		return nil, err
	}
	return &addrs, nil
}

// GetSupportedAssets retrieves the list of supported assets for bridging.
func (c *BridgeClient) GetSupportedAssets() ([]SupportedAsset, error) {
	var assets []SupportedAsset
	if err := c.getJSON("/supported-assets", &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

// GetDepositStatus retrieves the status of a deposit by transaction hash.
func (c *BridgeClient) GetDepositStatus(txHash string) (*DepositStatus, error) {
	var status DepositStatus
	if err := c.getJSON("/deposit-status?txHash="+txHash, &status); err != nil {
		return nil, err
	}
	return &status, nil
}
