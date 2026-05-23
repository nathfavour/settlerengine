package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type WebhookDispatcher struct {
	store      ports.DBStore
	httpClient *http.Client
	maxRetries int32
	wg         sync.WaitGroup
}

func NewWebhookDispatcher(store ports.DBStore, maxRetries int32) *WebhookDispatcher {
	return &WebhookDispatcher{
		store:      store,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		maxRetries: maxRetries,
	}
}

// QueueDelivery creates a new outbox entry in the database.
func (d *WebhookDispatcher) QueueDelivery(ctx context.Context, configID string, payload []byte, event string) error {
	id := uuid.New().String()
	delivery := &domain.WebhookDelivery{
		ID:            id,
		ConfigID:      configID,
		Payload:       string(payload),
		Event:         event,
		Status:        "PENDING",
		Attempts:      0,
		NextAttemptAt: time.Now(),
		CreatedAt:     time.Now(),
	}
	return d.store.SaveDelivery(ctx, delivery)
}

// Start runs a loop that scans and dispatches pending webhooks.
func (d *WebhookDispatcher) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.wg.Wait()
			return
		case <-ticker.C:
			d.dispatchPending(ctx)
		}
	}
}

func (d *WebhookDispatcher) dispatchPending(ctx context.Context) {
	deliveries, err := d.store.GetPendingDeliveries(ctx)
	if err != nil {
		fmt.Printf("⚠️ WebhookDispatcher: Failed to get pending deliveries: %v\n", err)
		return
	}

	configs, err := d.store.GetConfigs(ctx)
	if err != nil {
		fmt.Printf("⚠️ WebhookDispatcher: Failed to get webhook configs: %v\n", err)
		return
	}

	configMap := make(map[string]*domain.WebhookConfig)
	for _, cfg := range configs {
		configMap[cfg.ID] = cfg
	}

	for _, del := range deliveries {
		cfg, exists := configMap[del.ConfigID]
		if !exists {
			continue
		}

		d.wg.Add(1)
		go func(delivery *domain.WebhookDelivery, config *domain.WebhookConfig) {
			defer d.wg.Done()
			d.processDelivery(context.Background(), delivery, config)
		}(del, cfg)
	}
}

func (d *WebhookDispatcher) processDelivery(ctx context.Context, delivery *domain.WebhookDelivery, config *domain.WebhookConfig) {
	fmt.Printf("📡 WebhookDispatcher: Attempting to dispatch delivery %s to %s\n", delivery.ID, config.Url)

	req, err := http.NewRequestWithContext(ctx, "POST", config.Url, bytes.NewBuffer([]byte(delivery.Payload)))
	if err != nil {
		d.markFailed(ctx, delivery, fmt.Sprintf("invalid request setup: %v", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Settler-Event", delivery.Event)
	req.Header.Set("X-Settler-Delivery-ID", delivery.ID)

	// Sign the payload using HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(config.Secret))
	mac.Write([]byte(delivery.Payload))
	signature := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Settler-Signature", signature)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.markFailed(ctx, delivery, fmt.Sprintf("network error: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_ = d.store.UpdateDeliveryStatus(ctx, delivery.ID, "SUCCESS", delivery.Attempts+1, time.Now())
		fmt.Printf("✅ WebhookDispatcher: Successfully delivered %s\n", delivery.ID)
	} else {
		d.markFailed(ctx, delivery, fmt.Sprintf("status code error: %d", resp.StatusCode))
	}
}

func (d *WebhookDispatcher) markFailed(ctx context.Context, delivery *domain.WebhookDelivery, reason string) {
	attempts := delivery.Attempts + 1
	fmt.Printf("⚠️ WebhookDispatcher: Delivery %s failed (attempt %d): %s\n", delivery.ID, attempts, reason)

	if attempts >= d.maxRetries {
		_ = d.store.UpdateDeliveryStatus(ctx, delivery.ID, "FAILED", attempts, time.Now())
		return
	}

	// Exponential backoff retry time
	backoffSec := int64(attempts * attempts * 60)
	nextAttempt := time.Now().Add(time.Duration(backoffSec) * time.Second)
	_ = d.store.UpdateDeliveryStatus(ctx, delivery.ID, "PENDING", attempts, nextAttempt)
}
