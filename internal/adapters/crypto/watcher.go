package crypto

import (
	"context"
	"fmt"
	"time"

	"github.com/nathfavour/settlerengine/internal/adapters/crypto/bitcoin"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/monero"
	"github.com/nathfavour/settlerengine/internal/adapters/crypto/solana"
	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type MultiChainWatcher struct {
	utxoClient   *bitcoin.UTXOClient
	moneroClient *monero.MoneroClient
	tokenClient  *solana.TokenAccountsClient
}

func NewMultiChainWatcher(utxo *bitcoin.UTXOClient, monero *monero.MoneroClient, token *solana.TokenAccountsClient) *MultiChainWatcher {
	return &MultiChainWatcher{
		utxoClient:   utxo,
		moneroClient: monero,
		tokenClient:  token,
	}
}

func (w *MultiChainWatcher) GenerateAddress(ctx context.Context, network ports.ChainNetwork, invoiceID string) (string, error) {
	switch network {
	case ports.NetworkBitcoin, ports.NetworkLitecoin:
		return w.utxoClient.GenerateAddress(ctx, network)
	case ports.NetworkMonero:
		return w.moneroClient.GenerateSubaddress(ctx, invoiceID)
	case ports.NetworkSolana, ports.NetworkTron:
		return w.tokenClient.GenerateAddress(ctx, network)
	default:
		return "", fmt.Errorf("unsupported blockchain network: %s", network)
	}
}

func (w *MultiChainWatcher) VerifyPayment(ctx context.Context, network ports.ChainNetwork, address string, amount domain.Money) (bool, string, error) {
	switch network {
	case ports.NetworkBitcoin, ports.NetworkLitecoin:
		return w.utxoClient.VerifyPayment(ctx, network, address, amount)
	case ports.NetworkMonero:
		return w.moneroClient.ScanPayments(ctx, address, amount)
	case ports.NetworkSolana, ports.NetworkTron:
		return w.tokenClient.VerifyPayment(ctx, network, address, amount)
	default:
		return false, "", fmt.Errorf("unsupported blockchain network: %s", network)
	}
}

func (w *MultiChainWatcher) StartWatching(ctx context.Context, handler func(signal ports.InvoicePaymentSignal)) error {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Background event loop placeholder
			}
		}
	}()
	return nil
}

var _ ports.BlockchainWatcher = (*MultiChainWatcher)(nil)
