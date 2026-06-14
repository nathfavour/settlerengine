package ports

import (
	"context"
	"math/big"

	"github.com/nathfavour/settlerengine/internal/domain"
)

// AgentRegistry defines the interface for interacting with ERC-8004 registries.
type AgentRegistry interface {
	// ResolveAgent fetches the identity and metadata for a given agent ID.
	ResolveAgent(ctx context.Context, agentID *big.Int) (*domain.AgentIdentity, error)
	
	// GetReputation retrieves the current trust scores for an agent.
	GetReputation(ctx context.Context, agentID *big.Int) (*domain.AgentReputation, error)
	
	// PostFeedback submits a new feedback signal to the Reputation Registry.
	PostFeedback(ctx context.Context, agentID *big.Int, score *big.Int, tags []string, metadataURI string) error
	
	// RequestValidation triggers a validation task for an agent's work.
	RequestValidation(ctx context.Context, agentID *big.Int, taskHash string) (string, error)
}
