package router

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/health"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/metrics"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/pool"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/provider"
)

// Handler handles HTTP RPC requests
type Handler struct {
	pool         *pool.ProviderPool
	retryHandler *RetryHandler
	cacheHandler *CacheHandler
}

// NewHandler creates a new request handler
func NewHandler(pool *pool.ProviderPool, retryHandler *RetryHandler, cacheHandler *CacheHandler) *Handler {
	return &Handler{
		pool:         pool,
		retryHandler: retryHandler,
		cacheHandler: cacheHandler,
	}
}

// HandleRPC handles incoming JSON-RPC requests
func (h *Handler) HandleRPC(c *gin.Context) {
	start := time.Now()

	// Parse JSON-RPC request
	var rpcReq provider.RPCRequest
	if err := c.ShouldBindJSON(&rpcReq); err != nil {
		log.Printf("[ERROR] Invalid JSON-RPC request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32700,
				"message": "Parse error: invalid JSON",
			},
			"id": nil,
		})
		return
	}

	// Validate JSON-RPC request
	if rpcReq.JSONRPC != "2.0" {
		log.Printf("[ERROR] Invalid JSON-RPC version: %s", rpcReq.JSONRPC)
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request: jsonrpc must be 2.0",
			},
			"id": rpcReq.ID,
		})
		return
	}

	if rpcReq.Method == "" {
		log.Printf("[ERROR] Missing method in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request: method is required",
			},
			"id": rpcReq.ID,
		})
		return
	}

	// Check Cache (FR-7)
	if h.cacheHandler != nil {
		cachedResp, err := h.cacheHandler.GetCachedResponse(c.Request.Context(), &rpcReq)
		if err == nil && cachedResp != nil {
			log.Printf("[CACHE] Hit for method=%s id=%v", rpcReq.Method, rpcReq.ID)
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}

	// Forward request with retry and circuit breaking
	resp, providerName, err := h.retryHandler.ExecuteWithRetry(c.Request.Context(), &rpcReq)

	latency := time.Since(start)
	if err != nil {

		log.Printf("[ERROR] Failed to forward request: %v", err)

		// Record error metrics
		metrics.RequestsTotal.WithLabelValues(providerName, rpcReq.Method, "error").Inc()

		c.JSON(http.StatusInternalServerError, gin.H{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32603,
				"message": fmt.Sprintf("Internal error: %v", err),
			},
			"id": rpcReq.ID,
		})
		return
	}

	// Record success metrics
	metrics.RequestsTotal.WithLabelValues(providerName, rpcReq.Method, "success").Inc()
	metrics.RequestDuration.WithLabelValues(providerName).Observe(latency.Seconds())

	// Record cost (FR-4)
	// Find provider in pool to get its cost
	for _, p := range h.pool.GetAll() {
		if p.Name() == providerName {
			metrics.TotalCostUSD.WithLabelValues(providerName).Add(p.CostPerRequest())
			break
		}
	}

	// Log request details
	log.Printf("[REQUEST] method=%s provider=%s latency=%v", rpcReq.Method, providerName, latency)

	// Update latency in Redis for routing optimization (Phase 2)
	h.pool.UpdateLatency(c.Request.Context(), providerName, latency)

	// Store in Cache (FR-7)
	if h.cacheHandler != nil {
		h.cacheHandler.StoreResponse(c.Request.Context(), &rpcReq, resp)
	}

	// Return response
	c.JSON(http.StatusOK, resp)
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	providerCount := h.pool.Size()
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"providers": providerCount,
		"timestamp": time.Now().Unix(),
	})
}

// GetSystemStatus returns detailed status for all providers
func (h *Handler) GetSystemStatus(c *gin.Context) {
	providers := h.pool.GetAll()
	breakerStatuses := h.retryHandler.GetBreakerStatuses()

	type ProviderStatus struct {
		Name         string  `json:"name"`
		Healthy      bool    `json:"healthy"`
		Latency      int64   `json:"latency_ms"`
		BreakerState string  `json:"breaker_state"`
		Cost         float64 `json:"cost_per_req"`
	}

	var statusList []ProviderStatus
	for _, p := range providers {
		// Get health from Redis
		healthStatus, _ := health.GetProviderStatus(c.Request.Context(), h.pool.GetRedis(), p.Name())
		isHealthy := healthStatus != nil && healthStatus.Healthy

		// Get latency from Redis
		latency, _ := h.pool.GetLatency(c.Request.Context(), p.Name())

		statusList = append(statusList, ProviderStatus{
			Name:         p.Name(),
			Healthy:      isHealthy,
			Latency:      latency,
			BreakerState: breakerStatuses[p.Name()],
			Cost:         p.CostPerRequest(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": statusList,
		"timestamp": time.Now().Unix(),
	})
}

// TripProvider handles manual circuit breaker tripping for demo
func (h *Handler) TripProvider(c *gin.Context) {
	providerName := c.Query("provider")
	if providerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider name is required"})
		return
	}
	h.retryHandler.TripProvider(providerName)
	c.JSON(http.StatusOK, gin.H{"status": "tripped", "provider": providerName})
}

// ResetChaos handles clearing manual overrides for demo
func (h *Handler) ResetChaos(c *gin.Context) {
	h.retryHandler.ResetChaos()
	c.JSON(http.StatusOK, gin.H{"status": "reset"})
}

// TestRPC fires a dummy getSlot request through the system to show it in action
func (h *Handler) TestRPC(c *gin.Context) {
	req := provider.RPCRequest{
		JSONRPC: "2.0",
		ID:      time.Now().Unix(),
		Method:  "getSlot",
	}

	start := time.Now()
	resp, providerName, err := h.retryHandler.ExecuteWithRetry(c.Request.Context(), &req)
	latency := time.Since(start)

	if err != nil {
		metrics.RequestsTotal.WithLabelValues(providerName, req.Method, "error").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record success metrics & latency
	metrics.RequestsTotal.WithLabelValues(providerName, req.Method, "success").Inc()
	metrics.RequestDuration.WithLabelValues(providerName).Observe(latency.Seconds())
	h.pool.UpdateLatency(c.Request.Context(), providerName, latency)

	// Record cost
	for _, p := range h.pool.GetAll() {
		if p.Name() == providerName {
			metrics.TotalCostUSD.WithLabelValues(providerName).Add(p.CostPerRequest())
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": providerName,
		"latency":  latency.String(),
		"response": resp,
	})
}
