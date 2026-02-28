package service

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
)

type mockYieldProvider struct {
	depositedAmount money.Money
	harvested      bool
}

func (m *mockYieldProvider) DepositToYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	m.depositedAmount = amount
	return nil
}

func (m *mockYieldProvider) WithdrawFromYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	return nil
}

func (m *mockYieldProvider) GetAPY(ctx context.Context, vaultAddress string) (float64, error) {
	return 10.5, nil
}

func (m *mockYieldProvider) Harvest(ctx context.Context, strategy model.YieldStrategy) error {
	m.harvested = true
	return nil
}

type mockSettlementEngine struct {
	model.SettlementEngine
}

func TestYieldService_HandleSettlementConfirmed(t *testing.T) {
	provider := &mockYieldProvider{}
	engine := &mockSettlementEngine{}
	
	// Threshold: 10 units
	threshold := big.NewInt(10)
	svc := NewYieldService(engine, provider, threshold)

	strategy := model.YieldStrategy{
		ID:           "test_strategy",
		VaultAddress: "0xvault",
	}

	t.Run("Should route funds when above threshold", func(t *testing.T) {
		// Invoice for 100 units
		inv := &model.Invoice{
			ID:     "inv_1",
			Amount: money.New(big.NewInt(100), "USDT"),
			Status: model.StatusSettled,
		}

		// Route 50%
		err := svc.HandleSettlementConfirmed(context.Background(), inv, strategy, 50.0)
		if err != nil {
			t.Fatalf("HandleSettlementConfirmed failed: %v", err)
		}

		if provider.depositedAmount.Amount().Cmp(big.NewInt(50)) != 0 {
			t.Errorf("Expected 50 deposited, got %s", provider.depositedAmount.Amount().String())
		}
	})

	t.Run("Should skip routing when below threshold", func(t *testing.T) {
		provider.depositedAmount = money.New(big.NewInt(0), "USDT")
		
		// Invoice for 10 units, route 50% = 5 units (below 10 unit threshold)
		inv := &model.Invoice{
			ID:     "inv_2",
			Amount: money.New(big.NewInt(10), "USDT"),
			Status: model.StatusSettled,
		}

		err := svc.HandleSettlementConfirmed(context.Background(), inv, strategy, 50.0)
		if err != nil {
			t.Fatalf("HandleSettlementConfirmed failed: %v", err)
		}

		if provider.depositedAmount.Amount().Cmp(big.NewInt(0)) != 0 {
			t.Errorf("Expected 0 deposited, got %s", provider.depositedAmount.Amount().String())
		}
	})
}

func TestYieldService_AutoHarvestWorker(t *testing.T) {
	provider := &mockYieldProvider{}
	engine := &mockSettlementEngine{}
	svc := NewYieldService(engine, provider, big.NewInt(0))

	strategies := []model.YieldStrategy{
		{ID: "s1", AutoHarvest: true, VaultAddress: "0x1"},
		{ID: "s2", AutoHarvest: false, VaultAddress: "0x2"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// Run worker in background
	go svc.StartAutoHarvestWorker(ctx, 100*time.Millisecond, strategies)

	// Wait for ticker
	time.Sleep(200 * time.Millisecond)

	if !provider.harvested {
		t.Error("Expected harvest to be called")
	}
}
