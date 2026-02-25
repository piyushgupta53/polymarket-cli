package clob

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// testKey returns a deterministic ECDSA key for testing.
func testKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		t.Fatalf("creating test key: %v", err)
	}
	return key
}

func TestBuildSignedOrder_BUY(t *testing.T) {
	key := testKey(t)
	payload, err := BuildSignedOrder(OrderParams{
		TokenID:   "1234567890",
		Side:      "BUY",
		Price:     0.50,
		Size:      10,
		OrderType: "GTC",
	}, key, 137)
	if err != nil {
		t.Fatalf("BuildSignedOrder: %v", err)
	}

	if payload.OrderType != "GTC" {
		t.Errorf("OrderType = %q, want GTC", payload.OrderType)
	}
	if payload.Order.Side != "BUY" {
		t.Errorf("Side = %q, want BUY", payload.Order.Side)
	}
	if payload.Order.TokenID != "1234567890" {
		t.Errorf("TokenID = %q, want 1234567890", payload.Order.TokenID)
	}

	// BUY: makerAmount = price * size * 1e6 = 0.50 * 10 * 1e6 = 5000000
	if payload.Order.MakerAmount != "5000000" {
		t.Errorf("MakerAmount = %q, want 5000000", payload.Order.MakerAmount)
	}
	// BUY: takerAmount = size * 1e6 = 10 * 1e6 = 10000000
	if payload.Order.TakerAmount != "10000000" {
		t.Errorf("TakerAmount = %q, want 10000000", payload.Order.TakerAmount)
	}

	if !strings.HasPrefix(payload.Order.Signature, "0x") {
		t.Errorf("Signature should start with 0x, got %q", payload.Order.Signature[:10])
	}
	// 65-byte signature → 130 hex chars + "0x" prefix
	if len(payload.Order.Signature) != 132 {
		t.Errorf("Signature length = %d, want 132", len(payload.Order.Signature))
	}

	if payload.Order.Taker != ZeroAddress.Hex() {
		t.Errorf("Taker = %q, want zero address", payload.Order.Taker)
	}
	if payload.Order.Nonce != "0" {
		t.Errorf("Nonce = %q, want 0", payload.Order.Nonce)
	}
	if payload.Order.FeeRateBps != "0" {
		t.Errorf("FeeRateBps = %q, want 0", payload.Order.FeeRateBps)
	}
	if payload.Order.SignatureType != "0" {
		t.Errorf("SignatureType = %q, want 0", payload.Order.SignatureType)
	}

	// Owner and Maker should be the derived address.
	addr := crypto.PubkeyToAddress(key.PublicKey).Hex()
	if payload.Owner != addr {
		t.Errorf("Owner = %q, want %q", payload.Owner, addr)
	}
	if payload.Order.Maker != addr {
		t.Errorf("Maker = %q, want %q", payload.Order.Maker, addr)
	}
	if payload.Order.Signer != addr {
		t.Errorf("Signer = %q, want %q", payload.Order.Signer, addr)
	}
}

func TestBuildSignedOrder_SELL(t *testing.T) {
	key := testKey(t)
	payload, err := BuildSignedOrder(OrderParams{
		TokenID:   "9876543210",
		Side:      "SELL",
		Price:     0.75,
		Size:      20,
		OrderType: "FOK",
	}, key, 137)
	if err != nil {
		t.Fatalf("BuildSignedOrder: %v", err)
	}

	if payload.Order.Side != "SELL" {
		t.Errorf("Side = %q, want SELL", payload.Order.Side)
	}
	if payload.OrderType != "FOK" {
		t.Errorf("OrderType = %q, want FOK", payload.OrderType)
	}

	// SELL: makerAmount = size * 1e6 = 20 * 1e6 = 20000000
	if payload.Order.MakerAmount != "20000000" {
		t.Errorf("MakerAmount = %q, want 20000000", payload.Order.MakerAmount)
	}
	// SELL: takerAmount = price * size * 1e6 = 0.75 * 20 * 1e6 = 15000000
	if payload.Order.TakerAmount != "15000000" {
		t.Errorf("TakerAmount = %q, want 15000000", payload.Order.TakerAmount)
	}
}

func TestBuildSignedOrder_NegRisk(t *testing.T) {
	key := testKey(t)
	// Build two orders: one with negRisk=false, one with negRisk=true.
	// They should produce different signatures because the exchange address differs.
	p1, err := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "BUY", Price: 0.50, Size: 10, NegRisk: false,
	}, key, 137)
	if err != nil {
		t.Fatalf("negRisk=false: %v", err)
	}
	p2, err := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "BUY", Price: 0.50, Size: 10, NegRisk: true,
	}, key, 137)
	if err != nil {
		t.Fatalf("negRisk=true: %v", err)
	}
	// Different salts make signatures differ anyway, but both should be valid.
	if !strings.HasPrefix(p1.Order.Signature, "0x") {
		t.Error("negRisk=false signature missing 0x prefix")
	}
	if !strings.HasPrefix(p2.Order.Signature, "0x") {
		t.Error("negRisk=true signature missing 0x prefix")
	}
}

func TestBuildSignedOrder_Defaults(t *testing.T) {
	key := testKey(t)
	// Empty OrderType → defaults to GTC; empty Expiration → defaults to "0".
	payload, err := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "BUY", Price: 0.10, Size: 5,
	}, key, 137)
	if err != nil {
		t.Fatalf("BuildSignedOrder: %v", err)
	}
	if payload.OrderType != "GTC" {
		t.Errorf("default OrderType = %q, want GTC", payload.OrderType)
	}
	if payload.Order.Expiration != "0" {
		t.Errorf("default Expiration = %q, want 0", payload.Order.Expiration)
	}
}

func TestBuildSignedOrder_InvalidSide(t *testing.T) {
	key := testKey(t)
	_, err := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "HOLD", Price: 0.50, Size: 10,
	}, key, 137)
	if err == nil {
		t.Fatal("expected error for invalid side")
	}
	if !strings.Contains(err.Error(), "invalid side") {
		t.Errorf("error = %q, want to contain 'invalid side'", err.Error())
	}
}

func TestBuildSignedOrder_PriceOutOfRange(t *testing.T) {
	key := testKey(t)
	for _, price := range []float64{0.0, 1.0, -0.5, 1.5} {
		_, err := BuildSignedOrder(OrderParams{
			TokenID: "111", Side: "BUY", Price: price, Size: 10,
		}, key, 137)
		if err == nil {
			t.Errorf("expected error for price=%.2f", price)
		}
	}
}

func TestBuildSignedOrder_InvalidSize(t *testing.T) {
	key := testKey(t)
	for _, size := range []float64{0, -10} {
		_, err := BuildSignedOrder(OrderParams{
			TokenID: "111", Side: "BUY", Price: 0.50, Size: size,
		}, key, 137)
		if err == nil {
			t.Errorf("expected error for size=%.2f", size)
		}
	}
}

func TestBuildSignedOrder_AmountPrecision(t *testing.T) {
	key := testKey(t)
	// Test that fractional prices/sizes round correctly.
	payload, err := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "BUY", Price: 0.33, Size: 7.5,
	}, key, 137)
	if err != nil {
		t.Fatalf("BuildSignedOrder: %v", err)
	}
	// makerAmount = round(0.33 * 7.5 * 1e6) = round(2475000) = 2475000
	if payload.Order.MakerAmount != "2475000" {
		t.Errorf("MakerAmount = %q, want 2475000", payload.Order.MakerAmount)
	}
	// takerAmount = round(7.5 * 1e6) = 7500000
	if payload.Order.TakerAmount != "7500000" {
		t.Errorf("TakerAmount = %q, want 7500000", payload.Order.TakerAmount)
	}
}

func TestBuildSignedOrder_CaseSide(t *testing.T) {
	key := testKey(t)
	// Lowercase side should work.
	payload, err := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "buy", Price: 0.50, Size: 10,
	}, key, 137)
	if err != nil {
		t.Fatalf("BuildSignedOrder with lowercase: %v", err)
	}
	if payload.Order.Side != "BUY" {
		t.Errorf("Side = %q, want BUY", payload.Order.Side)
	}
}

func TestBuildSignedOrder_FeeRateAndExpiration(t *testing.T) {
	key := testKey(t)
	payload, err := BuildSignedOrder(OrderParams{
		TokenID:    "111",
		Side:       "BUY",
		Price:      0.50,
		Size:       10,
		FeeRateBps: "100",
		Expiration: "1700000000",
	}, key, 137)
	if err != nil {
		t.Fatalf("BuildSignedOrder: %v", err)
	}
	if payload.Order.FeeRateBps != "100" {
		t.Errorf("FeeRateBps = %q, want 100", payload.Order.FeeRateBps)
	}
	if payload.Order.Expiration != "1700000000" {
		t.Errorf("Expiration = %q, want 1700000000", payload.Order.Expiration)
	}
}

func TestAmountComputationSymmetry(t *testing.T) {
	// For the same price and size, the USDC side should be the same
	// whether it's a BUY or SELL.
	price := 0.65
	size := 100.0
	usdcAmount := int64(math.Round(price * size * usdcScale))
	tokenAmount := int64(math.Round(size * usdcScale))

	key := testKey(t)
	buy, _ := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "BUY", Price: price, Size: size,
	}, key, 137)
	sell, _ := BuildSignedOrder(OrderParams{
		TokenID: "111", Side: "SELL", Price: price, Size: size,
	}, key, 137)

	// BUY: maker=USDC, taker=tokens
	if buy.Order.MakerAmount != fmt.Sprintf("%d", usdcAmount) {
		t.Errorf("BUY makerAmount = %s, want %d", buy.Order.MakerAmount, usdcAmount)
	}
	if buy.Order.TakerAmount != fmt.Sprintf("%d", tokenAmount) {
		t.Errorf("BUY takerAmount = %s, want %d", buy.Order.TakerAmount, tokenAmount)
	}
	// SELL: maker=tokens, taker=USDC (swapped)
	if sell.Order.MakerAmount != fmt.Sprintf("%d", tokenAmount) {
		t.Errorf("SELL makerAmount = %s, want %d", sell.Order.MakerAmount, tokenAmount)
	}
	if sell.Order.TakerAmount != fmt.Sprintf("%d", usdcAmount) {
		t.Errorf("SELL takerAmount = %s, want %d", sell.Order.TakerAmount, usdcAmount)
	}
}
