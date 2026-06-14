package model

import (
	"time"

	"github.com/nathfavour/settlerengine/core/pkg/money"
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
	AgentID   string // ERC-8004 Global ID (e.g., eip155:1:0x8004...:42)
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
