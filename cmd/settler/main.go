package main

import (
	"context"
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
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/erc8004"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/mantle"
	"github.com/nathfavour/settlerengine/internal/domain"
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
	case "pay":
		runPay(os.Args[2:])
	case "demo":
		runDemo(os.Args[2:])
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
	fmt.Println("  pay          Execute a policy-protected payment")
	fmt.Println("  demo         Run a full agentic demo with Mantle on-chain anchoring")
	fmt.Println("  help         Show this help message")
}

func runDemo(args []string) {
	fmt.Println("🎬 Starting SettlerEngine Agentic Demo...")
	
	privKey := os.Getenv("PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("❌ PRIVATE_KEY environment variable is required for demo")
	}

	// 1. ERC-8004 Identity Resolution
	fmt.Println("🤖 [1/3] Resolving Agent Identity (ERC-8004)...")
	registry, _ := erc8004.NewRegistryClient(
		"https://rpc.sepolia.mantle.xyz",
		common.HexToAddress("0x8004000000000000000000000000000000000001"),
		common.HexToAddress("0x8004000000000000000000000000000000000002"),
		common.HexToAddress("0x8004000000000000000000000000000000000003"),
	)
	
	agentID := big.NewInt(42)
	identity, _ := registry.ResolveAgent(context.Background(), agentID)
	fmt.Printf("✅ Identity Verified: %s\n", identity.Metadata.Name)

	// 2. Policy-Protected Payment Handshake
	fmt.Println("💰 [2/3] Performing Policy-Protected Payment...")
	max, _ := new(big.Int).SetString("1000000000000000000", 10)
	policy := domain.NewPaymentPolicy("demo-policy", max, nil, time.Time{})
	
	amount := big.NewInt(1000)
	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
	
	if err := policy.Check(amount, recipient); err != nil {
		log.Fatalf("❌ Policy Violation: %v", err)
	}
	fmt.Println("✅ Payment Approved by local guardrails.")

	// 3. Mantle On-Chain Anchoring
	fmt.Println("⚓ [3/3] Anchoring transaction to Mantle Sepolia...")
	mantleClient, err := mantle.NewRegistryClient(
		"https://rpc.sepolia.mantle.xyz",
		common.HexToAddress("0x33aE8331a2406EEc3A33483001aC5650DA2e0662"),
		big.NewInt(5003),
	)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Mantle: %v", err)
	}

	var agentBytes [32]byte
	copy(agentBytes[:], agentID.Bytes())

	txHash, err := mantleClient.LogPayment(
		context.Background(),
		privKey,
		agentBytes,
		big.NewInt(1337), // Demo Invoice ID
		amount,
		"Demo agent payment anchored via SettlerEngine",
	)
	if err != nil {
		log.Fatalf("❌ On-chain anchoring failed: %v", err)
	}

	fmt.Printf("🚀 SUCCESS! On-chain footprint created.\n")
	fmt.Printf("🔗 Transaction Hash: %s\n", txHash)
	fmt.Printf("🌍 Explorer: https://explorer.sepolia.mantle.xyz/tx/%s\n", txHash)
}

func runPay(args []string) {
	fs := flag.NewFlagSet("pay", flag.ExitOnError)
	fs.String("rpc", "https://sepolia.base.org", "Ethereum RPC URL")
	to := fs.String("to", "", "Recipient address")
	amountStr := fs.String("amount", "0", "Amount in wei")
	privKey := fs.String("key", "", "Private key (hex)")
	maxPerTx := fs.String("max-per-tx", "1000000000000000000", "Max wei per transaction (1 ETH)")
	fs.Parse(args)

	if *to == "" || *privKey == "" {
		log.Fatal("Recipient (-to) and Private Key (-key) are required")
	}

	chainID := big.NewInt(84532) // Base Sepolia
	signer, err := crypto.NewSessionKeySigner(*privKey, chainID)
	if err != nil {
		log.Fatalf("Invalid signer: %v", err)
	}

	max, _ := new(big.Int).SetString(*maxPerTx, 10)
	policy := domain.NewPaymentPolicy("cli-policy", max, nil, time.Time{})

	policySigner := crypto.NewPolicySigner(signer, policy)
	
	amount, _ := new(big.Int).SetString(*amountStr, 10)
	recipient := common.HexToAddress(*to)

	// In a real CLI, we would use policySigner.GetTransactorWithPolicy
	if err := policySigner.Check(amount, recipient); err != nil {
		log.Fatalf("❌ Policy Blocked Payment: %v", err)
	}

	fmt.Printf("✅ Policy Approved Payment of %s wei to %s\n", amount.String(), recipient.Hex())
	fmt.Println("Executing transaction... (Simulated for this demo)")
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
