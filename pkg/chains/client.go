package chains

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
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

// BroadcastTransaction sends a signed transaction to the network.
func (mc *MultiClient) BroadcastTransaction(ctx context.Context, tx interface{}) (string, error) {
	ethTx, ok := tx.(*types.Transaction)
	if !ok {
		return "", fmt.Errorf("invalid transaction type: %T", tx)
	}

	chainID := ethTx.ChainId()
	client, err := mc.GetClient(ChainID(chainID.Uint64()))
	if err != nil {
		return "", err
	}

	if err := client.SendTransaction(ctx, ethTx); err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return ethTx.Hash().Hex(), nil
}

// GetBalance returns the balance of an address in its native asset or a specific token.
func (mc *MultiClient) GetBalance(ctx context.Context, address string, asset string) (money.Money, error) {
	// Simple implementation for native balance for now.
	// For tokens, this would require calling balanceOf(address).
	
	// Default to first client for now, or we might need ChainID in the signature
	// Let's assume BSC for now as requested
	client, err := mc.GetClient(ChainIDBSC)
	if err != nil {
		return money.Money{}, err
	}

	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(ctx, account, nil)
	if err != nil {
		return money.Money{}, fmt.Errorf("failed to get balance: %w", err)
	}

	return money.New(balance, "BNB"), nil
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

// Ensure implementation of model.BlockchainClient.
var _ model.BlockchainClient = (*MultiClient)(nil)
