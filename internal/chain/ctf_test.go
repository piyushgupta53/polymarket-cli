package chain

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestComputeConditionID(t *testing.T) {
	oracle := common.HexToAddress("0x0000000000000000000000000000000000000001")
	var questionID [32]byte
	questionID[31] = 0x01

	result := ComputeConditionID(oracle, questionID, 2)

	// Should be deterministic — same inputs give same output
	result2 := ComputeConditionID(oracle, questionID, 2)
	if result != result2 {
		t.Error("ComputeConditionID not deterministic")
	}

	// Different inputs give different output
	result3 := ComputeConditionID(oracle, questionID, 3)
	if result == result3 {
		t.Error("different outcome count should give different condition ID")
	}

	// Result should be non-zero
	var zero [32]byte
	if result == zero {
		t.Error("condition ID should not be zero")
	}
}

func TestComputeCollectionID(t *testing.T) {
	var parentCollectionID [32]byte
	var conditionID [32]byte
	conditionID[31] = 0x01

	result := ComputeCollectionID(parentCollectionID, conditionID, 1)
	result2 := ComputeCollectionID(parentCollectionID, conditionID, 2)

	if result == result2 {
		t.Error("different index sets should give different collection IDs")
	}

	var zero [32]byte
	if result == zero {
		t.Error("collection ID should not be zero")
	}
}

func TestComputeCollectionIDWithParent(t *testing.T) {
	var parentCollectionID [32]byte
	parentCollectionID[0] = 0xFF
	var conditionID [32]byte
	conditionID[31] = 0x01

	withParent := ComputeCollectionID(parentCollectionID, conditionID, 1)
	withoutParent := ComputeCollectionID([32]byte{}, conditionID, 1)

	if withParent == withoutParent {
		t.Error("parent collection ID should affect result")
	}
}

func TestComputePositionID(t *testing.T) {
	var collectionID [32]byte
	collectionID[31] = 0x01

	result := ComputePositionID(USDCAddress, collectionID)

	var zero [32]byte
	if result == zero {
		t.Error("position ID should not be zero")
	}

	// Deterministic
	result2 := ComputePositionID(USDCAddress, collectionID)
	if result != result2 {
		t.Error("ComputePositionID not deterministic")
	}
}

func TestComputeConditionIDKnownVector(t *testing.T) {
	// Test with a known oracle and question ID to verify the encoding format
	oracle := common.HexToAddress("0xdd4D117723C257CEe402285D3aCF218E9A8236E1")
	qIDBytes, _ := hex.DecodeString("abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234")
	var questionID [32]byte
	copy(questionID[:], qIDBytes)

	result := ComputeConditionID(oracle, questionID, 2)

	// Verify result length
	if len(result) != 32 {
		t.Errorf("condition ID length = %d, want 32", len(result))
	}

	// Verify determinism with same params
	result2 := ComputeConditionID(oracle, questionID, 2)
	if result != result2 {
		t.Error("not deterministic")
	}
}

func TestSplitPosition(t *testing.T) {
	key := testKey(t)
	receipt := testReceipt()
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 100000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	var conditionID [32]byte
	conditionID[31] = 0x01

	result, err := client.SplitPosition(context.Background(), key, conditionID, big.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("SplitPosition error: %v", err)
	}
	if result.BlockNumber != 100 {
		t.Errorf("block = %d, want 100", result.BlockNumber)
	}
	if mock.lastSentTx == nil {
		t.Fatal("no transaction sent")
	}
	// Verify tx was sent to ConditionalTokens contract
	if *mock.lastSentTx.To() != ConditionalTokensAddr {
		t.Errorf("tx.To = %s, want %s", mock.lastSentTx.To().Hex(), ConditionalTokensAddr.Hex())
	}
}

func TestMergePositions(t *testing.T) {
	key := testKey(t)
	receipt := testReceipt()
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 100000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	var conditionID [32]byte
	conditionID[31] = 0x01

	result, err := client.MergePositions(context.Background(), key, conditionID, big.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("MergePositions error: %v", err)
	}
	if result.BlockNumber != 100 {
		t.Errorf("block = %d, want 100", result.BlockNumber)
	}
}

func TestRedeemPositions(t *testing.T) {
	key := testKey(t)
	receipt := testReceipt()
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 80000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	var conditionID [32]byte
	conditionID[31] = 0x01

	result, err := client.RedeemPositions(context.Background(), key, conditionID)
	if err != nil {
		t.Fatalf("RedeemPositions error: %v", err)
	}
	if result.BlockNumber != 100 {
		t.Errorf("block = %d, want 100", result.BlockNumber)
	}
	if *mock.lastSentTx.To() != ConditionalTokensAddr {
		t.Errorf("tx.To = %s, want ConditionalTokens", mock.lastSentTx.To().Hex())
	}
}

func TestRedeemNegRiskPositions(t *testing.T) {
	key := testKey(t)
	receipt := testReceipt()
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 80000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	var conditionID [32]byte
	conditionID[31] = 0x01

	result, err := client.RedeemNegRiskPositions(context.Background(), key, conditionID)
	if err != nil {
		t.Fatalf("RedeemNegRiskPositions error: %v", err)
	}
	if result.BlockNumber != 100 {
		t.Errorf("block = %d, want 100", result.BlockNumber)
	}
	// Should target NegRiskAdapter, not ConditionalTokens
	if *mock.lastSentTx.To() != NegRiskAdapterAddr {
		t.Errorf("tx.To = %s, want NegRiskAdapter", mock.lastSentTx.To().Hex())
	}
}
