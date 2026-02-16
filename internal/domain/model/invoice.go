package model

import (
	"time"

	"github.com/nathfavour/settlerengine/pkg/money"
)

type InvoiceStatus string

const (
	StatusNew       InvoiceStatus = "NEW"
	StatusDetected  InvoiceStatus = "DETECTED"
	StatusConfirmed InvoiceStatus = "CONFIRMED"
	StatusSettled   InvoiceStatus = "SETTLED"
	StatusExpired   InvoiceStatus = "EXPIRED"
)

type Invoice struct {
	ID        string
	Amount    money.Money
	Status    InvoiceStatus
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewInvoice(id string, amount money.Money, duration time.Duration) *Invoice {
	now := time.Now()
	return &Invoice{
		ID:        id,
		Amount:    amount,
		Status:    StatusNew,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
	}
}
