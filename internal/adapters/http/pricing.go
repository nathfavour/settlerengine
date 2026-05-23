package http

import (
	"context"
	"math/big"

	"github.com/nathfavour/settlerengine/internal/domain"
	"github.com/nathfavour/settlerengine/internal/ports"
)

type DynamicPricingEngine struct {
	store ports.DBStore
}

func NewDynamicPricingEngine(store ports.DBStore) *DynamicPricingEngine {
	return &DynamicPricingEngine{store: store}
}

// CalculatePrice resolves the customized pricing model for a resource path and client.
func (e *DynamicPricingEngine) CalculatePrice(ctx context.Context, resourcePath string, clientAddress string) (domain.Money, error) {
	policy, err := e.store.GetPolicy(ctx, resourcePath)
	if err != nil || policy == nil {
		defaultPrice := domain.NewMoney(big.NewInt(1000000), "USDC") // Default $1.00
		return defaultPrice, nil
	}

	priceBig := policy.BasePrice.Amount()

	// Apply surge multiplier
	if policy.SurgeMultiplier > 1.0 {
		surgeBig := big.NewFloat(policy.SurgeMultiplier)
		priceFloat := new(big.Float).SetInt(priceBig)
		priceFloat.Mul(priceFloat, surgeBig)
		priceBig, _ = priceFloat.Int(nil)
	}

	// Query client reputation to apply discounts or premiums
	reputation, err := e.store.GetReputation(ctx, clientAddress)
	if err == nil && reputation != nil {
		discountBps := int64(0)
		if reputation.Score >= 80 {
			discountBps = 2000 // 20% discount
		} else if reputation.Score >= 50 {
			discountBps = 1000 // 10% discount
		} else if reputation.Score <= 20 {
			discountBps = -1000 // 10% premium
		}

		if discountBps != 0 {
			multiplier := big.NewInt(10000 - discountBps)
			priceBig.Mul(priceBig, multiplier)
			priceBig.Div(priceBig, big.NewInt(10000))
		}
	}

	return domain.NewMoney(priceBig, policy.BasePrice.Currency()), nil
}
