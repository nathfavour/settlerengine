package yield

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/metrics"
	"github.com/nathfavour/settlerengine/pkg/yield/contracts"
)

// RiquidAdapter implements the model.YieldProvider interface for the Riquid Yield Engine.
type RiquidAdapter struct {
	client *ethclient.Client
	signer *crypto.SessionKeySigner
}

func NewRiquidAdapter(client *ethclient.Client, signer *crypto.SessionKeySigner) (*RiquidAdapter, error) {
	return &RiquidAdapter{
		client: client,
		signer: signer,
	}, nil
}

// DepositToYield transfers assets from the main settlement balance to a yield-generating vault.
func (a *RiquidAdapter) DepositToYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	if a.signer == nil {
		return fmt.Errorf("no signer configured for automated deposit")
	}

	vault, err := contracts.NewRiquidVault(common.HexToAddress(strategy.VaultAddress), a.client)
	if err != nil {
		return fmt.Errorf("failed to load vault contract: %w", err)
	}

	auth, err := a.signer.GetTransactor(ctx, a.client)
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	fmt.Printf("💰 Depositing %s %s to %s\n", amount.Amount().String(), amount.Currency(), strategy.VaultAddress)
	
	tx, err := vault.Deposit(auth, amount.Amount())
	if err != nil {
		return fmt.Errorf("failed to broadcast deposit: %w", err)
	}
	_ = tx

	// Update Metrics
	metrics.YieldTVL.WithLabelValues(strategy.ID, amount.Currency()).Set(float64(amount.Amount().Int64()))

	return nil
}

// WithdrawFromYield pulls assets from a yield-generating vault back to the main settlement balance.
func (a *RiquidAdapter) WithdrawFromYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	if a.signer == nil {
		return fmt.Errorf("no signer configured for automated withdrawal")
	}

	vault, err := contracts.NewRiquidVault(common.HexToAddress(strategy.VaultAddress), a.client)
	if err != nil {
		return fmt.Errorf("failed to load vault contract: %w", err)
	}

	auth, err := a.signer.GetTransactor(ctx, a.client)
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	fmt.Printf("💸 Withdrawing %s %s from %s\n", amount.Amount().String(), amount.Currency(), strategy.VaultAddress)
	
	// Note: RiquidVault uses 'shares' for withdrawal, here we assume 1:1 for simplicity or that 'amount' refers to shares.
	tx, err := vault.Withdraw(auth, amount.Amount())
	if err != nil {
		return fmt.Errorf("failed to broadcast withdrawal: %w", err)
	}
	_ = tx

	return nil
}

// GetAPY returns the current Annual Percentage Yield for a given vault.
func (a *RiquidAdapter) GetAPY(ctx context.Context, vaultAddress string) (float64, error) {
	vault, err := contracts.NewRiquidVault(common.HexToAddress(vaultAddress), a.client)
	if err != nil {
		return 0, fmt.Errorf("failed to load vault contract: %w", err)
	}

	apy, err := vault.GetAPY(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call getAPY: %w", err)
	}

	apyFloat := float64(apy.Int64()) / 100.0 // Assuming APY is in basis points
	
	// Update Metrics
	metrics.YieldAPY.WithLabelValues(vaultAddress, vaultAddress).Set(apyFloat)

	return apyFloat, nil
}

// Harvest triggers the claiming and reinvesting of accrued yield.
func (a *RiquidAdapter) Harvest(ctx context.Context, strategy model.YieldStrategy) error {
	if a.signer == nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_NO_SIGNER").Inc()
		return fmt.Errorf("no signer configured for automated harvest")
	}

	vault, err := contracts.NewRiquidVault(common.HexToAddress(strategy.VaultAddress), a.client)
	if err != nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_CONTRACT_ERROR").Inc()
		return fmt.Errorf("failed to load vault contract: %w", err)
	}

	auth, err := a.signer.GetTransactor(ctx, a.client)
	if err != nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_SIGNER_ERROR").Inc()
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	fmt.Printf("🚜 Harvesting yield from %s using Session Key %s\n", strategy.VaultAddress, a.signer.Address().Hex())
	
	tx, err := vault.Harvest(auth)
	if err != nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_TX_ERROR").Inc()
		return fmt.Errorf("failed to broadcast harvest: %w", err)
	}
	_ = tx

	metrics.YieldHarvests.WithLabelValues(strategy.ID, "SUCCESS").Inc()
	return nil
}

// Ensure RiquidAdapter implements model.YieldProvider.
var _ model.YieldProvider = (*RiquidAdapter)(nil)
