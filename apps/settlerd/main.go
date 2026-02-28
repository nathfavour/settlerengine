package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/domain/service"
	"github.com/nathfavour/settlerengine/pkg/chains"
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/storage"
	"github.com/nathfavour/settlerengine/pkg/yield"
)

func main() {
	fmt.Println("SettlerEngine: The Payment Engine for the Agentic Era")
	log.Println("Starting settlement daemon...")

	// 1. Initialize Storage
	db, err := storage.OpenDefault()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()
	log.Printf("📂 Data Directory: %s", db.DataDir)

	// 2. Initialize Blockchain Multi-Client
	mc := chains.NewMultiClient()
	defer mc.Close()

	// 3. Initialize Signer for Automation (Load from env in production)
	// Using a placeholder key for demonstration.
	signer, err := crypto.NewSessionKeySigner("0x0000000000000000000000000000000000000000000000000000000000000001", big.NewInt(56))
	if err != nil {
		log.Fatalf("Failed to initialize signer: %v", err)
	}

	// 4. Initialize Riquid Adapter
	bscClient, err := mc.GetClient(chains.ChainIDBSC)
	if err != nil {
		log.Printf("⚠️ BSC Client not available: %v", err)
	}
	
	riquid, err := yield.NewRiquidAdapter(bscClient, signer)
	if err != nil {
		log.Fatalf("Failed to initialize Riquid adapter: %v", err)
	}

	// 5. Initialize Event Bus
	bus := service.NewLocalBus()

	// 6. Initialize Settlement Engine
	engine := service.NewDefaultSettlementEngine(db, mc, riquid, bus)

	// 7. Initialize Yield Service
	// Threshold: 0.1 BNB (demonstration)
	threshold := big.NewInt(100000000000000000) 
	yieldSvc := service.NewYieldService(engine, riquid, threshold)

	// 8. Start Background Workers
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	strategies := []model.YieldStrategy{
		{
			ID:           "riquid_bnb_vault",
			Provider:     "Riquid",
			AutoHarvest:   true,
			VaultAddress: "0x0000000000000000000000000000000000000000", // placeholder
		},
	}
	
	// Start Auto-Harvesting
	go yieldSvc.StartAutoHarvestWorker(ctx, 1*time.Hour, strategies)

	// Listen for new settlements and route 100% to Riquid
	go yieldSvc.ListenForSettlements(ctx, bus, strategies[0], 100.0)

	log.Println("✅ Settlement daemon is running")
	
	// Keep alive until interrupt
	<-ctx.Done()
	log.Println("Shutting down...")
}
