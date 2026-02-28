package yield

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
	"github.com/nathfavour/settlerengine/pkg/crypto"
)

// RiquidAdapter implements the model.YieldProvider interface for the Riquid Yield Engine.
type RiquidAdapter struct {
	client *ethclient.Client
	signer *crypto.SessionKeySigner
}

func NewRiquidAdapter(client *ethclient.Client, signer *crypto.SessionKeySigner) *RiquidAdapter {
	return &RiquidAdapter{
		client: client,
		signer: signer,
	}
}

// DepositToYield transfers assets from the main settlement balance to a yield-generating vault.
func (a *RiquidAdapter) DepositToYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	// TODO: Implement contract call to Riquid Vault (Deposit)
	// 1. Check allowance
	// 2. Encode deposit(amount)
	// 3. Sign and broadcast via Session Key or AA account
	return fmt.Errorf("not implemented")
}

// WithdrawFromYield pulls assets from a yield-generating vault back to the main settlement balance.
func (a *RiquidAdapter) WithdrawFromYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	// TODO: Implement contract call to Riquid Vault (Withdraw)
	return fmt.Errorf("not implemented")
}

// GetAPY returns the current Annual Percentage Yield for a given vault.
func (a *RiquidAdapter) GetAPY(ctx context.Context, vaultAddress string) (float64, error) {
	// TODO: Query contract for current yield rates
	return 0.0, nil
}

// Harvest triggers the claiming and reinvesting of accrued yield.
func (a *RiquidAdapter) Harvest(ctx context.Context, strategy model.YieldStrategy) error {
	if a.signer == nil {
		return fmt.Errorf("no signer configured for automated harvest")
	}

	auth, err := a.signer.GetTransactor(ctx, a.client)
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	fmt.Printf("🚜 Harvesting yield from %s using Session Key %s\n", strategy.VaultAddress, a.signer.Address().Hex())
	// TODO: Actually broadcast harvest() call to the contract
	_ = auth
	return nil
}

// Ensure RiquidAdapter implements model.YieldProvider.
var _ model.YieldProvider = (*RiquidAdapter)(nil)
