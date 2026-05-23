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

-- name: SaveWebhookConfig :exec
INSERT INTO webhook_configs (id, url, secret, events, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetWebhookConfigs :many
SELECT id, url, secret, events, created_at
FROM webhook_configs;

-- name: SaveWebhookDelivery :exec
INSERT INTO webhook_deliveries (id, config_id, payload, event, status, attempts, next_attempt_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPendingWebhookDeliveries :many
SELECT id, config_id, payload, event, status, attempts, next_attempt_at, created_at
FROM webhook_deliveries
WHERE status = 'PENDING' AND next_attempt_at < ?;

-- name: UpdateWebhookDelivery :exec
UPDATE webhook_deliveries
SET status = ?, attempts = ?, next_attempt_at = ?
WHERE id = ?;

-- name: SaveClientReputation :exec
INSERT OR REPLACE INTO client_reputations (client_address, score, total_payments, last_payment_at)
VALUES (?, ?, ?, ?);

-- name: GetClientReputation :one
SELECT client_address, score, total_payments, last_payment_at
FROM client_reputations
WHERE client_address = ?;

-- name: SavePricingPolicy :exec
INSERT OR REPLACE INTO pricing_policies (resource_path, base_price, currency, surge_multiplier)
VALUES (?, ?, ?, ?);

-- name: GetPricingPolicy :one
SELECT resource_path, base_price, currency, surge_multiplier
FROM pricing_policies
WHERE resource_path = ?;

-- name: SaveEscrow :exec
INSERT INTO escrows (id, invoice_id, amount, currency, status, delivery_hash, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetEscrow :one
SELECT id, invoice_id, amount, currency, status, delivery_hash, created_at
FROM escrows
WHERE id = ?;

-- name: UpdateEscrowStatus :exec
UPDATE escrows
SET status = ?
WHERE id = ?;

-- name: SaveLsatChallenge :exec
INSERT INTO lsat_challenges (macaroon_id, preimage_hash, preimage, resource_path, amount, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetLsatChallenge :one
SELECT macaroon_id, preimage_hash, preimage, resource_path, amount, created_at
FROM lsat_challenges
WHERE macaroon_id = ?;

-- name: UpdateLsatPreimage :exec
UPDATE lsat_challenges
SET preimage = ?
WHERE macaroon_id = ?;
