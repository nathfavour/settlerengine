package ports

import (
	"context"
	"time"

	"github.com/nathfavour/settlerengine/internal/domain"
)

type InvoiceStore interface {
	Save(ctx context.Context, invoice *domain.Invoice) error
	FindByID(ctx context.Context, id string) (*domain.Invoice, error)
	UpdateStatus(ctx context.Context, id string, status domain.InvoiceStatus) error
	DeleteExpired(ctx context.Context, expiryThreshold time.Duration) (int64, error)
	GetSettledBefore(ctx context.Context, threshold time.Time) ([]*domain.Invoice, error)
	DeleteInvoices(ctx context.Context, ids []string) error
}

type VerifiedPaymentStore interface {
	RecordPayment(ctx context.Context, payment *domain.VerifiedPayment) error
	CheckPayment(ctx context.Context, signature string) (string, error)
}

type WebhookStore interface {
	SaveConfig(ctx context.Context, config *domain.WebhookConfig) error
	GetConfigs(ctx context.Context) ([]*domain.WebhookConfig, error)
	SaveDelivery(ctx context.Context, delivery *domain.WebhookDelivery) error
	GetPendingDeliveries(ctx context.Context) ([]*domain.WebhookDelivery, error)
	UpdateDeliveryStatus(ctx context.Context, id string, status string, attempts int32, nextAttemptAt time.Time) error
}

type PricingStore interface {
	SaveReputation(ctx context.Context, rep *domain.ClientReputation) error
	GetReputation(ctx context.Context, clientAddress string) (*domain.ClientReputation, error)
	SavePolicy(ctx context.Context, policy *domain.PricingPolicy) error
	GetPolicy(ctx context.Context, resourcePath string) (*domain.PricingPolicy, error)
}

type EscrowStore interface {
	SaveEscrow(ctx context.Context, escrow *domain.Escrow) error
	GetEscrow(ctx context.Context, id string) (*domain.Escrow, error)
	UpdateEscrow(ctx context.Context, id string, status string) error
}

type LsatStore interface {
	SaveLsatChallenge(ctx context.Context, challenge *domain.LsatChallenge) error
	GetLsatChallenge(ctx context.Context, macaroonID string) (*domain.LsatChallenge, error)
	UpdateLsatChallengePreimage(ctx context.Context, macaroonID string, preimage string) error
}

type DBStore interface {
	InvoiceStore
	VerifiedPaymentStore
	WebhookStore
	PricingStore
	EscrowStore
	LsatStore
	Close() error
	Vacuum(ctx context.Context) error
}
