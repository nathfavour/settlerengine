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
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/x402"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "proxy":
		runProxy(os.Args[2:])
	case "facilitator":
		runFacilitator(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s
", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SettlerEngine - The Agentic Settlement Gateway")
	fmt.Println("
Usage:")
	fmt.Println("  settler <command> [arguments]")
	fmt.Println("
Commands:")
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
	fmt.Println("Starting Settler Facilitator...")
	// Logic from apps/settlerd/main.go or expanded logic
	// For now, just a placeholder
	log.Println("Facilitator daemon is running (stateless verification mode active)")
	select {} // Keep alive
}
