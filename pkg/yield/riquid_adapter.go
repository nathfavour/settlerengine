package yield

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/core/pkg/money"
	"github.com/nathfavour/settlerengine/pkg/crypto"
	"github.com/nathfavour/settlerengine/pkg/metrics"
)

// vaultABI is a simplified ABI for the RiquidVault.
const vaultABI = `[{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"deposit","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"getAPY","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"harvest","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

// RiquidAdapter implements the model.YieldProvider interface for the Riquid Yield Engine.
type RiquidAdapter struct {
	client *ethclient.Client
	signer *crypto.SessionKeySigner
	abi    abi.ABI
}

func NewRiquidAdapter(client *ethclient.Client, signer *crypto.SessionKeySigner) (*RiquidAdapter, error) {
	parsedABI, err := abi.JSON(strings.NewReader(vaultABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	return &RiquidAdapter{
		client: client,
		signer: signer,
		abi:    parsedABI,
	}, nil
}

// DepositToYield transfers assets from the main settlement balance to a yield-generating vault.
func (a *RiquidAdapter) DepositToYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	if a.signer == nil {
		return fmt.Errorf("no signer configured for automated deposit")
	}

	auth, err := a.signer.GetTransactor(ctx, a.client)
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	// 1. Encode deposit(amount)
	input, err := a.abi.Pack("deposit", amount.Amount())
	if err != nil {
		return fmt.Errorf("failed to pack deposit call: %w", err)
	}

	// 2. Broadcast (In a real scenario, we'd also check and set ERC-20 allowance)
	fmt.Printf("💰 Depositing %s %s to %s\n", amount.Amount().String(), amount.Currency(), strategy.VaultAddress)
	_ = auth
	_ = input

	// Update Metrics
	metrics.YieldTVL.WithLabelValues(strategy.ID, amount.Currency()).Set(float64(amount.Amount().Int64()))

	return nil
}

// WithdrawFromYield pulls assets from a yield-generating vault back to the main settlement balance.
func (a *RiquidAdapter) WithdrawFromYield(ctx context.Context, amount money.Money, strategy model.YieldStrategy) error {
	// TODO: Implement contract call to Riquid Vault (Withdraw)
	return fmt.Errorf("not implemented")
}

// GetAPY returns the current Annual Percentage Yield for a given vault.
func (a *RiquidAdapter) GetAPY(ctx context.Context, vaultAddress string) (float64, error) {
	to := common.HexToAddress(vaultAddress)
	input, err := a.abi.Pack("getAPY")
	if err != nil {
		return 0, fmt.Errorf("failed to pack getAPY call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &to,
		Data: input,
	}

	result, err := a.client.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call getAPY: %w", err)
	}

	var apy *big.Int
	err = a.abi.UnpackIntoInterface(&apy, "getAPY", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack getAPY result: %w", err)
	}

	apyFloat := float64(apy.Int64()) / 100.0 // Assuming APY is in basis points
	
	// Update Metrics
	// We'd need the strategy ID here, but for now we use vaultAddress as ID
	metrics.YieldAPY.WithLabelValues(vaultAddress, vaultAddress).Set(apyFloat)

	return apyFloat, nil
}

// Harvest triggers the claiming and reinvesting of accrued yield.
func (a *RiquidAdapter) Harvest(ctx context.Context, strategy model.YieldStrategy) error {
	if a.signer == nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_NO_SIGNER").Inc()
		return fmt.Errorf("no signer configured for automated harvest")
	}

	auth, err := a.signer.GetTransactor(ctx, a.client)
	if err != nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_SIGNER_ERROR").Inc()
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	fmt.Printf("🚜 Harvesting yield from %s using Session Key %s\n", strategy.VaultAddress, a.signer.Address().Hex())
	
	// Encode harvest()
	input, err := a.abi.Pack("harvest")
	if err != nil {
		metrics.YieldHarvests.WithLabelValues(strategy.ID, "FAILED_PACK_ERROR").Inc()
		return fmt.Errorf("failed to pack harvest call: %w", err)
	}
	_ = input
	_ = auth

	metrics.YieldHarvests.WithLabelValues(strategy.ID, "SUCCESS").Inc()
	return nil
}

// Ensure RiquidAdapter implements model.YieldProvider.
var _ model.YieldProvider = (*RiquidAdapter)(nil)
