package chain

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// MaxUint256 is the maximum value for a uint256.
var MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

var (
	erc20ABI  abi.ABI
	erc1155ABI abi.ABI
	ctfABI    abi.ABI
)

const erc20ABIJSON = `[
	{"name":"balanceOf","type":"function","stateMutability":"view","inputs":[{"name":"account","type":"address"}],"outputs":[{"name":"","type":"uint256"}]},
	{"name":"approve","type":"function","stateMutability":"nonpayable","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]},
	{"name":"allowance","type":"function","stateMutability":"view","inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"outputs":[{"name":"","type":"uint256"}]}
]`

const erc1155ABIJSON = `[
	{"name":"balanceOf","type":"function","stateMutability":"view","inputs":[{"name":"account","type":"address"},{"name":"id","type":"uint256"}],"outputs":[{"name":"","type":"uint256"}]},
	{"name":"setApprovalForAll","type":"function","stateMutability":"nonpayable","inputs":[{"name":"operator","type":"address"},{"name":"approved","type":"bool"}],"outputs":[]},
	{"name":"isApprovedForAll","type":"function","stateMutability":"view","inputs":[{"name":"account","type":"address"},{"name":"operator","type":"address"}],"outputs":[{"name":"","type":"bool"}]}
]`

const ctfABIJSON = `[
	{"name":"splitPosition","type":"function","stateMutability":"nonpayable","inputs":[{"name":"collateralToken","type":"address"},{"name":"parentCollectionId","type":"bytes32"},{"name":"conditionId","type":"bytes32"},{"name":"partition","type":"uint256[]"},{"name":"amount","type":"uint256"}],"outputs":[]},
	{"name":"mergePositions","type":"function","stateMutability":"nonpayable","inputs":[{"name":"collateralToken","type":"address"},{"name":"parentCollectionId","type":"bytes32"},{"name":"conditionId","type":"bytes32"},{"name":"partition","type":"uint256[]"},{"name":"amount","type":"uint256"}],"outputs":[]},
	{"name":"redeemPositions","type":"function","stateMutability":"nonpayable","inputs":[{"name":"collateralToken","type":"address"},{"name":"parentCollectionId","type":"bytes32"},{"name":"conditionId","type":"bytes32"},{"name":"indexSets","type":"uint256[]"}],"outputs":[]}
]`

func init() {
	var err error
	erc20ABI, err = abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		panic("failed to parse ERC-20 ABI: " + err.Error())
	}
	erc1155ABI, err = abi.JSON(strings.NewReader(erc1155ABIJSON))
	if err != nil {
		panic("failed to parse ERC-1155 ABI: " + err.Error())
	}
	ctfABI, err = abi.JSON(strings.NewReader(ctfABIJSON))
	if err != nil {
		panic("failed to parse CTF ABI: " + err.Error())
	}
}

// PackERC20BalanceOf encodes a balanceOf(address) call.
func PackERC20BalanceOf(account common.Address) ([]byte, error) {
	return erc20ABI.Pack("balanceOf", account)
}

// PackERC20Approve encodes an approve(address,uint256) call.
func PackERC20Approve(spender common.Address, amount *big.Int) ([]byte, error) {
	return erc20ABI.Pack("approve", spender, amount)
}

// PackERC20Allowance encodes an allowance(address,address) call.
func PackERC20Allowance(owner, spender common.Address) ([]byte, error) {
	return erc20ABI.Pack("allowance", owner, spender)
}

// PackERC1155BalanceOf encodes a balanceOf(address,uint256) call.
func PackERC1155BalanceOf(account common.Address, id *big.Int) ([]byte, error) {
	return erc1155ABI.Pack("balanceOf", account, id)
}

// PackSetApprovalForAll encodes a setApprovalForAll(address,bool) call.
func PackSetApprovalForAll(operator common.Address, approved bool) ([]byte, error) {
	return erc1155ABI.Pack("setApprovalForAll", operator, approved)
}

// PackIsApprovedForAll encodes an isApprovedForAll(address,address) call.
func PackIsApprovedForAll(account, operator common.Address) ([]byte, error) {
	return erc1155ABI.Pack("isApprovedForAll", account, operator)
}

// PackSplitPosition encodes a splitPosition call.
func PackSplitPosition(collateral common.Address, parentCollectionID [32]byte, conditionID [32]byte, partition []*big.Int, amount *big.Int) ([]byte, error) {
	return ctfABI.Pack("splitPosition", collateral, parentCollectionID, conditionID, partition, amount)
}

// PackMergePositions encodes a mergePositions call.
func PackMergePositions(collateral common.Address, parentCollectionID [32]byte, conditionID [32]byte, partition []*big.Int, amount *big.Int) ([]byte, error) {
	return ctfABI.Pack("mergePositions", collateral, parentCollectionID, conditionID, partition, amount)
}

// PackRedeemPositions encodes a redeemPositions call.
func PackRedeemPositions(collateral common.Address, parentCollectionID [32]byte, conditionID [32]byte, indexSets []*big.Int) ([]byte, error) {
	return ctfABI.Pack("redeemPositions", collateral, parentCollectionID, conditionID, indexSets)
}

// UnpackUint256 decodes a single uint256 from ABI-encoded output.
func UnpackUint256(data []byte) (*big.Int, error) {
	results, err := abi.Arguments{{Type: mustType("uint256")}}.Unpack(data)
	if err != nil {
		return nil, err
	}
	return results[0].(*big.Int), nil
}

// UnpackBool decodes a single bool from ABI-encoded output.
func UnpackBool(data []byte) (bool, error) {
	results, err := abi.Arguments{{Type: mustType("bool")}}.Unpack(data)
	if err != nil {
		return false, err
	}
	return results[0].(bool), nil
}

func mustType(t string) abi.Type {
	typ, err := abi.NewType(t, "", nil)
	if err != nil {
		panic("invalid ABI type: " + t)
	}
	return typ
}
