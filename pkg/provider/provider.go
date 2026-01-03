package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthStatus represents the health state of a provider
type HealthStatus struct {
	Healthy      bool      `json:"healthy"`
	LastCheck    time.Time `json:"last_check"`
	LatencyMs    int64     `json:"latency_ms"`
	SuccessRate  float64   `json:"success_rate"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Provider interface defines the contract for RPC providers
type Provider interface {
	// Name returns the provider name (e.g., "helius", "alchemy")
	Name() string
	
	// URL returns the provider's RPC endpoint URL
	URL() string
	
	// CostPerRequest returns the cost in USD for each request
	CostPerRequest() float64
	
	// ForwardRequest forwards an RPC request to the provider
	ForwardRequest(ctx context.Context, req *RPCRequest) (*RPCResponse, error)
	
	// CheckHealth performs a health check on the provider
	CheckHealth(ctx context.Context) (*HealthStatus, error)
}

// BaseProvider implements common functionality for all providers
type BaseProvider struct {
	name           string
	url            string
	costPerRequest float64
	client         *http.Client
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name, url string, costPerRequest float64) *BaseProvider {
	return &BaseProvider{
		name:           name,
		url:            url,
		costPerRequest: costPerRequest,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *BaseProvider) Name() string {
	return p.name
}

// URL returns the provider URL
func (p *BaseProvider) URL() string {
	return p.url
}

// CostPerRequest returns the cost per request
func (p *BaseProvider) CostPerRequest() float64 {
	return p.costPerRequest
}

// ForwardRequest forwards an RPC request to the provider
func (p *BaseProvider) ForwardRequest(ctx context.Context, req *RPCRequest) (*RPCResponse, error) {
	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	start := time.Now()
	httpResp, err := p.client.Do(httpReq)
	latency := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Log latency (can be removed or replaced with proper logger)
	_ = latency // Placeholder for metrics

	return &rpcResp, nil
}

// CheckHealth performs a basic health check by calling getHealth
func (p *BaseProvider) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()

	// Create health check request
	healthReq := &RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getHealth",
	}

	// Forward request
	resp, err := p.ForwardRequest(ctx, healthReq)
	latency := time.Since(start)

	status := &HealthStatus{
		LastCheck:   time.Now(),
		LatencyMs:   latency.Milliseconds(),
		SuccessRate: 1.0,
	}

	if err != nil {
		status.Healthy = false
		status.ErrorMessage = err.Error()
		return status, nil
	}

	if resp.Error != nil {
		status.Healthy = false
		status.ErrorMessage = resp.Error.Message
		return status, nil
	}

	status.Healthy = true
	return status, nil
}
