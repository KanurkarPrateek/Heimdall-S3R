package provider

import (
	"context"
)

// AlchemyProvider implements the Provider interface for Alchemy
type AlchemyProvider struct {
	*BaseProvider
}

// NewAlchemyProvider creates a new Alchemy provider
func NewAlchemyProvider(url string, costPerRequest float64) Provider {
	return &AlchemyProvider{
		BaseProvider: NewBaseProvider("alchemy", url, costPerRequest),
	}
}

// CheckHealth performs Alchemy-specific health check
func (a *AlchemyProvider) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	// Use base implementation for now
	// Can be customized later with Alchemy-specific checks
	return a.BaseProvider.CheckHealth(ctx)
}
