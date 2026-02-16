package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

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
	`
	_, err := db.Exec(query)
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
