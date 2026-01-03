package router

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/config"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/provider"
)

// CacheHandler handles caching of RPC responses
type CacheHandler struct {
	redis  *redis.Client
	config config.CachingConfig
}

// NewCacheHandler creates a new cache handler
func NewCacheHandler(redisClient *redis.Client, cfg config.CachingConfig) *CacheHandler {
	return &CacheHandler{
		redis:  redisClient,
		config: cfg,
	}
}

// GetCachedResponse attempts to retrieve a cached response for the given request
func (h *CacheHandler) GetCachedResponse(ctx context.Context, req *provider.RPCRequest) (*provider.RPCResponse, error) {
	if !h.config.Enabled {
		return nil, nil
	}

	ttl, exists := h.config.Methods[req.Method]
	if !exists || ttl <= 0 {
		return nil, nil
	}

	key := h.generateKey(req)
	val, err := h.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var resp provider.RPCResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// StoreResponse caches a response for the given request if the method is cacheable
func (h *CacheHandler) StoreResponse(ctx context.Context, req *provider.RPCRequest, resp *provider.RPCResponse) error {
	if !h.config.Enabled {
		return nil
	}

	ttl, exists := h.config.Methods[req.Method]
	if !exists || ttl <= 0 {
		return nil
	}

	key := h.generateKey(req)
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return h.redis.Set(ctx, key, data, ttl).Err()
}

// generateKey creates a unique cache key based on the RPC method and parameters
func (h *CacheHandler) generateKey(req *provider.RPCRequest) string {
	paramsJSON, _ := json.Marshal(req.Params)
	hash := sha256.Sum256(paramsJSON)
	return fmt.Sprintf("rpc:cache:%s:%x", req.Method, hash[:8])
}
