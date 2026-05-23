package db

import (
	"context"
	"database/sql"
)

const saveInvoice = `
INSERT INTO invoices (id, amount, currency, status, created_at, expires_at)
VALUES (?, ?, ?, ?, ?, ?);
`

type SaveInvoiceParams struct {
	ID        string
	Amount    string
	Currency  string
	Status    string
	CreatedAt int64
	ExpiresAt int64
}

func (q *Queries) SaveInvoice(ctx context.Context, arg SaveInvoiceParams) error {
	_, err := q.db.ExecContext(ctx, saveInvoice,
		arg.ID,
		arg.Amount,
		arg.Currency,
		arg.Status,
		arg.CreatedAt,
		arg.ExpiresAt,
	)
	return err
}

const findInvoiceByID = `
SELECT id, amount, currency, status, created_at, expires_at
FROM invoices
WHERE id = ?;
`

func (q *Queries) FindInvoiceByID(ctx context.Context, id string) (Invoice, error) {
	row := q.db.QueryRowContext(ctx, findInvoiceByID, id)
	var i Invoice
	err := row.Scan(
		&i.ID,
		&i.Amount,
		&i.Currency,
		&i.Status,
		&i.CreatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const updateInvoiceStatus = `
UPDATE invoices
SET status = ?
WHERE id = ?;
`

type UpdateInvoiceStatusParams struct {
	Status string
	ID     string
}

func (q *Queries) UpdateInvoiceStatus(ctx context.Context, arg UpdateInvoiceStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateInvoiceStatus, arg.Status, arg.ID)
	return err
}

const deleteExpiredInvoices = `
DELETE FROM invoices
WHERE status = 'EXPIRED' AND created_at < ?;
`

func (q *Queries) DeleteExpiredInvoices(ctx context.Context, createdAt int64) (int64, error) {
	res, err := q.db.ExecContext(ctx, deleteExpiredInvoices, createdAt)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

const findSettledInvoicesBefore = `
SELECT id, amount, currency, status, created_at, expires_at
FROM invoices
WHERE status = 'SETTLED' AND created_at < ?;
`

func (q *Queries) FindSettledInvoicesBefore(ctx context.Context, createdAt int64) ([]Invoice, error) {
	rows, err := q.db.QueryContext(ctx, findSettledInvoicesBefore, createdAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Invoice
	for rows.Next() {
		var i Invoice
		if err := rows.Scan(
			&i.ID,
			&i.Amount,
			&i.Currency,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteInvoicesByIDs = `
DELETE FROM invoices
WHERE id = ?;
`

func (q *Queries) DeleteInvoicesByIDs(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteInvoicesByIDs, id)
	return err
}

const recordVerifiedPayment = `
INSERT INTO verified_payments (signature, signer, amount, asset, nonce, verified_at)
VALUES (?, ?, ?, ?, ?, ?);
`

type RecordVerifiedPaymentParams struct {
	Signature  string
	Signer     string
	Amount     string
	Asset      string
	Nonce      string
	VerifiedAt int64
}

func (q *Queries) RecordVerifiedPayment(ctx context.Context, arg RecordVerifiedPaymentParams) error {
	_, err := q.db.ExecContext(ctx, recordVerifiedPayment,
		arg.Signature,
		arg.Signer,
		arg.Amount,
		arg.Asset,
		arg.Nonce,
		arg.VerifiedAt,
	)
	return err
}

const checkVerifiedPayment = `
SELECT signer
FROM verified_payments
WHERE signature = ?;
`

func (q *Queries) CheckVerifiedPayment(ctx context.Context, signature string) (string, error) {
	row := q.db.QueryRowContext(ctx, checkVerifiedPayment, signature)
	var signer string
	err := row.Scan(&signer)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return signer, err
}

const saveWebhookConfig = `
INSERT INTO webhook_configs (id, url, secret, events, created_at)
VALUES (?, ?, ?, ?, ?);
`

type SaveWebhookConfigParams struct {
	ID        string
	Url       string
	Secret    string
	Events    string
	CreatedAt int64
}

func (q *Queries) SaveWebhookConfig(ctx context.Context, arg SaveWebhookConfigParams) error {
	_, err := q.db.ExecContext(ctx, saveWebhookConfig,
		arg.ID,
		arg.Url,
		arg.Secret,
		arg.Events,
		arg.CreatedAt,
	)
	return err
}

const getWebhookConfigs = `
SELECT id, url, secret, events, created_at
FROM webhook_configs;
`

func (q *Queries) GetWebhookConfigs(ctx context.Context) ([]WebhookConfig, error) {
	rows, err := q.db.QueryContext(ctx, getWebhookConfigs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebhookConfig
	for rows.Next() {
		var i WebhookConfig
		if err := rows.Scan(
			&i.ID,
			&i.Url,
			&i.Secret,
			&i.Events,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const saveWebhookDelivery = `
INSERT INTO webhook_deliveries (id, config_id, payload, event, status, attempts, next_attempt_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
`

type SaveWebhookDeliveryParams struct {
	ID            string
	ConfigID      string
	Payload       string
	Event         string
	Status        string
	Attempts      int32
	NextAttemptAt int64
	CreatedAt     int64
}

func (q *Queries) SaveWebhookDelivery(ctx context.Context, arg SaveWebhookDeliveryParams) error {
	_, err := q.db.ExecContext(ctx, saveWebhookDelivery,
		arg.ID,
		arg.ConfigID,
		arg.Payload,
		arg.Event,
		arg.Status,
		arg.Attempts,
		arg.NextAttemptAt,
		arg.CreatedAt,
	)
	return err
}

const getPendingWebhookDeliveries = `
SELECT id, config_id, payload, event, status, attempts, next_attempt_at, created_at
FROM webhook_deliveries
WHERE status = 'PENDING' AND next_attempt_at < ?;
`

func (q *Queries) GetPendingWebhookDeliveries(ctx context.Context, nextAttemptAt int64) ([]WebhookDelivery, error) {
	rows, err := q.db.QueryContext(ctx, getPendingWebhookDeliveries, nextAttemptAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebhookDelivery
	for rows.Next() {
		var i WebhookDelivery
		if err := rows.Scan(
			&i.ID,
			&i.ConfigID,
			&i.Payload,
			&i.Event,
			&i.Status,
			&i.Attempts,
			&i.NextAttemptAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateWebhookDelivery = `
UPDATE webhook_deliveries
SET status = ?, attempts = ?, next_attempt_at = ?
WHERE id = ?;
`

type UpdateWebhookDeliveryParams struct {
	Status        string
	Attempts      int32
	NextAttemptAt int64
	ID            string
}

func (q *Queries) UpdateWebhookDelivery(ctx context.Context, arg UpdateWebhookDeliveryParams) error {
	_, err := q.db.ExecContext(ctx, updateWebhookDelivery,
		arg.Status,
		arg.Attempts,
		arg.NextAttemptAt,
		arg.ID,
	)
	return err
}

const saveClientReputation = `
INSERT OR REPLACE INTO client_reputations (client_address, score, total_payments, last_payment_at)
VALUES (?, ?, ?, ?);
`

type SaveClientReputationParams struct {
	ClientAddress string
	Score         int32
	TotalPayments string
	LastPaymentAt int64
}

func (q *Queries) SaveClientReputation(ctx context.Context, arg SaveClientReputationParams) error {
	_, err := q.db.ExecContext(ctx, saveClientReputation,
		arg.ClientAddress,
		arg.Score,
		arg.TotalPayments,
		arg.LastPaymentAt,
	)
	return err
}

const getClientReputation = `
SELECT client_address, score, total_payments, last_payment_at
FROM client_reputations
WHERE client_address = ?;
`

func (q *Queries) GetClientReputation(ctx context.Context, clientAddress string) (ClientReputation, error) {
	row := q.db.QueryRowContext(ctx, getClientReputation, clientAddress)
	var i ClientReputation
	err := row.Scan(
		&i.ClientAddress,
		&i.Score,
		&i.TotalPayments,
		&i.LastPaymentAt,
	)
	return i, err
}

const savePricingPolicy = `
INSERT OR REPLACE INTO pricing_policies (resource_path, base_price, currency, surge_multiplier)
VALUES (?, ?, ?, ?);
`

type SavePricingPolicyParams struct {
	ResourcePath    string
	BasePrice       string
	Currency        string
	SurgeMultiplier float64
}

func (q *Queries) SavePricingPolicy(ctx context.Context, arg SavePricingPolicyParams) error {
	_, err := q.db.ExecContext(ctx, savePricingPolicy,
		arg.ResourcePath,
		arg.BasePrice,
		arg.Currency,
		arg.SurgeMultiplier,
	)
	return err
}

const getPricingPolicy = `
SELECT resource_path, base_price, currency, surge_multiplier
FROM pricing_policies
WHERE resource_path = ?;
`

func (q *Queries) GetPricingPolicy(ctx context.Context, resourcePath string) (PricingPolicy, error) {
	row := q.db.QueryRowContext(ctx, getPricingPolicy, resourcePath)
	var i PricingPolicy
	err := row.Scan(
		&i.ResourcePath,
		&i.BasePrice,
		&i.Currency,
		&i.SurgeMultiplier,
	)
	return i, err
}

const saveEscrow = `
INSERT INTO escrows (id, invoice_id, amount, currency, status, delivery_hash, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
`

type SaveEscrowParams struct {
	ID           string
	InvoiceID    string
	Amount       string
	Currency     string
	Status       string
	DeliveryHash string
	CreatedAt    int64
}

func (q *Queries) SaveEscrow(ctx context.Context, arg SaveEscrowParams) error {
	_, err := q.db.ExecContext(ctx, saveEscrow,
		arg.ID,
		arg.InvoiceID,
		arg.Amount,
		arg.Currency,
		arg.Status,
		arg.DeliveryHash,
		arg.CreatedAt,
	)
	return err
}

const getEscrow = `
SELECT id, invoice_id, amount, currency, status, delivery_hash, created_at
FROM escrows
WHERE id = ?;
`

func (q *Queries) GetEscrow(ctx context.Context, id string) (Escrow, error) {
	row := q.db.QueryRowContext(ctx, getEscrow, id)
	var i Escrow
	err := row.Scan(
		&i.ID,
		&i.InvoiceID,
		&i.Amount,
		&i.Currency,
		&i.Status,
		&i.DeliveryHash,
		&i.CreatedAt,
	)
	return i, err
}

const updateEscrowStatus = `
UPDATE escrows
SET status = ?
WHERE id = ?;
`

type UpdateEscrowStatusParams struct {
	Status string
	ID     string
}

func (q *Queries) UpdateEscrowStatus(ctx context.Context, arg UpdateEscrowStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateEscrowStatus, arg.Status, arg.ID)
	return err
}
