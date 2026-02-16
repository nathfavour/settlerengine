package x402

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nathfavour/settlerengine/pkg/crypto"
)

// Config defines the configuration for the x402 middleware.
type Config struct {
	DomainParams crypto.DomainParams
	NonceExpiry  time.Duration
	Recipient    string
	Asset        string
	Amount       string
}

// Middleware handles the x402 handshake.
type Middleware struct {
	config Config
	nonces *NonceManager
}

func NewMiddleware(cfg Config) *Middleware {
	return &Middleware{
		config: cfg,
		nonces: NewNonceManager(),
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Try to parse payment header
		payload, err := ParseHeader(r)
		if err == nil {
			// 2. Validate Nonce
			if m.nonces.Verify(payload.Intent.Nonce) {
				// 3. Verify Signature
				recovered, err := crypto.VerifyIntentToPay(payload.Intent, payload.Signature, m.config.DomainParams)
				if err == nil && recovered.Hex() != "" {
					// Authorized!
					// TODO: Add the recovered signer address to context if needed
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// 4. Fail and issue challenge (HTTP 402)
		nonce, _ := m.nonces.Generate(m.config.NonceExpiry)
		
		resp := ChallengeResponse{
			Status:      http.StatusPaymentRequired,
			Title:       "Payment Required",
			Description: "This resource requires a valid x402 payment signature.",
			Payment: PaymentDescriptor{
				Amount:    m.config.Amount,
				Asset:     m.config.Asset,
				Network:   m.config.DomainParams.ChainID.String(),
				Recipient: m.config.Recipient,
				Nonce:     nonce,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(resp)
	})
}
