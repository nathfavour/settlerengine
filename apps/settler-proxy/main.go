package main

import (
	"flag"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/x402"
)

func main() {
	target := flag.String("target", "http://localhost:8081", "Target URL to proxy to")
	listen := flag.String("listen", ":8080", "Listen address")
	recipient := flag.String("recipient", "0x1234567890AbcdEF1234567890aBcdef12345678", "Merchant recipient address")
	chainID := flag.Int64("chain-id", 84532, "Chain ID (default Base Sepolia)")
	asset := flag.String("asset", "0x036CbD53842c5426634e7929541eC2318f3dCF7e", "Asset address (USDC)")
	amount := flag.String("amount", "1000000", "Amount in atomic units")

	flag.Parse()

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
	log.Printf("💰 Policy: %s %s on Chain %d", *amount, *asset, *chainID)
	
	if err := http.ListenAndServe(*listen, handler); err != nil {
		log.Fatal(err)
	}
}
