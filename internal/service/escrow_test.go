package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
	"github.com/nathfavour/settlerengine/internal/domain"
)

func TestEscrowEngineFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "escrow_test_*")
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

	engine := NewEscrowEngine(store)
	ctx := context.Background()

	amount := domain.NewMoney(big.NewInt(500000), "USDC")
	invoice := domain.NewInvoice("invoice-escrow-1", amount, 1*time.Hour)
	invoice.Status = domain.StatusNew
	if err := store.Save(ctx, invoice); err != nil {
		t.Fatalf("failed to save invoice: %v", err)
	}

	preimage := "my_delivered_digital_file_data"
	hasher := sha256.New()
	hasher.Write([]byte(preimage))
	deliveryHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Lock funds in Escrow
	escrow, err := engine.LockFunds(ctx, invoice.ID, amount, deliveryHash)
	if err != nil {
		t.Fatalf("failed to lock funds: %v", err)
	}

	if escrow.Status != "LOCKED" {
		t.Errorf("expected escrow status LOCKED, got %s", escrow.Status)
	}

	// Try to release with invalid preimage proof
	_, err = engine.VerifyAndRelease(ctx, escrow.ID, "wrong_file_data")
	if err == nil {
		t.Errorf("expected error with invalid proof, but got none")
	}

	// Release with correct preimage proof
	success, err := engine.VerifyAndRelease(ctx, escrow.ID, preimage)
	if err != nil {
		t.Fatalf("release failed: %v", err)
	}

	if !success {
		t.Errorf("expected release to be successful")
	}

	// Check DB records status
	retrievedEscrow, err := store.GetEscrow(ctx, escrow.ID)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if retrievedEscrow.Status != "RELEASED" {
		t.Errorf("expected escrow status RELEASED, got %s", retrievedEscrow.Status)
	}

	retrievedInvoice, err := store.FindByID(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if retrievedInvoice.Status != domain.StatusSettled {
		t.Errorf("expected invoice status SETTLED, got %s", retrievedInvoice.Status)
	}
}
