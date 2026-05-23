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
