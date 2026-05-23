package bitcoin

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type UTXOClient struct {
	RPCURL string
}

func NewUTXOClient(rpcURL string) *UTXOClient {
	return &UTXOClient{RPCURL: rpcURL}
}

func (c *UTXOClient) GenerateAddress(ctx context.Context, network ports.ChainNetwork) (string, error) {
	// In production, this calls bitcoind/litecoind getnewaddress or derives from an xpub
	prefix := "bc1q"
	if network == ports.NetworkLitecoin {
		prefix = "ltc1q"
	}
	randomBytes := make([]byte, 20)
	rand.Read(randomBytes)
	return fmt.Sprintf("%s%x", prefix, randomBytes), nil
}

func (c *UTXOClient) VerifyPayment(ctx context.Context, network ports.ChainNetwork, address string, amount domain.Money) (bool, string, error) {
	// In production, queries the node via listtransactions or scantxoutset to check for balance
	fmt.Printf("🔍 BTC/LTC UTXO: Scanning address %s for %s %s\n", address, amount.Amount().String(), amount.Currency())
	return true, "mock_tx_hash_" + address[:6], nil
}
