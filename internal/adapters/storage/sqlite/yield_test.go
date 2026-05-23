package sqlite

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/internal/domain"
)

func TestCompleteMigrationAndTableVerification(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migration_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "migrate.db")
	repo, err := NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// 1. Verify we can save an Invoice (Invoices table check)
	amount := domain.NewMoney(big.NewInt(100), "USDC")
	invoice := domain.NewInvoice("inv-migrate-test", amount, 1*time.Hour)
	if err := repo.Save(ctx, invoice); err != nil {
		t.Errorf("Failed to save invoice: %v (Is invoices table missing?)", err)
	}

	// 2. Verify Webhooks Config table
	config := &domain.WebhookConfig{
		ID:        "cfg-test",
		Url:       "http://localhost:8080/callback",
		Secret:    "shhh",
		Events:    "payment.confirmed",
		CreatedAt: time.Now(),
	}
	if err := repo.SaveConfig(ctx, config); err != nil {
		t.Errorf("Failed to save webhook config: %v (Is webhook_configs table missing?)", err)
	}

	// 3. Verify Escrows table
	escrow := &domain.Escrow{
		ID:           "escrow-test",
		InvoiceID:    "inv-migrate-test",
		Amount:       amount,
		Status:       "LOCKED",
		DeliveryHash: "hash-proof-123",
		CreatedAt:    time.Now(),
	}
	if err := repo.SaveEscrow(ctx, escrow); err != nil {
		t.Errorf("Failed to save escrow: %v (Is escrows table missing?)", err)
	}

	// 4. Verify LSAT Challenges table
	lsat := &domain.LsatChallenge{
		MacaroonID:   "mac-123",
		PreimageHash: "hash-pre",
		Preimage:     "",
		ResourcePath: "/metrics",
		Amount:       1000,
		CreatedAt:    time.Now(),
	}
	if err := repo.SaveLsatChallenge(ctx, lsat); err != nil {
		t.Errorf("Failed to save lsat challenge: %v (Is lsat_challenges table missing?)", err)
	}

	// 5. Verify Client Reputations table
	rep := &domain.ClientReputation{
		ClientAddress: "0xAddress",
		Score:         85,
		TotalPayments: amount,
		LastPaymentAt: time.Now(),
	}
	if err := repo.SaveReputation(ctx, rep); err != nil {
		t.Errorf("Failed to save client reputation: %v (Is client_reputations table missing?)", err)
	}

	// 6. Verify Pricing Policies table
	policy := &domain.PricingPolicy{
		ResourcePath:    "/proxy/endpoint",
		BasePrice:       amount,
		SurgeMultiplier: 1.5,
	}
	if err := repo.SavePolicy(ctx, policy); err != nil {
		t.Errorf("Failed to save pricing policy: %v (Is pricing_policies table missing?)", err)
	}
}

func TestYieldStrategyAndHarvestPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yield_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "yield.db")
	repo, err := NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// 1. Assert initial strategies query is empty
	strategies, err := repo.GetYieldStrategies(ctx)
	if err != nil {
		t.Fatalf("GetYieldStrategies failed: %v", err)
	}
	if len(strategies) != 0 {
		t.Errorf("Expected 0 strategies initially, got: %d", len(strategies))
	}

	// 2. Save new Yield Strategy
	strategyID := "riquid_bnb_vault"
	initialTVL := domain.NewMoney(big.NewInt(5000000000), "BNB") // 5 BNB in gwei
	strategy := &domain.YieldStrategy{
		ID:           strategyID,
		Provider:     "Riquid",
		VaultAddress: "0x1234567890123456789012345678901234567890",
		Asset:        "BNB",
		TVL:          initialTVL,
		APY:          12.85,
		LastHarvest:  time.Now().Round(time.Second),
		Status:       "ACTIVE",
	}

	if err := repo.SaveYieldStrategy(ctx, strategy); err != nil {
		t.Fatalf("SaveYieldStrategy failed: %v", err)
	}

	// 3. Query single strategy and assert equality
	stored, err := repo.GetYieldStrategy(ctx, strategyID)
	if err != nil {
		t.Fatalf("GetYieldStrategy failed: %v", err)
	}
	if stored == nil {
		t.Fatal("Expected stored strategy, got nil")
	}

	if stored.ID != strategy.ID || stored.Provider != strategy.Provider || stored.VaultAddress != strategy.VaultAddress {
		t.Errorf("Strategy mismatch: %+v vs %+v", stored, strategy)
	}
	if stored.TVL.Amount().Cmp(strategy.TVL.Amount()) != 0 || stored.TVL.Currency() != strategy.TVL.Currency() {
		t.Errorf("Strategy TVL mismatch: %s vs %s", stored.TVL.Amount(), strategy.TVL.Amount())
	}
	if stored.APY != strategy.APY || stored.Status != strategy.Status {
		t.Errorf("Strategy APY/Status mismatch")
	}

	// 4. Update Strategy TVL & APY and verify updates persist
	strategy.TVL = domain.NewMoney(big.NewInt(6500000000), "BNB")
	strategy.APY = 14.20
	strategy.LastHarvest = time.Now().Add(1 * time.Hour).Round(time.Second)

	if err := repo.SaveYieldStrategy(ctx, strategy); err != nil {
		t.Fatalf("Failed to update yield strategy: %v", err)
	}

	updated, err := repo.GetYieldStrategy(ctx, strategyID)
	if err != nil {
		t.Fatalf("Failed to query updated strategy: %v", err)
	}
	if updated.TVL.Amount().Cmp(strategy.TVL.Amount()) != 0 {
		t.Errorf("Expected updated TVL of %s, got: %s", strategy.TVL.Amount(), updated.TVL.Amount())
	}
	if updated.APY != 14.20 {
		t.Errorf("Expected updated APY of 14.20, got: %f", updated.APY)
	}

	// 5. Record Yield Harvest Logs
	harvest := &domain.YieldHarvest{
		ID:          "harvest-001",
		StrategyID:  strategyID,
		Amount:      domain.NewMoney(big.NewInt(150000000), "BNB"), // 0.15 BNB
		TxHash:      "0xhashbnbharvestabc",
		Status:      "SUCCESS",
		HarvestedAt: time.Now().Round(time.Second),
	}

	if err := repo.RecordYieldHarvest(ctx, harvest); err != nil {
		t.Fatalf("RecordYieldHarvest failed: %v", err)
	}

	// 6. Query harvests and verify results
	harvests, err := repo.GetYieldHarvests(ctx, strategyID)
	if err != nil {
		t.Fatalf("GetYieldHarvests failed: %v", err)
	}
	if len(harvests) != 1 {
		t.Fatalf("Expected 1 harvest record, got: %d", len(harvests))
	}

	h := harvests[0]
	if h.ID != harvest.ID || h.StrategyID != harvest.StrategyID || h.TxHash != harvest.TxHash || h.Status != harvest.Status {
		t.Errorf("Harvest log fields mismatch")
	}
	if h.Amount.Amount().Cmp(harvest.Amount.Amount()) != 0 {
		t.Errorf("Harvest log amount mismatch: %s vs %s", h.Amount.Amount(), harvest.Amount.Amount())
	}
}
