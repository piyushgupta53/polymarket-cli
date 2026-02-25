package clob

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`"OK"`))
	}))
	defer server.Close()

	c := NewClient(server.URL)
	status, err := c.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck() error: %v", err)
	}
	if status != "OK" {
		t.Errorf("HealthCheck() = %q, want %q", status, "OK")
	}
}

func TestClientServerTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/time" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("1234567890"))
	}))
	defer server.Close()

	c := NewClient(server.URL)
	time, err := c.ServerTime()
	if err != nil {
		t.Fatalf("ServerTime() error: %v", err)
	}
	if time != "1234567890" {
		t.Errorf("ServerTime() = %q, want %q", time, "1234567890")
	}
}

func TestClientGetPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "token1" {
			t.Errorf("unexpected token_id: %s", r.URL.Query().Get("token_id"))
		}
		if r.URL.Query().Get("side") != "buy" {
			t.Errorf("unexpected side: %s", r.URL.Query().Get("side"))
		}
		json.NewEncoder(w).Encode(PriceResponse{Price: "0.65"})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetPrice("token1", "buy")
	if err != nil {
		t.Fatalf("GetPrice() error: %v", err)
	}
	if resp.Price != "0.65" {
		t.Errorf("GetPrice() = %q, want %q", resp.Price, "0.65")
	}
}

func TestClientGetMidpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(MidpointResponse{Mid: "0.50"})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetMidpoint("token1")
	if err != nil {
		t.Fatalf("GetMidpoint() error: %v", err)
	}
	if resp.Mid != "0.50" {
		t.Errorf("GetMidpoint() = %q, want %q", resp.Mid, "0.50")
	}
}

func TestClientGetSpread(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SpreadResponse{Spread: "0.02"})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetSpread("token1")
	if err != nil {
		t.Fatalf("GetSpread() error: %v", err)
	}
	if resp.Spread != "0.02" {
		t.Errorf("GetSpread() = %q, want %q", resp.Spread, "0.02")
	}
}

func TestClientGetBook(t *testing.T) {
	book := OrderBook{
		Market:  "0xabc",
		AssetID: "token1",
		Bids:    []OrderBookEntry{{Price: "0.50", Size: "100"}},
		Asks:    []OrderBookEntry{{Price: "0.51", Size: "150"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(book)
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetBook("token1")
	if err != nil {
		t.Fatalf("GetBook() error: %v", err)
	}
	if len(resp.Bids) != 1 {
		t.Errorf("Bids len = %d, want 1", len(resp.Bids))
	}
	if len(resp.Asks) != 1 {
		t.Errorf("Asks len = %d, want 1", len(resp.Asks))
	}
}

func TestClientGetLastTradePrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(LastTradeResponse{Price: "0.55", Side: "BUY"})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetLastTradePrice("token1")
	if err != nil {
		t.Fatalf("GetLastTradePrice() error: %v", err)
	}
	if resp.Price != "0.55" {
		t.Errorf("Price = %q, want %q", resp.Price, "0.55")
	}
	if resp.Side != "BUY" {
		t.Errorf("Side = %q, want %q", resp.Side, "BUY")
	}
}

func TestClientGetTickSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TickSizeResponse{MinimumTickSize: 0.001})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetTickSize("token1")
	if err != nil {
		t.Fatalf("GetTickSize() error: %v", err)
	}
	if resp.MinimumTickSize != 0.001 {
		t.Errorf("MinimumTickSize = %f, want 0.001", resp.MinimumTickSize)
	}
}

func TestClientGetFeeRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(FeeRateResponse{BaseFee: 0.01})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetFeeRate("token1")
	if err != nil {
		t.Fatalf("GetFeeRate() error: %v", err)
	}
	if resp.BaseFee != 0.01 {
		t.Errorf("BaseFee = %f, want 0.01", resp.BaseFee)
	}
}

func TestClientGetNegRisk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(NegRiskResponse{NegRisk: true})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetNegRisk("token1")
	if err != nil {
		t.Fatalf("GetNegRisk() error: %v", err)
	}
	if !resp.NegRisk {
		t.Error("NegRisk = false, want true")
	}
}

func TestClientGetMarket(t *testing.T) {
	market := ClobMarket{
		ConditionID: "0xabc",
		Question:    "Will BTC hit 100k?",
		Active:      true,
		Tokens:      []ClobToken{{TokenID: "111", Outcome: "Yes", Price: 0.65}},
		Rewards:     ClobRewards{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/0xabc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(market)
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetMarket("0xabc")
	if err != nil {
		t.Fatalf("GetMarket() error: %v", err)
	}
	if resp.Question != "Will BTC hit 100k?" {
		t.Errorf("Question = %q, want %q", resp.Question, "Will BTC hit 100k?")
	}
}

func TestClientListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ClobMarketsResponse{
			Data:       []ClobMarket{{Question: "Test"}},
			NextCursor: "abc",
			Count:      1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.ListMarkets("")
	if err != nil {
		t.Fatalf("ListMarkets() error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(resp.Data))
	}
	if resp.NextCursor != "abc" {
		t.Errorf("NextCursor = %q, want %q", resp.NextCursor, "abc")
	}
}

func TestClientGetPriceHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("market") == "" {
			t.Error("missing market param")
		}
		if r.URL.Query().Get("interval") != "1d" {
			t.Errorf("interval = %q, want %q", r.URL.Query().Get("interval"), "1d")
		}
		json.NewEncoder(w).Encode(PriceHistoryResponse{
			History: []PriceHistoryPoint{{Timestamp: 12345, Price: 0.55}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetPriceHistory("token1", "1d", 10)
	if err != nil {
		t.Fatalf("GetPriceHistory() error: %v", err)
	}
	if len(resp.History) != 1 {
		t.Errorf("History len = %d, want 1", len(resp.History))
	}
}

func TestClientBatchPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify body is JSON array
		body, _ := io.ReadAll(r.Body)
		var reqs []BatchTokenRequest
		if err := json.Unmarshal(body, &reqs); err != nil {
			t.Errorf("invalid request body: %v", err)
		}
		if len(reqs) != 2 {
			t.Errorf("request body len = %d, want 2", len(reqs))
		}

		resp := map[string]BatchPriceEntry{
			"token1": {Buy: "0.65"},
			"token2": {Buy: "0.35"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetBatchPrices([]string{"token1", "token2"}, "buy")
	if err != nil {
		t.Fatalf("GetBatchPrices() error: %v", err)
	}
	if resp["token1"].Buy != "0.65" {
		t.Errorf("token1 Buy = %q, want %q", resp["token1"].Buy, "0.65")
	}
}

func TestClientBatchMidpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"token1": "0.50", "token2": "0.60"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetBatchMidpoints([]string{"token1", "token2"})
	if err != nil {
		t.Fatalf("GetBatchMidpoints() error: %v", err)
	}
	if resp["token1"] != "0.50" {
		t.Errorf("token1 = %q, want %q", resp["token1"], "0.50")
	}
}

func TestClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL)

	_, err := c.GetPrice("token1", "buy")
	if err == nil {
		t.Error("expected error for 500 response")
	}

	_, err = c.GetBook("token1")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestClientSamplingMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ClobMarketsResponse{
			Data:  []ClobMarket{{Question: "Sampling market"}},
			Count: 1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL)
	resp, err := c.GetSamplingMarkets()
	if err != nil {
		t.Fatalf("GetSamplingMarkets() error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len = %d, want 1", len(resp))
	}
}
