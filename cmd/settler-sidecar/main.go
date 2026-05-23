package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto"
)

type ChallengeResponse struct {
	Status      int                 `json:"status"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Accepts     []PaymentDescriptor `json:"accepts"`
}

type PaymentDescriptor struct {
	Scheme  string `json:"scheme"`
	Price   string `json:"price"`
	Asset   string `json:"asset"`
	Network string `json:"network"`
	Nonce   string `json:"nonce"`
	PayTo   string `json:"pay_to"`
}

func main() {
	listenAddr := flag.String("listen", ":8090", "Local listen address")
	targetAddr := flag.String("target", "http://localhost:8080", "Gated SettlerProxy address")
	keystorePath := flag.String("keystore", "", "Path to encrypted session key keystore")
	password := flag.String("password", "", "Keystore decryption password")

	flag.Parse()

	if *keystorePath == "" || *password == "" {
		log.Fatal("Error: --keystore and --password are required parameters.")
	}

	// 1. Decrypt and load session key
	privKey, err := crypto.LoadSessionKey(*keystorePath, *password)
	if err != nil {
		log.Fatalf("Keystore decryption failed: %v", err)
	}

	log.Printf("🤖 SettlerSidecar: Loaded session key wallet address %s", ethcrypto.PubkeyToAddress(privKey.PublicKey).Hex())

	httpClient := &http.Client{Timeout: 15 * time.Second}

	// 2. Local proxy gateway handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetURL := *targetAddr + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		bodyBytes, _ := io.ReadAll(r.Body)

		log.Printf("📡 Sidecar proxying request -> %s", targetURL)
		req, _ := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(bodyBytes))
		
		for name, values := range r.Header {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Intercept x402 402 challenge
		if resp.StatusCode == http.StatusPaymentRequired {
			log.Println("💰 Sidecar: Intercepted x402 challenge. Auto-solving challenge...")

			var challenge ChallengeResponse
			if err := json.NewDecoder(resp.Body).Decode(&challenge); err == nil && len(challenge.Accepts) > 0 {
				desc := challenge.Accepts[0]
				
				// Hash EIP-712 Intent
				hasher := sha256.New()
				hasher.Write([]byte(fmt.Sprintf("%s:%s:%s", desc.Price, desc.Asset, desc.Nonce)))
				sigBytes, err := ethcrypto.Sign(hasher.Sum(nil), privKey)
				if err != nil {
					http.Error(w, "signing failed", http.StatusInternalServerError)
					return
				}
				sigHex := hexutil.Encode(sigBytes)

				log.Printf("🤖 Sidecar: Derived signature: %s. Resubmitting request...", sigHex[:16]+"...")

				// Re-dispatch original request with signature header
				retryReq, _ := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(bodyBytes))
				for name, values := range r.Header {
					for _, value := range values {
						retryReq.Header.Add(name, value)
					}
				}
				retryReq.Header.Set("X-Payment-Signature", sigHex)

				retryResp, err := httpClient.Do(retryReq)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadGateway)
					return
				}
				defer retryResp.Body.Close()

				for name, values := range retryResp.Header {
					for _, value := range values {
						w.Header().Add(name, value)
					}
				}
				w.WriteHeader(retryResp.StatusCode)
				io.Copy(w, retryResp.Body)
				return
			}
		}

		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	log.Printf("🤖 SettlerSidecar: Listening for local agent requests on %s\n", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}
