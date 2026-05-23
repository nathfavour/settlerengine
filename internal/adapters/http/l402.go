package http

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type L402Middleware struct {
	store        ports.DBStore
	serverSecret []byte
}

func NewL402Middleware(store ports.DBStore, secret string) *L402Middleware {
	return &L402Middleware{
		store:        store,
		serverSecret: []byte(secret),
	}
}

type Macaroon struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	ExpiresAt int64  `json:"expires_at"`
	Signature string `json:"signature"`
}

// GenerateMacaroon creates and signs a new macaroon for L402 resource gating.
func (m *L402Middleware) GenerateMacaroon(id, path string, duration time.Duration) (string, error) {
	mac := Macaroon{
		ID:        id,
		Path:      path,
		ExpiresAt: time.Now().Add(duration).Unix(),
	}

	// Compute signature over metadata
	h := hmac.New(sha256.New, m.serverSecret)
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", mac.ID, mac.Path, mac.ExpiresAt)))
	mac.Signature = hex.EncodeToString(h.Sum(nil))

	bytes, err := json.Marshal(mac)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// VerifyMacaroon decodes and cryptographically validates the macaroon signature and caveats.
func (m *L402Middleware) VerifyMacaroon(tokenStr, currentPath string) (*Macaroon, error) {
	bytes, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}

	var mac Macaroon
	if err := json.Unmarshal(bytes, &mac); err != nil {
		return nil, fmt.Errorf("invalid macaroon structure: %w", err)
	}

	// Verify signature
	h := hmac.New(sha256.New, m.serverSecret)
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", mac.ID, mac.Path, mac.ExpiresAt)))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(mac.Signature), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid macaroon signature")
	}

	// Verify caveats
	if time.Now().Unix() > mac.ExpiresAt {
		return nil, fmt.Errorf("macaroon expired")
	}

	if mac.Path != currentPath {
		return nil, fmt.Errorf("macaroon path mismatch: expected %s, got %s", mac.Path, currentPath)
	}

	return &mac, nil
}

func (m *L402Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if strings.HasPrefix(authHeader, "L402 ") {
			tokenParts := strings.Split(strings.TrimPrefix(authHeader, "L402 "), ":")
			if len(tokenParts) == 2 {
				macBase64 := tokenParts[0]
				preimageHex := tokenParts[1]

				// 1. Verify Macaroon Caveats
				mac, err := m.VerifyMacaroon(macBase64, r.URL.Path)
				if err == nil {
					// 2. Validate Preimage signature hash
					preimageBytes, err := hex.DecodeString(preimageHex)
					if err == nil {
						h := sha256.New()
						h.Write(preimageBytes)
						derivedHash := hex.EncodeToString(h.Sum(nil))

						// Retrieve challenge from active database
						challenge, err := m.store.GetLsatChallenge(r.Context(), mac.ID)
						if err == nil && challenge != nil {
							if challenge.PreimageHash == derivedHash {
								// Success! Record paid preimage (idempotency) and authorize request
								_ = m.store.UpdateLsatChallengePreimage(r.Context(), mac.ID, preimageHex)
								next.ServeHTTP(w, r)
								return
							}
						}
					}
				}
			}
		}

		// 3. Issue L402 / LSAT challenge (HTTP 402 Payment Required)
		macID := uuid.New().String()
		
		// Generate random preimage
		preimageBytes := make([]byte, 32)
		_, _ = rand.Read(preimageBytes)
		preimageHex := hex.EncodeToString(preimageBytes)
		
		h := sha256.New()
		h.Write(preimageBytes)
		preimageHashHex := hex.EncodeToString(h.Sum(nil))

		// Save LSAT challenge in active store
		challenge := &domain.LsatChallenge{
			MacaroonID:   macID,
			PreimageHash: preimageHashHex,
			Preimage:     "", // unpaid
			ResourcePath: r.URL.Path,
			Amount:       1000, // 1000 satoshis default
			CreatedAt:    time.Now(),
		}
		_ = m.store.SaveLsatChallenge(r.Context(), challenge)

		// Create Macaroon
		macaroonBase64, err := m.GenerateMacaroon(macID, r.URL.Path, 1*time.Hour)
		if err != nil {
			http.Error(w, "macaroon generation failed", http.StatusInternalServerError)
			return
		}

		// Mock lightning invoice representing preimages matching the payment hash
		mockInvoice := fmt.Sprintf("lnbc10u1pv5s...mock_invoice_hash_%s", preimageHashHex[:8])

		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`L402 macaroon="%s", invoice="%s"`, macaroonBase64, mockInvoice))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		
		// Return preimage details to allow offline programmatic solving in tests
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":            http.StatusPaymentRequired,
			"title":             "Payment Required",
			"description":       "L402 / LSAT token required. Complete payment of the invoice to obtain preimage.",
			"preimage_for_test": preimageHex,
		})
	})
}
