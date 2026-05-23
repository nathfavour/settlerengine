package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto"
)

func TestSidecarAutoSolveChallenge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidecar_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keystorePath := filepath.Join(tmpDir, "session.key")
	password := "correcthorsebatterystaple"

	address, err := crypto.GenerateSessionKey(keystorePath, password)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	privKey, err := crypto.LoadSessionKey(keystorePath, password)
	if err != nil {
		t.Fatalf("failed to load key: %v", err)
	}

	callAttempts := 0
	gatedGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callAttempts++
		
		sig := r.Header.Get("X-Payment-Signature")
		if sig == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(ChallengeResponse{
				Status:      http.StatusPaymentRequired,
				Title:       "Payment Required",
				Description: "Signature needed",
				Accepts: []PaymentDescriptor{
					{
						Scheme:  "x402",
						Price:   "1000",
						Asset:   "USDC",
						Network: "84532",
						Nonce:   "nonce_123",
					},
				},
			})
			return
		}

		hasher := sha256.New()
		hasher.Write([]byte("1000:USDC:nonce_123"))
		sigBytes, _ := hexutil.Decode(sig)
		
		pubKey, err := ethcrypto.SigToPub(hasher.Sum(nil), sigBytes)
		if err != nil {
			http.Error(w, "invalid signature bytes", http.StatusBadRequest)
			return
		}
		recoveredAddr := ethcrypto.PubkeyToAddress(*pubKey).Hex()

		if recoveredAddr != address {
			http.Error(w, "signer mismatch", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SUCCESS_DATA"))
	}))
	defer gatedGateway.Close()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	sidecarMux := http.NewServeMux()
	sidecarMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetURL := gatedGateway.URL + r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)

		req, _ := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(bodyBytes))
		resp, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusPaymentRequired {
			var challenge ChallengeResponse
			if err := json.NewDecoder(resp.Body).Decode(&challenge); err == nil && len(challenge.Accepts) > 0 {
				desc := challenge.Accepts[0]

				hasher := sha256.New()
				hasher.Write([]byte(fmt.Sprintf("%s:%s:%s", desc.Price, desc.Asset, desc.Nonce)))
				sigBytes, _ := ethcrypto.Sign(hasher.Sum(nil), privKey)
				sigHex := hexutil.Encode(sigBytes)

				retryReq, _ := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(bodyBytes))
				retryReq.Header.Set("X-Payment-Signature", sigHex)

				retryResp, err := httpClient.Do(retryReq)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadGateway)
					return
				}
				defer retryResp.Body.Close()

				w.WriteHeader(retryResp.StatusCode)
				io.Copy(w, retryResp.Body)
				return
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	sidecarServer := httptest.NewServer(sidecarMux)
	defer sidecarServer.Close()

	resp, err := httpClient.Get(sidecarServer.URL + "/data")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "SUCCESS_DATA" {
		t.Errorf("expected 200 OK and SUCCESS_DATA, got %d and %s", resp.StatusCode, string(body))
	}

	if callAttempts != 2 {
		t.Errorf("expected sidecar to retry and hit gateway exactly 2 times, got %d", callAttempts)
	}
}
