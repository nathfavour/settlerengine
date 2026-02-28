package crypto

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SessionKeySigner manages a local private key for automated transaction signing.
// In a production environment, this should wrap a secure enclave or KMS.
type SessionKeySigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    *big.Int
}

func NewSessionKeySigner(hexKey string, chainID *big.Int) (*SessionKeySigner, error) {
	privateKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &SessionKeySigner{
		privateKey: privateKey,
		address:    address,
		chainID:    chainID,
	}, nil
}

func (s *SessionKeySigner) Address() common.Address {
	return s.address
}

// GetTransactor returns a bind.Transopts configured with the session key.
func (s *SessionKeySigner) GetTransactor(ctx context.Context, client *ethclient.Client) (*bind.TransactOpts, error) {
	nonce, err := client.PendingNonceAt(ctx, s.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // default to 0
	auth.GasLimit = uint64(300000) // standard limit for simple contract calls
	auth.GasPrice = gasPrice

	return auth, nil
}
