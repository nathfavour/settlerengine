package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type Pruner struct {
	store   ports.DBStore
	dataDir string
}

func NewPruner(store ports.DBStore, dataDir string) *Pruner {
	return &Pruner{
		store:   store,
		dataDir: dataDir,
	}
}

func (p *Pruner) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial check on startup
	p.RunPrune(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.RunPrune(ctx)
		}
	}
}

func (p *Pruner) RunPrune(ctx context.Context) {
	fmt.Println("🧹 Pruner: Starting scheduled maintenance and archiving loop...")

	// 1. Expired Row Purging
	deleted, err := p.store.DeleteExpired(ctx, 48*time.Hour)
	if err != nil {
		fmt.Printf("⚠️ Pruner: Failed to purge expired invoices: %v\n", err)
	} else {
		fmt.Printf("🧹 Pruner: Hard deleted %d expired invoices older than 48 hours.\n", deleted)
	}

	// 2. Incremental OS Page Release
	if deleted > 0 {
		if err := p.store.Vacuum(ctx); err != nil {
			fmt.Printf("⚠️ Pruner: Failed to vacuum database pages: %v\n", err)
		} else {
			fmt.Println("🧹 Pruner: Incremental vacuum complete, released pages to OS.")
		}
	}

	// 3. Flat-File Ledger Archiving
	threshold := time.Now().AddDate(-1, 0, 0) // 1 year threshold
	oldInvoices, err := p.store.GetSettledBefore(ctx, threshold)
	if err != nil {
		fmt.Printf("⚠️ Pruner: Failed to query settled invoices for archiving: %v\n", err)
		return
	}

	if len(oldInvoices) == 0 {
		fmt.Println("🧹 Pruner: No transaction logs exceeding 365 days found for archiving.")
		return
	}

	fmt.Printf("🧹 Pruner: Found %d settled invoices older than 365 days to archive.\n", len(oldInvoices))

	// Group by Year
	byYear := make(map[int][]*domain.Invoice)
	for _, inv := range oldInvoices {
		year := inv.CreatedAt.Year()
		byYear[year] = append(byYear[year], inv)
	}

	for year, invoices := range byYear {
		if err := p.archiveYear(year, invoices); err != nil {
			fmt.Printf("⚠️ Pruner: Failed to archive year %d: %v\n", year, err)
			continue
		}

		// Delete from active DB
		var ids []string
		for _, inv := range invoices {
			ids = append(ids, inv.ID)
		}
		if err := p.store.DeleteInvoices(ctx, ids); err != nil {
			fmt.Printf("⚠️ Pruner: Failed to delete archived invoices from DB: %v\n", err)
		} else {
			fmt.Printf("🧹 Pruner: Successfully archived and dropped %d transactions from %d.\n", len(invoices), year)
		}
	}
}

type ArchiveRecord struct {
	ID        string `json:"id"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	CreatedAt int64  `json:"created_at"`
}

func (p *Pruner) archiveYear(year int, invoices []*domain.Invoice) error {
	archivePath := filepath.Join(p.dataDir, fmt.Sprintf("archive_%d.json", year))
	
	var existing []*ArchiveRecord
	if data, err := os.ReadFile(archivePath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	for _, inv := range invoices {
		existing = append(existing, &ArchiveRecord{
			ID:        inv.ID,
			Amount:    inv.Amount.Amount().String(),
			Currency:  inv.Amount.Currency(),
			CreatedAt: inv.CreatedAt.Unix(),
		})
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(archivePath, data, 0644)
}
