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

type WebhookConfig struct {
	ID        string
	Url       string
	Secret    string
	Events    string
	CreatedAt time.Time
}

type WebhookDelivery struct {
	ID            string
	ConfigID      string
	Payload       string
	Event         string
	Status        string
	Attempts      int32
	NextAttemptAt time.Time
	CreatedAt     time.Time
}

type ClientReputation struct {
	ClientAddress string
	Score         int32
	TotalPayments Money
	LastPaymentAt time.Time
}

type PricingPolicy struct {
	ResourcePath    string
	BasePrice       Money
	SurgeMultiplier float64
}
