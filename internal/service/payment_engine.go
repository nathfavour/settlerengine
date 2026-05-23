package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type PaymentEngine struct {
	store   ports.DBStore
	watcher ports.BlockchainWatcher
}

func NewPaymentEngine(store ports.DBStore, watcher ports.BlockchainWatcher) *PaymentEngine {
	return &PaymentEngine{
		store:   store,
		watcher: watcher,
	}
}

func (e *PaymentEngine) CreateInvoice(ctx context.Context, amount domain.Money, duration time.Duration) (*domain.Invoice, error) {
	id := uuid.New().String()
	invoice := domain.NewInvoice(id, amount, duration)

	if err := e.store.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return invoice, nil
}

func (e *PaymentEngine) GetInvoice(ctx context.Context, id string) (*domain.Invoice, error) {
	return e.store.FindByID(ctx, id)
}

func (e *PaymentEngine) VerifyInvoicePayment(ctx context.Context, id string, network ports.ChainNetwork, address string) (bool, string, error) {
	invoice, err := e.store.FindByID(ctx, id)
	if err != nil {
		return false, "", err
	}
	if invoice == nil {
		return false, "", fmt.Errorf("invoice not found")
	}

	if invoice.Status == domain.StatusSettled {
		return true, "", nil
	}

	if time.Now().After(invoice.ExpiresAt) {
		_ = e.store.UpdateStatus(ctx, id, domain.StatusExpired)
		return false, "", nil
	}

	success, txHash, err := e.watcher.VerifyPayment(ctx, network, address, invoice.Amount)
	if err != nil {
		return false, "", err
	}

	if success {
		_ = e.store.UpdateStatus(ctx, id, domain.StatusSettled)
		return true, txHash, nil
	}

	return false, "", nil
}
