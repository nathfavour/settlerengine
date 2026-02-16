package x402

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nathfavour/settlerengine/pkg/crypto"
)

type contextKey string

const (
	SignerContextKey contextKey = "x402-signer"
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
	verified sync.Map // Map of signature hash to Address
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

// GetSigner returns the recovered signer address from the request context.
func GetSigner(ctx context.Context) (common.Address, bool) {
	addr, ok := ctx.Value(SignerContextKey).(common.Address)
	return addr, ok
}

// Handler handles the x402 handshake.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Try to parse payment header
		payload, err := ParseHeader(r)
		if err == nil {
			// 2. Check Cache (Idempotency)
			if addr, ok := m.verified.Load(payload.Signature); ok {
				ctx := context.WithValue(r.Context(), SignerContextKey, addr.(common.Address))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 3. Validate Nonce
			if m.nonces.Verify(payload.Intent.Nonce) {
				// 4. Verify Signature
				recovered, err := crypto.VerifyIntentToPay(payload.Intent, payload.Signature, m.config.DomainParams)
				if err == nil && recovered.Hex() != "" {
					// Authorized!
					m.verified.Store(payload.Signature, recovered)
					
					ctx := context.WithValue(r.Context(), SignerContextKey, recovered)
					next.ServeHTTP(w, r.WithContext(ctx))
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
