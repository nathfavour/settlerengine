package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
)

func TestL402MiddlewareChallengeAndSolve(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "l402_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := sqlite.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize sqlite: %v", err)
	}
	defer store.Close()

	middleware := NewL402Middleware(store, "supersecretkey")

	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PROTECTED_CONTENT"))
	})

	handlerToTest := middleware.Handler(targetHandler)

	// 1. First Request (No Auth -> Expect Challenge)
	req1 := httptest.NewRequest("GET", "/v1/api/data", nil)
	rr1 := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusPaymentRequired {
		t.Errorf("expected status 402, got %d", rr1.Code)
	}

	wwwAuth := rr1.Header().Get("WWW-Authenticate")
	if !strings.HasPrefix(wwwAuth, "L402 ") {
		t.Errorf("expected WWW-Authenticate starting with L402, got %s", wwwAuth)
	}

	var macaroonBase64 string
	parts := strings.Split(wwwAuth, ",")
	for _, p := range parts {
		if strings.Contains(p, "macaroon=") {
			macaroonBase64 = strings.Trim(strings.Split(p, "macaroon=")[1], "\" ")
		}
	}

	if macaroonBase64 == "" {
		t.Fatalf("failed to extract macaroon from header: %s", wwwAuth)
	}

	var challengeResp map[string]any
	if err := json.NewDecoder(rr1.Body).Decode(&challengeResp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	preimageHex := challengeResp["preimage_for_test"].(string)

	// 2. Second Request: Resubmit with Valid L402 Header
	req2 := httptest.NewRequest("GET", "/v1/api/data", nil)
	req2.Header.Set("Authorization", fmt.Sprintf("L402 %s:%s", macaroonBase64, preimageHex))
	rr2 := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected status 200 OK after payment, got %d. Body: %s", rr2.Code, rr2.Body.String())
	}

	if rr2.Body.String() != "PROTECTED_CONTENT" {
		t.Errorf("expected body PROTECTED_CONTENT, got %s", rr2.Body.String())
	}

	// 3. Third Request: Resubmit with Invalid Preimage -> Expect Rejection
	req3 := httptest.NewRequest("GET", "/v1/api/data", nil)
	req3.Header.Set("Authorization", fmt.Sprintf("L402 %s:wrongpreimagehex", macaroonBase64))
	rr3 := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusPaymentRequired {
		t.Errorf("expected status 402 for invalid preimage, got %d", rr3.Code)
	}
}
