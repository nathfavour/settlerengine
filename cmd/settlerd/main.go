package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/crypto"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/bitcoin"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/monero"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/solana"
	"github.com/nathfavour/settlerengine/internal/adapters/http"
	"github.com/nathfavour/settlerengine/internal/adapters/storage/postgres"
	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
	"github.com/nathfavour/settlerengine/internal/ports"
	"github.com/nathfavour/settlerengine/internal/service"
)

func InitializeStore(ctx context.Context) ports.DBStore {
	dbType := os.Getenv("SETTLER_DB_TYPE")

	switch dbType {
	case "postgres":
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://postgres:postgres@localhost:5432/settler?sslmode=disable"
		}
		repo, err := postgres.NewRepository(dbURL)
		if err != nil {
			log.Fatalf("Failed to initialize postgres: %v", err)
		}
		return repo
	case "sqlite", "":
		dbPath := os.Getenv("SQLITE_DB_PATH")
		if dbPath == "" {
			configDir, err := os.UserConfigDir()
			if err != nil {
				log.Fatalf("Failed to get user config dir: %v", err)
			}
			dataDir := filepath.Join(configDir, "settlerengine")
			_ = os.MkdirAll(dataDir, 0755)
			dbPath = filepath.Join(dataDir, "settler.db")
		}
		repo, err := sqlite.NewRepository(dbPath)
		if err != nil {
			log.Fatalf("Failed to initialize sqlite: %v", err)
		}
		return repo
	default:
		log.Fatalf("Invalid database implementation selection: %s", dbType)
		return nil
	}
}

func main() {
	fmt.Println("======================================================")
	fmt.Println("   SettlerEngine Architecture Specification Gateway   ")
	fmt.Println("======================================================")
	log.Println("Starting settlement gateway daemon (settlerd)...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Initialize DB Store dynamic port
	store := InitializeStore(ctx)
	defer store.Close()
	log.Println("📂 Database initialized successfully")

	// 2. Initialize Blockchain watcher adapters
	utxo := bitcoin.NewUTXOClient("http://localhost:8332")
	moneroScanner := monero.NewMoneroClient("http://localhost:18082", "mock_view_key", "mock_public_address")
	tokenTracker := solana.NewTokenAccountsClient("http://localhost:8899", "http://localhost:50051")
	
	watcher := crypto.NewMultiChainWatcher(utxo, moneroScanner, tokenTracker)
	_ = watcher.StartWatching(ctx, func(signal ports.InvoicePaymentSignal) {
		log.Printf("🔔 Payment Signal: Detected %s %s on %s for invoice %s\n", 
			signal.Amount.Amount().String(), 
			signal.Amount.Currency(), 
			signal.Network, 
			signal.InvoiceID,
		)
	})
	log.Println("🔗 Multi-Chain watcher clients active")

	// 3. Initialize Payment Engine & Background Maintenance Ticker
	engine := service.NewPaymentEngine(store, watcher)

	// UserConfigDir for flat-file JSON archives
	configDir, _ := os.UserConfigDir()
	dataDir := filepath.Join(configDir, "settlerengine")
	pruner := service.NewPruner(store, dataDir)
	
	// Start 24-Hour Pruning background task
	go pruner.Start(ctx, 24*time.Hour)
	log.Println("🧹 Background 24h Pruning daemon worker started")

	// 4. Start HTTP Server proxy & dashboard API
	listenAddr := os.Getenv("SETTLER_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	server := http.NewServer(listenAddr, engine, store)

	go func() {
		if err := server.Start(ctx); err != nil {
			log.Fatalf("HTTP server failure: %v", err)
		}
	}()

	log.Println("✅ SettlerEngine is running. Press Ctrl+C to terminate.")
	
	<-ctx.Done()
	log.Println("Daemon received termination signal. Gracefully shutting down...")
	time.Sleep(1 * time.Second) // wait for server closure
	log.Println("Shutdown complete. Farewell!")
}
