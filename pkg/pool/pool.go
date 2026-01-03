package pool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/health"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/provider"
)

// ProviderPool manages a pool of RPC providers with round-robin selection and health filtering
type ProviderPool struct {
	providers []provider.Provider
	redis     *redis.Client
	current   int
	mu        sync.Mutex
}

// NewProviderPool creates a new provider pool
func NewProviderPool(providers []provider.Provider, redisClient *redis.Client) *ProviderPool {
	return &ProviderPool{
		providers: providers,
		redis:     redisClient,
		current:   0,
	}
}

// Next returns the next provider using a latency-optimized strategy
func (p *ProviderPool) Next(ctx context.Context) (provider.Provider, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// 1. Filter healthy providers
	var healthyProviders []provider.Provider
	for _, prov := range p.providers {
		status, err := health.GetProviderStatus(ctx, p.redis, prov.Name())
		if err != nil || status == nil || status.Healthy {
			healthyProviders = append(healthyProviders, prov)
		}
	}

	// 2. Prioritize discovery: find providers without latency data
	for _, prov := range healthyProviders {
		_, err := p.GetLatency(ctx, prov.Name())
		if err != nil {
			log.Printf("[ROUTING] Discovery: Selected healthy provider without latency data: %s", prov.Name())
			return prov, nil
		}
	}

	// 3. Find provider with lowest latency
	var bestProv provider.Provider
	minLatency := int64(999999)

	for _, prov := range healthyProviders {
		latency, _ := p.GetLatency(ctx, prov.Name())
		if latency > 0 && latency < minLatency {
			minLatency = latency
			bestProv = prov
		}
	}

	// 4. Select provider
	if bestProv != nil {
		log.Printf("[ROUTING] Selected least-latency provider: %s (%dms)", bestProv.Name(), minLatency)
		return bestProv, nil
	}

	// 4. Fallback to round-robin if no latency data
	selected := healthyProviders[p.current%len(healthyProviders)]
	p.current = (p.current + 1) % len(p.providers)
	log.Printf("[ROUTING] Selected healthy provider (round-robin): %s", selected.Name())

	return selected, nil
}

// GetLatency retrieves the latest recorded latency for a provider from Redis
func (p *ProviderPool) GetLatency(ctx context.Context, name string) (int64, error) {
	if p.redis == nil {
		return 0, fmt.Errorf("redis not initialized")
	}
	key := fmt.Sprintf("latency:%s", name)
	val, err := p.redis.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	var latency int64
	fmt.Sscanf(val, "%d", &latency)
	return latency, nil
}

// GetAll returns all providers in the pool
func (p *ProviderPool) GetAll() []provider.Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.providers
}

// Size returns the number of providers in the pool
func (p *ProviderPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.providers)
}

// ForwardRequest forwards a request using the next available provider
// UpdateLatency stores the latest latency of a provider in Redis
func (p *ProviderPool) UpdateLatency(ctx context.Context, name string, latency time.Duration) {
	if p.redis == nil {
		return
	}
	key := fmt.Sprintf("latency:%s", name)
	// Store latency in milliseconds as a string for easy retrieval
	p.redis.Set(ctx, key, fmt.Sprintf("%d", latency.Milliseconds()), 10*time.Minute)
}

func (p *ProviderPool) ForwardRequest(ctx context.Context, req *provider.RPCRequest) (*provider.RPCResponse, string, error) {
	// Get next provider
	prov, err := p.Next(ctx)
	if err != nil {
		return nil, "", err
	}

	// Forward request
	resp, err := prov.ForwardRequest(ctx, req)
	if err != nil {
		return nil, prov.Name(), fmt.Errorf("provider %s failed: %w", prov.Name(), err)
	}

	return resp, prov.Name(), nil
}
