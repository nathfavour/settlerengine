package x402

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nathfavour/settlerengine/pkg/crypto"
)

// Config defines the configuration for the x402 middleware.
type Config struct {
	DomainParams  crypto.DomainParams
	NonceExpiry   time.Duration
	Recipient     string
	Asset         string
	Amount        string
	PriceResolver PriceResolver
}

// PriceResolver dynamically determines the payment requirements for a request.
type PriceResolver func(r *http.Request) (amount, asset, recipient string, err error)

// Middleware handles the x402 handshake.
type Middleware struct {
	config   Config
	nonces   *NonceManager
	verified sync.Map // Map of signature hash to expiry time.Time
}

func NewMiddleware(cfg Config) *Middleware {
	if cfg.PriceResolver == nil {
		cfg.PriceResolver = func(r *http.Request) (string, string, string, error) {
			return cfg.Amount, cfg.Asset, cfg.Recipient, nil
		}
	}

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
			// 2. Check Cache (Idempotency)
			if _, ok := m.verified.Load(payload.Signature); ok {
				next.ServeHTTP(w, r)
				return
			}

			// 3. Validate Nonce
			if m.nonces.Verify(payload.Intent.Nonce) {
				// 4. Verify Signature
				recovered, err := crypto.VerifyIntentToPay(payload.Intent, payload.Signature, m.config.DomainParams)
				if err == nil && recovered.Hex() != "" {
					// Authorized!
					m.verified.Store(payload.Signature, time.Now().Add(m.config.NonceExpiry))
					// TODO: Add the recovered signer address to context if needed
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// 5. Fail and issue challenge (HTTP 402)
		amount, asset, recipient, err := m.config.PriceResolver(r)
		if err != nil {
			http.Error(w, "Failed to resolve price", http.StatusInternalServerError)
			return
		}

		nonce, _ := m.nonces.Generate(m.config.NonceExpiry)

		resp := ChallengeResponse{
			Status:      http.StatusPaymentRequired,
			Title:       "Payment Required",
			Description: "This resource requires a valid x402 payment signature.",
			Payment: PaymentDescriptor{
				Amount:    amount,
				Asset:     asset,
				Network:   m.config.DomainParams.ChainID.String(),
				Recipient: recipient,
				Nonce:     nonce,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(resp)
	})
}
