package router

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
