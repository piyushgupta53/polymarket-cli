package clob

import (
	"encoding/json"
	"testing"
)

func TestOrderBookUnmarshal(t *testing.T) {
	raw := `{
		"market": "0xabc",
		"asset_id": "12345",
		"timestamp": "1234567890",
		"hash": "abc123",
		"bids": [{"price": "0.50", "size": "100"}, {"price": "0.49", "size": "200"}],
		"asks": [{"price": "0.51", "size": "150"}],
		"min_order_size": "5",
		"tick_size": "0.001",
		"neg_risk": true,
		"last_trade_price": "0.505"
	}`

	var book OrderBook
	if err := json.Unmarshal([]byte(raw), &book); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if book.Market != "0xabc" {
		t.Errorf("Market = %q, want %q", book.Market, "0xabc")
	}
	if len(book.Bids) != 2 {
		t.Errorf("Bids len = %d, want 2", len(book.Bids))
	}
	if len(book.Asks) != 1 {
		t.Errorf("Asks len = %d, want 1", len(book.Asks))
	}
	if book.Bids[0].Price != "0.50" {
		t.Errorf("Bids[0].Price = %q, want %q", book.Bids[0].Price, "0.50")
	}
	if !book.NegRisk {
		t.Error("NegRisk = false, want true")
	}
	if book.LastTradePrice != "0.505" {
		t.Errorf("LastTradePrice = %q, want %q", book.LastTradePrice, "0.505")
	}
}

func TestClobMarketUnmarshal(t *testing.T) {
	raw := `{
		"enable_order_book": true,
		"active": true,
		"closed": false,
		"archived": false,
		"accepting_orders": true,
		"accepting_order_timestamp": "2025-01-05T18:48:34Z",
		"minimum_order_size": 5,
		"minimum_tick_size": 0.001,
		"condition_id": "0xabc123",
		"question_id": "0xdef456",
		"question": "Will BTC hit 100k?",
		"description": "Test market",
		"market_slug": "btc-100k",
		"end_date_iso": "2025-12-31T00:00:00Z",
		"seconds_delay": 3,
		"fpmm": "0x12345",
		"maker_base_fee": 0,
		"taker_base_fee": 0,
		"notifications_enabled": true,
		"neg_risk": false,
		"neg_risk_market_id": "",
		"neg_risk_request_id": "",
		"icon": "",
		"image": "",
		"rewards": {"rates": null, "min_size": 0, "max_spread": 0},
		"is_50_50_outcome": false,
		"tokens": [
			{"token_id": "111", "outcome": "Yes", "price": 0.65, "winner": false},
			{"token_id": "222", "outcome": "No", "price": 0.35, "winner": false}
		],
		"tags": ["crypto"]
	}`

	var m ClobMarket
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if m.Question != "Will BTC hit 100k?" {
		t.Errorf("Question = %q, want %q", m.Question, "Will BTC hit 100k?")
	}
	if m.MinimumTickSize != 0.001 {
		t.Errorf("MinimumTickSize = %f, want 0.001", m.MinimumTickSize)
	}
	if len(m.Tokens) != 2 {
		t.Errorf("Tokens len = %d, want 2", len(m.Tokens))
	}
	if m.Tokens[0].Price != 0.65 {
		t.Errorf("Tokens[0].Price = %f, want 0.65", m.Tokens[0].Price)
	}
}

func TestPriceHistoryUnmarshal(t *testing.T) {
	raw := `{"history": [{"t": 1234567890, "p": 0.55}, {"t": 1234567900, "p": 0.56}]}`

	var resp PriceHistoryResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp.History) != 2 {
		t.Errorf("History len = %d, want 2", len(resp.History))
	}
	if resp.History[0].Timestamp != 1234567890 {
		t.Errorf("History[0].Timestamp = %d, want 1234567890", resp.History[0].Timestamp)
	}
	if resp.History[1].Price != 0.56 {
		t.Errorf("History[1].Price = %f, want 0.56", resp.History[1].Price)
	}
}

func TestBatchPriceEntryUnmarshal(t *testing.T) {
	raw := `{"token1": {"BUY": "0.65"}, "token2": {"BUY": "0.35", "SELL": "0.36"}}`

	var resp map[string]BatchPriceEntry
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp["token1"].Buy != "0.65" {
		t.Errorf("token1 Buy = %q, want %q", resp["token1"].Buy, "0.65")
	}
	if resp["token2"].Sell != "0.36" {
		t.Errorf("token2 Sell = %q, want %q", resp["token2"].Sell, "0.36")
	}
}

func TestClobMarketsResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": [{"question": "test", "condition_id": "0xabc", "tokens": [], "rewards": {"rates": null, "min_size": 0, "max_spread": 0}, "tags": []}],
		"next_cursor": "MTA=",
		"limit": 100,
		"count": 1
	}`

	var resp ClobMarketsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(resp.Data))
	}
	if resp.NextCursor != "MTA=" {
		t.Errorf("NextCursor = %q, want %q", resp.NextCursor, "MTA=")
	}
}
