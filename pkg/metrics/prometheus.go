package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal tracks total RPC requests by provider, method, and status
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_requests_total",
			Help: "Total RPC requests by provider, method, and status",
		},
		[]string{"provider", "method", "status"},
	)

	// RequestDuration tracks RPC request latency
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rpc_request_duration_seconds",
			Help:    "RPC request latency",
			Buckets: []float64{.01, .05, .1, .5, 1, 2, 5, 10},
		},
		[]string{"provider"},
	)

	// ProviderHealthStatus tracks the current health state of each provider
	// 1 = healthy, 0.5 = degraded, 0 = unhealthy
	ProviderHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rpc_provider_health_status",
			Help: "Provider health status (1=healthy, 0=unhealthy)",
		},
		[]string{"provider"},
	)

	// TotalCostUSD tracks the cumulative cost incurred per provider
	TotalCostUSD = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_total_cost_usd",
			Help: "Total cost in USD by provider",
		},
		[]string{"provider"},
	)
)
