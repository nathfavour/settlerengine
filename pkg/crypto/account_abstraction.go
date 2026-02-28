package crypto

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SmartAccount represents an ERC-4337 Account Abstraction wallet.
type SmartAccount struct {
	Address     common.Address
	Owner       common.Address
	Entrypoint  common.Address
	Paymaster   common.Address
}

// UserOperation represents an ERC-4337 user operation.
type UserOperation struct {
	Sender             common.Address
	Nonce              *big.Int
	InitCode           []byte
	CallData           []byte
	CallGasLimit       *big.Int
	VerificationGas    *big.Int
	PreVerificationGas *big.Int
	MaxFeePerGas       *big.Int
	MaxPriorityFeePerGas *big.Int
	PaymasterAndData   []byte
	Signature          []byte
}

// AAProvider defines the port for interacting with ERC-4337 bundlers and paymasters.
type AAProvider interface {
	// SendUserOperation submits a signed user operation to the bundler.
	SendUserOperation(ctx context.Context, op UserOperation) (string, error)
	
	// EstimateUserOperationGas estimates gas for a user operation.
	EstimateUserOperationGas(ctx context.Context, op UserOperation) (*UserOperation, error)
	
	// GetSmartAccountAddress computes or retrieves the address of a smart account.
	GetSmartAccountAddress(ctx context.Context, owner common.Address) (common.Address, error)
}

// TransactionManager handles the orchestration of AA or EOA transactions.
type TransactionManager struct {
	client *ethclient.Client
	aa     AAProvider
}

func NewTransactionManager(client *ethclient.Client, aa AAProvider) *TransactionManager {
	return &TransactionManager{
		client: client,
		aa:     aa,
	}
}

// Broadcast handles the dispatch of a transaction, using AA if available.
func (m *TransactionManager) Broadcast(ctx context.Context, tx *types.Transaction) error {
	if m.aa != nil {
		// TODO: Wrap tx in UserOperation and send via AAProvider
		return nil
	}
	return m.client.SendTransaction(ctx, tx)
}
