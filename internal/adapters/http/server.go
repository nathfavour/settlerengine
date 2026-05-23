package http

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
	"github.com/nathfavour/settlerengine/internal/service"
)

type Server struct {
	listenAddr string
	engine     *service.PaymentEngine
	store      ports.DBStore
}

func NewServer(listenAddr string, engine *service.PaymentEngine, store ports.DBStore) *Server {
	return &Server{
		listenAddr: listenAddr,
		engine:     engine,
		store:      store,
	}
}

type CreateInvoiceRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// 1. Merchant Dashboard API
	mux.HandleFunc("/api/invoices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req CreateInvoiceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			amountBig := new(big.Int)
			amountBig.SetString(req.Amount, 10)
			money := domain.NewMoney(amountBig, req.Currency)

			invoice, err := s.engine.CreateInvoice(r.Context(), money, 1*time.Hour)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(invoice)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// 2. x402 reverse proxy gating endpoint
	targetURL, _ := url.Parse("http://localhost:8081") // Default proxy target
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	mux.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		sig := r.Header.Get("X-Payment-Signature")
		if sig != "" {
			signer, err := s.store.CheckPayment(r.Context(), sig)
			if err == nil && signer != "" {
				proxy.ServeHTTP(w, r)
				return
			}
		}

		// Challenge: HTTP 402 Payment Required
		nonce := fmt.Sprintf("nonce_%d", time.Now().UnixNano())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":      http.StatusPaymentRequired,
			"title":       "Payment Required",
			"description": "A valid EIP-712 / x402 payment signature is required to view this resource.",
			"accepts": []map[string]any{
				{
					"scheme":  "x402",
					"price":   "1000000",
					"asset":   "USDC",
					"network": "84532",
					"nonce":   nonce,
				},
			},
		})
	})

	server := &http.Server{
		Addr:    s.listenAddr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	fmt.Printf("🚀 HTTP Server: Listening on %s\n", s.listenAddr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
