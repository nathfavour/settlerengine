package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/nathfavour/settlerengine/core/domain/model"
)

// Event is a simple representation of a domain event.
type Event struct {
	Type string
	Data interface{}
}

const (
	EventSettlementConfirmed = "SETTLEMENT_CONFIRMED"
)

// LocalBus is a simple, in-memory event bus for decoupled communication.
// In a production scenario, this would be replaced by Watermill/Redis/NATS.
type LocalBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
}

func NewLocalBus() *LocalBus {
	return &LocalBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (b *LocalBus) Subscribe(eventType string) chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan Event, 10)
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	return ch
}

func (b *LocalBus) Publish(eventType string, data interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	event := Event{Type: eventType, Data: data}
	for _, ch := range b.subscribers[eventType] {
		select {
		case ch <- event:
		default:
			// Buffer full, skip for now
		}
	}
}
