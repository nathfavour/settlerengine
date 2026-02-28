package model

import "context"

// InvoiceRepository defines the port for persisting invoice data.
type InvoiceRepository interface {
	Save(ctx context.Context, invoice *Invoice) error
	FindByID(ctx context.Context, id string) (*Invoice, error)
	UpdateStatus(ctx context.Context, id string, status InvoiceStatus) error
}
