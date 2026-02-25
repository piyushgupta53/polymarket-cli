package auth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Hardhat test key #0 — well-known, never use in production.
const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestPrivateKeyToAddress(t *testing.T) {
	addr, err := PrivateKeyToAddress(testPrivateKey)
	if err != nil {
		t.Fatalf("PrivateKeyToAddress() error: %v", err)
	}
	expected := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	if addr != expected {
		t.Errorf("address = %s, want %s", addr.Hex(), expected.Hex())
	}
}

func TestPrivateKeyToAddress_WithPrefix(t *testing.T) {
	addr, err := PrivateKeyToAddress("0x" + testPrivateKey)
	if err != nil {
		t.Fatalf("PrivateKeyToAddress() error: %v", err)
	}
	expected := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	if addr != expected {
		t.Errorf("address = %s, want %s", addr.Hex(), expected.Hex())
	}
}

func TestPrivateKeyToAddress_Invalid(t *testing.T) {
	_, err := PrivateKeyToAddress("not-a-key")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestBuildClobAuthTypedData(t *testing.T) {
	td := BuildClobAuthTypedData("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "1234567890", "0", 137)

	if td.PrimaryType != "ClobAuth" {
		t.Errorf("PrimaryType = %q, want ClobAuth", td.PrimaryType)
	}
	if td.Domain.Name != "ClobAuthDomain" {
		t.Errorf("Domain.Name = %q, want ClobAuthDomain", td.Domain.Name)
	}
	if td.Message["message"] != "This message attests that I control the given wallet" {
		t.Error("unexpected message field")
	}
}

func TestSignTypedData_RoundTrip(t *testing.T) {
	key, err := crypto.HexToECDSA(testPrivateKey)
	if err != nil {
		t.Fatalf("HexToECDSA() error: %v", err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey)
	td := BuildClobAuthTypedData(address.Hex(), "1234567890", "0", 137)

	sig, err := SignTypedData(td, key)
	if err != nil {
		t.Fatalf("SignTypedData() error: %v", err)
	}

	if len(sig) != 65 {
		t.Errorf("signature length = %d, want 65", len(sig))
	}

	// v should be 27 or 28
	v := sig[64]
	if v != 27 && v != 28 {
		t.Errorf("v = %d, want 27 or 28", v)
	}

	// Verify we can recover the signer address
	domainSeparator, _ := td.HashStruct("EIP712Domain", td.Domain.Map())
	messageHash, _ := td.HashStruct(td.PrimaryType, td.Message)
	rawData := append([]byte("\x19\x01"), append(domainSeparator, messageHash...)...)
	hash := crypto.Keccak256Hash(rawData)

	sigForRecover := make([]byte, 65)
	copy(sigForRecover, sig)
	sigForRecover[64] -= 27 // Convert back to 0/1 for recovery

	pubKey, err := crypto.SigToPub(hash.Bytes(), sigForRecover)
	if err != nil {
		t.Fatalf("SigToPub() error: %v", err)
	}

	recovered := crypto.PubkeyToAddress(*pubKey)
	if recovered != address {
		t.Errorf("recovered = %s, want %s", recovered.Hex(), address.Hex())
	}
}

func TestBuildOrderTypedData(t *testing.T) {
	order := OrderData{
		Salt:          "123",
		Maker:         "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Signer:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       "12345",
		MakerAmount:   "100000000",
		TakerAmount:   "50000000",
		Expiration:    "0",
		Nonce:         "0",
		FeeRateBps:    "0",
		Side:          0,
		SignatureType: 0,
	}

	td := BuildOrderTypedData(order, 137, false)
	if td.PrimaryType != "Order" {
		t.Errorf("PrimaryType = %q, want Order", td.PrimaryType)
	}
	if td.Domain.Name != "Polymarket CTF Exchange" {
		t.Errorf("Domain.Name = %q, want 'Polymarket CTF Exchange'", td.Domain.Name)
	}
	if td.Domain.VerifyingContract != CTFExchangeAddress {
		t.Errorf("VerifyingContract = %q, want %q", td.Domain.VerifyingContract, CTFExchangeAddress)
	}

	// Test neg risk uses different contract
	tdNeg := BuildOrderTypedData(order, 137, true)
	if tdNeg.Domain.VerifyingContract != NegRiskCTFExchangeAddress {
		t.Errorf("neg risk VerifyingContract = %q, want %q", tdNeg.Domain.VerifyingContract, NegRiskCTFExchangeAddress)
	}
}

func TestSignOrder(t *testing.T) {
	key, err := crypto.HexToECDSA(testPrivateKey)
	if err != nil {
		t.Fatalf("HexToECDSA() error: %v", err)
	}

	order := OrderData{
		Salt:          "123",
		Maker:         "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Signer:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       "12345",
		MakerAmount:   "100000000",
		TakerAmount:   "50000000",
		Expiration:    "0",
		Nonce:         "0",
		FeeRateBps:    "0",
		Side:          0,
		SignatureType: 0,
	}

	sig, err := SignOrder(order, key, 137, false)
	if err != nil {
		t.Fatalf("SignOrder() error: %v", err)
	}

	if len(sig) < 4 || sig[:2] != "0x" {
		t.Errorf("signature should start with 0x, got %q", sig[:10])
	}
}

func TestGenerateOrderSalt(t *testing.T) {
	salt1, err := GenerateOrderSalt()
	if err != nil {
		t.Fatalf("GenerateOrderSalt: %v", err)
	}
	salt2, err := GenerateOrderSalt()
	if err != nil {
		t.Fatalf("GenerateOrderSalt: %v", err)
	}

	if salt1.Cmp(salt2) == 0 {
		t.Error("expected different salts")
	}
	if salt1.Sign() <= 0 {
		t.Error("salt should be positive")
	}
}

func TestStripHexPrefix(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"0xdeadbeef", "deadbeef"},
		{"deadbeef", "deadbeef"},
		{"0x", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripHexPrefix(tt.input)
		if got != tt.want {
			t.Errorf("stripHexPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
