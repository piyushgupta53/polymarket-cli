package clob

import (
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildHMACSignature(t *testing.T) {
	// Known test vector: compute HMAC-SHA256 manually
	secret := base64.URLEncoding.EncodeToString([]byte("test-secret"))
	timestamp := "1234567890"
	method := "GET"
	path := "/orders"
	body := ""

	sig, err := BuildHMACSignature(secret, timestamp, method, path, body)
	if err != nil {
		t.Fatalf("BuildHMACSignature() error: %v", err)
	}

	// Verify manually
	message := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(message))
	expected := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	if sig != expected {
		t.Errorf("sig = %q, want %q", sig, expected)
	}
}

func TestBuildHMACSignature_WithBody(t *testing.T) {
	secret := base64.URLEncoding.EncodeToString([]byte("my-secret"))
	sig1, _ := BuildHMACSignature(secret, "123", "POST", "/order", `{"id":"abc"}`)
	sig2, _ := BuildHMACSignature(secret, "123", "POST", "/order", `{"id":"xyz"}`)

	if sig1 == sig2 {
		t.Error("different bodies should produce different signatures")
	}
}

func TestBuildHMACSignature_InvalidSecret(t *testing.T) {
	// Invalid base64 should error
	_, err := BuildHMACSignature("not-valid-base64!!!", "123", "GET", "/", "")
	if err == nil {
		t.Error("expected error for invalid base64 secret")
	}
}

func TestApplyAuthHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	auth := &AuthCredentials{
		Key:        "api-key-123",
		Secret:     "secret-456",
		Passphrase: "pass-789",
		Address:    "0xdeadbeef",
	}

	ApplyAuthHeaders(req, auth, "1234567890", "sig-abc")

	tests := []struct {
		header, want string
	}{
		{"POLY_ADDRESS", "0xdeadbeef"},
		{"POLY_API_KEY", "api-key-123"},
		{"POLY_SIGNATURE", "sig-abc"},
		{"POLY_TIMESTAMP", "1234567890"},
		{"POLY_PASSPHRASE", "pass-789"},
	}

	for _, tt := range tests {
		got := req.Header.Get(tt.header)
		if got != tt.want {
			t.Errorf("header %s = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestDoAuthRequest_NilAuth(t *testing.T) {
	c := NewClient("")
	_, err := c.doAuthRequest("GET", "/test", nil)
	if err == nil {
		t.Error("expected error for nil auth")
	}
}

func TestDoAuthRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth headers are present
		if r.Header.Get("POLY_ADDRESS") == "" {
			t.Error("missing POLY_ADDRESS")
		}
		if r.Header.Get("POLY_API_KEY") == "" {
			t.Error("missing POLY_API_KEY")
		}
		if r.Header.Get("POLY_SIGNATURE") == "" {
			t.Error("missing POLY_SIGNATURE")
		}
		if r.Header.Get("POLY_TIMESTAMP") == "" {
			t.Error("missing POLY_TIMESTAMP")
		}
		if r.Header.Get("POLY_PASSPHRASE") == "" {
			t.Error("missing POLY_PASSPHRASE")
		}

		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	// Use a valid base64 secret
	secret := base64.URLEncoding.EncodeToString([]byte("test-secret"))
	auth := &AuthCredentials{
		Key:        "test-key",
		Secret:     secret,
		Passphrase: "test-pass",
		Address:    "0x1234",
	}

	c := NewAuthenticatedClient(server.URL, auth)
	data, err := c.doAuthGet("/test")
	if err != nil {
		t.Fatalf("doAuthGet() error: %v", err)
	}

	if string(data) != `{"ok":true}` {
		t.Errorf("response = %q, want {\"ok\":true}", string(data))
	}
}

func TestDoAuthRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	secret := base64.URLEncoding.EncodeToString([]byte("test-secret"))
	auth := &AuthCredentials{
		Key:        "test-key",
		Secret:     secret,
		Passphrase: "test-pass",
		Address:    "0x1234",
	}

	c := NewAuthenticatedClient(server.URL, auth)
	_, err := c.doAuthGet("/fail")
	if err == nil {
		t.Error("expected error for 403 response")
	}
}

func TestNewAuthenticatedClient(t *testing.T) {
	auth := &AuthCredentials{Key: "k", Secret: "s", Passphrase: "p", Address: "a"}
	c := NewAuthenticatedClient("", auth)

	if c.baseURL != DefaultClobBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultClobBaseURL)
	}
	if !c.IsAuthenticated() {
		t.Error("expected IsAuthenticated() to be true")
	}
}

func TestNewClient_NotAuthenticated(t *testing.T) {
	c := NewClient("")
	if c.IsAuthenticated() {
		t.Error("expected IsAuthenticated() to be false")
	}
}
