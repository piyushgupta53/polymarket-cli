package data

import "encoding/json"

// Position represents an open position from the Data API.
type Position struct {
	Asset         string  `json:"asset"`
	ConditionID   string  `json:"conditionId"`
	CurPrice      float64 `json:"curPrice"`
	InitialValue  float64 `json:"initialValue"`
	CurrentValue  float64 `json:"currentValue"`
	CashPnl       float64 `json:"cashPnl"`
	PercentPnl    float64 `json:"percentPnl"`
	TotalBought   float64 `json:"totalBought"`
	TotalSold     float64 `json:"totalSold"`
	RealizedPnl   float64 `json:"realizedPnl"`
	Size          float64 `json:"size"`
	AvgPrice      float64 `json:"avgPrice"`
	EventSlug     string  `json:"eventSlug"`
	EventTitle    string  `json:"eventTitle"`
	Outcome       string  `json:"outcome"`
	OutcomeIndex  int     `json:"outcomeIndex"`
	ProxyWallet   string  `json:"proxyWallet,omitempty"`
}

// PortfolioValue represents portfolio value data from the Data API.
type PortfolioValue struct {
	Timestamp string  `json:"t"`
	Value     float64 `json:"v"`
}

// Activity represents an account activity record from the Data API.
type Activity struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Asset         string  `json:"asset"`
	ConditionID   string  `json:"conditionId"`
	EventSlug     string  `json:"eventSlug"`
	EventTitle    string  `json:"eventTitle"`
	Outcome       string  `json:"outcome"`
	Size          float64 `json:"size"`
	Price         float64 `json:"price"`
	UsdcSize      float64 `json:"usdcSize"`
	Fee           float64 `json:"fee"`
	Side          string  `json:"side"`
	Timestamp     string  `json:"timestamp"`
	TransactionID string  `json:"transactionId,omitempty"`
}

// TradeRecord represents a single trade from the Data API.
type TradeRecord struct {
	ID            string  `json:"id"`
	Taker         string  `json:"taker"`
	Maker         string  `json:"maker"`
	Market        string  `json:"market"`
	Asset         string  `json:"asset"`
	Side          string  `json:"side"`
	Size          float64 `json:"size"`
	Price         float64 `json:"price"`
	UsdcSize      float64 `json:"usdcSize"`
	Fee           float64 `json:"fee"`
	Outcome       string  `json:"outcome"`
	EventSlug     string  `json:"eventSlug"`
	Timestamp     string  `json:"timestamp"`
	TransactionID string  `json:"transactionId,omitempty"`
}

// Holder represents a top holder for a market token.
type Holder struct {
	Address  string  `json:"address"`
	Position float64 `json:"position"`
	Value    float64 `json:"value"`
	Rank     int     `json:"rank"`
}

// OpenInterest represents open interest data for a market.
type OpenInterest struct {
	Market        string  `json:"market"`
	ConditionID   string  `json:"conditionId"`
	OpenInterest  float64 `json:"openInterest"`
}

// EventVolume represents volume data for an event.
type EventVolume struct {
	EventID    string  `json:"eventId"`
	EventSlug  string  `json:"eventSlug"`
	EventTitle string  `json:"eventTitle"`
	Volume     float64 `json:"volume"`
	Volume24hr float64 `json:"volume24hr"`
}

// LeaderboardEntry represents a leaderboard entry from the Data API.
type LeaderboardEntry struct {
	Address       string  `json:"address"`
	Rank          int     `json:"rank"`
	Volume        float64 `json:"volume"`
	ProfitLoss    float64 `json:"profitLoss"`
	MarketsTraded int     `json:"marketsTraded"`
	DisplayName   string  `json:"displayName,omitempty"`
}

// BuilderLeaderboardEntry represents a builder leaderboard entry.
type BuilderLeaderboardEntry = json.RawMessage

// BuilderVolume represents builder volume data.
type BuilderVolume = json.RawMessage
