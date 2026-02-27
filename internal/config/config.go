package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultGammaAPIURL = "https://gamma-api.polymarket.com"
	DefaultCLOBAPIURL  = "https://clob.polymarket.com"
	DefaultRPCURL      = "https://polygon-rpc.com"
	DefaultDataAPIURL  = "https://data-api.polymarket.com"
	DefaultBridgeAPIURL = "https://bridge.polymarket.com"
	DefaultChainID     = 137
)

// Config holds the CLI configuration.
type Config struct {
	PrivateKey    string `json:"private_key,omitempty"`
	ChainID       int    `json:"chain_id,omitempty"`
	SignatureType string `json:"signature_type,omitempty"`
	GammaAPIURL   string `json:"gamma_api_url,omitempty"`
	CLOBAPIURL    string `json:"clob_api_url,omitempty"`
	RPCURL        string `json:"rpc_url,omitempty"`
	DataAPIURL    string `json:"data_api_url,omitempty"`
	BridgeAPIURL  string `json:"bridge_api_url,omitempty"`
	APIKey        string `json:"api_key,omitempty"`
	APISecret     string `json:"api_secret,omitempty"`
	Passphrase    string `json:"passphrase,omitempty"`
}

// HasWallet returns true if a private key is configured.
func (c *Config) HasWallet() bool {
	return c.PrivateKey != ""
}

// HasAPIKeys returns true if derived API credentials are cached.
func (c *Config) HasAPIKeys() bool {
	return c.APIKey != "" && c.APISecret != "" && c.Passphrase != ""
}

// ClearWallet removes all wallet and API key fields.
func (c *Config) ClearWallet() {
	c.PrivateKey = ""
	c.SignatureType = ""
	c.APIKey = ""
	c.APISecret = ""
	c.Passphrase = ""
}

// GetSignatureTypeInt returns the signature type as an integer.
// 0=EOA (default), 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE
func (c *Config) GetSignatureTypeInt() int {
	switch c.SignatureType {
	case "POLY_PROXY":
		return 1
	case "POLY_GNOSIS_SAFE":
		return 2
	default:
		return 0
	}
}

// DefaultConfigDir returns the default config directory path.
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "polymarket")
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.json")
}

// Load reads the config from the default path.
func Load() (*Config, error) {
	return LoadFrom(DefaultConfigPath())
}

// LoadFrom reads the config from a specific path.
func LoadFrom(path string) (*Config, error) {
	cfg := &Config{
		ChainID:     DefaultChainID,
		GammaAPIURL: DefaultGammaAPIURL,
		CLOBAPIURL:  DefaultCLOBAPIURL,
		RPCURL:       DefaultRPCURL,
		DataAPIURL:   DefaultDataAPIURL,
		BridgeAPIURL: DefaultBridgeAPIURL,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply defaults for empty fields
	if cfg.GammaAPIURL == "" {
		cfg.GammaAPIURL = DefaultGammaAPIURL
	}
	if cfg.CLOBAPIURL == "" {
		cfg.CLOBAPIURL = DefaultCLOBAPIURL
	}
	if cfg.ChainID == 0 {
		cfg.ChainID = DefaultChainID
	}
	if cfg.RPCURL == "" {
		cfg.RPCURL = DefaultRPCURL
	}
	if cfg.DataAPIURL == "" {
		cfg.DataAPIURL = DefaultDataAPIURL
	}
	if cfg.BridgeAPIURL == "" {
		cfg.BridgeAPIURL = DefaultBridgeAPIURL
	}

	return cfg, nil
}

// Save writes the config to the default path.
func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath())
}

// SaveTo writes the config to a specific path using atomic write
// (write to temp file + rename) to prevent corruption from concurrent
// access or crashes mid-write.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file in the same directory, then rename for atomicity.
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, path)
}
