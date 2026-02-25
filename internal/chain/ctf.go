package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// DefaultPartition is the standard [1, 2] partition for binary outcomes.
var DefaultPartition = []*big.Int{big.NewInt(1), big.NewInt(2)}

// DefaultIndexSets is the standard [1, 2] index sets for redeeming binary outcomes.
var DefaultIndexSets = []*big.Int{big.NewInt(1), big.NewInt(2)}

// ComputeConditionID computes the condition ID from oracle, questionID, and outcome count.
// conditionID = keccak256(abi.encodePacked(oracle, questionID, outcomeSlotCount))
func ComputeConditionID(oracle common.Address, questionID [32]byte, outcomeSlotCount uint) [32]byte {
	// abi.encodePacked(address, bytes32, uint256)
	data := make([]byte, 0, 20+32+32)
	data = append(data, oracle.Bytes()...)
	data = append(data, questionID[:]...)
	countBytes := common.LeftPadBytes(big.NewInt(int64(outcomeSlotCount)).Bytes(), 32)
	data = append(data, countBytes...)

	hash := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], hash)
	return result
}

// ComputeCollectionID computes a collection ID.
// collectionID = keccak256(abi.encodePacked(conditionID, indexSet)) XOR parentCollectionID
func ComputeCollectionID(parentCollectionID [32]byte, conditionID [32]byte, indexSet uint) [32]byte {
	// keccak256(conditionID, indexSet)
	indexBytes := common.LeftPadBytes(big.NewInt(int64(indexSet)).Bytes(), 32)
	data := make([]byte, 0, 64)
	data = append(data, conditionID[:]...)
	data = append(data, indexBytes...)
	hash := crypto.Keccak256(data)

	var result [32]byte
	copy(result[:], hash)

	// XOR with parentCollectionID
	for i := 0; i < 32; i++ {
		result[i] ^= parentCollectionID[i]
	}
	return result
}

// ComputePositionID computes a position ID from collateral token and collection ID.
// positionID = keccak256(abi.encodePacked(collateralToken, collectionID))
func ComputePositionID(collateralToken common.Address, collectionID [32]byte) [32]byte {
	data := make([]byte, 0, 20+32)
	data = append(data, collateralToken.Bytes()...)
	data = append(data, collectionID[:]...)

	hash := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], hash)
	return result
}

// SplitPosition splits USDC collateral into YES and NO conditional tokens.
func (c *ChainClient) SplitPosition(ctx context.Context, privateKey *ecdsa.PrivateKey, conditionID [32]byte, amount *big.Int) (*TxResult, error) {
	var zeroParent [32]byte
	data, err := PackSplitPosition(USDCAddress, zeroParent, conditionID, DefaultPartition, amount)
	if err != nil {
		return nil, fmt.Errorf("packing splitPosition: %w", err)
	}
	return c.SendTx(ctx, privateKey, ConditionalTokensAddr, data, nil)
}

// MergePositions merges YES and NO conditional tokens back into USDC collateral.
func (c *ChainClient) MergePositions(ctx context.Context, privateKey *ecdsa.PrivateKey, conditionID [32]byte, amount *big.Int) (*TxResult, error) {
	var zeroParent [32]byte
	data, err := PackMergePositions(USDCAddress, zeroParent, conditionID, DefaultPartition, amount)
	if err != nil {
		return nil, fmt.Errorf("packing mergePositions: %w", err)
	}
	return c.SendTx(ctx, privateKey, ConditionalTokensAddr, data, nil)
}

// RedeemPositions redeems winning conditional tokens for USDC collateral.
func (c *ChainClient) RedeemPositions(ctx context.Context, privateKey *ecdsa.PrivateKey, conditionID [32]byte) (*TxResult, error) {
	var zeroParent [32]byte
	data, err := PackRedeemPositions(USDCAddress, zeroParent, conditionID, DefaultIndexSets)
	if err != nil {
		return nil, fmt.Errorf("packing redeemPositions: %w", err)
	}
	return c.SendTx(ctx, privateKey, ConditionalTokensAddr, data, nil)
}

// RedeemNegRiskPositions redeems neg-risk conditional tokens via the NegRiskAdapter.
func (c *ChainClient) RedeemNegRiskPositions(ctx context.Context, privateKey *ecdsa.PrivateKey, conditionID [32]byte) (*TxResult, error) {
	var zeroParent [32]byte
	data, err := PackRedeemPositions(USDCAddress, zeroParent, conditionID, DefaultIndexSets)
	if err != nil {
		return nil, fmt.Errorf("packing redeemPositions: %w", err)
	}
	return c.SendTx(ctx, privateKey, NegRiskAdapterAddr, data, nil)
}
