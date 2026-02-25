package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestDeriveAPIKey(t *testing.T) {
	expectedCreds := &APICredentials{
		APIKey:     "test-api-key",
		Secret:     "dGVzdC1zZWNyZXQ=", // base64 of "test-secret"
		Passphrase: "test-passphrase",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/derive-api-key" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		// Verify auth headers are present
		if r.Header.Get("POLY_ADDRESS") == "" {
			t.Error("missing POLY_ADDRESS header")
		}
		if r.Header.Get("POLY_SIGNATURE") == "" {
			t.Error("missing POLY_SIGNATURE header")
		}
		if r.Header.Get("POLY_TIMESTAMP") == "" {
			t.Error("missing POLY_TIMESTAMP header")
		}
		if r.Header.Get("POLY_NONCE") == "" {
			t.Error("missing POLY_NONCE header")
		}

		// Signature should be hex with 0x prefix
		sig := r.Header.Get("POLY_SIGNATURE")
		if len(sig) < 4 || sig[:2] != "0x" {
			t.Errorf("signature should start with 0x, got %q", sig[:10])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedCreds)
	}))
	defer server.Close()

	key, err := crypto.HexToECDSA(testPrivateKey)
	if err != nil {
		t.Fatalf("HexToECDSA() error: %v", err)
	}

	creds, err := DeriveAPIKey(server.URL, key, 137, 0)
	if err != nil {
		t.Fatalf("DeriveAPIKey() error: %v", err)
	}

	if creds.APIKey != expectedCreds.APIKey {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, expectedCreds.APIKey)
	}
	if creds.Secret != expectedCreds.Secret {
		t.Errorf("Secret = %q, want %q", creds.Secret, expectedCreds.Secret)
	}
	if creds.Passphrase != expectedCreds.Passphrase {
		t.Errorf("Passphrase = %q, want %q", creds.Passphrase, expectedCreds.Passphrase)
	}
}

func TestDeriveAPIKey_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	key, err := crypto.HexToECDSA(testPrivateKey)
	if err != nil {
		t.Fatalf("HexToECDSA() error: %v", err)
	}

	_, err = DeriveAPIKey(server.URL, key, 137, 0)
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestCreateAPIKey(t *testing.T) {
	expectedCreds := &APICredentials{
		APIKey:     "new-api-key",
		Secret:     "bmV3LXNlY3JldA==",
		Passphrase: "new-passphrase",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/api-key" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		if r.Method != "POST" {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		// Verify auth headers
		if r.Header.Get("POLY_ADDRESS") == "" {
			t.Error("missing POLY_ADDRESS header")
		}
		if r.Header.Get("POLY_SIGNATURE") == "" {
			t.Error("missing POLY_SIGNATURE header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedCreds)
	}))
	defer server.Close()

	key, err := crypto.HexToECDSA(testPrivateKey)
	if err != nil {
		t.Fatalf("HexToECDSA() error: %v", err)
	}

	creds, err := CreateAPIKey(server.URL, key, 137)
	if err != nil {
		t.Fatalf("CreateAPIKey() error: %v", err)
	}

	if creds.APIKey != expectedCreds.APIKey {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, expectedCreds.APIKey)
	}
}
