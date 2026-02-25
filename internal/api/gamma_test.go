package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGammaClientGetMarket(t *testing.T) {
	market := Market{
		ID:       "123",
		Question: "Will BTC hit 100k?",
		Slug:     "btc-100k",
		Active:   true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(market)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetMarket("123")
	if err != nil {
		t.Fatalf("GetMarket() error: %v", err)
	}
	if got.ID != "123" {
		t.Errorf("GetMarket() ID = %q, want %q", got.ID, "123")
	}
	if got.Question != "Will BTC hit 100k?" {
		t.Errorf("GetMarket() Question = %q, want %q", got.Question, "Will BTC hit 100k?")
	}
}

func TestGammaClientGetMarketBySlug(t *testing.T) {
	markets := []Market{
		{ID: "456", Question: "Test Market", Slug: "test-market"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("slug") != "test-market" {
			t.Errorf("unexpected slug param: %s", r.URL.Query().Get("slug"))
		}
		_ = json.NewEncoder(w).Encode(markets)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetMarketBySlug("test-market")
	if err != nil {
		t.Fatalf("GetMarketBySlug() error: %v", err)
	}
	if got.ID != "456" {
		t.Errorf("GetMarketBySlug() ID = %q, want %q", got.ID, "456")
	}
}

func TestGammaClientListMarkets(t *testing.T) {
	markets := []Market{
		{ID: "1", Question: "Market 1"},
		{ID: "2", Question: "Market 2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(markets)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.ListMarkets(MarketListParams{Limit: 10})
	if err != nil {
		t.Fatalf("ListMarkets() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListMarkets() returned %d markets, want 2", len(got))
	}
}

func TestGammaClientGetEvent(t *testing.T) {
	event := Event{
		ID:    "789",
		Title: "2024 Election",
		Slug:  "2024-election",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/789" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetEvent("789")
	if err != nil {
		t.Fatalf("GetEvent() error: %v", err)
	}
	if got.Title != "2024 Election" {
		t.Errorf("GetEvent() Title = %q, want %q", got.Title, "2024 Election")
	}
}

func TestGammaClientListTags(t *testing.T) {
	tags := []Tag{
		{ID: "1", Label: "Politics", Slug: "politics"},
		{ID: "2", Label: "Crypto", Slug: "crypto"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(tags)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.ListTags()
	if err != nil {
		t.Fatalf("ListTags() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListTags() returned %d tags, want 2", len(got))
	}
}

func TestGammaClientPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]Market{})
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	if err := client.Ping(); err != nil {
		t.Errorf("Ping() error: %v", err)
	}
}

func TestGammaClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	_, err := client.GetMarket("123")
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestGammaClientSlugNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]Market{})
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	_, err := client.GetMarketBySlug("nonexistent")
	if err == nil {
		t.Error("expected error for empty market list, got nil")
	}
}

func TestGammaClientListSeries(t *testing.T) {
	series := []Series{
		{ID: "s1", Title: "US Elections", Slug: "us-elections"},
		{ID: "s2", Title: "Crypto Prices", Slug: "crypto-prices"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/series" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.ListSeries(10, 0)
	if err != nil {
		t.Fatalf("ListSeries() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListSeries() returned %d series, want 2", len(got))
	}
	if got[0].Title != "US Elections" {
		t.Errorf("ListSeries()[0].Title = %q, want %q", got[0].Title, "US Elections")
	}
}

func TestGammaClientGetSeries(t *testing.T) {
	series := Series{
		ID:    "s1",
		Title: "US Elections",
		Slug:  "us-elections",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/series/s1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetSeries("s1")
	if err != nil {
		t.Fatalf("GetSeries() error: %v", err)
	}
	if got.Title != "US Elections" {
		t.Errorf("GetSeries() Title = %q, want %q", got.Title, "US Elections")
	}
}

func TestGammaClientListComments(t *testing.T) {
	comments := []Comment{
		{ID: "c1", Author: "0xabc", Body: "Great market!", EntityType: "event", EntityID: "e1", CreatedAt: "2024-01-01T00:00:00Z"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("entity_type") != "event" {
			t.Errorf("unexpected entity_type: %s", r.URL.Query().Get("entity_type"))
		}
		if r.URL.Query().Get("entity_id") != "e1" {
			t.Errorf("unexpected entity_id: %s", r.URL.Query().Get("entity_id"))
		}
		_ = json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.ListComments("event", "e1")
	if err != nil {
		t.Fatalf("ListComments() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("ListComments() returned %d comments, want 1", len(got))
	}
	if got[0].Body != "Great market!" {
		t.Errorf("ListComments()[0].Body = %q, want %q", got[0].Body, "Great market!")
	}
}

func TestGammaClientGetComment(t *testing.T) {
	comment := Comment{
		ID:     "c1",
		Author: "0xabc",
		Body:   "Great market!",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comments/c1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(comment)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetComment("c1")
	if err != nil {
		t.Fatalf("GetComment() error: %v", err)
	}
	if got.Author != "0xabc" {
		t.Errorf("GetComment() Author = %q, want %q", got.Author, "0xabc")
	}
}

func TestGammaClientGetUserComments(t *testing.T) {
	comments := []Comment{
		{ID: "c1", Author: "0xabc", Body: "My comment"},
		{ID: "c2", Author: "0xabc", Body: "Another comment"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comments/user/0xabc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetUserComments("0xabc")
	if err != nil {
		t.Fatalf("GetUserComments() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetUserComments() returned %d comments, want 2", len(got))
	}
}

func TestGammaClientGetProfile(t *testing.T) {
	profile := Profile{
		Address:  "0xabc",
		Username: "trader1",
		Volume:   12345.67,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profiles/0xabc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(profile)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetProfile("0xabc")
	if err != nil {
		t.Fatalf("GetProfile() error: %v", err)
	}
	if got.Username != "trader1" {
		t.Errorf("GetProfile() Username = %q, want %q", got.Username, "trader1")
	}
	if got.Volume != 12345.67 {
		t.Errorf("GetProfile() Volume = %f, want %f", got.Volume, 12345.67)
	}
}

func TestGammaClientListSports(t *testing.T) {
	sports := []Sport{
		{ID: "sp1", Label: "Football", Slug: "football"},
		{ID: "sp2", Label: "Basketball", Slug: "basketball"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sports" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(sports)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.ListSports()
	if err != nil {
		t.Fatalf("ListSports() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListSports() returned %d sports, want 2", len(got))
	}
	if got[0].Label != "Football" {
		t.Errorf("ListSports()[0].Label = %q, want %q", got[0].Label, "Football")
	}
}

func TestGammaClientGetMarketTypes(t *testing.T) {
	types := []SportMarketType{
		{ID: "mt1", Label: "Moneyline", Sport: "football"},
		{ID: "mt2", Label: "Spread", Sport: "basketball"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sports/market-types" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(types)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.GetMarketTypes()
	if err != nil {
		t.Fatalf("GetMarketTypes() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetMarketTypes() returned %d types, want 2", len(got))
	}
	if got[0].Label != "Moneyline" {
		t.Errorf("GetMarketTypes()[0].Label = %q, want %q", got[0].Label, "Moneyline")
	}
}

func TestGammaClientListTeams(t *testing.T) {
	teams := []Team{
		{ID: "t1", Name: "Lakers", League: "NBA", Slug: "lakers"},
		{ID: "t2", Name: "Patriots", League: "NFL", Slug: "patriots"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/teams" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("league") != "NBA" {
			t.Errorf("unexpected league: %s", r.URL.Query().Get("league"))
		}
		if r.URL.Query().Get("limit") != "25" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(teams)
	}))
	defer server.Close()

	client := NewGammaClient(server.URL)
	got, err := client.ListTeams("NBA", 25)
	if err != nil {
		t.Fatalf("ListTeams() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListTeams() returned %d teams, want 2", len(got))
	}
	if got[0].Name != "Lakers" {
		t.Errorf("ListTeams()[0].Name = %q, want %q", got[0].Name, "Lakers")
	}
}
