package data

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDataClientGetPositions(t *testing.T) {
	positions := []Position{
		{Asset: "token1", ConditionID: "cond1", Size: 100, CurPrice: 0.65, Outcome: "Yes"},
		{Asset: "token2", ConditionID: "cond2", Size: 50, CurPrice: 0.30, Outcome: "No"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("address") != "0xABC" {
			t.Errorf("unexpected address param: %s", r.URL.Query().Get("address"))
		}
		_ = json.NewEncoder(w).Encode(positions)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetPositions("0xABC")
	if err != nil {
		t.Fatalf("GetPositions() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetPositions() returned %d positions, want 2", len(got))
	}
	if got[0].Asset != "token1" {
		t.Errorf("GetPositions()[0].Asset = %q, want %q", got[0].Asset, "token1")
	}
	if got[1].Size != 50 {
		t.Errorf("GetPositions()[1].Size = %v, want 50", got[1].Size)
	}
}

func TestDataClientGetClosedPositions(t *testing.T) {
	positions := []Position{
		{Asset: "token3", ConditionID: "cond3", Size: 0, RealizedPnl: 25.5, Outcome: "Yes"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions/closed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("address") != "0xDEF" {
			t.Errorf("unexpected address param: %s", r.URL.Query().Get("address"))
		}
		_ = json.NewEncoder(w).Encode(positions)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetClosedPositions("0xDEF")
	if err != nil {
		t.Fatalf("GetClosedPositions() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("GetClosedPositions() returned %d positions, want 1", len(got))
	}
	if got[0].RealizedPnl != 25.5 {
		t.Errorf("GetClosedPositions()[0].RealizedPnl = %v, want 25.5", got[0].RealizedPnl)
	}
}

func TestDataClientGetPortfolioValue(t *testing.T) {
	values := []PortfolioValue{
		{Timestamp: "2024-01-01T00:00:00Z", Value: 1000.0},
		{Timestamp: "2024-01-02T00:00:00Z", Value: 1050.0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/value" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(values)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetPortfolioValue("0xABC")
	if err != nil {
		t.Fatalf("GetPortfolioValue() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetPortfolioValue() returned %d values, want 2", len(got))
	}
	if got[1].Value != 1050.0 {
		t.Errorf("GetPortfolioValue()[1].Value = %v, want 1050", got[1].Value)
	}
}

func TestDataClientGetPortfolioTraded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/traded" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"totalTraded": 50000}`))
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetPortfolioTraded("0xABC")
	if err != nil {
		t.Fatalf("GetPortfolioTraded() error: %v", err)
	}
	if string(got) != `{"totalTraded": 50000}` {
		t.Errorf("GetPortfolioTraded() = %s, want %s", string(got), `{"totalTraded": 50000}`)
	}
}

func TestDataClientGetActivity(t *testing.T) {
	activities := []Activity{
		{ID: "act1", Type: "trade", Side: "BUY", Size: 10, Price: 0.55, Outcome: "Yes"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activity" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(activities)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetActivity("0xABC", 50, 0)
	if err != nil {
		t.Fatalf("GetActivity() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("GetActivity() returned %d activities, want 1", len(got))
	}
	if got[0].Side != "BUY" {
		t.Errorf("GetActivity()[0].Side = %q, want %q", got[0].Side, "BUY")
	}
}

func TestDataClientGetTrades(t *testing.T) {
	trades := []TradeRecord{
		{ID: "tr1", Side: "SELL", Size: 20, Price: 0.70, Outcome: "Yes"},
		{ID: "tr2", Side: "BUY", Size: 15, Price: 0.45, Outcome: "No"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "25" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "10" {
			t.Errorf("unexpected offset: %s", r.URL.Query().Get("offset"))
		}
		_ = json.NewEncoder(w).Encode(trades)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetTrades("0xABC", 25, 10)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetTrades() returned %d trades, want 2", len(got))
	}
	if got[0].ID != "tr1" {
		t.Errorf("GetTrades()[0].ID = %q, want %q", got[0].ID, "tr1")
	}
}

func TestDataClientGetHolders(t *testing.T) {
	holders := []Holder{
		{Address: "0x111", Position: 5000, Value: 3250, Rank: 1},
		{Address: "0x222", Position: 2000, Value: 1300, Rank: 2},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/holders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("conditionId") != "cond1" {
			t.Errorf("unexpected conditionId: %s", r.URL.Query().Get("conditionId"))
		}
		_ = json.NewEncoder(w).Encode(holders)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetHolders("cond1")
	if err != nil {
		t.Fatalf("GetHolders() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetHolders() returned %d holders, want 2", len(got))
	}
	if got[0].Rank != 1 {
		t.Errorf("GetHolders()[0].Rank = %d, want 1", got[0].Rank)
	}
}

func TestDataClientGetOpenInterest(t *testing.T) {
	oi := OpenInterest{Market: "mkt1", ConditionID: "cond1", OpenInterest: 150000}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-interest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(oi)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetOpenInterest("cond1")
	if err != nil {
		t.Fatalf("GetOpenInterest() error: %v", err)
	}
	if got.OpenInterest != 150000 {
		t.Errorf("GetOpenInterest().OpenInterest = %v, want 150000", got.OpenInterest)
	}
}

func TestDataClientGetEventVolume(t *testing.T) {
	vol := EventVolume{EventID: "ev1", EventSlug: "test-event", Volume: 500000, Volume24hr: 25000}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/volume" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("eventId") != "ev1" {
			t.Errorf("unexpected eventId: %s", r.URL.Query().Get("eventId"))
		}
		_ = json.NewEncoder(w).Encode(vol)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetEventVolume("ev1")
	if err != nil {
		t.Fatalf("GetEventVolume() error: %v", err)
	}
	if got.Volume != 500000 {
		t.Errorf("GetEventVolume().Volume = %v, want 500000", got.Volume)
	}
}

func TestDataClientGetLeaderboard(t *testing.T) {
	entries := []LeaderboardEntry{
		{Address: "0xAAA", Rank: 1, Volume: 1000000, ProfitLoss: 50000, MarketsTraded: 100},
		{Address: "0xBBB", Rank: 2, Volume: 800000, ProfitLoss: 30000, MarketsTraded: 80},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/leaderboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetLeaderboard(10, 0)
	if err != nil {
		t.Fatalf("GetLeaderboard() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetLeaderboard() returned %d entries, want 2", len(got))
	}
	if got[0].Volume != 1000000 {
		t.Errorf("GetLeaderboard()[0].Volume = %v, want 1000000", got[0].Volume)
	}
}

func TestDataClientGetBuilderLeaderboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/builder-leaderboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"builder":"app1","volume":100000}]`))
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetBuilderLeaderboard()
	if err != nil {
		t.Fatalf("GetBuilderLeaderboard() error: %v", err)
	}
	if string(got) != `[{"builder":"app1","volume":100000}]` {
		t.Errorf("GetBuilderLeaderboard() = %s", string(got))
	}
}

func TestDataClientGetBuilderVolume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/builder-volume" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"totalVolume":999999}`))
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	got, err := client.GetBuilderVolume()
	if err != nil {
		t.Fatalf("GetBuilderVolume() error: %v", err)
	}
	if string(got) != `{"totalVolume":999999}` {
		t.Errorf("GetBuilderVolume() = %s", string(got))
	}
}

func TestDataClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := NewDataClient(server.URL)
	_, err := client.GetPositions("0xABC")
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestDataClientDefaultBaseURL(t *testing.T) {
	client := NewDataClient("")
	if client.baseURL != DefaultDataBaseURL {
		t.Errorf("NewDataClient(\"\").baseURL = %q, want %q", client.baseURL, DefaultDataBaseURL)
	}
}
