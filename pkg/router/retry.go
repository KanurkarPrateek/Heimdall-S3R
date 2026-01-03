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
	forcedStates    map[string]string // "open" or "" (normal)
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
		forcedStates:    make(map[string]string),
	}
}

// ExecuteWithRetry executes an RPC request with up to 3 retries and exponential backoff
func (r *RetryHandler) ExecuteWithRetry(ctx context.Context, req *provider.RPCRequest) (*provider.RPCResponse, string, error) {
	var lastErr error
	maxRetries := 3
	backoff := 100 * time.Millisecond

	tried := make(map[string]bool)

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get next healthy provider, excluding already tried ones in this request
		prov, err := r.pool.NextWithExclude(ctx, tried)
		if err != nil {
			return nil, "", fmt.Errorf("failed to select provider: %w", err)
		}

		tried[prov.Name()] = true

		// Check if forced into open state (Demo Chaos)
		if r.forcedStates[prov.Name()] == "open" {
			log.Printf("[CHAOS] Skipping provider %s (Forced Open)", prov.Name())
			continue
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

// GetBreakerStatuses returns the current state of all circuit breakers
func (r *RetryHandler) GetBreakerStatuses() map[string]string {
	statuses := make(map[string]string)
	for name, cb := range r.circuitBreakers {
		state := cb.State().String()
		if r.forcedStates[name] == "open" {
			state = "FORCED OPEN"
		}
		statuses[name] = state
	}
	return statuses
}

// TripProvider manually forces a provider's circuit breaker to open (for demo)
func (r *RetryHandler) TripProvider(name string) {
	r.forcedStates[name] = "open"
	log.Printf("[CHAOS] Provider %s manually TRIPPED", name)
}

// ResetChaos clears all manual overrides
func (r *RetryHandler) ResetChaos() {
	r.forcedStates = make(map[string]string)
	log.Printf("[CHAOS] All manual overrides RESET")
}
