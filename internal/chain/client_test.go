package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// mockEthClient implements EthClient for testing.
type mockEthClient struct {
	chainID            *big.Int
	balance            *big.Int
	nonce              uint64
	gasPrice           *big.Int
	gasEstimate        uint64
	callResult         []byte
	callErr            error
	sendErr            error
	receipt            *types.Receipt
	receiptErr         error
	estimateErr        error
	lastSentTx         *types.Transaction
	lastCallMsg        ethereum.CallMsg
}

func (m *mockEthClient) ChainID(ctx context.Context) (*big.Int, error) {
	return m.chainID, nil
}

func (m *mockEthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return m.balance, nil
}

func (m *mockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return m.nonce, nil
}

func (m *mockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return m.gasPrice, nil
}

func (m *mockEthClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if m.estimateErr != nil {
		return 0, m.estimateErr
	}
	return m.gasEstimate, nil
}

func (m *mockEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	m.lastCallMsg = msg
	return m.callResult, m.callErr
}

func (m *mockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	m.lastSentTx = tx
	return m.sendErr
}

func (m *mockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if m.receiptErr != nil {
		return nil, m.receiptErr
	}
	return m.receipt, nil
}

func (m *mockEthClient) Close() {}

func newTestClient(mock *mockEthClient) *ChainClient {
	return NewChainClientFromEth(mock, big.NewInt(137))
}

func testKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	return key
}

func TestGetMATICBalance(t *testing.T) {
	mock := &mockEthClient{
		balance: big.NewInt(1e18),
	}
	client := newTestClient(mock)

	bal, err := client.GetMATICBalance(context.Background(), common.HexToAddress("0x1234"))
	if err != nil {
		t.Fatalf("GetMATICBalance error: %v", err)
	}
	if bal.Cmp(big.NewInt(1e18)) != 0 {
		t.Errorf("balance = %s, want 1000000000000000000", bal.String())
	}
}

func TestCallContract(t *testing.T) {
	expected := common.LeftPadBytes(big.NewInt(42).Bytes(), 32)
	mock := &mockEthClient{
		callResult: expected,
	}
	client := newTestClient(mock)

	result, err := client.CallContract(context.Background(), USDCAddress, []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("CallContract error: %v", err)
	}
	if len(result) != 32 {
		t.Errorf("result length = %d, want 32", len(result))
	}
}

func TestSendTx(t *testing.T) {
	key := testKey(t)
	receipt := &types.Receipt{
		TxHash:      common.HexToHash("0xabc"),
		BlockNumber: big.NewInt(12345),
		GasUsed:     21000,
	}
	mock := &mockEthClient{
		nonce:       5,
		gasPrice:    big.NewInt(30e9),
		gasEstimate: 50000,
		receipt:     receipt,
	}
	client := newTestClient(mock)

	result, err := client.SendTx(context.Background(), key, USDCAddress, []byte{0x01}, nil)
	if err != nil {
		t.Fatalf("SendTx error: %v", err)
	}
	if result.BlockNumber != 12345 {
		t.Errorf("block = %d, want 12345", result.BlockNumber)
	}
	if result.GasUsed != 21000 {
		t.Errorf("gas = %d, want 21000", result.GasUsed)
	}
	if mock.lastSentTx == nil {
		t.Fatal("no transaction was sent")
	}
}

func TestSendTxInsufficientFunds(t *testing.T) {
	key := testKey(t)
	mock := &mockEthClient{
		nonce:       0,
		gasPrice:    big.NewInt(30e9),
		estimateErr: fmt.Errorf("insufficient funds for gas * price + value"),
	}
	client := newTestClient(mock)

	_, err := client.SendTx(context.Background(), key, USDCAddress, []byte{0x01}, nil)
	if err == nil {
		t.Fatal("expected error for insufficient funds")
	}
	if !contains(err.Error(), "insufficient MATIC") {
		t.Errorf("error = %q, should mention insufficient MATIC", err.Error())
	}
}

func testReceipt() *types.Receipt {
	return &types.Receipt{
		TxHash:      common.HexToHash("0xabc123"),
		BlockNumber: big.NewInt(100),
		GasUsed:     21000,
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
