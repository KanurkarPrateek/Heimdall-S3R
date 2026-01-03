package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/metrics"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/provider"
)

const (
	healthKeyPrefix = "health:"
	healthTTL       = 30 * time.Second
)

// HealthMonitor probes providers periodically and stores their status in Redis
type HealthMonitor struct {
	providers []provider.Provider
	redis     *redis.Client
	interval  time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(providers []provider.Provider, redisClient *redis.Client, interval time.Duration) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthMonitor{
		providers: providers,
		redis:     redisClient,
		interval:  interval,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins the background health monitoring
func (m *HealthMonitor) Start() {
	log.Printf("[HEALTH] Starting health monitor with interval %v", m.interval)
	ticker := time.NewTicker(m.interval)

	// Initial check
	m.checkAll()

	go func() {
		for {
			select {
			case <-ticker.C:
				m.checkAll()
			case <-m.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the health monitor
func (m *HealthMonitor) Stop() {
	m.cancel()
}

func (m *HealthMonitor) checkAll() {
	for _, p := range m.providers {
		go m.checkProvider(p)
	}
}

func (m *HealthMonitor) checkProvider(p provider.Provider) {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	status, err := p.CheckHealth(ctx)
	if err != nil {
		log.Printf("[HEALTH] Error checking provider %s: %v", p.Name(), err)
		// Update metrics
		metrics.ProviderHealthStatus.WithLabelValues(p.Name()).Set(0)
		return
	}

	// Update Prometheus metrics
	healthVal := 1.0
	if !status.Healthy {
		healthVal = 0.0
	}
	metrics.ProviderHealthStatus.WithLabelValues(p.Name()).Set(healthVal)

	// Update Redis
	if err := m.updateStatus(p.Name(), status); err != nil {
		log.Printf("[HEALTH] Error updating status in Redis for %s: %v", p.Name(), err)
	}

	if !status.Healthy {
		log.Printf("[HEALTH] Provider %s is UNHEALTHY: %s", p.Name(), status.ErrorMessage)
	}
}

func (m *HealthMonitor) updateStatus(name string, status *provider.HealthStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	key := healthKeyPrefix + name
	err = m.redis.Set(m.ctx, key, data, healthTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set status in Redis: %w", err)
	}

	return nil
}

// GetProviderStatus retrieves the health status of a provider from Redis
func GetProviderStatus(ctx context.Context, redisClient *redis.Client, name string) (*provider.HealthStatus, error) {
	key := healthKeyPrefix + name
	data, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get status from Redis: %w", err)
	}

	var status provider.HealthStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &status, nil
}
