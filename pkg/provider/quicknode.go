package provider

import (
	"context"
)

// QuickNodeProvider implements the Provider interface for QuickNode
type QuickNodeProvider struct {
	*BaseProvider
}

// NewQuickNodeProvider creates a new QuickNode provider
func NewQuickNodeProvider(url string, costPerRequest float64) Provider {
	return &QuickNodeProvider{
		BaseProvider: NewBaseProvider("quicknode", url, costPerRequest),
	}
}

// CheckHealth performs QuickNode-specific health check
func (q *QuickNodeProvider) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	// Use base implementation for now
	// Can be customized later with QuickNode-specific checks
	return q.BaseProvider.CheckHealth(ctx)
}
