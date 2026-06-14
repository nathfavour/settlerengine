package domain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// PaymentPolicy defines constraints for automated payment signing.
type PaymentPolicy struct {
	ID                 string
	MaxAmountPerTx     *big.Int
	DailyBudget        *big.Int
	SpentToday         *big.Int
	Expiration         time.Time
	WhitelistedRecipients []common.Address
	LastReset          time.Time
}

// NewPaymentPolicy creates a new policy with defaults.
func NewPaymentPolicy(id string, maxPerTx, dailyBudget *big.Int, expiration time.Time) *PaymentPolicy {
	return &PaymentPolicy{
		ID:             id,
		MaxAmountPerTx: maxPerTx,
		DailyBudget:    dailyBudget,
		SpentToday:     big.NewInt(0),
		Expiration:     expiration,
		LastReset:      time.Now(),
	}
}

// Check validates if a payment conforms to the policy.
func (p *PaymentPolicy) Check(amount *big.Int, recipient common.Address) error {
	// 1. Check expiration
	if !p.Expiration.IsZero() && time.Now().After(p.Expiration) {
		return PolicyError{Message: "policy has expired"}
	}

	// 2. Check Whitelist
	if len(p.WhitelistedRecipients) > 0 {
		allowed := false
		for _, addr := range p.WhitelistedRecipients {
			if addr == recipient {
				allowed = true
				break
			}
		}
		if !allowed {
			return PolicyError{Message: "recipient is not whitelisted"}
		}
	}

	// 3. Check Max per Tx
	if p.MaxAmountPerTx != nil && amount.Cmp(p.MaxAmountPerTx) > 0 {
		return PolicyError{Message: "amount exceeds max per transaction"}
	}

	// 4. Reset daily budget if needed
	now := time.Now()
	if now.Year() != p.LastReset.Year() || now.YearDay() != p.LastReset.YearDay() {
		p.SpentToday = big.NewInt(0)
		p.LastReset = now
	}

	// 5. Check Daily Budget
	if p.DailyBudget != nil {
		newSpent := new(big.Int).Add(p.SpentToday, amount)
		if newSpent.Cmp(p.DailyBudget) > 0 {
			return PolicyError{Message: "daily budget exceeded"}
		}
	}

	return nil
}

// RecordUpdate updates the spent amount.
func (p *PaymentPolicy) RecordUpdate(amount *big.Int) {
	p.SpentToday = new(big.Int).Add(p.SpentToday, amount)
}

type PolicyError struct {
	Message string
}

func (e PolicyError) Error() string {
	return e.Message
}
