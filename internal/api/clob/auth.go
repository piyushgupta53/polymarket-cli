package clob

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
)

// AuthCredentials holds HMAC-based CLOB API credentials.
type AuthCredentials struct {
	Key        string
	Secret     string
	Passphrase string
	Address    string
}

// BuildHMACSignature constructs an HMAC-SHA256 signature for CLOB auth.
func BuildHMACSignature(secret, timestamp, method, path, body string) (string, error) {
	secretBytes, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decoding secret: %w", err)
	}

	message := timestamp + method + path + body
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(message))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return sig, nil
}

// ApplyAuthHeaders sets CLOB API auth headers on a request.
func ApplyAuthHeaders(req *http.Request, auth *AuthCredentials, timestamp, signature string) {
	req.Header.Set("POLY_ADDRESS", auth.Address)
	req.Header.Set("POLY_API_KEY", auth.Key)
	req.Header.Set("POLY_SIGNATURE", signature)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_PASSPHRASE", auth.Passphrase)
}

// doAuthRequest performs an authenticated HTTP request.
func (c *Client) doAuthRequest(method, path string, body []byte) ([]byte, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required: no credentials configured")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	bodyStr := ""
	if body != nil {
		bodyStr = string(body)
	}

	sig, err := BuildHMACSignature(c.auth.Secret, timestamp, method, path, bodyStr)
	if err != nil {
		return nil, fmt.Errorf("building signature: %w", err)
	}

	u := c.baseURL + path
	log.Debug("CLOB auth request", "method", method, "url", u)

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	ApplyAuthHeaders(req, c.auth, timestamp, sig)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("CLOB API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// doAuthGet performs an authenticated GET request.
func (c *Client) doAuthGet(path string) ([]byte, error) {
	return c.doAuthRequest("GET", path, nil)
}

// doAuthPost performs an authenticated POST request.
func (c *Client) doAuthPost(path string, body []byte) ([]byte, error) {
	return c.doAuthRequest("POST", path, body)
}

// doAuthDelete performs an authenticated DELETE request.
func (c *Client) doAuthDelete(path string) ([]byte, error) {
	return c.doAuthRequest("DELETE", path, nil)
}

// doAuthDeleteWithBody performs an authenticated DELETE request with a body.
func (c *Client) doAuthDeleteWithBody(path string, body []byte) ([]byte, error) {
	return c.doAuthRequest("DELETE", path, body)
}
