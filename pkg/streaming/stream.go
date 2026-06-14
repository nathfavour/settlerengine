package streaming

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nathfavour/settlerengine/pkg/crypto"
)

// Stream represents a persistent payment channel between an agent and a service.
type Stream struct {
	ID        string
	Recipient common.Address
	Asset     common.Address
	Rate      *big.Int      // Amount per interval
	Interval  time.Duration
	
	mu        sync.RWMutex
	LastPulse time.Time
	TotalPaid *big.Int
	Active    bool
}

// NewStream creates a new payment stream.
func NewStream(id string, recipient, asset common.Address, rate *big.Int, interval time.Duration) *Stream {
	return &Stream{
		ID:        id,
		Recipient: recipient,
		Asset:     asset,
		Rate:      rate,
		Interval:  interval,
		TotalPaid: big.NewInt(0),
		Active:    true,
	}
}

// ValidatePulse checks if a pulse signature is valid and covers the required interval.
func (s *Stream) ValidatePulse(intent crypto.IntentToPay, signature string, domain crypto.DomainParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Active {
		return fmt.Errorf("stream is inactive")
	}

	// 1. Verify Signature
	_, err := crypto.VerifyIntentToPay(intent, signature, domain)
	if err != nil {
		return fmt.Errorf("invalid pulse signature: %w", err)
	}

	// 2. Verify Recipient and Asset
	if intent.Recipient != s.Recipient.Hex() {
		return fmt.Errorf("wrong recipient in pulse")
	}
	if intent.Asset != s.Asset.Hex() {
		return fmt.Errorf("wrong asset in pulse")
	}

	// 3. Verify Amount (must be >= Rate)
	amount, ok := new(big.Int).SetString(intent.Amount, 10)
	if !ok || amount.Cmp(s.Rate) < 0 {
		return fmt.Errorf("insufficient amount in pulse")
	}

	// 4. Update state
	s.LastPulse = time.Now()
	s.TotalPaid = new(big.Int).Add(s.TotalPaid, amount)

	return nil
}

// Monitor monitors the stream and deactivates it if a pulse is missed.
func (s *Stream) Monitor(ctx context.Context, gracePeriod time.Duration) {
	ticker := time.NewTicker(s.Interval / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			lastPulse := s.LastPulse
			active := s.Active
			s.mu.RUnlock()

			if !active {
				return
			}

			if !lastPulse.IsZero() && time.Since(lastPulse) > (s.Interval + gracePeriod) {
				s.mu.Lock()
				s.Active = false
				s.mu.Unlock()
				fmt.Printf("🔴 Stream %s deactivated: pulse timeout\n", s.ID)
				return
			}
		}
	}
}
