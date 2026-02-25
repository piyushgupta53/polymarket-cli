package clob

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestAuthClient(handler http.Handler) *Client {
	server := httptest.NewServer(handler)
	secret := base64.URLEncoding.EncodeToString([]byte("test-secret"))
	auth := &AuthCredentials{
		Key:        "test-key",
		Secret:     secret,
		Passphrase: "test-pass",
		Address:    "0x1234",
	}
	return NewAuthenticatedClient(server.URL, auth)
}

func TestPostOrder(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/order" {
			t.Errorf("path = %s, want /order", r.URL.Path)
		}
		json.NewEncoder(w).Encode(OrderResponse{
			Success: true,
			OrderID: "order-123",
		})
	}))

	order := &OrderPayload{
		OrderType: "GTC",
		Order: SignedOrder{
			Salt:      "123",
			Maker:     "0x1234",
			Signer:    "0x1234",
			TokenID:   "token-abc",
			Signature: "0xsig",
		},
	}

	resp, err := c.PostOrder(order)
	if err != nil {
		t.Fatalf("PostOrder() error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.OrderID != "order-123" {
		t.Errorf("OrderID = %q, want order-123", resp.OrderID)
	}
}

func TestCancelOrder(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		json.NewEncoder(w).Encode(CancelResponse{
			Canceled: []string{"order-123"},
		})
	}))

	resp, err := c.CancelOrder("order-123")
	if err != nil {
		t.Fatalf("CancelOrder() error: %v", err)
	}
	if len(resp.Canceled) != 1 || resp.Canceled[0] != "order-123" {
		t.Errorf("Canceled = %v, want [order-123]", resp.Canceled)
	}
}

func TestCancelAll(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cancel-all" {
			t.Errorf("path = %s, want /cancel-all", r.URL.Path)
		}
		json.NewEncoder(w).Encode(CancelResponse{
			Canceled: []string{"a", "b", "c"},
		})
	}))

	resp, err := c.CancelAll()
	if err != nil {
		t.Fatalf("CancelAll() error: %v", err)
	}
	if len(resp.Canceled) != 3 {
		t.Errorf("Canceled count = %d, want 3", len(resp.Canceled))
	}
}

func TestGetOrder(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/abc-123" {
			t.Errorf("path = %s, want /order/abc-123", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Order{
			ID:     "abc-123",
			Status: "LIVE",
			Side:   "BUY",
			Price:  "0.50",
		})
	}))

	order, err := c.GetOrder("abc-123")
	if err != nil {
		t.Fatalf("GetOrder() error: %v", err)
	}
	if order.ID != "abc-123" {
		t.Errorf("ID = %q, want abc-123", order.ID)
	}
	if order.Status != "LIVE" {
		t.Errorf("Status = %q, want LIVE", order.Status)
	}
}

func TestGetOpenOrders(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Errorf("path = %s, want /orders", r.URL.Path)
		}
		// Check query params
		if r.URL.Query().Get("market") != "cond-123" {
			t.Errorf("market = %q, want cond-123", r.URL.Query().Get("market"))
		}
		json.NewEncoder(w).Encode([]Order{
			{ID: "o1", Side: "BUY"},
			{ID: "o2", Side: "SELL"},
		})
	}))

	orders, err := c.GetOpenOrders(&OpenOrdersParams{Market: "cond-123"})
	if err != nil {
		t.Fatalf("GetOpenOrders() error: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("len(orders) = %d, want 2", len(orders))
	}
}

func TestGetTrades(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Errorf("path = %s, want /trades", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Trade{
			{ID: "t1", Side: "BUY", Price: "0.55"},
		})
	}))

	trades, err := c.GetTrades(nil)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}
	if len(trades) != 1 {
		t.Errorf("len(trades) = %d, want 1", len(trades))
	}
	if trades[0].Price != "0.55" {
		t.Errorf("Price = %q, want 0.55", trades[0].Price)
	}
}

func TestGetBalanceAllowance(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/balance-allowance" {
			t.Errorf("path = %s, want /balance-allowance", r.URL.Path)
		}
		json.NewEncoder(w).Encode(BalanceAllowance{
			Balance:   "1000.00",
			Allowance: "5000.00",
		})
	}))

	bal, err := c.GetBalanceAllowance(nil)
	if err != nil {
		t.Fatalf("GetBalanceAllowance() error: %v", err)
	}
	if bal.Balance != "1000.00" {
		t.Errorf("Balance = %q, want 1000.00", bal.Balance)
	}
	if bal.Allowance != "5000.00" {
		t.Errorf("Allowance = %q, want 5000.00", bal.Allowance)
	}
}

func TestUpdateBalanceAllowance(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))

	err := c.UpdateBalanceAllowance()
	if err != nil {
		t.Fatalf("UpdateBalanceAllowance() error: %v", err)
	}
}

func TestCancelOrders(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/orders" {
			t.Errorf("path = %s, want /orders", r.URL.Path)
		}
		json.NewEncoder(w).Encode(CancelResponse{
			Canceled: []string{"o1", "o2"},
		})
	}))

	resp, err := c.CancelOrders([]string{"o1", "o2"})
	if err != nil {
		t.Fatalf("CancelOrders() error: %v", err)
	}
	if len(resp.Canceled) != 2 {
		t.Errorf("Canceled count = %d, want 2", len(resp.Canceled))
	}
}

func TestCancelMarketOrders(t *testing.T) {
	c := newTestAuthClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CancelResponse{
			Canceled: []string{"o1"},
		})
	}))

	resp, err := c.CancelMarketOrders("cond-123")
	if err != nil {
		t.Fatalf("CancelMarketOrders() error: %v", err)
	}
	if len(resp.Canceled) != 1 {
		t.Errorf("Canceled count = %d, want 1", len(resp.Canceled))
	}
}
