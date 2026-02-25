package clob

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/piyushgupta/polymarket-cli/internal/auth"
)

const (
	// usdcScale is 10^6, matching USDC and conditional token decimals on Polygon.
	usdcScale = 1_000_000
)

// ZeroAddress is the Ethereum zero address (open order, any taker).
var ZeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")

// OrderParams holds user-friendly order parameters.
type OrderParams struct {
	TokenID    string
	Side       string  // "BUY" or "SELL"
	Price      float64 // 0.01 – 0.99
	Size       float64 // number of shares
	OrderType  string  // "GTC", "FOK", "GTD"
	Expiration string  // unix timestamp string, "0" for no expiry
	NegRisk    bool
	FeeRateBps string // basis points; default "0"
}

// BuildSignedOrder constructs, signs, and returns an order payload ready for POST /order.
func BuildSignedOrder(params OrderParams, privateKey *ecdsa.PrivateKey, chainID int) (*OrderPayload, error) {
	side := strings.ToUpper(params.Side)
	if side != "BUY" && side != "SELL" {
		return nil, fmt.Errorf("invalid side %q: must be BUY or SELL", params.Side)
	}
	if params.Price < 0.01 || params.Price > 0.99 {
		return nil, fmt.Errorf("price %.4f out of range [0.01, 0.99]", params.Price)
	}
	if params.Size <= 0 {
		return nil, fmt.Errorf("size must be greater than 0")
	}

	// Compute raw amounts (both USDC and conditional tokens use 6 decimals).
	//   BUY:  maker pays USDC, receives conditional tokens.
	//   SELL: maker gives conditional tokens, receives USDC.
	var makerAmount, takerAmount int64
	sideInt := 0 // BUY
	if side == "BUY" {
		makerAmount = int64(math.Round(params.Price * params.Size * usdcScale))
		takerAmount = int64(math.Round(params.Size * usdcScale))
	} else {
		sideInt = 1 // SELL
		makerAmount = int64(math.Round(params.Size * usdcScale))
		takerAmount = int64(math.Round(params.Price * params.Size * usdcScale))
	}
	if makerAmount <= 0 || takerAmount <= 0 {
		return nil, fmt.Errorf("computed amounts too small (maker=%d, taker=%d)", makerAmount, takerAmount)
	}

	salt, err := auth.GenerateOrderSalt()
	if err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	feeRateBps := params.FeeRateBps
	if feeRateBps == "" {
		feeRateBps = "0"
	}
	expiration := params.Expiration
	if expiration == "" {
		expiration = "0"
	}

	orderData := auth.OrderData{
		Salt:          salt.String(),
		Maker:         address.Hex(),
		Signer:        address.Hex(),
		Taker:         ZeroAddress.Hex(),
		TokenID:       params.TokenID,
		MakerAmount:   strconv.FormatInt(makerAmount, 10),
		TakerAmount:   strconv.FormatInt(takerAmount, 10),
		Expiration:    expiration,
		Nonce:         "0",
		FeeRateBps:    feeRateBps,
		Side:          sideInt,
		SignatureType: 0, // EOA
	}

	signature, err := auth.SignOrder(orderData, privateKey, chainID, params.NegRisk)
	if err != nil {
		return nil, fmt.Errorf("signing order: %w", err)
	}

	signedOrder := SignedOrder{
		Salt:          orderData.Salt,
		Maker:         orderData.Maker,
		Signer:        orderData.Signer,
		Taker:         orderData.Taker,
		TokenID:       orderData.TokenID,
		MakerAmount:   orderData.MakerAmount,
		TakerAmount:   orderData.TakerAmount,
		Expiration:    orderData.Expiration,
		Nonce:         orderData.Nonce,
		FeeRateBps:    orderData.FeeRateBps,
		Side:          side,
		SignatureType: fmt.Sprintf("%d", orderData.SignatureType),
		Signature:     signature,
	}

	orderType := strings.ToUpper(params.OrderType)
	if orderType == "" {
		orderType = "GTC"
	}

	return &OrderPayload{
		Order:     signedOrder,
		Owner:     address.Hex(),
		OrderType: orderType,
	}, nil
}
