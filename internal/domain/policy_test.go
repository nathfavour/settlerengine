package domain

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func TestPaymentPolicy(t *testing.T) {
	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
	otherRecipient := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	t.Run("MaxAmountPerTx", func(t *testing.T) {
		policy := NewPaymentPolicy("test", big.NewInt(100), nil, time.Time{})
		
		err := policy.Check(big.NewInt(50), recipient)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = policy.Check(big.NewInt(150), recipient)
		if err == nil {
			t.Error("expected error for exceeding max amount per tx")
		}
	})

	t.Run("DailyBudget", func(t *testing.T) {
		policy := NewPaymentPolicy("test", nil, big.NewInt(100), time.Time{})
		
		err := policy.Check(big.NewInt(60), recipient)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		policy.RecordUpdate(big.NewInt(60))

		err = policy.Check(big.NewInt(50), recipient)
		if err == nil {
			t.Error("expected error for exceeding daily budget")
		}
	})

	t.Run("Whitelist", func(t *testing.T) {
		policy := NewPaymentPolicy("test", nil, nil, time.Time{})
		policy.WhitelistedRecipients = []common.Address{recipient}

		err := policy.Check(big.NewInt(50), recipient)
		if err != nil {
			t.Errorf("expected no error for whitelisted recipient, got %v", err)
		}

		err = policy.Check(big.NewInt(50), otherRecipient)
		if err == nil {
			t.Error("expected error for non-whitelisted recipient")
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		policy := NewPaymentPolicy("test", nil, nil, time.Now().Add(-1*time.Hour))
		
		err := policy.Check(big.NewInt(50), recipient)
		if err == nil {
			t.Error("expected error for expired policy")
		}
	})
}
