package auth

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"


	"github.com/ethereum/go-ethereum/crypto"
)

// APICredentials holds derived CLOB API credentials.
type APICredentials struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// DeriveAPIKey derives CLOB API credentials using L1 auth headers.
// sigType: 0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE
func DeriveAPIKey(baseURL string, privateKey *ecdsa.PrivateKey, chainID, sigType int) (*APICredentials, error) {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "0"

	typedData := BuildClobAuthTypedData(address.Hex(), timestamp, nonce, chainID)
	sig, err := SignTypedData(typedData, privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing auth message: %w", err)
	}

	sigHex := fmt.Sprintf("0x%x", sig)

	url := strings.TrimSuffix(baseURL, "/") + "/auth/derive-api-key"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("POLY_ADDRESS", address.Hex())
	req.Header.Set("POLY_SIGNATURE", sigHex)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_NONCE", nonce)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("derive-api-key request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("derive-api-key failed (status %d): %s", resp.StatusCode, string(body))
	}

	var creds APICredentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	return &creds, nil
}

// CreateAPIKey creates a new CLOB API key using L1 auth headers.
func CreateAPIKey(baseURL string, privateKey *ecdsa.PrivateKey, chainID int) (*APICredentials, error) {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "0"

	typedData := BuildClobAuthTypedData(address.Hex(), timestamp, nonce, chainID)
	sig, err := SignTypedData(typedData, privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing auth message: %w", err)
	}

	sigHex := fmt.Sprintf("0x%x", sig)

	url := strings.TrimSuffix(baseURL, "/") + "/auth/api-key"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("POLY_ADDRESS", address.Hex())
	req.Header.Set("POLY_SIGNATURE", sigHex)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_NONCE", nonce)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create-api-key request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create-api-key failed (status %d): %s", resp.StatusCode, string(body))
	}

	var creds APICredentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	return &creds, nil
}
