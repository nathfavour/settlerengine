package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nathfavour/settlerengine/pkg/anyisland"
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/storage"
	"github.com/nathfavour/settlerengine/pkg/uds"
	"github.com/nathfavour/settlerengine/pkg/x402"
)

const Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Anyisland Integration
	_ = anyisland.Register("settler", Version)
	if pulse, err := anyisland.CheckManaged(); err == nil && pulse.Status == "MANAGED" {
		log.Printf("🏝️  Anyisland: Managed by %s", pulse.AnyislandVersion)
	}

	switch os.Args[1] {
	case "proxy":
		runProxy(os.Args[2:])
	case "facilitator":
		runFacilitator(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SettlerEngine - The Agentic Settlement Gateway")
	fmt.Println("\nUsage:")
	fmt.Println("  settler <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  proxy        Start the x402 reverse proxy")
	fmt.Println("  facilitator  Start the settlement facilitator daemon")
	fmt.Println("  help         Show this help message")
}

func runProxy(args []string) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)
	target := fs.String("target", "http://localhost:8081", "Target URL to proxy to")
	listen := fs.String("listen", ":8080", "Listen address")
	recipient := fs.String("recipient", "0x1234567890AbcdEF1234567890aBcdef12345678", "Merchant recipient address")
	chainID := fs.Int64("chain-id", 84532, "Chain ID (default Base Sepolia)")
	asset := fs.String("asset", "0x036CbD53842c5426634e7929541eC2318f3dCF7e", "Asset address (USDC)")
	amount := fs.String("amount", "1000000", "Amount in atomic units")
	fs.Parse(args)

	// 1. Initialize Storage
	db, err := storage.OpenDefault()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()
	log.Printf("📂 Data Directory: %s", db.DataDir)

	// 2. Start UDS Server
	udsServer := uds.NewServer(db.SocketPath())
	if err := udsServer.Start(); err != nil {
		log.Printf("⚠️  UDS Server failed to start: %v", err)
	}
	defer udsServer.Close()

	targetURL, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	cfg := x402.Config{
		DomainParams: crypto.DomainParams{
			ChainID:           big.NewInt(*chainID),
			VerifyingContract: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		NonceExpiry: 5 * time.Minute,
		Recipient:   *recipient,
		Asset:       *asset,
		Amount:      *amount,
		DB:          db,
	}

	mw := x402.NewMiddleware(cfg)
	handler := mw.Handler(proxy)

	log.Printf("🚀 SettlerProxy: Listening on %s", *listen)
	log.Printf("🔗 Proxying to: %s", *target)
	if err := http.ListenAndServe(*listen, handler); err != nil {
		log.Fatal(err)
	}
}

func runFacilitator(args []string) {
	db, err := storage.OpenDefault()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()

	fmt.Println("Starting Settler Facilitator...")
	log.Printf("📂 Data Directory: %s", db.DataDir)
	log.Println("Facilitator daemon is running (stateless verification mode active)")
	select {} // Keep alive
}
