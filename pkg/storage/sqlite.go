package storage

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	DataDir string
}

func OpenDefault() (*DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir: %w", err)
	}

	dataDir := filepath.Join(configDir, "settlerengine")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "settler.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	s := &DB{DB: db, DataDir: dataDir}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	return s, nil
}

func (db *DB) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS verified_payments (
		signature TEXT PRIMARY KEY,
		signer TEXT NOT NULL,
		amount TEXT NOT NULL,
		asset TEXT NOT NULL,
		nonce TEXT NOT NULL,
		verified_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS invoices (
		id TEXT PRIMARY KEY,
		amount TEXT NOT NULL,
		currency TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// Save implements model.InvoiceRepository.
func (db *DB) Save(ctx context.Context, inv *model.Invoice) error {
	query := `INSERT INTO invoices (id, amount, currency, status, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query, inv.ID, inv.Amount.Amount().String(), inv.Amount.Currency(), inv.Status, inv.CreatedAt, inv.ExpiresAt)
	return err
}

// FindByID implements model.InvoiceRepository.
func (db *DB) FindByID(ctx context.Context, id string) (*model.Invoice, error) {
	var amountStr, currency, status string
	var createdAt, expiresAt time.Time
	query := `SELECT amount, currency, status, created_at, expires_at FROM invoices WHERE id = ?`
	err := db.QueryRowContext(ctx, query, id).Scan(&amountStr, &currency, &status, &createdAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amount := new(big.Int)
	amount.SetString(amountStr, 10)

	return &model.Invoice{
		ID:        id,
		Amount:    money.New(amount, currency),
		Status:    model.InvoiceStatus(status),
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}, nil
}

// UpdateStatus implements model.InvoiceRepository.
func (db *DB) UpdateStatus(ctx context.Context, id string, status model.InvoiceStatus) error {
	query := `UPDATE invoices SET status = ? WHERE id = ?`
	_, err := db.ExecContext(ctx, query, status, id)
	return err
}

func (db *DB) RecordPayment(signature, signer, amount, asset, nonce string) error {
	query := `INSERT OR REPLACE INTO verified_payments (signature, signer, amount, asset, nonce) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, signature, signer, amount, asset, nonce)
	return err
}

func (db *DB) CheckPayment(signature string) (string, error) {
	var signer string
	query := `SELECT signer FROM verified_payments WHERE signature = ?`
	err := db.QueryRow(query, signature).Scan(&signer)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return signer, err
}

func (db *DB) SocketPath() string {
	return filepath.Join(db.DataDir, "settler.sock")
}
