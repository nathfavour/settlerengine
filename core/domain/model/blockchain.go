package model

import (
	"context"
	"github.com/nathfavour/settlerengine/core/pkg/money"
)

// BlockchainClient defines the port for interacting with various blockchain networks.
type BlockchainClient interface {
	// BroadcastTransaction sends a signed transaction to the network.
	BroadcastTransaction(ctx context.Context, tx interface{}) (string, error)
	
	// GetBalance returns the balance of an address in its native asset or a specific token.
	GetBalance(ctx context.Context, address string, asset string) (money.Money, error)
}
