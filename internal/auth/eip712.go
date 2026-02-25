package auth

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// PrivateKeyToAddress parses a hex private key and returns the derived address.
func PrivateKeyToAddress(hexKey string) (common.Address, error) {
	key, err := crypto.HexToECDSA(stripHexPrefix(hexKey))
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid private key: %w", err)
	}
	return crypto.PubkeyToAddress(key.PublicKey), nil
}

// BuildClobAuthTypedData returns EIP-712 typed data for CLOB auth.
func BuildClobAuthTypedData(address, timestamp, nonce string, chainID int) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    "ClobAuthDomain",
			Version: "1",
			ChainId: math.NewHexOrDecimal256(int64(chainID)),
		},
		Message: apitypes.TypedDataMessage{
			"address":   address,
			"timestamp": timestamp,
			"nonce":     nonce,
			"message":   "This message attests that I control the given wallet",
		},
	}
}

// SignTypedData hashes and signs EIP-712 typed data. Returns raw signature bytes.
func SignTypedData(typedData apitypes.TypedData, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("hashing domain: %w", err)
	}

	messageHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("hashing message: %w", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(messageHash)))
	hash := crypto.Keccak256Hash(rawData)

	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing: %w", err)
	}

	// Adjust v value: go-ethereum returns 0/1, EIP-712 expects 27/28
	if sig[64] < 27 {
		sig[64] += 27
	}

	return sig, nil
}

// OrderData holds order fields needed for EIP-712 signing.
type OrderData struct {
	Salt          string
	Maker         string
	Signer        string
	Taker         string
	TokenID       string
	MakerAmount   string
	TakerAmount   string
	Expiration    string
	Nonce         string
	FeeRateBps    string
	Side          int // 0=BUY, 1=SELL
	SignatureType int // 0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE
}

// CTFExchangeAddress is the Polymarket CTF exchange contract on Polygon.
const CTFExchangeAddress = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"

// NegRiskCTFExchangeAddress is the neg-risk exchange contract on Polygon.
const NegRiskCTFExchangeAddress = "0xC5d563A36AE78145C45a50134d48A1215220f80a"

// BuildOrderTypedData returns EIP-712 typed data for signing an order.
func BuildOrderTypedData(order OrderData, chainID int, negRisk bool) apitypes.TypedData {
	exchange := CTFExchangeAddress
	if negRisk {
		exchange = NegRiskCTFExchangeAddress
	}

	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "expiration", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "feeRateBps", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(int64(chainID)),
			VerifyingContract: exchange,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          order.Salt,
			"maker":         order.Maker,
			"signer":        order.Signer,
			"taker":         order.Taker,
			"tokenId":       order.TokenID,
			"makerAmount":   order.MakerAmount,
			"takerAmount":   order.TakerAmount,
			"expiration":    order.Expiration,
			"nonce":         order.Nonce,
			"feeRateBps":    order.FeeRateBps,
			"side":          fmt.Sprintf("%d", order.Side),
			"signatureType": fmt.Sprintf("%d", order.SignatureType),
		},
	}
}

// SignOrder signs an order using EIP-712 and returns the hex-encoded signature.
func SignOrder(order OrderData, privateKey *ecdsa.PrivateKey, chainID int, negRisk bool) (string, error) {
	typedData := BuildOrderTypedData(order, chainID, negRisk)
	sig, err := SignTypedData(typedData, privateKey)
	if err != nil {
		return "", fmt.Errorf("signing order: %w", err)
	}
	return fmt.Sprintf("0x%x", sig), nil
}

// GenerateOrderSalt returns a random salt for order signing.
func GenerateOrderSalt() (*big.Int, error) {
	salt, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generating random key for salt: %w", err)
	}
	return new(big.Int).SetBytes(crypto.Keccak256(crypto.FromECDSA(salt))[:20]), nil
}

func stripHexPrefix(s string) string {
	if len(s) >= 2 && s[0:2] == "0x" {
		return s[2:]
	}
	return s
}
