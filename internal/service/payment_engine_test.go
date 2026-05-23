package service

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type MockWatcher struct{}

func (w *MockWatcher) GenerateAddress(ctx context.Context, network ports.ChainNetwork, invoiceID string) (string, error) {
	return "mock_addr", nil
}

func (w *MockWatcher) VerifyPayment(ctx context.Context, network ports.ChainNetwork, address string, amount domain.Money) (bool, string, error) {
	return true, "tx_hash_123", nil
}

func (w *MockWatcher) StartWatching(ctx context.Context, handler func(signal ports.InvoicePaymentSignal)) error {
	return nil
}

func TestPaymentEngineFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "settler_test_*")
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

	watcher := &MockWatcher{}
	engine := NewPaymentEngine(store, watcher)

	ctx := context.Background()
	amount := domain.NewMoney(big.NewInt(1000000), "USDC")

	invoice, err := engine.CreateInvoice(ctx, amount, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}

	if invoice.Status != domain.StatusNew {
		t.Errorf("expected status NEW, got %s", invoice.Status)
	}

	// Verify payment
	ok, txHash, err := engine.VerifyInvoicePayment(ctx, invoice.ID, ports.NetworkSolana, "mock_addr")
	if err != nil {
		t.Fatalf("payment verification failed: %v", err)
	}

	if !ok || txHash != "tx_hash_123" {
		t.Errorf("expected success and tx_hash_123, got %t and %s", ok, txHash)
	}

	// Retrieve invoice and check status
	retrieved, err := engine.GetInvoice(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("failed to retrieve invoice: %v", err)
	}

	if retrieved.Status != domain.StatusSettled {
		t.Errorf("expected status SETTLED, got %s", retrieved.Status)
	}
}
