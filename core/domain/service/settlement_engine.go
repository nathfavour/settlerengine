package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
)

// DefaultSettlementEngine is the standard implementation of the SettlementEngine port.
type DefaultSettlementEngine struct {
	repo          model.InvoiceRepository
	chainClient   model.BlockchainClient
	yieldProvider model.YieldProvider
	bus           *LocalBus
}

func NewDefaultSettlementEngine(
	repo model.InvoiceRepository,
	chainClient model.BlockchainClient,
	yieldProvider model.YieldProvider,
	bus *LocalBus,
) *DefaultSettlementEngine {
	return &DefaultSettlementEngine{
		repo:          repo,
		chainClient:   chainClient,
		yieldProvider: yieldProvider,
		bus:           bus,
	}
}

func (s *DefaultSettlementEngine) CreateInvoice(ctx context.Context, amount money.Money) (*model.Invoice, error) {
	id := uuid.New().String()
	invoice := model.NewInvoice(id, amount, 1*time.Hour) // Default 1h expiry

	if err := s.repo.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return invoice, nil
}

func (s *DefaultSettlementEngine) GetInvoice(ctx context.Context, id string) (*model.Invoice, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DefaultSettlementEngine) DepositToYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	// In a real implementation, this might involve checking balances,
	// preparing a transaction, and broadcasting it.
	return s.yieldProvider.DepositToYield(ctx, amount, strategy)
}

func (s *DefaultSettlementEngine) WithdrawFromYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	return s.yieldProvider.WithdrawFromYield(ctx, amount, strategy)
}

// MarkAsSettled is an internal helper to transition status and trigger yield logic.
func (s *DefaultSettlementEngine) MarkAsSettled(ctx context.Context, id string) error {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if invoice.Status == model.StatusSettled {
		return nil
	}

	if err := s.repo.UpdateStatus(ctx, id, model.StatusSettled); err != nil {
		return err
	}

	// Publish event to the bus
	if s.bus != nil {
		s.bus.Publish(EventSettlementConfirmed, invoice)
	}

	return nil
}

// Ensure implementation of SettlementEngine.
var _ model.SettlementEngine = (*DefaultSettlementEngine)(nil)
