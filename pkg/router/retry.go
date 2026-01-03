package router

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kanurkarprateek/rpc-load-balancer/pkg/pool"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/provider"
	"github.com/sony/gobreaker"
)

// RetryHandler handles requests with retries and circuit breaking
type RetryHandler struct {
	pool            *pool.ProviderPool
	circuitBreakers map[string]*gobreaker.CircuitBreaker
}

// NewRetryHandler creates a new retry handler
func NewRetryHandler(providerPool *pool.ProviderPool, providerNames []string) *RetryHandler {
	cbs := make(map[string]*gobreaker.CircuitBreaker)

	for _, name := range providerNames {
		st := gobreaker.Settings{
			Name:        name,
			MaxRequests: 5,
			Interval:    time.Minute,
			Timeout:     time.Minute,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 5
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				log.Printf("[CIRCUIT-BREAKER] Provider %s state changed from %s to %s", name, from, to)
			},
		}
		cbs[name] = gobreaker.NewCircuitBreaker(st)
	}

	return &RetryHandler{
		pool:            providerPool,
		circuitBreakers: cbs,
	}
}

// ExecuteWithRetry executes an RPC request with up to 3 retries and exponential backoff
func (r *RetryHandler) ExecuteWithRetry(ctx context.Context, req *provider.RPCRequest) (*provider.RPCResponse, string, error) {
	var lastErr error
	maxRetries := 3
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get next healthy provider
		prov, err := r.pool.Next(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to select provider: %w", err)
		}

		cb, ok := r.circuitBreakers[prov.Name()]
		if !ok {
			// Fallback if CB not initialized for some reason
			resp, err := prov.ForwardRequest(ctx, req)
			if err == nil {
				return resp, prov.Name(), nil
			}
			lastErr = err
		} else {
			// Execute through circuit breaker
			result, err := cb.Execute(func() (interface{}, error) {
				return prov.ForwardRequest(ctx, req)
			})

			if err == nil {
				return result.(*provider.RPCResponse), prov.Name(), nil
			}
			lastErr = err
		}

		log.Printf("[RETRY] Attempt %d failed for provider %s: %v", attempt+1, prov.Name(), lastErr)

		// Exponential backoff
		if attempt < maxRetries-1 {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return nil, "", ctx.Err()
			}
		}
	}

	return nil, "", fmt.Errorf("max retries exceeded, last error: %v", lastErr)
}
