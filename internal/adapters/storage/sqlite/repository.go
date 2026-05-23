package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/db"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type SqliteRepository struct {
	database *sql.DB
	queries  *db.Queries
}

func NewRepository(dbPath string) (*SqliteRepository, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Apply WAL and busy timeout configuration
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA auto_vacuum = INCREMENTAL;",
	}
	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	repo := &SqliteRepository{
		database: conn,
		queries:  db.New(conn),
	}

	if err := repo.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

func (r *SqliteRepository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS invoices (
		id VARCHAR(255) PRIMARY KEY,
		amount VARCHAR(255) NOT NULL,
		currency VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		created_at BIGINT NOT NULL,
		expires_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS verified_payments (
		signature VARCHAR(255) PRIMARY KEY,
		signer VARCHAR(255) NOT NULL,
		amount VARCHAR(255) NOT NULL,
		asset VARCHAR(255) NOT NULL,
		nonce VARCHAR(255) NOT NULL,
		verified_at BIGINT NOT NULL
	);
	`
	_, err := r.database.Exec(schema)
	return err
}

func (r *SqliteRepository) Close() error {
	return r.database.Close()
}

func (r *SqliteRepository) Vacuum(ctx context.Context) error {
	_, err := r.database.ExecContext(ctx, "PRAGMA incremental_vacuum(100);")
	return err
}

func (r *SqliteRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
	return r.queries.SaveInvoice(ctx, db.SaveInvoiceParams{
		ID:        invoice.ID,
		Amount:    invoice.Amount.Amount().String(),
		Currency:  invoice.Amount.Currency(),
		Status:    string(invoice.Status),
		CreatedAt: invoice.CreatedAt.Unix(),
		ExpiresAt: invoice.ExpiresAt.Unix(),
	})
}

func (r *SqliteRepository) FindByID(ctx context.Context, id string) (*domain.Invoice, error) {
	row, err := r.queries.FindInvoiceByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amountBig := new(big.Int)
	amountBig.SetString(row.Amount, 10)

	return &domain.Invoice{
		ID:        row.ID,
		Amount:    domain.NewMoney(amountBig, row.Currency),
		Status:    domain.InvoiceStatus(row.Status),
		CreatedAt: time.Unix(row.CreatedAt, 0),
		ExpiresAt: time.Unix(row.ExpiresAt, 0),
	}, nil
}

func (r *SqliteRepository) UpdateStatus(ctx context.Context, id string, status domain.InvoiceStatus) error {
	return r.queries.UpdateInvoiceStatus(ctx, db.UpdateInvoiceStatusParams{
		Status: string(status),
		ID:     id,
	})
}

func (r *SqliteRepository) DeleteExpired(ctx context.Context, expiryThreshold time.Duration) (int64, error) {
	limitTime := time.Now().Add(-expiryThreshold).Unix()
	return r.queries.DeleteExpiredInvoices(ctx, limitTime)
}

func (r *SqliteRepository) GetSettledBefore(ctx context.Context, threshold time.Time) ([]*domain.Invoice, error) {
	rows, err := r.queries.FindSettledInvoicesBefore(ctx, threshold.Unix())
	if err != nil {
		return nil, err
	}

	invoices := make([]*domain.Invoice, len(rows))
	for i, row := range rows {
		amountBig := new(big.Int)
		amountBig.SetString(row.Amount, 10)

		invoices[i] = &domain.Invoice{
			ID:        row.ID,
			Amount:    domain.NewMoney(amountBig, row.Currency),
			Status:    domain.InvoiceStatus(row.Status),
			CreatedAt: time.Unix(row.CreatedAt, 0),
			ExpiresAt: time.Unix(row.ExpiresAt, 0),
		}
	}
	return invoices, nil
}

func (r *SqliteRepository) DeleteInvoices(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := r.queries.DeleteInvoicesByIDs(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *SqliteRepository) RecordPayment(ctx context.Context, payment *domain.VerifiedPayment) error {
	return r.queries.RecordVerifiedPayment(ctx, db.RecordVerifiedPaymentParams{
		Signature:  payment.Signature,
		Signer:     payment.Signer,
		Amount:     payment.Amount.Amount().String(),
		Asset:      payment.Amount.Currency(),
		Nonce:      payment.Nonce,
		VerifiedAt: payment.VerifiedAt.Unix(),
	})
}

func (r *SqliteRepository) CheckPayment(ctx context.Context, signature string) (string, error) {
	return r.queries.CheckVerifiedPayment(ctx, signature)
}

func (r *SqliteRepository) SaveConfig(ctx context.Context, config *domain.WebhookConfig) error {
	return r.queries.SaveWebhookConfig(ctx, db.SaveWebhookConfigParams{
		ID:        config.ID,
		Url:       config.Url,
		Secret:    config.Secret,
		Events:    config.Events,
		CreatedAt: config.CreatedAt.Unix(),
	})
}

func (r *SqliteRepository) GetConfigs(ctx context.Context) ([]*domain.WebhookConfig, error) {
	rows, err := r.queries.GetWebhookConfigs(ctx)
	if err != nil {
		return nil, err
	}
	configs := make([]*domain.WebhookConfig, len(rows))
	for i, row := range rows {
		configs[i] = &domain.WebhookConfig{
			ID:        row.ID,
			Url:       row.Url,
			Secret:    row.Secret,
			Events:    row.Events,
			CreatedAt: time.Unix(row.CreatedAt, 0),
		}
	}
	return configs, nil
}

func (r *SqliteRepository) SaveDelivery(ctx context.Context, delivery *domain.WebhookDelivery) error {
	return r.queries.SaveWebhookDelivery(ctx, db.SaveWebhookDeliveryParams{
		ID:            delivery.ID,
		ConfigID:      delivery.ConfigID,
		Payload:       delivery.Payload,
		Event:         delivery.Event,
		Status:        delivery.Status,
		Attempts:      delivery.Attempts,
		NextAttemptAt: delivery.NextAttemptAt.Unix(),
		CreatedAt:     delivery.CreatedAt.Unix(),
	})
}

func (r *SqliteRepository) GetPendingDeliveries(ctx context.Context) ([]*domain.WebhookDelivery, error) {
	rows, err := r.queries.GetPendingWebhookDeliveries(ctx, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	deliveries := make([]*domain.WebhookDelivery, len(rows))
	for i, row := range rows {
		deliveries[i] = &domain.WebhookDelivery{
			ID:            row.ID,
			ConfigID:      row.ConfigID,
			Payload:       row.Payload,
			Event:         row.Event,
			Status:        row.Status,
			Attempts:      row.Attempts,
			NextAttemptAt: time.Unix(row.NextAttemptAt, 0),
			CreatedAt:     time.Unix(row.CreatedAt, 0),
		}
	}
	return deliveries, nil
}

func (r *SqliteRepository) UpdateDeliveryStatus(ctx context.Context, id string, status string, attempts int32, nextAttemptAt time.Time) error {
	return r.queries.UpdateWebhookDelivery(ctx, db.UpdateWebhookDeliveryParams{
		Status:        status,
		Attempts:      attempts,
		NextAttemptAt: nextAttemptAt.Unix(),
		ID:            id,
	})
}

func (r *SqliteRepository) SaveReputation(ctx context.Context, rep *domain.ClientReputation) error {
	return r.queries.SaveClientReputation(ctx, db.SaveClientReputationParams{
		ClientAddress: rep.ClientAddress,
		Score:         rep.Score,
		TotalPayments: rep.TotalPayments.Amount().String(),
		LastPaymentAt: rep.LastPaymentAt.Unix(),
	})
}

func (r *SqliteRepository) GetReputation(ctx context.Context, clientAddress string) (*domain.ClientReputation, error) {
	row, err := r.queries.GetClientReputation(ctx, clientAddress)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amountBig := new(big.Int)
	amountBig.SetString(row.TotalPayments, 10)

	return &domain.ClientReputation{
		ClientAddress: row.ClientAddress,
		Score:         row.Score,
		TotalPayments: domain.NewMoney(amountBig, "USDC"),
		LastPaymentAt: time.Unix(row.LastPaymentAt, 0),
	}, nil
}

func (r *SqliteRepository) SavePolicy(ctx context.Context, policy *domain.PricingPolicy) error {
	return r.queries.SavePricingPolicy(ctx, db.SavePricingPolicyParams{
		ResourcePath:    policy.ResourcePath,
		BasePrice:       policy.BasePrice.Amount().String(),
		Currency:        policy.BasePrice.Currency(),
		SurgeMultiplier: policy.SurgeMultiplier,
	})
}

func (r *SqliteRepository) GetPolicy(ctx context.Context, resourcePath string) (*domain.PricingPolicy, error) {
	row, err := r.queries.GetPricingPolicy(ctx, resourcePath)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amountBig := new(big.Int)
	amountBig.SetString(row.BasePrice, 10)

	return &domain.PricingPolicy{
		ResourcePath:    row.ResourcePath,
		BasePrice:       domain.NewMoney(amountBig, row.Currency),
		SurgeMultiplier: row.SurgeMultiplier,
	}, nil
}

var _ ports.DBStore = (*SqliteRepository)(nil)
