package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// USDCToRaw converts a USDC amount (e.g., 10.5) to raw units (10500000).
func USDCToRaw(amount float64) *big.Int {
	micro := int64(math.Round(amount * 1e6))
	return big.NewInt(micro)
}

// RawToUSDC converts raw USDC units (e.g., 10500000) to a human-readable string ("10.500000").
func RawToUSDC(raw *big.Int) string {
	if raw == nil {
		return "0.000000"
	}
	// Handle negative values by converting the absolute value and prepending "-"
	if raw.Sign() < 0 {
		return "-" + RawToUSDC(new(big.Int).Abs(raw))
	}
	divisor := big.NewInt(1e6)
	whole := new(big.Int).Div(raw, divisor)
	remainder := new(big.Int).Mod(raw, divisor)
	return fmt.Sprintf("%s.%06d", whole.String(), remainder.Int64())
}

// GetUSDCBalance returns the USDC balance for an address.
func (c *ChainClient) GetUSDCBalance(ctx context.Context, owner common.Address) (*big.Int, error) {
	data, err := PackERC20BalanceOf(owner)
	if err != nil {
		return nil, fmt.Errorf("packing balanceOf: %w", err)
	}
	result, err := c.CallContract(ctx, USDCAddress, data)
	if err != nil {
		return nil, fmt.Errorf("calling USDC balanceOf: %w", err)
	}
	return UnpackUint256(result)
}

// GetUSDCAllowance returns the USDC allowance for owner → spender.
func (c *ChainClient) GetUSDCAllowance(ctx context.Context, owner, spender common.Address) (*big.Int, error) {
	data, err := PackERC20Allowance(owner, spender)
	if err != nil {
		return nil, fmt.Errorf("packing allowance: %w", err)
	}
	result, err := c.CallContract(ctx, USDCAddress, data)
	if err != nil {
		return nil, fmt.Errorf("calling USDC allowance: %w", err)
	}
	return UnpackUint256(result)
}

// ApproveUSDC sends an ERC-20 approve transaction for USDC.
// If amount is nil, approves MaxUint256 (unlimited).
func (c *ChainClient) ApproveUSDC(ctx context.Context, privateKey *ecdsa.PrivateKey, spender common.Address, amount *big.Int) (*TxResult, error) {
	if amount == nil {
		amount = MaxUint256()
	}
	data, err := PackERC20Approve(spender, amount)
	if err != nil {
		return nil, fmt.Errorf("packing approve: %w", err)
	}
	return c.SendTx(ctx, privateKey, USDCAddress, data, nil)
}

// GetCTFBalance returns the ERC-1155 balance for a conditional token.
func (c *ChainClient) GetCTFBalance(ctx context.Context, owner common.Address, tokenID *big.Int) (*big.Int, error) {
	data, err := PackERC1155BalanceOf(owner, tokenID)
	if err != nil {
		return nil, fmt.Errorf("packing ERC1155 balanceOf: %w", err)
	}
	result, err := c.CallContract(ctx, ConditionalTokensAddr, data)
	if err != nil {
		return nil, fmt.Errorf("calling CTF balanceOf: %w", err)
	}
	return UnpackUint256(result)
}

// IsApprovedForAll checks if an operator is approved for all tokens on the CTF contract.
func (c *ChainClient) IsApprovedForAll(ctx context.Context, owner, operator common.Address) (bool, error) {
	data, err := PackIsApprovedForAll(owner, operator)
	if err != nil {
		return false, fmt.Errorf("packing isApprovedForAll: %w", err)
	}
	result, err := c.CallContract(ctx, ConditionalTokensAddr, data)
	if err != nil {
		return false, fmt.Errorf("calling isApprovedForAll: %w", err)
	}
	return UnpackBool(result)
}

// SetApprovalForAll sends an ERC-1155 setApprovalForAll transaction on the CTF contract.
func (c *ChainClient) SetApprovalForAll(ctx context.Context, privateKey *ecdsa.PrivateKey, operator common.Address, approved bool) (*TxResult, error) {
	data, err := PackSetApprovalForAll(operator, approved)
	if err != nil {
		return nil, fmt.Errorf("packing setApprovalForAll: %w", err)
	}
	return c.SendTx(ctx, privateKey, ConditionalTokensAddr, data, nil)
}

// AddressFromKey derives the Ethereum address from a private key.
func AddressFromKey(key *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(key.PublicKey)
}
