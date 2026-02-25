package chain

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

var (
	USDCAddress           = common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
	ConditionalTokensAddr = common.HexToAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045")
	CTFExchangeAddr       = common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	NegRiskExchangeAddr   = common.HexToAddress("0xC5d563A36AE78145C45a50134d48A1215220f80a")
	NegRiskAdapterAddr    = common.HexToAddress("0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296")
)

// SpenderFromName maps a human-readable spender name to a contract address.
func SpenderFromName(name string) (common.Address, error) {
	switch name {
	case "exchange":
		return CTFExchangeAddr, nil
	case "neg-risk-exchange":
		return NegRiskExchangeAddr, nil
	default:
		return common.Address{}, fmt.Errorf("unknown spender %q (use \"exchange\" or \"neg-risk-exchange\")", name)
	}
}
