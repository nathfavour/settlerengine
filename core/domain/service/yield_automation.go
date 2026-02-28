package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
)

// YieldService manages the automated routing of funds to yield strategies.
type YieldService struct {
	settlementEngine model.SettlementEngine
	yieldProvider    model.YieldProvider
	
	// Configurable threshold for gas efficiency (e.g., don't deposit less than $10)
	minDepositThreshold *big.Int
}

func NewYieldService(se model.SettlementEngine, yp model.YieldProvider, threshold *big.Int) *YieldService {
	return &YieldService{
		settlementEngine:    se,
		yieldProvider:       yp,
		minDepositThreshold: threshold,
	}
}

// HandleSettlementConfirmed is called when a settlement is confirmed.
// It automatically routes a portion of the funds to the configured yield strategy.
func (s *YieldService) HandleSettlementConfirmed(ctx context.Context, invoice *model.Invoice, strategy model.YieldStrategy, percentage float64) error {
	if invoice.Status != model.StatusSettled {
		return fmt.Errorf("invoice %s is not settled, status: %s", invoice.ID, invoice.Status)
	}

	amount := invoice.Amount.Amount()
	
	// Calculate amount to route: amount * percentage / 100
	routeAmountBig := new(big.Int).Mul(amount, big.NewInt(int64(percentage*100)))
	routeAmountBig.Div(routeAmountBig, big.NewInt(10000))

	// Check threshold
	if routeAmountBig.Cmp(s.minDepositThreshold) < 0 {
		return nil // Skip due to gas efficiency
	}

	routeMoney := money.New(routeAmountBig, invoice.Amount.Currency())

	// Route to yield
	return s.yieldProvider.DepositToYield(ctx, routeMoney, strategy)
}

// Rebalance checks APY and moves funds if a better strategy is available.
func (s *YieldService) Rebalance(ctx context.Context, currentStrategy model.YieldStrategy, newStrategy model.YieldStrategy) error {
	// TODO: Implement cross-vault rebalancing
	return nil
}

// StartAutoHarvestWorker runs a background loop to periodically trigger harvesting.
func (s *YieldService) StartAutoHarvestWorker(ctx context.Context, interval time.Duration, strategies []model.YieldStrategy) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, strategy := range strategies {
				if !strategy.AutoHarvest {
					continue
				}
				
				fmt.Printf("🤖 YieldService: Triggering harvest for strategy %s\n", strategy.ID)
				if err := s.yieldProvider.Harvest(ctx, strategy); err != nil {
					fmt.Printf("⚠️ YieldService: Failed to harvest %s: %v\n", strategy.ID, err)
				}
			}
		}
	}
}
