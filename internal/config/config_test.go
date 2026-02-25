package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Load from a non-existent path should return defaults
	cfg, err := LoadFrom("/tmp/nonexistent/polymarket/config.json")
	if err != nil {
		t.Fatalf("LoadFrom() error: %v", err)
	}

	if cfg.GammaAPIURL != DefaultGammaAPIURL {
		t.Errorf("GammaAPIURL = %q, want %q", cfg.GammaAPIURL, DefaultGammaAPIURL)
	}
	if cfg.CLOBAPIURL != DefaultCLOBAPIURL {
		t.Errorf("CLOBAPIURL = %q, want %q", cfg.CLOBAPIURL, DefaultCLOBAPIURL)
	}
	if cfg.ChainID != DefaultChainID {
		t.Errorf("ChainID = %d, want %d", cfg.ChainID, DefaultChainID)
	}
	if cfg.RPCURL != DefaultRPCURL {
		t.Errorf("RPCURL = %q, want %q", cfg.RPCURL, DefaultRPCURL)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		PrivateKey:    "0xdeadbeef",
		ChainID:       137,
		SignatureType: "EOA",
		GammaAPIURL:   DefaultGammaAPIURL,
		CLOBAPIURL:    DefaultCLOBAPIURL,
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load it back
	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error: %v", err)
	}

	if loaded.PrivateKey != "0xdeadbeef" {
		t.Errorf("PrivateKey = %q, want %q", loaded.PrivateKey, "0xdeadbeef")
	}
	if loaded.ChainID != 137 {
		t.Errorf("ChainID = %d, want %d", loaded.ChainID, 137)
	}
	if loaded.SignatureType != "EOA" {
		t.Errorf("SignatureType = %q, want %q", loaded.SignatureType, "EOA")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nested", "dir", "config.json")

	cfg := &Config{ChainID: DefaultChainID, GammaAPIURL: DefaultGammaAPIURL, CLOBAPIURL: DefaultCLOBAPIURL}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created in nested directory")
	}
}

func TestHasWallet(t *testing.T) {
	cfg := &Config{}
	if cfg.HasWallet() {
		t.Error("HasWallet() should be false with no key")
	}

	cfg.PrivateKey = "deadbeef"
	if !cfg.HasWallet() {
		t.Error("HasWallet() should be true with key set")
	}
}

func TestHasAPIKeys(t *testing.T) {
	cfg := &Config{}
	if cfg.HasAPIKeys() {
		t.Error("HasAPIKeys() should be false with no keys")
	}

	cfg.APIKey = "key"
	if cfg.HasAPIKeys() {
		t.Error("HasAPIKeys() should be false with partial keys")
	}

	cfg.APISecret = "secret"
	cfg.Passphrase = "pass"
	if !cfg.HasAPIKeys() {
		t.Error("HasAPIKeys() should be true with all keys set")
	}
}

func TestClearWallet(t *testing.T) {
	cfg := &Config{
		PrivateKey:    "key",
		SignatureType: "EOA",
		APIKey:        "ak",
		APISecret:     "as",
		Passphrase:    "pp",
	}

	cfg.ClearWallet()

	if cfg.PrivateKey != "" {
		t.Errorf("PrivateKey not cleared: %q", cfg.PrivateKey)
	}
	if cfg.SignatureType != "" {
		t.Errorf("SignatureType not cleared: %q", cfg.SignatureType)
	}
	if cfg.APIKey != "" || cfg.APISecret != "" || cfg.Passphrase != "" {
		t.Error("API credentials not cleared")
	}
}

func TestGetSignatureTypeInt(t *testing.T) {
	tests := []struct {
		sigType string
		want    int
	}{
		{"", 0},
		{"EOA", 0},
		{"POLY_PROXY", 1},
		{"POLY_GNOSIS_SAFE", 2},
		{"unknown", 0},
	}

	for _, tt := range tests {
		cfg := &Config{SignatureType: tt.sigType}
		got := cfg.GetSignatureTypeInt()
		if got != tt.want {
			t.Errorf("GetSignatureTypeInt(%q) = %d, want %d", tt.sigType, got, tt.want)
		}
	}
}

func TestRPCURLRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		ChainID:     137,
		GammaAPIURL: DefaultGammaAPIURL,
		CLOBAPIURL:  DefaultCLOBAPIURL,
		RPCURL:      "https://custom-rpc.example.com",
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo() error: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error: %v", err)
	}

	if loaded.RPCURL != "https://custom-rpc.example.com" {
		t.Errorf("RPCURL = %q, want custom URL", loaded.RPCURL)
	}
}

func TestRPCURLDefault(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	// Save config without RPCURL
	cfg := &Config{ChainID: 137, GammaAPIURL: DefaultGammaAPIURL, CLOBAPIURL: DefaultCLOBAPIURL}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo() error: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error: %v", err)
	}

	if loaded.RPCURL != DefaultRPCURL {
		t.Errorf("RPCURL = %q, want default %q", loaded.RPCURL, DefaultRPCURL)
	}
}

func TestAPIKeyRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		PrivateKey:    "deadbeef",
		ChainID:       137,
		SignatureType: "EOA",
		GammaAPIURL:   DefaultGammaAPIURL,
		CLOBAPIURL:    DefaultCLOBAPIURL,
		APIKey:        "my-api-key",
		APISecret:     "my-api-secret",
		Passphrase:    "my-passphrase",
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo() error: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error: %v", err)
	}

	if loaded.APIKey != "my-api-key" {
		t.Errorf("APIKey = %q, want my-api-key", loaded.APIKey)
	}
	if loaded.APISecret != "my-api-secret" {
		t.Errorf("APISecret = %q, want my-api-secret", loaded.APISecret)
	}
	if loaded.Passphrase != "my-passphrase" {
		t.Errorf("Passphrase = %q, want my-passphrase", loaded.Passphrase)
	}
	if !loaded.HasAPIKeys() {
		t.Error("HasAPIKeys() should be true after round-trip")
	}
}
