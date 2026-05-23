package domain

import (
	"time"
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
	Amount    Money
	Status    InvoiceStatus
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewInvoice(id string, amount Money, duration time.Duration) *Invoice {
	now := time.Now()
	return &Invoice{
		ID:        id,
		Amount:    amount,
		Status:    StatusNew,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
	}
}

type VerifiedPayment struct {
	Signature  string
	Signer     string
	Amount     Money
	Nonce      string
	VerifiedAt time.Time
}
