package chain

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestUSDCToRaw(t *testing.T) {
	tests := []struct {
		input float64
		want  int64
	}{
		{0, 0},
		{1.0, 1_000_000},
		{10.5, 10_500_000},
		{0.000001, 1},
		{100.123456, 100_123_456},
	}

	for _, tt := range tests {
		got := USDCToRaw(tt.input)
		if got.Int64() != tt.want {
			t.Errorf("USDCToRaw(%f) = %d, want %d", tt.input, got.Int64(), tt.want)
		}
	}
}

func TestRawToUSDC(t *testing.T) {
	tests := []struct {
		input *big.Int
		want  string
	}{
		{big.NewInt(0), "0.000000"},
		{big.NewInt(1_000_000), "1.000000"},
		{big.NewInt(10_500_000), "10.500000"},
		{big.NewInt(1), "0.000001"},
		{big.NewInt(100_123_456), "100.123456"},
		{nil, "0.000000"},
	}

	for _, tt := range tests {
		got := RawToUSDC(tt.input)
		if got != tt.want {
			t.Errorf("RawToUSDC(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetUSDCBalance(t *testing.T) {
	expected := big.NewInt(5_000_000) // 5 USDC
	mock := &mockEthClient{
		callResult: common.LeftPadBytes(expected.Bytes(), 32),
	}
	client := newTestClient(mock)

	bal, err := client.GetUSDCBalance(context.Background(), common.HexToAddress("0x1234"))
	if err != nil {
		t.Fatalf("GetUSDCBalance error: %v", err)
	}
	if bal.Cmp(expected) != 0 {
		t.Errorf("balance = %s, want %s", bal.String(), expected.String())
	}
}

func TestGetUSDCAllowance(t *testing.T) {
	expected := MaxUint256
	mock := &mockEthClient{
		callResult: common.LeftPadBytes(expected.Bytes(), 32),
	}
	client := newTestClient(mock)

	allowance, err := client.GetUSDCAllowance(
		context.Background(),
		common.HexToAddress("0x1234"),
		CTFExchangeAddr,
	)
	if err != nil {
		t.Fatalf("GetUSDCAllowance error: %v", err)
	}
	if allowance.Cmp(expected) != 0 {
		t.Errorf("allowance = %s, want max uint256", allowance.String())
	}
}

func TestApproveUSDC(t *testing.T) {
	key := testKey(t)
	receipt := testReceipt()
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 50000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	result, err := client.ApproveUSDC(context.Background(), key, CTFExchangeAddr, nil)
	if err != nil {
		t.Fatalf("ApproveUSDC error: %v", err)
	}
	if result.BlockNumber != 100 {
		t.Errorf("block = %d, want 100", result.BlockNumber)
	}
}

func TestGetCTFBalance(t *testing.T) {
	expected := big.NewInt(1000)
	mock := &mockEthClient{
		callResult: common.LeftPadBytes(expected.Bytes(), 32),
	}
	client := newTestClient(mock)

	bal, err := client.GetCTFBalance(context.Background(), common.HexToAddress("0x1234"), big.NewInt(42))
	if err != nil {
		t.Fatalf("GetCTFBalance error: %v", err)
	}
	if bal.Cmp(expected) != 0 {
		t.Errorf("balance = %s, want %s", bal.String(), expected.String())
	}
}

func TestIsApprovedForAll(t *testing.T) {
	// true
	mock := &mockEthClient{
		callResult: common.LeftPadBytes([]byte{1}, 32),
	}
	client := newTestClient(mock)

	approved, err := client.IsApprovedForAll(context.Background(), common.HexToAddress("0x1234"), CTFExchangeAddr)
	if err != nil {
		t.Fatalf("IsApprovedForAll error: %v", err)
	}
	if !approved {
		t.Error("expected approved=true")
	}
}

func TestSetApprovalForAll(t *testing.T) {
	key := testKey(t)
	receipt := testReceipt()
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 50000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	result, err := client.SetApprovalForAll(context.Background(), key, CTFExchangeAddr, true)
	if err != nil {
		t.Fatalf("SetApprovalForAll error: %v", err)
	}
	if result.GasUsed != 21000 {
		t.Errorf("gas = %d, want 21000", result.GasUsed)
	}
}
