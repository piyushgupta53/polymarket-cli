package clob

// OrderBookEntry represents a single bid or ask in the order book.
type OrderBookEntry struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// OrderBook represents a full order book for a token.
type OrderBook struct {
	Market         string           `json:"market"`
	AssetID        string           `json:"asset_id"`
	Timestamp      string           `json:"timestamp"`
	Hash           string           `json:"hash"`
	Bids           []OrderBookEntry `json:"bids"`
	Asks           []OrderBookEntry `json:"asks"`
	MinOrderSize   string           `json:"min_order_size"`
	TickSize       string           `json:"tick_size"`
	NegRisk        bool             `json:"neg_risk"`
	LastTradePrice string           `json:"last_trade_price"`
}

// ClobToken represents a token within a CLOB market.
type ClobToken struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
	Winner  bool    `json:"winner"`
}

// ClobRewards holds reward configuration for a market.
type ClobRewards struct {
	Rates     []any   `json:"rates"`
	MinSize   float64 `json:"min_size"`
	MaxSpread float64 `json:"max_spread"`
}

// ClobMarket represents a market from the CLOB API.
type ClobMarket struct {
	EnableOrderBook          bool        `json:"enable_order_book"`
	Active                   bool        `json:"active"`
	Closed                   bool        `json:"closed"`
	Archived                 bool        `json:"archived"`
	AcceptingOrders          bool        `json:"accepting_orders"`
	AcceptingOrderTimestamp   *string     `json:"accepting_order_timestamp"`
	MinimumOrderSize         float64     `json:"minimum_order_size"`
	MinimumTickSize          float64     `json:"minimum_tick_size"`
	ConditionID              string      `json:"condition_id"`
	QuestionID               string      `json:"question_id"`
	Question                 string      `json:"question"`
	Description              string      `json:"description"`
	MarketSlug               string      `json:"market_slug"`
	EndDateISO               string      `json:"end_date_iso"`
	GameStartTime            string      `json:"game_start_time,omitempty"`
	SecondsDelay             int         `json:"seconds_delay"`
	Fpmm                     string      `json:"fpmm"`
	MakerBaseFee             float64     `json:"maker_base_fee"`
	TakerBaseFee             float64     `json:"taker_base_fee"`
	NotificationsEnabled     bool        `json:"notifications_enabled"`
	NegRisk                  bool        `json:"neg_risk"`
	NegRiskMarketID          string      `json:"neg_risk_market_id"`
	NegRiskRequestID         string      `json:"neg_risk_request_id"`
	Icon                     string      `json:"icon"`
	Image                    string      `json:"image"`
	Rewards                  ClobRewards `json:"rewards"`
	Is5050Outcome            bool        `json:"is_50_50_outcome"`
	Tokens                   []ClobToken `json:"tokens"`
	Tags                     []string    `json:"tags"`
}

// ClobMarketsResponse is the paginated response from /markets.
type ClobMarketsResponse struct {
	Data       []ClobMarket `json:"data"`
	NextCursor string       `json:"next_cursor"`
	Limit      int          `json:"limit"`
	Count      int          `json:"count"`
}

// PriceResponse is the response from /price.
type PriceResponse struct {
	Price string `json:"price"`
}

// MidpointResponse is the response from /midpoint.
type MidpointResponse struct {
	Mid string `json:"mid"`
}

// SpreadResponse is the response from /spread.
type SpreadResponse struct {
	Spread string `json:"spread"`
}

// LastTradeResponse is the response from /last-trade-price.
type LastTradeResponse struct {
	Price string `json:"price"`
	Side  string `json:"side"`
}

// TickSizeResponse is the response from /tick-size.
type TickSizeResponse struct {
	MinimumTickSize float64 `json:"minimum_tick_size"`
}

// FeeRateResponse is the response from /fee-rate.
type FeeRateResponse struct {
	BaseFee float64 `json:"base_fee"`
}

// NegRiskResponse is the response from /neg-risk.
type NegRiskResponse struct {
	NegRisk bool `json:"neg_risk"`
}

// PriceHistoryPoint represents a single point in price history.
type PriceHistoryPoint struct {
	Timestamp int64   `json:"t"`
	Price     float64 `json:"p"`
}

// PriceHistoryResponse is the response from /prices-history.
type PriceHistoryResponse struct {
	History []PriceHistoryPoint `json:"history"`
}

// BatchPriceEntry holds buy/sell prices for a batch query.
type BatchPriceEntry struct {
	Buy  string `json:"BUY,omitempty"`
	Sell string `json:"SELL,omitempty"`
}

// BatchTokenRequest is the request body for batch endpoints.
type BatchTokenRequest struct {
	TokenID string `json:"token_id"`
	Side    string `json:"side,omitempty"`
}

// --- Trading Types ---

// SignedOrder represents a signed order ready for submission.
type SignedOrder struct {
	Salt          string `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          string `json:"side"`
	SignatureType string `json:"signatureType"`
	Signature     string `json:"signature"`
}

// OrderPayload is the request body for posting an order.
type OrderPayload struct {
	Order     SignedOrder `json:"order"`
	Owner     string      `json:"owner,omitempty"`
	OrderType string      `json:"orderType"`
}

// OrderResponse is the response from posting an order.
type OrderResponse struct {
	Success bool   `json:"success"`
	OrderID string `json:"orderID"`
	Status  string `json:"status,omitempty"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

// Order represents an order from the CLOB API.
type Order struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	Owner           string  `json:"owner"`
	Market          string  `json:"market"`
	AssetID         string  `json:"asset_id"`
	Side            string  `json:"side"`
	OriginalSize    string  `json:"original_size"`
	SizeMatched     string  `json:"size_matched"`
	Price           string  `json:"price"`
	Outcome         string  `json:"outcome"`
	OrderType       string  `json:"type"`
	CreatedAt       string  `json:"created_at"`
	ExpiresAt       string  `json:"expiration,omitempty"`
	Associate       string  `json:"associate_trades,omitempty"`
}

// Trade represents a trade from the CLOB API.
type Trade struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	Market      string  `json:"market"`
	AssetID     string  `json:"asset_id"`
	Side        string  `json:"side"`
	Size        string  `json:"size"`
	Price       string  `json:"price"`
	Outcome     string  `json:"outcome"`
	Fee         string  `json:"fee"`
	TradeOwner  string  `json:"owner"`
	Maker       string  `json:"maker_address"`
	MatchTime   string  `json:"match_time"`
	CreatedAt   string  `json:"created_at"`
	BucketIndex int     `json:"bucket_index"`
	Type        string  `json:"type"`
}

// BalanceAllowance holds balance and allowance info.
type BalanceAllowance struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

// CancelResponse is the response from a cancel operation.
type CancelResponse struct {
	Canceled []string `json:"canceled"`
	NotCanceled []CancelFailure `json:"not_canceled,omitempty"`
}

// CancelFailure represents a single cancel failure.
type CancelFailure struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// OpenOrdersParams holds optional params for listing open orders.
type OpenOrdersParams struct {
	Market string
	Asset  string
}

// TradesParams holds optional params for listing trades.
type TradesParams struct {
	Market string
	Asset  string
}

// BalanceParams holds optional params for balance queries.
type BalanceParams struct {
	AssetType string
	TokenID   string
}
