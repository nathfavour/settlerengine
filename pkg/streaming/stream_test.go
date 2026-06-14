package streaming

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func TestStream(t *testing.T) {
	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
	asset := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	rate := big.NewInt(1000)
	interval := 100 * time.Millisecond

	s := NewStream("test-stream", recipient, asset, rate, interval)

	t.Run("InitialState", func(t *testing.T) {
		if !s.Active {
			t.Error("expected stream to be active")
		}
	})

	t.Run("MonitorTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		s.LastPulse = time.Now()
		go s.Monitor(ctx, 50*time.Millisecond)

		time.Sleep(300 * time.Millisecond)

		s.mu.RLock()
		active := s.Active
		s.mu.RUnlock()

		if active {
			t.Error("expected stream to be deactivated after timeout")
		}
	})
}
