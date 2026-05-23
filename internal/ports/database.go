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

type DBStore interface {
	InvoiceStore
	VerifiedPaymentStore
	Close() error
	Vacuum(ctx context.Context) error
}
