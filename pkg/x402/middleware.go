package x402

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nathfavour/settlerengine/internal/ports"
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/storage"
)

type contextKey string

const (
	SignerContextKey contextKey = "x402-signer"
	AgentContextKey  contextKey = "x402-agent"
)

// Config defines the configuration for the x402 middleware.
type Config struct {
	DomainParams  crypto.DomainParams
	NonceExpiry   time.Duration
	Recipient     string
	Asset         string
	Amount        string
	PriceResolver PriceResolver
	DB            *storage.DB
	Registry      ports.AgentRegistry
	MinReputation *big.Int // Optional: block agents with score below this
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

// GetAgent returns the recovered agent identity from the request context.
func GetAgent(ctx context.Context) (*common.Address, bool) {
	addr, ok := ctx.Value(AgentContextKey).(*common.Address)
	return addr, ok
}

// Handler handles the x402 handshake.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Try to parse payment header
		payload, err := ParseHeader(r)
		if err == nil {
			// 2. Check Cache & DB (Idempotency)
			if addr, ok := m.verified.Load(payload.Signature); ok {
				ctx := context.WithValue(r.Context(), SignerContextKey, addr.(common.Address))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if m.config.DB != nil {
				signer, err := m.config.DB.CheckPayment(payload.Signature)
				if err == nil && signer != "" {
					recovered := common.HexToAddress(signer)
					m.verified.Store(payload.Signature, recovered)
					ctx := context.WithValue(r.Context(), SignerContextKey, recovered)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// 3. Validate Nonce
			if m.nonces.Verify(payload.Intent.Nonce) {
				// 4. Verify Signature
				recovered, err := crypto.VerifyIntentToPay(payload.Intent, payload.Signature, m.config.DomainParams)
				if err == nil && recovered.Hex() != "" {
					// 4a. ERC-8004 Trust Checks
					if m.config.Registry != nil {
						// For this demo, we assume the agent ID is passed or mapped from the recovered address.
						// In reality, we'd lookup the tokenId owned by 'recovered'.
						// Placeholder: assume tokenId 42
						agentID := big.NewInt(42) 
						
						rep, err := m.config.Registry.GetReputation(r.Context(), agentID)
						if err == nil && m.config.MinReputation != nil {
							if rep.Score.Cmp(m.config.MinReputation) < 0 {
								http.Error(w, fmt.Sprintf("Agent reputation too low: %s", rep.Score.String()), http.StatusForbidden)
								return
							}
						}
						fmt.Printf("🔍 ERC-8004: Agent %s verified with reputation score %s\n", agentID.String(), rep.Score.String())
					}

					// Authorized!
					m.verified.Store(payload.Signature, recovered)

					if m.config.DB != nil {
						_ = m.config.DB.RecordPayment(
							payload.Signature,
							recovered.Hex(),
							payload.Intent.Amount,
							payload.Intent.Asset,
							payload.Intent.Nonce,
						)
					}
					
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
			Accepts: []PaymentDescriptor{
				{
					Scheme:  "x402",
					Price:   amount,
					Asset:   asset,
					Network: m.config.DomainParams.ChainID.String(),
					PayTo:   recipient,
					Nonce:   nonce,
				},
			},
			Resource: r.URL.Path,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(resp)
	})
}
