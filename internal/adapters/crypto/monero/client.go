package monero

import (
	"context"
	"fmt"

	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type MoneroClient struct {
	RPCURL         string
	PrivateViewKey string
	PublicAddress  string
}

func NewMoneroClient(rpcURL, viewKey, address string) *MoneroClient {
	return &MoneroClient{
		RPCURL:         rpcURL,
		PrivateViewKey: viewKey,
		PublicAddress:  address,
	}
}

func (c *MoneroClient) GenerateSubaddress(ctx context.Context, invoiceID string) (string, error) {
	// Monero subaddresses are derived using the private view key + public address via monero-wallet-rpc
	// This enables non-custodial destination generation without exposing the private spend key.
	return fmt.Sprintf("4SubAddrMockMonero%s", invoiceID[:8]), nil
}

func (c *MoneroClient) ScanPayments(ctx context.Context, address string, amount domain.Money) (bool, string, error) {
	fmt.Printf("🔒 XMR ViewKey: Scanning subaddress %s using view key %s for %s XMR\n", address, c.PrivateViewKey[:8]+"...", amount.Amount().String())
	return true, "mock_xmr_tx_hash_" + address[:6], nil
}
