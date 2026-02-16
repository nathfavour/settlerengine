package chains

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

// MultiClient manages multiple RPC clients for different chains.
type MultiClient struct {
	clients map[ChainID]*ethclient.Client
	mu      sync.RWMutex
}

func NewMultiClient() *MultiClient {
	return &MultiClient{
		clients: make(map[ChainID]*ethclient.Client),
	}
}

// GetClient returns an ethclient for the given chain ID, initializing it if necessary.
func (mc *MultiClient) GetClient(id ChainID) (*ethclient.Client, error) {
	mc.mu.RLock()
	client, ok := mc.clients[id]
	mc.mu.RUnlock()
	if ok {
		return client, nil
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check again in case another goroutine initialized it
	if client, ok = mc.clients[id]; ok {
		return client, nil
	}

	cfg, err := GetChainConfig(id)
	if err != nil {
		return nil, err
	}

	client, err = ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC for chain %d: %w", id, err)
	}

	mc.clients[id] = client
	return client, nil
}

// Close closes all managed clients.
func (mc *MultiClient) Close() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, client := range mc.clients {
		client.Close()
	}
	mc.clients = make(map[ChainID]*ethclient.Client)
}
