package provider

import (
	"context"
)

// HeliusProvider implements the Provider interface for Helius
type HeliusProvider struct {
	*BaseProvider
}

// NewHeliusProvider creates a new Helius provider
func NewHeliusProvider(url string, costPerRequest float64) Provider {
	return &HeliusProvider{
		BaseProvider: NewBaseProvider("helius", url, costPerRequest),
	}
}

// CheckHealth performs Helius-specific health check
func (h *HeliusProvider) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	// Use base implementation for now
	// Can be customized later with Helius-specific checks
	return h.BaseProvider.CheckHealth(ctx)
}
