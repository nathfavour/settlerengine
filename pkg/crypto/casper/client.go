package casper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PaymentDetails represents the payment transaction parameters verified/settled by CSPR.cloud x402 Facilitator.
type PaymentDetails struct {
	Recipient string `json:"recipient"` // Casper public key hex
	Amount    string `json:"amount"`    // Amount in motes (1 CSPR = 10^9 motes)
	Asset     string `json:"asset"`     // e.g. "CSPR"
	Nonce     string `json:"nonce"`     // Unique session UUID for replay protection
	Network   string `json:"network"`   // e.g. "casper-testnet"
}

// CasperFacilitatorClient represents the client wrapper for interacting with CSPR.cloud x402 Facilitator.
type CasperFacilitatorClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewCasperFacilitatorClient initializes a new Casper x402 Facilitator client.
func NewCasperFacilitatorClient(baseURL, apiKey string) *CasperFacilitatorClient {
	if baseURL == "" {
		baseURL = "https://x402-facilitator.cspr.cloud"
	}
	return &CasperFacilitatorClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type verifyRequest struct {
	Signature string         `json:"signature"` // base64-encoded Casper Ed25519 signature
	Details   PaymentDetails `json:"details"`
}

type verifyResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type settleRequest struct {
	Signature string         `json:"signature"` // base64-encoded Casper Ed25519 signature
	Details   PaymentDetails `json:"details"`
}

type settleResponse struct {
	TxHash string `json:"txHash"`
	Error  string `json:"error,omitempty"`
}

// VerifyPayload sends the payment signature and details to the facilitator for verification.
func (c *CasperFacilitatorClient) VerifyPayload(signature string, details PaymentDetails) (bool, error) {
	url := fmt.Sprintf("%s/verify", c.BaseURL)
	reqBody, err := json.Marshal(verifyRequest{
		Signature: signature,
		Details:   details,
	})
	if err != nil {
		return false, fmt.Errorf("failed to marshal verify request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return false, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send verify request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, fmt.Errorf("failed to decode verify response: %w", err)
	}

	if res.Error != "" {
		return false, fmt.Errorf("facilitator error: %s", res.Error)
	}

	return res.Valid, nil
}

// SettlePayload forwards the signature and details to the facilitator to execute the Casper payment.
func (c *CasperFacilitatorClient) SettlePayload(signature string, details PaymentDetails) (string, error) {
	url := fmt.Sprintf("%s/settle", c.BaseURL)
	reqBody, err := json.Marshal(settleRequest{
		Signature: signature,
		Details:   details,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal settle request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send settle request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res settleResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode settle response: %w", err)
	}

	if res.Error != "" {
		return "", fmt.Errorf("facilitator error: %s", res.Error)
	}

	return res.TxHash, nil
}
