package bridge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBridgeClientGetDepositAddresses(t *testing.T) {
	addrs := DepositAddresses{
		EVMAddress:     "0x1234567890abcdef1234567890abcdef12345678",
		SolanaAddress:  "5FHwkrdxNTbparnKHACyXLqYt8q2JMKBgjJfQGqfXzMb",
		BitcoinAddress: "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deposit-addresses" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(addrs)
	}))
	defer server.Close()

	client := NewBridgeClient(server.URL)
	got, err := client.GetDepositAddresses()
	if err != nil {
		t.Fatalf("GetDepositAddresses() error: %v", err)
	}
	if got.EVMAddress != addrs.EVMAddress {
		t.Errorf("EVMAddress = %q, want %q", got.EVMAddress, addrs.EVMAddress)
	}
	if got.SolanaAddress != addrs.SolanaAddress {
		t.Errorf("SolanaAddress = %q, want %q", got.SolanaAddress, addrs.SolanaAddress)
	}
	if got.BitcoinAddress != addrs.BitcoinAddress {
		t.Errorf("BitcoinAddress = %q, want %q", got.BitcoinAddress, addrs.BitcoinAddress)
	}
}

func TestBridgeClientGetSupportedAssets(t *testing.T) {
	assets := []SupportedAsset{
		{Chain: "ethereum", Token: "USDC", Address: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"},
		{Chain: "solana", Token: "USDC", Address: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"},
		{Chain: "polygon", Token: "USDC", Address: "0x2791bca1f2de4661ed88a30c99a7a9449aa84174"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/supported-assets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(assets)
	}))
	defer server.Close()

	client := NewBridgeClient(server.URL)
	got, err := client.GetSupportedAssets()
	if err != nil {
		t.Fatalf("GetSupportedAssets() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("GetSupportedAssets() returned %d assets, want 3", len(got))
	}
	if got[0].Chain != "ethereum" {
		t.Errorf("asset[0].Chain = %q, want %q", got[0].Chain, "ethereum")
	}
	if got[1].Token != "USDC" {
		t.Errorf("asset[1].Token = %q, want %q", got[1].Token, "USDC")
	}
	if got[2].Address != "0x2791bca1f2de4661ed88a30c99a7a9449aa84174" {
		t.Errorf("asset[2].Address = %q, want %q", got[2].Address, "0x2791bca1f2de4661ed88a30c99a7a9449aa84174")
	}
}

func TestBridgeClientGetDepositStatus(t *testing.T) {
	status := DepositStatus{
		TxHash: "0xabc123",
		Status: "completed",
		Amount: "100.00",
		Chain:  "ethereum",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deposit-status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("txHash") != "0xabc123" {
			t.Errorf("unexpected txHash param: %s", r.URL.Query().Get("txHash"))
		}
		_ = json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	client := NewBridgeClient(server.URL)
	got, err := client.GetDepositStatus("0xabc123")
	if err != nil {
		t.Fatalf("GetDepositStatus() error: %v", err)
	}
	if got.TxHash != "0xabc123" {
		t.Errorf("TxHash = %q, want %q", got.TxHash, "0xabc123")
	}
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
	if got.Amount != "100.00" {
		t.Errorf("Amount = %q, want %q", got.Amount, "100.00")
	}
	if got.Chain != "ethereum" {
		t.Errorf("Chain = %q, want %q", got.Chain, "ethereum")
	}
}

func TestBridgeClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := NewBridgeClient(server.URL)
	_, err := client.GetDepositAddresses()
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestBridgeClientDefaultURL(t *testing.T) {
	client := NewBridgeClient("")
	if client.baseURL != DefaultBridgeBaseURL {
		t.Errorf("default baseURL = %q, want %q", client.baseURL, DefaultBridgeBaseURL)
	}
}

func TestBridgeClientCustomURL(t *testing.T) {
	client := NewBridgeClient("https://custom.api.com")
	if client.baseURL != "https://custom.api.com" {
		t.Errorf("custom baseURL = %q, want %q", client.baseURL, "https://custom.api.com")
	}
}

func TestBridgeClientGetDepositStatusEmptyOptionalFields(t *testing.T) {
	status := DepositStatus{
		TxHash: "0xdef456",
		Status: "pending",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	client := NewBridgeClient(server.URL)
	got, err := client.GetDepositStatus("0xdef456")
	if err != nil {
		t.Fatalf("GetDepositStatus() error: %v", err)
	}
	if got.Amount != "" {
		t.Errorf("Amount = %q, want empty", got.Amount)
	}
	if got.Chain != "" {
		t.Errorf("Chain = %q, want empty", got.Chain)
	}
}
