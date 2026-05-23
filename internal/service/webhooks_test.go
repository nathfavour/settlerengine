package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/storage/sqlite"
	"github.com/nathfavour/settlerengine/internal/domain"
)

func TestWebhookDispatcherProcessing(t *testing.T) {
	receivedPayload := ""
	receivedSig := ""
	receivedEvent := ""
	receivedDeliveryID := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedPayload = string(body)
		receivedSig = r.Header.Get("X-Settler-Signature")
		receivedEvent = r.Header.Get("X-Settler-Event")
		receivedDeliveryID = r.Header.Get("X-Settler-Delivery-ID")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "webhooks_test_*")
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

	dispatcher := NewWebhookDispatcher(store, 3)

	ctx := context.Background()

	configSecret := "supersecret"
	config := &domain.WebhookConfig{
		ID:        "cfg-1",
		Url:       server.URL,
		Secret:    configSecret,
		Events:    "invoice.settled",
		CreatedAt: time.Now(),
	}
	if err := store.SaveConfig(ctx, config); err != nil {
		t.Fatalf("failed to save webhook config: %v", err)
	}

	testPayload := `{"invoice_id":"123","status":"SETTLED"}`
	err = dispatcher.QueueDelivery(ctx, "cfg-1", []byte(testPayload), "invoice.settled")
	if err != nil {
		t.Fatalf("failed to queue webhook delivery: %v", err)
	}

	dispatcher.dispatchPending(ctx)

	time.Sleep(100 * time.Millisecond)

	if receivedPayload != testPayload {
		t.Errorf("expected payload %s, got %s", testPayload, receivedPayload)
	}

	if receivedEvent != "invoice.settled" {
		t.Errorf("expected event invoice.settled, got %s", receivedEvent)
	}

	if receivedDeliveryID == "" {
		t.Errorf("expected received delivery ID to not be empty")
	}

	mac := hmac.New(sha256.New, []byte(configSecret))
	mac.Write([]byte(testPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if receivedSig != expectedSig {
		t.Errorf("expected signature %s, got %s", expectedSig, receivedSig)
	}
}
