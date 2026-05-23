package http

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
	"github.com/nathfavour/settlerengine/internal/domain"
)

func TestDynamicPricingEngine(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pricing_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := sqlite.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize sqlite: %v", err)
	}
	defer store.Close()

	engine := NewDynamicPricingEngine(store)
	ctx := context.Background()

	// 1. Fallback to default pricing
	price, err := engine.CalculatePrice(ctx, "/v1/data", "0xclient1")
	if err != nil {
		t.Fatalf("pricing calculation failed: %v", err)
	}
	if price.Amount().Int64() != 1000000 || price.Currency() != "USDC" {
		t.Errorf("expected fallback price 1000000 USDC, got %s %s", price.Amount().String(), price.Currency())
	}

	// 2. Setup pricing policy: $2.00 USDC, 1.5x surge multiplier (target: $3.00 USDC)
	policy := &domain.PricingPolicy{
		ResourcePath:    "/v1/compute",
		BasePrice:       domain.NewMoney(big.NewInt(2000000), "USDC"),
		SurgeMultiplier: 1.5,
	}
	if err := store.SavePolicy(ctx, policy); err != nil {
		t.Fatalf("failed to save policy: %v", err)
	}

	price, err = engine.CalculatePrice(ctx, "/v1/compute", "0xclient1")
	if err != nil {
		t.Fatalf("pricing calculation failed: %v", err)
	}
	if price.Amount().Int64() != 3000000 {
		t.Errorf("expected surge price 3000000, got %s", price.Amount().String())
	}

	// 3. Setup client reputation: High score (90) -> 20% discount (target: $3.00 * 0.8 = $2.40 USDC)
	repHigh := &domain.ClientReputation{
		ClientAddress: "0xhighrep",
		Score:         90,
		TotalPayments: domain.NewMoney(big.NewInt(50000000), "USDC"),
		LastPaymentAt: time.Now(),
	}
	if err := store.SaveReputation(ctx, repHigh); err != nil {
		t.Fatalf("failed to save reputation: %v", err)
	}

	price, err = engine.CalculatePrice(ctx, "/v1/compute", "0xhighrep")
	if err != nil {
		t.Fatalf("pricing calculation failed: %v", err)
	}
	if price.Amount().Int64() != 2400000 {
		t.Errorf("expected discounted price 2400000, got %s", price.Amount().String())
	}

	// 4. Setup client reputation: Low score (10) -> 10% premium (target: $3.00 * 1.1 = $3.30 USDC)
	repLow := &domain.ClientReputation{
		ClientAddress: "0xlowrep",
		Score:         10,
		TotalPayments: domain.NewMoney(big.NewInt(0), "USDC"),
		LastPaymentAt: time.Now(),
	}
	if err := store.SaveReputation(ctx, repLow); err != nil {
		t.Fatalf("failed to save reputation: %v", err)
	}

	price, err = engine.CalculatePrice(ctx, "/v1/compute", "0xlowrep")
	if err != nil {
		t.Fatalf("pricing calculation failed: %v", err)
	}
	if price.Amount().Int64() != 3300000 {
		t.Errorf("expected premium price 3300000, got %s", price.Amount().String())
	}
}
