package x402

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nathfavour/settlerengine/pkg/crypto"
)

const (
	HeaderPayment          = "X-Payment"
	HeaderPaymentSignature = "X-Payment-Signature"
)

// PaymentPayload represents the data extracted from the payment header.
type PaymentPayload struct {
	Intent    crypto.IntentToPay `json:"intent"`
	Signature string             `json:"signature"`
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

	// Fallback to separate signature header (simplified version)
	// This would require the intent to be reconstructible or passed elsewhere.
	// For MVP, we'll focus on the self-contained JSON payload.
	
	return nil, fmt.Errorf("no payment header found")
}
