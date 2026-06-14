package domain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// AgentIdentity represents an ERC-8004 Agent.
type AgentIdentity struct {
	AgentID   *big.Int       `json:"agent_id"`
	Registry  common.Address `json:"registry"`
	ChainID   *big.Int       `json:"chain_id"`
	Owner     common.Address `json:"owner"`
	Wallet    common.Address `json:"wallet"`
	Metadata  AgentMetadata  `json:"metadata"`
}

// AgentMetadata is the content of the agent-card.json (ERC-8004 Agent Registration File).
type AgentMetadata struct {
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Image        string            `json:"image"`
	Services     []AgentService    `json:"services"`
	Capabilities map[string]any    `json:"capabilities"`
}

type AgentService struct {
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
}

// AgentReputation represents the trust score for an agent.
type AgentReputation struct {
	AgentID  *big.Int       `json:"agent_id"`
	Score    *big.Int       `json:"score"`    // int128 in spec, big.Int here
	Decimals uint8          `json:"decimals"`
	Tags     []string       `json:"tags"`
}

// GlobalAgentID returns a CAIP-10 style identifier.
func (a *AgentIdentity) GlobalID() string {
	return "eip155:" + a.ChainID.String() + ":" + a.Registry.Hex() + ":" + a.AgentID.String()
}
