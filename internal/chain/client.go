package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient is the subset of ethclient.Client methods used by ChainClient.
type EthClient interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	Close()
}

// ChainClient wraps an Ethereum client for Polygon RPC interactions.
type ChainClient struct {
	eth     EthClient
	chainID *big.Int
}

// NewChainClient dials an RPC endpoint and verifies the chain ID is 137 (Polygon).
func NewChainClient(rpcURL string) (*ChainClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dialing RPC %s: %w", rpcURL, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("fetching chain ID: %w", err)
	}

	if chainID.Int64() != 137 {
		client.Close()
		return nil, fmt.Errorf("unexpected chain ID %s, expected 137 (Polygon)", chainID.String())
	}

	return &ChainClient{eth: client, chainID: chainID}, nil
}

// NewChainClientFromEth creates a ChainClient from an existing EthClient (for testing).
func NewChainClientFromEth(eth EthClient, chainID *big.Int) *ChainClient {
	return &ChainClient{eth: eth, chainID: chainID}
}

// Close closes the underlying RPC connection.
func (c *ChainClient) Close() {
	c.eth.Close()
}

// GetMATICBalance returns the native MATIC balance in wei.
func (c *ChainClient) GetMATICBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	return c.eth.BalanceAt(ctx, addr, nil)
}

// TxResult holds the result of a sent transaction.
type TxResult struct {
	TxHash      common.Hash
	BlockNumber uint64
	GasUsed     uint64
}

// SendTx builds, signs, sends a transaction and waits for the receipt.
func (c *ChainClient) SendTx(ctx context.Context, privateKey *ecdsa.PrivateKey, to common.Address, data []byte, value *big.Int) (*TxResult, error) {
	from := crypto.PubkeyToAddress(privateKey.PublicKey)

	nonce, err := c.eth.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("fetching nonce: %w", err)
	}

	gasPrice, err := c.eth.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggesting gas price: %w", err)
	}

	if value == nil {
		value = big.NewInt(0)
	}

	msg := ethereum.CallMsg{
		From:     from,
		To:       &to,
		GasPrice: gasPrice,
		Value:    value,
		Data:     data,
	}

	gasLimit, err := c.eth.EstimateGas(ctx, msg)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			return nil, fmt.Errorf("insufficient MATIC for gas: %w", err)
		}
		return nil, fmt.Errorf("estimating gas: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &to,
		Value:    value,
		Data:     data,
	})

	signer := types.NewEIP155Signer(c.chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing transaction: %w", err)
	}

	if err := c.eth.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("sending transaction: %w", err)
	}

	receipt, err := c.WaitForReceipt(ctx, signedTx.Hash())
	if err != nil {
		return nil, err
	}

	return &TxResult{
		TxHash:      receipt.TxHash,
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
	}, nil
}

// CallContract performs a read-only contract call.
func (c *ChainClient) CallContract(ctx context.Context, to common.Address, data []byte) ([]byte, error) {
	msg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
	return c.eth.CallContract(ctx, msg, nil)
}

// WaitForReceipt polls for a transaction receipt with a 2s interval and 120s timeout.
func (c *ChainClient) WaitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		receipt, err := c.eth.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("timeout waiting for receipt of tx %s", txHash.Hex())
			}
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
