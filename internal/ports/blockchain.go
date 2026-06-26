package ports

import (
	"context"

	"github.com/nathfavour/settlerengine/internal/domain"
)

type ChainNetwork string

const (
	NetworkBitcoin  ChainNetwork = "BTC"
	NetworkLitecoin ChainNetwork = "LTC"
	NetworkMonero   ChainNetwork = "XMR"
	NetworkSolana   ChainNetwork = "SOL"
	NetworkTron     ChainNetwork = "TRX"
	NetworkCasper   ChainNetwork = "CSPR"
)

type InvoicePaymentSignal struct {
	InvoiceID string
	TxHash    string
	Amount    domain.Money
	Network   ChainNetwork
}

type BlockchainWatcher interface {
	// StartWatching begins monitoring the blockchain for incoming transaction events.
	StartWatching(ctx context.Context, handler func(signal InvoicePaymentSignal)) error
	
	// GenerateAddress creates a new payment destination address for the invoice.
	GenerateAddress(ctx context.Context, network ChainNetwork, invoiceID string) (string, error)
	
	// VerifyPayment checks if a specific transaction or address contains a settled payment.
	VerifyPayment(ctx context.Context, network ChainNetwork, address string, amount domain.Money) (bool, string, error)
}
