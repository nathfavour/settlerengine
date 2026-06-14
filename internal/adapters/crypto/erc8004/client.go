package erc8004

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type RegistryClient struct {
	client            *ethclient.Client
	identityRegistry  common.Address
	reputationRegistry common.Address
	validationRegistry common.Address
}

func NewRegistryClient(rpcURL string, identity, reputation, validation common.Address) (*RegistryClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &RegistryClient{
		client:             client,
		identityRegistry:   identity,
		reputationRegistry: reputation,
		validationRegistry: validation,
	}, nil
}

func (c *RegistryClient) ResolveAgent(ctx context.Context, agentID *big.Int) (*domain.AgentIdentity, error) {
	// In a real implementation, we would call the Identity Registry contract
	// For now, return a placeholder that follows the spec
	return &domain.AgentIdentity{
		AgentID:  agentID,
		Registry: c.identityRegistry,
		ChainID:  big.NewInt(1), // Assume Mainnet for placeholder
		Metadata: domain.AgentMetadata{
			Name: "Settler-Managed Agent",
		},
	}, nil
}

func (c *RegistryClient) GetReputation(ctx context.Context, agentID *big.Int) (*domain.AgentReputation, error) {
	// Call Reputation Registry
	return &domain.AgentReputation{
		AgentID:  agentID,
		Score:    big.NewInt(100),
		Decimals: 2,
	}, nil
}

func (c *RegistryClient) PostFeedback(ctx context.Context, agentID *big.Int, score *big.Int, tags []string, metadataURI string) error {
	fmt.Printf("📝 ERC-8004: Posting feedback for agent %s: score=%s tags=%v\n", agentID.String(), score.String(), tags)
	return nil
}

func (c *RegistryClient) RequestValidation(ctx context.Context, agentID *big.Int, taskHash string) (string, error) {
	fmt.Printf("🔍 ERC-8004: Requesting validation for agent %s, task=%s\n", agentID.String(), taskHash)
	return "0x_validation_request_id", nil
}

// Ensure interface compliance
var _ ports.AgentRegistry = (*RegistryClient)(nil)
