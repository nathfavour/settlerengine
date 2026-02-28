package model

import (
	"context"
	"time"

	"github.com/nathfavour/settlerengine/core/pkg/money"
)

// YieldStrategy represents the configuration for automated yield generation.
type YieldStrategy struct {
	ID            string
	Provider      string // e.g., "Riquid"
	TargetAPY     float64
	AutoHarvest   bool
	LastHarvested time.Time
	VaultAddress  string
}

// YieldProvider defines the port for interacting with external yield engines.
type YieldProvider interface {
	// DepositToYield transfers assets from the main settlement balance to a yield-generating vault.
	DepositToYield(ctx context.Context, amount money.Money, strategy YieldStrategy) error

	// WithdrawFromYield pulls assets from a yield-generating vault back to the main settlement balance.
	WithdrawFromYield(ctx context.Context, amount money.Money, strategy YieldStrategy) error

	// GetAPY returns the current Annual Percentage Yield for a given vault.
	GetAPY(ctx context.Context, vaultAddress string) (float64, error)

	// Harvest triggers the claiming and reinvesting of accrued yield.
	Harvest(ctx context.Context, strategy YieldStrategy) error
}
