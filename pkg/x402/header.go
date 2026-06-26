package x402

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nathfavour/settlerengine/pkg/crypto"
)

const (
	HeaderPayment          = "X-Payment"
	HeaderPaymentSignature = "X-Payment-Signature"
	HeaderPaymentRequired  = "Payment-Required"
)

// PaymentPayload represents the data extracted from the payment header.
type PaymentPayload struct {
	Intent    crypto.IntentToPay `json:"intent"`
	Signature string             `json:"signature"`
	Scheme    string             `json:"scheme,omitempty"`
}

// ParseHeader extracts and decodes the payment information from a request.
func ParseHeader(r *http.Request) (*PaymentPayload, error) {
	// Try X-Payment first (JSON payload)
	if val := r.Header.Get(HeaderPayment); val != "" {
		var payload PaymentPayload
		if err := json.Unmarshal([]byte(val), &payload); err != nil {
			return nil, fmt.Errorf("failed to decode %s header: %w", HeaderPayment, err)
		}
		return &payload, nil
	}

	// Fallback to separate signature header (X-Payment-Signature)
	if sig := r.Header.Get(HeaderPaymentSignature); sig != "" {
		// Attempt to reconstruct or build a minimal payload using URL query parameters or default headers.
		return &PaymentPayload{
			Signature: sig,
			Scheme:    "casper-native",
		}, nil
	}
	
	return nil, fmt.Errorf("no payment header found")
}
