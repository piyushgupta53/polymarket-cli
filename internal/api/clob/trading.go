package clob

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// PostOrder submits a signed order to the CLOB.
func (c *Client) PostOrder(order *OrderPayload) (*OrderResponse, error) {
	body, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("marshaling order: %w", err)
	}
	data, err := c.doAuthPost("/order", body)
	if err != nil {
		return nil, err
	}
	var resp OrderResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing order response: %w", err)
	}
	return &resp, nil
}

// CancelOrder cancels a single order by ID.
func (c *Client) CancelOrder(orderID string) (*CancelResponse, error) {
	body, err := json.Marshal(map[string]string{"orderID": orderID})
	if err != nil {
		return nil, fmt.Errorf("marshaling cancel: %w", err)
	}
	data, err := c.doAuthDeleteWithBody("/order", body)
	if err != nil {
		return nil, err
	}
	var resp CancelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing cancel response: %w", err)
	}
	return &resp, nil
}

// CancelOrders cancels multiple orders by ID.
func (c *Client) CancelOrders(orderIDs []string) (*CancelResponse, error) {
	body, err := json.Marshal(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("marshaling cancel: %w", err)
	}
	data, err := c.doAuthDeleteWithBody("/orders", body)
	if err != nil {
		return nil, err
	}
	var resp CancelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing cancel response: %w", err)
	}
	return &resp, nil
}

// CancelMarketOrders cancels all orders for a market (condition ID).
func (c *Client) CancelMarketOrders(conditionID string) (*CancelResponse, error) {
	body, err := json.Marshal(map[string]string{"market": conditionID})
	if err != nil {
		return nil, fmt.Errorf("marshaling cancel: %w", err)
	}
	data, err := c.doAuthDeleteWithBody("/cancel-market-orders", body)
	if err != nil {
		return nil, err
	}
	var resp CancelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing cancel response: %w", err)
	}
	return &resp, nil
}

// CancelAll cancels all open orders.
func (c *Client) CancelAll() (*CancelResponse, error) {
	data, err := c.doAuthDelete("/cancel-all")
	if err != nil {
		return nil, err
	}
	var resp CancelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing cancel response: %w", err)
	}
	return &resp, nil
}

// GetOrder returns a single order by ID.
func (c *Client) GetOrder(orderID string) (*Order, error) {
	data, err := c.doAuthGet("/order/" + orderID)
	if err != nil {
		return nil, err
	}
	var resp Order
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing order: %w", err)
	}
	return &resp, nil
}

// GetOpenOrders returns open orders with optional filtering.
func (c *Client) GetOpenOrders(params *OpenOrdersParams) ([]Order, error) {
	path := "/orders"
	if params != nil {
		q := url.Values{}
		if params.Market != "" {
			q.Set("market", params.Market)
		}
		if params.Asset != "" {
			q.Set("asset_id", params.Asset)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}
	data, err := c.doAuthGet(path)
	if err != nil {
		return nil, err
	}
	var resp []Order
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing orders: %w", err)
	}
	return resp, nil
}

// GetTrades returns trade history with optional filtering.
func (c *Client) GetTrades(params *TradesParams) ([]Trade, error) {
	path := "/trades"
	if params != nil {
		q := url.Values{}
		if params.Market != "" {
			q.Set("market", params.Market)
		}
		if params.Asset != "" {
			q.Set("asset_id", params.Asset)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}
	data, err := c.doAuthGet(path)
	if err != nil {
		return nil, err
	}
	var resp []Trade
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing trades: %w", err)
	}
	return resp, nil
}

// GetBalanceAllowance returns balance and allowance info for a token.
func (c *Client) GetBalanceAllowance(params *BalanceParams) (*BalanceAllowance, error) {
	q := url.Values{}
	if params != nil {
		if params.AssetType != "" {
			q.Set("asset_type", params.AssetType)
		}
		if params.TokenID != "" {
			q.Set("token_id", params.TokenID)
		}
	}
	path := "/balance-allowance"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	data, err := c.doAuthGet(path)
	if err != nil {
		return nil, err
	}
	var resp BalanceAllowance
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing balance: %w", err)
	}
	return &resp, nil
}

// UpdateBalanceAllowance triggers a balance/allowance refresh.
func (c *Client) UpdateBalanceAllowance() error {
	_, err := c.doAuthPost("/balance-allowance", nil)
	return err
}
