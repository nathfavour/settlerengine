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
	store    ports.DBStore
	watcher  ports.BlockchainWatcher
	registry ports.AgentRegistry
}

func NewPaymentEngine(store ports.DBStore, watcher ports.BlockchainWatcher, registry ports.AgentRegistry) *PaymentEngine {
	return &PaymentEngine{
		store:    store,
		watcher:  watcher,
		registry: registry,
	}
}

// CloseSettlementLoop handles post-settlement actions like ERC-8004 reputation updates.
func (e *PaymentEngine) CloseSettlementLoop(ctx context.Context, invoiceID string, success bool) error {
	invoice, err := e.store.FindByID(ctx, invoiceID)
	if err != nil || invoice == nil {
		return err
	}

	if success && e.registry != nil {
		// In a real scenario, we would retrieve the AgentID associated with the invoice.
		// For now, we'll simulate posting a positive feedback if it was an agentic payment.
		fmt.Printf("🏆 ERC-8004: Closing settlement loop for invoice %s. Posting reputation update...\n", invoiceID)
		
		// Placeholder: Assume agent ID 42
		agentID := big.NewInt(42)
		score := big.NewInt(1) // +1 point for successful settlement
		_ = e.registry.PostFeedback(ctx, agentID, score, []string{"payment", "settled"}, "")
	}

	return nil
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
		_ = e.CloseSettlementLoop(ctx, id, true)
		return true, txHash, nil
	}

	return false, "", nil
}
