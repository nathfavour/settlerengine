package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type EscrowEngine struct {
	store ports.DBStore
}

func NewEscrowEngine(store ports.DBStore) *EscrowEngine {
	return &EscrowEngine{store: store}
}

// LockFunds initiates a locked SLA escrow for an invoice.
func (e *EscrowEngine) LockFunds(ctx context.Context, invoiceID string, amount domain.Money, deliveryHash string) (*domain.Escrow, error) {
	id := uuid.New().String()
	escrow := &domain.Escrow{
		ID:           id,
		InvoiceID:    invoiceID,
		Amount:       amount,
		Status:       "LOCKED",
		DeliveryHash: deliveryHash,
		CreatedAt:    time.Now(),
	}

	if err := e.store.SaveEscrow(ctx, escrow); err != nil {
		return nil, fmt.Errorf("failed to save escrow: %w", err)
	}

	return escrow, nil
}

// VerifyAndRelease unlocks the funds to the merchant upon proving resource delivery.
// Verification checks that sha256(preimageProof) matches the configured DeliveryHash.
func (e *EscrowEngine) VerifyAndRelease(ctx context.Context, escrowID string, preimageProof string) (bool, error) {
	escrow, err := e.store.GetEscrow(ctx, escrowID)
	if err != nil {
		return false, err
	}
	if escrow == nil {
		return false, fmt.Errorf("escrow record not found")
	}

	if escrow.Status != "LOCKED" {
		return false, fmt.Errorf("escrow is not locked, status: %s", escrow.Status)
	}

	// Hash the preimage proof
	hasher := sha256.New()
	hasher.Write([]byte(preimageProof))
	derivedHash := fmt.Sprintf("%x", hasher.Sum(nil))

	if derivedHash != escrow.DeliveryHash {
		return false, fmt.Errorf("cryptographic proof of delivery invalid")
	}

	// Update escrow status to RELEASED
	if err := e.store.UpdateEscrow(ctx, escrowID, "RELEASED"); err != nil {
		return false, err
	}

	// Mark invoice as SETTLED upon successful release
	if err := e.store.UpdateStatus(ctx, escrow.InvoiceID, domain.StatusSettled); err != nil {
		return false, err
	}

	return true, nil
}

// Dispute initiates a dispute state on the locked funds.
func (e *EscrowEngine) Dispute(ctx context.Context, escrowID string) error {
	escrow, err := e.store.GetEscrow(ctx, escrowID)
	if err != nil {
		return err
	}
	if escrow == nil {
		return fmt.Errorf("escrow record not found")
	}

	if escrow.Status != "LOCKED" {
		return fmt.Errorf("escrow is not locked, status: %s", escrow.Status)
	}

	return e.store.UpdateEscrow(ctx, escrowID, "DISPUTED")
}
