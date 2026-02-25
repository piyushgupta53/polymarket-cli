package chain

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestFunctionSelectors(t *testing.T) {
	tests := []struct {
		name     string
		packFn   func() ([]byte, error)
		selector string // first 4 bytes as hex
	}{
		{
			name: "ERC20 balanceOf",
			packFn: func() ([]byte, error) {
				return PackERC20BalanceOf(common.HexToAddress("0x0000000000000000000000000000000000000001"))
			},
			selector: "70a08231",
		},
		{
			name: "ERC20 approve",
			packFn: func() ([]byte, error) {
				return PackERC20Approve(common.HexToAddress("0x0000000000000000000000000000000000000001"), big.NewInt(100))
			},
			selector: "095ea7b3",
		},
		{
			name: "ERC20 allowance",
			packFn: func() ([]byte, error) {
				return PackERC20Allowance(
					common.HexToAddress("0x0000000000000000000000000000000000000001"),
					common.HexToAddress("0x0000000000000000000000000000000000000002"),
				)
			},
			selector: "dd62ed3e",
		},
		{
			name: "ERC1155 balanceOf",
			packFn: func() ([]byte, error) {
				return PackERC1155BalanceOf(common.HexToAddress("0x0000000000000000000000000000000000000001"), big.NewInt(1))
			},
			selector: "00fdd58e",
		},
		{
			name: "setApprovalForAll",
			packFn: func() ([]byte, error) {
				return PackSetApprovalForAll(common.HexToAddress("0x0000000000000000000000000000000000000001"), true)
			},
			selector: "a22cb465",
		},
		{
			name: "isApprovedForAll",
			packFn: func() ([]byte, error) {
				return PackIsApprovedForAll(
					common.HexToAddress("0x0000000000000000000000000000000000000001"),
					common.HexToAddress("0x0000000000000000000000000000000000000002"),
				)
			},
			selector: "e985e9c5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.packFn()
			if err != nil {
				t.Fatalf("pack error: %v", err)
			}
			got := hex.EncodeToString(data[:4])
			if got != tt.selector {
				t.Errorf("selector = %s, want %s", got, tt.selector)
			}
		})
	}
}

func TestUnpackUint256RoundTrip(t *testing.T) {
	val := big.NewInt(123456789)
	// ABI-encode a uint256
	packed := common.LeftPadBytes(val.Bytes(), 32)
	result, err := UnpackUint256(packed)
	if err != nil {
		t.Fatalf("UnpackUint256 error: %v", err)
	}
	if result.Cmp(val) != 0 {
		t.Errorf("got %s, want %s", result.String(), val.String())
	}
}

func TestUnpackBoolRoundTrip(t *testing.T) {
	// true = 0x...01
	packed := common.LeftPadBytes([]byte{1}, 32)
	result, err := UnpackBool(packed)
	if err != nil {
		t.Fatalf("UnpackBool error: %v", err)
	}
	if !result {
		t.Error("expected true, got false")
	}

	// false = 0x...00
	packed = make([]byte, 32)
	result, err = UnpackBool(packed)
	if err != nil {
		t.Fatalf("UnpackBool error: %v", err)
	}
	if result {
		t.Error("expected false, got true")
	}
}

func TestMaxUint256(t *testing.T) {
	// MaxUint256 should be 2^256 - 1
	expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	if MaxUint256.Cmp(expected) != 0 {
		t.Errorf("MaxUint256 = %s, want %s", MaxUint256.String(), expected.String())
	}
}

func TestCTFPackSelectors(t *testing.T) {
	zero := [32]byte{}
	partition := []*big.Int{big.NewInt(1), big.NewInt(2)}
	amount := big.NewInt(1000000)

	tests := []struct {
		name     string
		packFn   func() ([]byte, error)
		selector string
	}{
		{
			name: "splitPosition",
			packFn: func() ([]byte, error) {
				return PackSplitPosition(USDCAddress, zero, zero, partition, amount)
			},
			selector: "72ce4275",
		},
		{
			name: "mergePositions",
			packFn: func() ([]byte, error) {
				return PackMergePositions(USDCAddress, zero, zero, partition, amount)
			},
			selector: "9e7212ad",
		},
		{
			name: "redeemPositions",
			packFn: func() ([]byte, error) {
				return PackRedeemPositions(USDCAddress, zero, zero, partition)
			},
			selector: "01b7037c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.packFn()
			if err != nil {
				t.Fatalf("pack error: %v", err)
			}
			got := hex.EncodeToString(data[:4])
			if got != tt.selector {
				t.Errorf("selector = %s, want %s", got, tt.selector)
			}
		})
	}
}
