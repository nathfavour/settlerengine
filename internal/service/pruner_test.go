package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
	"github.com/nathfavour/settlerengine/internal/domain"
)

func TestPrunerMaintenanceAndArchiving(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pruner_test_*")
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

	pruner := NewPruner(store, tmpDir)

	ctx := context.Background()

	// 1. Create an expired invoice older than 48 hours
	expiredOldInvoice := &domain.Invoice{
		ID:        "expired-old",
		Amount:    domain.NewMoney(big.NewInt(500), "USDC"),
		Status:    domain.StatusExpired,
		CreatedAt: time.Now().Add(-50 * time.Hour),
		ExpiresAt: time.Now().Add(-49 * time.Hour),
	}
	if err := store.Save(ctx, expiredOldInvoice); err != nil {
		t.Fatalf("failed to save old expired invoice: %v", err)
	}

	// 2. Create an expired invoice NOT older than 48 hours
	expiredNewInvoice := &domain.Invoice{
		ID:        "expired-new",
		Amount:    domain.NewMoney(big.NewInt(500), "USDC"),
		Status:    domain.StatusExpired,
		CreatedAt: time.Now().Add(-20 * time.Hour),
		ExpiresAt: time.Now().Add(-19 * time.Hour),
	}
	if err := store.Save(ctx, expiredNewInvoice); err != nil {
		t.Fatalf("failed to save new expired invoice: %v", err)
	}

	// 3. Create a settled invoice older than 365 days (to be archived)
	settledOldInvoice := &domain.Invoice{
		ID:        "settled-old",
		Amount:    domain.NewMoney(big.NewInt(1000), "USDC"),
		Status:    domain.StatusSettled,
		CreatedAt: time.Now().AddDate(-2, 0, 0), // 2 years old
		ExpiresAt: time.Now().AddDate(-2, 0, 0).Add(1 * time.Hour),
	}
	if err := store.Save(ctx, settledOldInvoice); err != nil {
		t.Fatalf("failed to save settled old invoice: %v", err)
	}

	// Run Pruner Maintenance
	pruner.RunPrune(ctx)

	// Check if "expired-old" was deleted
	retrievedOld, err := store.FindByID(ctx, "expired-old")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if retrievedOld != nil {
		t.Errorf("expected expired-old to be deleted, but it exists")
	}

	// Check if "expired-new" was NOT deleted
	retrievedNew, err := store.FindByID(ctx, "expired-new")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if retrievedNew == nil {
		t.Errorf("expected expired-new to not be deleted, but it was")
	}

	// Check if "settled-old" was deleted from active DB after archiving
	retrievedSettled, err := store.FindByID(ctx, "settled-old")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if retrievedSettled != nil {
		t.Errorf("expected settled-old to be dropped from DB, but it exists")
	}

	// Verify the archive file was created and contains the record
	archiveYear := settledOldInvoice.CreatedAt.Year()
	archivePath := filepath.Join(tmpDir, fmt.Sprintf("archive_%d.json", archiveYear))
	
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatalf("expected archive file %s to exist, but it does not", archivePath)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("failed to read archive file: %v", err)
	}

	var records []*ArchiveRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("failed to unmarshal archive data: %v", err)
	}

	if len(records) != 1 || records[0].ID != "settled-old" {
		t.Errorf("expected archive to contain 'settled-old', got %+v", records)
	}
}
