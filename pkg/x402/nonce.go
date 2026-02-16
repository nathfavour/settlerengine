package x402

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// NonceManager handles the generation and validation of cryptographic nonces.
type NonceManager struct {
	nonces sync.Map // Map of string nonce to expiry time.Time
}

func NewNonceManager() *NonceManager {
	return &NonceManager{}
}

// Generate creates a new nonce that expires after a certain duration.
func (nm *NonceManager) Generate(expiry time.Duration) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}
	nonce := hex.EncodeToString(b)
	nm.nonces.Store(nonce, time.Now().Add(expiry))
	return nonce, nil
}

// Verify checks if a nonce is valid and has not expired.
func (nm *NonceManager) Verify(nonce string) bool {
	val, ok := nm.nonces.Load(nonce)
	if !ok {
		return false
	}
	expiry := val.(time.Time)
	if time.Now().After(expiry) {
		nm.nonces.Delete(nonce)
		return false
	}
	// Once verified, we should probably invalidate it to prevent reuse if needed,
	// but for x402 handshake, a session nonce might be reused within its TTL.
	// For strict single-use, we would nm.nonces.Delete(nonce) here.
	return true
}

// Cleanup removes expired nonces from the map.
func (nm *NonceManager) Cleanup() {
	nm.nonces.Range(func(key, value interface{}) bool {
		if time.Now().After(value.(time.Time)) {
			nm.nonces.Delete(key)
		}
		return true
	})
}
