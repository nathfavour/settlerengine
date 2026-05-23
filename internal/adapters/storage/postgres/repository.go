package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/db"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type PostgresRepository struct {
	database *sql.DB
	queries  *db.Queries
}

func NewRepository(databaseURL string) (*PostgresRepository, error) {
	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}

	repo := &PostgresRepository{
		database: conn,
		queries:  db.New(conn),
	}

	if err := repo.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepository) migrate() error {
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

	CREATE TABLE IF NOT EXISTS webhook_configs (
		id VARCHAR(255) PRIMARY KEY,
		url VARCHAR(512) NOT NULL,
		secret VARCHAR(255) NOT NULL,
		events VARCHAR(255) NOT NULL,
		created_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS webhook_deliveries (
		id VARCHAR(255) PRIMARY KEY,
		config_id VARCHAR(255) NOT NULL,
		payload TEXT NOT NULL,
		event VARCHAR(100) NOT NULL,
		status VARCHAR(50) NOT NULL,
		attempts INT NOT NULL,
		next_attempt_at BIGINT NOT NULL,
		created_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS client_reputations (
		client_address VARCHAR(255) PRIMARY KEY,
		score INT NOT NULL,
		total_payments VARCHAR(255) NOT NULL,
		last_payment_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS pricing_policies (
		resource_path VARCHAR(255) PRIMARY KEY,
		base_price VARCHAR(255) NOT NULL,
		currency VARCHAR(50) NOT NULL,
		surge_multiplier DOUBLE PRECISION NOT NULL
	);

	CREATE TABLE IF NOT EXISTS escrows (
		id VARCHAR(255) PRIMARY KEY,
		invoice_id VARCHAR(255) NOT NULL,
		amount VARCHAR(255) NOT NULL,
		currency VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		delivery_hash VARCHAR(255) NOT NULL,
		created_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS lsat_challenges (
		macaroon_id VARCHAR(255) PRIMARY KEY,
		preimage_hash VARCHAR(255) NOT NULL,
		preimage VARCHAR(255) NOT NULL,
		resource_path VARCHAR(255) NOT NULL,
		amount BIGINT NOT NULL,
		created_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS yield_strategies (
		id VARCHAR(255) PRIMARY KEY,
		provider VARCHAR(255) NOT NULL,
		vault_address VARCHAR(255) NOT NULL,
		asset VARCHAR(50) NOT NULL,
		tvl VARCHAR(255) NOT NULL,
		apy DOUBLE PRECISION NOT NULL,
		last_harvest_at BIGINT NOT NULL,
		status VARCHAR(50) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS yield_harvests (
		id VARCHAR(255) PRIMARY KEY,
		strategy_id VARCHAR(255) NOT NULL,
		amount VARCHAR(255) NOT NULL,
		asset VARCHAR(50) NOT NULL,
		tx_hash VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		harvested_at BIGINT NOT NULL
	);
	`
	_, err := r.database.Exec(schema)
	return err
}

func (r *PostgresRepository) Close() error {
	return r.database.Close()
}

func (r *PostgresRepository) Vacuum(ctx context.Context) error {
	_, err := r.database.ExecContext(ctx, "VACUUM invoices;")
	return err
}

func (r *PostgresRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
	return r.queries.SaveInvoice(ctx, db.SaveInvoiceParams{
		ID:        invoice.ID,
		Amount:    invoice.Amount.Amount().String(),
		Currency:  invoice.Amount.Currency(),
		Status:    string(invoice.Status),
		CreatedAt: invoice.CreatedAt.Unix(),
		ExpiresAt: invoice.ExpiresAt.Unix(),
	})
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*domain.Invoice, error) {
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

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.InvoiceStatus) error {
	return r.queries.UpdateInvoiceStatus(ctx, db.UpdateInvoiceStatusParams{
		Status: string(status),
		ID:     id,
	})
}

func (r *PostgresRepository) DeleteExpired(ctx context.Context, expiryThreshold time.Duration) (int64, error) {
	limitTime := time.Now().Add(-expiryThreshold).Unix()
	return r.queries.DeleteExpiredInvoices(ctx, limitTime)
}

func (r *PostgresRepository) GetSettledBefore(ctx context.Context, threshold time.Time) ([]*domain.Invoice, error) {
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

func (r *PostgresRepository) DeleteInvoices(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := r.queries.DeleteInvoicesByIDs(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) RecordPayment(ctx context.Context, payment *domain.VerifiedPayment) error {
	return r.queries.RecordVerifiedPayment(ctx, db.RecordVerifiedPaymentParams{
		Signature:  payment.Signature,
		Signer:     payment.Signer,
		Amount:     payment.Amount.Amount().String(),
		Asset:      payment.Amount.Currency(),
		Nonce:      payment.Nonce,
		VerifiedAt: payment.VerifiedAt.Unix(),
	})
}

func (r *PostgresRepository) CheckPayment(ctx context.Context, signature string) (string, error) {
	return r.queries.CheckVerifiedPayment(ctx, signature)
}

func (r *PostgresRepository) SaveConfig(ctx context.Context, config *domain.WebhookConfig) error {
	return r.queries.SaveWebhookConfig(ctx, db.SaveWebhookConfigParams{
		ID:        config.ID,
		Url:       config.Url,
		Secret:    config.Secret,
		Events:    config.Events,
		CreatedAt: config.CreatedAt.Unix(),
	})
}

func (r *PostgresRepository) GetConfigs(ctx context.Context) ([]*domain.WebhookConfig, error) {
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

func (r *PostgresRepository) SaveDelivery(ctx context.Context, delivery *domain.WebhookDelivery) error {
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

func (r *PostgresRepository) GetPendingDeliveries(ctx context.Context) ([]*domain.WebhookDelivery, error) {
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

func (r *PostgresRepository) UpdateDeliveryStatus(ctx context.Context, id string, status string, attempts int32, nextAttemptAt time.Time) error {
	return r.queries.UpdateWebhookDelivery(ctx, db.UpdateWebhookDeliveryParams{
		Status:        status,
		Attempts:      attempts,
		NextAttemptAt: nextAttemptAt.Unix(),
		ID:            id,
	})
}

func (r *PostgresRepository) SaveReputation(ctx context.Context, rep *domain.ClientReputation) error {
	return r.queries.SaveClientReputation(ctx, db.SaveClientReputationParams{
		ClientAddress: rep.ClientAddress,
		Score:         rep.Score,
		TotalPayments: rep.TotalPayments.Amount().String(),
		LastPaymentAt: rep.LastPaymentAt.Unix(),
	})
}

func (r *PostgresRepository) GetReputation(ctx context.Context, clientAddress string) (*domain.ClientReputation, error) {
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

func (r *PostgresRepository) SavePolicy(ctx context.Context, policy *domain.PricingPolicy) error {
	return r.queries.SavePricingPolicy(ctx, db.SavePricingPolicyParams{
		ResourcePath:    policy.ResourcePath,
		BasePrice:       policy.BasePrice.Amount().String(),
		Currency:        policy.BasePrice.Currency(),
		SurgeMultiplier: policy.SurgeMultiplier,
	})
}

func (r *PostgresRepository) GetPolicy(ctx context.Context, resourcePath string) (*domain.PricingPolicy, error) {
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

func (r *PostgresRepository) SaveEscrow(ctx context.Context, escrow *domain.Escrow) error {
	return r.queries.SaveEscrow(ctx, db.SaveEscrowParams{
		ID:           escrow.ID,
		InvoiceID:    escrow.InvoiceID,
		Amount:       escrow.Amount.Amount().String(),
		Currency:     escrow.Amount.Currency(),
		Status:       escrow.Status,
		DeliveryHash: escrow.DeliveryHash,
		CreatedAt:    escrow.CreatedAt.Unix(),
	})
}

func (r *PostgresRepository) GetEscrow(ctx context.Context, id string) (*domain.Escrow, error) {
	row, err := r.queries.GetEscrow(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amountBig := new(big.Int)
	amountBig.SetString(row.Amount, 10)

	return &domain.Escrow{
		ID:           row.ID,
		InvoiceID:    row.InvoiceID,
		Amount:       domain.NewMoney(amountBig, row.Currency),
		Status:       row.Status,
		DeliveryHash: row.DeliveryHash,
		CreatedAt:    time.Unix(row.CreatedAt, 0),
	}, nil
}

func (r *PostgresRepository) UpdateEscrow(ctx context.Context, id string, status string) error {
	return r.queries.UpdateEscrowStatus(ctx, db.UpdateEscrowStatusParams{
		Status: status,
		ID:     id,
	})
}

func (r *PostgresRepository) SaveLsatChallenge(ctx context.Context, challenge *domain.LsatChallenge) error {
	return r.queries.SaveLsatChallenge(ctx, db.SaveLsatChallengeParams{
		MacaroonID:   challenge.MacaroonID,
		PreimageHash: challenge.PreimageHash,
		Preimage:     challenge.Preimage,
		ResourcePath: challenge.ResourcePath,
		Amount:       challenge.Amount,
		CreatedAt:    challenge.CreatedAt.Unix(),
	})
}

func (r *PostgresRepository) GetLsatChallenge(ctx context.Context, macaroonID string) (*domain.LsatChallenge, error) {
	row, err := r.queries.GetLsatChallenge(ctx, macaroonID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.LsatChallenge{
		MacaroonID:   row.MacaroonID,
		PreimageHash: row.PreimageHash,
		Preimage:     row.Preimage,
		ResourcePath: row.ResourcePath,
		Amount:       row.Amount,
		CreatedAt:    time.Unix(row.CreatedAt, 0),
	}, nil
}

func (r *PostgresRepository) UpdateLsatChallengePreimage(ctx context.Context, macaroonID string, preimage string) error {
	return r.queries.UpdateLsatPreimage(ctx, db.UpdateLsatPreimageParams{
		Preimage:   preimage,
		MacaroonID: macaroonID,
	})
}

func (r *PostgresRepository) SaveYieldStrategy(ctx context.Context, strategy *domain.YieldStrategy) error {
	return r.queries.SaveYieldStrategy(ctx, db.SaveYieldStrategyParams{
		ID:            strategy.ID,
		Provider:      strategy.Provider,
		VaultAddress:  strategy.VaultAddress,
		Asset:         strategy.Asset,
		Tvl:           strategy.TVL.Amount().String(),
		Apy:           strategy.APY,
		LastHarvestAt: strategy.LastHarvest.Unix(),
		Status:        strategy.Status,
	})
}

func (r *PostgresRepository) GetYieldStrategy(ctx context.Context, id string) (*domain.YieldStrategy, error) {
	row, err := r.queries.GetYieldStrategy(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amountBig := new(big.Int)
	amountBig.SetString(row.Tvl, 10)

	return &domain.YieldStrategy{
		ID:           row.ID,
		Provider:     row.Provider,
		VaultAddress: row.VaultAddress,
		Asset:        row.Asset,
		TVL:          domain.NewMoney(amountBig, row.Asset),
		APY:          row.Apy,
		LastHarvest:  time.Unix(row.LastHarvestAt, 0),
		Status:       row.Status,
	}, nil
}

func (r *PostgresRepository) GetYieldStrategies(ctx context.Context) ([]*domain.YieldStrategy, error) {
	rows, err := r.queries.GetYieldStrategies(ctx)
	if err != nil {
		return nil, err
	}

	strategies := make([]*domain.YieldStrategy, len(rows))
	for i, row := range rows {
		amountBig := new(big.Int)
		amountBig.SetString(row.Tvl, 10)

		strategies[i] = &domain.YieldStrategy{
			ID:           row.ID,
			Provider:     row.Provider,
			VaultAddress: row.VaultAddress,
			Asset:        row.Asset,
			TVL:          domain.NewMoney(amountBig, row.Asset),
			APY:          row.Apy,
			LastHarvest:  time.Unix(row.LastHarvestAt, 0),
			Status:       row.Status,
		}
	}
	return strategies, nil
}

func (r *PostgresRepository) RecordYieldHarvest(ctx context.Context, harvest *domain.YieldHarvest) error {
	return r.queries.RecordYieldHarvest(ctx, db.RecordYieldHarvestParams{
		ID:          harvest.ID,
		StrategyID:  harvest.StrategyID,
		Amount:      harvest.Amount.Amount().String(),
		Asset:       harvest.Amount.Currency(),
		TxHash:      harvest.TxHash,
		Status:      harvest.Status,
		HarvestedAt: harvest.HarvestedAt.Unix(),
	})
}

func (r *PostgresRepository) GetYieldHarvests(ctx context.Context, strategyID string) ([]*domain.YieldHarvest, error) {
	rows, err := r.queries.GetYieldHarvests(ctx, strategyID)
	if err != nil {
		return nil, err
	}

	harvests := make([]*domain.YieldHarvest, len(rows))
	for i, row := range rows {
		amountBig := new(big.Int)
		amountBig.SetString(row.Amount, 10)

		harvests[i] = &domain.YieldHarvest{
			ID:          row.ID,
			StrategyID:  row.StrategyID,
			Amount:      domain.NewMoney(amountBig, row.Asset),
			TxHash:      row.TxHash,
			Status:      row.Status,
			HarvestedAt: time.Unix(row.HarvestedAt, 0),
		}
	}
	return harvests, nil
}

var _ ports.DBStore = (*PostgresRepository)(nil)

