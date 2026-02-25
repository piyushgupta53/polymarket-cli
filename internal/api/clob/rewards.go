package clob

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// GetCurrentRewards returns current reward rates.
func (c *Client) GetCurrentRewards() (json.RawMessage, error) {
	data, err := c.doAuthGet("/rewards")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetMarketRewards returns reward info for a specific market.
func (c *Client) GetMarketRewards(conditionID string) (json.RawMessage, error) {
	data, err := c.doAuthGet("/rewards/markets/" + conditionID)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetEpochEarnings returns earnings for the current epoch.
func (c *Client) GetEpochEarnings() (json.RawMessage, error) {
	data, err := c.doAuthGet("/rewards/earnings")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetEpochTotal returns total earnings.
func (c *Client) GetEpochTotal() (json.RawMessage, error) {
	data, err := c.doAuthGet("/rewards/earnings/total")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetRewardPercentages returns reward percentage configuration.
func (c *Client) GetRewardPercentages() (json.RawMessage, error) {
	data, err := c.doAuthGet("/rewards/percentages")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetOrderScoring returns the scoring info for a single order.
func (c *Client) GetOrderScoring(orderID string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("order_id", orderID)
	data, err := c.doAuthGet("/order-scoring?" + params.Encode())
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetOrdersScoring returns scoring info for multiple orders.
func (c *Client) GetOrdersScoring(orderIDs []string) (json.RawMessage, error) {
	body, err := json.Marshal(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("marshaling order IDs: %w", err)
	}
	// This endpoint uses GET with body
	data, err := c.doAuthRequest("GET", "/orders-scoring", body)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetAPIKeys returns API keys for the authenticated user.
func (c *Client) GetAPIKeys() (json.RawMessage, error) {
	data, err := c.doAuthGet("/auth/api-keys")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// DeleteAPIKey deletes an API key.
func (c *Client) DeleteAPIKey(apiKey string) error {
	body, err := json.Marshal(map[string]string{"api_key": apiKey})
	if err != nil {
		return fmt.Errorf("marshaling delete: %w", err)
	}
	_, err = c.doAuthDeleteWithBody("/auth/api-key", body)
	return err
}

// GetNotifications returns user notifications.
func (c *Client) GetNotifications() (json.RawMessage, error) {
	data, err := c.doAuthGet("/notifications")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// DeleteNotifications deletes all user notifications.
func (c *Client) DeleteNotifications() error {
	_, err := c.doAuthDelete("/notifications")
	return err
}
