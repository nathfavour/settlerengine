package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// YieldAPY tracks the current Annual Percentage Yield per strategy
	YieldAPY = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settler_yield_apy",
		Help: "Current Annual Percentage Yield for a strategy",
	}, []string{"strategy_id", "vault_address"})

	// YieldTVL tracks the total value locked in yield strategies
	YieldTVL = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settler_yield_tvl_atomic",
		Help: "Total Value Locked in a yield strategy (in atomic units)",
	}, []string{"strategy_id", "currency"})

	// YieldHarvests counts the number of harvest operations
	YieldHarvests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settler_yield_harvest_total",
		Help: "Total number of yield harvest operations",
	}, []string{"strategy_id", "status"})
)
