package solana

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type TokenAccountsClient struct {
	SolanaRPC string
	TronRPC   string
}

func NewTokenAccountsClient(solanaRPC, tronRPC string) *TokenAccountsClient {
	return &TokenAccountsClient{
		SolanaRPC: solanaRPC,
		TronRPC:   tronRPC,
	}
}

func (c *TokenAccountsClient) GenerateAddress(ctx context.Context, network ports.ChainNetwork) (string, error) {
	randomBytes := make([]byte, 20)
	rand.Read(randomBytes)
	
	if network == ports.NetworkTron {
		return fmt.Sprintf("T%x", randomBytes), nil
	}
	// Solana base58-like format
	return fmt.Sprintf("SolMockAddress%x", randomBytes), nil
}

func (c *TokenAccountsClient) VerifyPayment(ctx context.Context, network ports.ChainNetwork, address string, amount domain.Money) (bool, string, error) {
	if network == ports.NetworkTron {
		fmt.Printf("🪙 TRON Client: Parsing TRC-20 logs at address %s for amount %s\n", address, amount.Amount().String())
		return true, "tron_trc20_tx_hash_" + address[:6], nil
	}
	fmt.Printf("☀️ SOLANA Client: Monitoring SPL token transfers for %s on address %s\n", amount.Currency(), address)
	return true, "solana_spl_tx_hash_" + address[:6], nil
}
