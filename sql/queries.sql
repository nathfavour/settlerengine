-- name: SaveInvoice :exec
INSERT INTO invoices (id, amount, currency, status, created_at, expires_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: FindInvoiceByID :one
SELECT id, amount, currency, status, created_at, expires_at
FROM invoices
WHERE id = ?;

-- name: UpdateInvoiceStatus :exec
UPDATE invoices
SET status = ?
WHERE id = ?;

-- name: DeleteExpiredInvoices :execresult
DELETE FROM invoices
WHERE status = 'EXPIRED' AND created_at < ?;

-- name: FindSettledInvoicesBefore :many
SELECT id, amount, currency, status, created_at, expires_at
FROM invoices
WHERE status = 'SETTLED' AND created_at < ?;

-- name: DeleteInvoicesByIDs :exec
DELETE FROM invoices
WHERE id = ?;

-- name: RecordVerifiedPayment :exec
INSERT INTO verified_payments (signature, signer, amount, asset, nonce, verified_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: CheckVerifiedPayment :one
SELECT signer
FROM verified_payments
WHERE signature = ?;
