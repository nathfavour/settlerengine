package model

import (
	"context"

	"github.com/nathfavour/settlerengine/pkg/money"
)

// SettlementEngine defines the core driving port for settlement operations.
type SettlementEngine interface {
	// CreateInvoice initiates a new settlement request.
	CreateInvoice(ctx context.Context, amount money.Money) (*Invoice, error)

	// GetInvoice retrieves the current state of an invoice.
	GetInvoice(ctx context.Context, id string) (*Invoice, error)

	// DepositToYield moves idle settlement funds into a yield-generating strategy.
	DepositToYield(ctx context.Context, amount money.Money, strategy YieldStrategy) error

	// WithdrawFromYield retrieves funds from a yield strategy back to the settlement account.
	WithdrawFromYield(ctx context.Context, amount money.Money, strategy YieldStrategy) error
}
