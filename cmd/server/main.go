package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/config"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/health"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/pool"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/provider"
	"github.com/kanurkarprateek/rpc-load-balancer/pkg/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting RPC Load Balancer...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found or error loading it. Relying on environment variables.")
	}

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Loaded configuration with %d providers", len(cfg.Providers))

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.URL,
		DB:   cfg.Redis.DB,
	})

	// Test Redis connection
	ctx_redis, cancel_redis := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel_redis()
	if err := redisClient.Ping(ctx_redis).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis at %s: %v. Health monitoring may not work correctly.", cfg.Redis.URL, err)
	} else {
		log.Printf("Connected to Redis at %s", cfg.Redis.URL)
	}

	// Initialize providers
	providers := make([]provider.Provider, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		var prov provider.Provider
		switch p.Name {
		case "helius":
			prov = provider.NewHeliusProvider(p.URL, p.CostPerRequest)
		case "alchemy":
			prov = provider.NewAlchemyProvider(p.URL, p.CostPerRequest)
		case "quicknode":
			prov = provider.NewQuickNodeProvider(p.URL, p.CostPerRequest)
		default:
			log.Printf("Warning: unknown provider type '%s', using base provider", p.Name)
			prov = provider.NewBaseProvider(p.Name, p.URL, p.CostPerRequest)
		}
		providers = append(providers, prov)

		// Log masked URL for debugging
		url := prov.URL()
		maskedURL := url
		if len(url) > 20 {
			maskedURL = url[:20] + "..."
		}
		log.Printf("Initialized provider: %s (url: %s, cost: $%.6f/req)", prov.Name(), maskedURL, prov.CostPerRequest())
	}

	// Create provider pool
	providerPool := pool.NewProviderPool(providers, redisClient)
	log.Printf("Provider pool created with %d providers", providerPool.Size())

	// Initialize RetryHandler with circuit breakers for each provider
	providerNames := make([]string, 0, len(providers))
	for _, p := range providers {
		providerNames = append(providerNames, p.Name())
	}
	retryHandler := router.NewRetryHandler(providerPool, providerNames)

	// Start health monitor
	healthMonitor := health.NewHealthMonitor(providers, redisClient, cfg.Health.CheckInterval)
	healthMonitor.Start()
	defer healthMonitor.Stop()

	// Initialize CacheHandler
	cacheHandler := router.NewCacheHandler(redisClient, cfg.Caching)

	// Create HTTP handler
	handler := router.NewHandler(providerPool, retryHandler, cacheHandler)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode) // Use gin.DebugMode for development
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(customLogger())

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Register routes
	r.POST("/", handler.HandleRPC)                   // Main RPC endpoint
	r.GET("/health", handler.HealthCheck)            // Health check endpoint
	r.GET("/api/v1/status", handler.GetSystemStatus) // Dashboard status API
	r.POST("/api/v1/chaos/trip", handler.TripProvider)
	r.POST("/api/v1/chaos/reset", handler.ResetChaos)
	r.POST("/api/v1/test-rpc", handler.TestRPC)      // Test RPC endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler())) // Real Prometheus metrics endpoint

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting HTTP server on %s", addr)

	// Start server in goroutine
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("✓ RPC Load Balancer is running!")
	log.Printf("  - RPC Endpoint: http://localhost%s/", addr)
	log.Printf("  - Health Check: http://localhost%s/health", addr)
	log.Printf("  - Metrics: http://localhost%s/metrics", addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// customLogger is a custom Gin middleware for logging
func customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		log.Printf("[HTTP] %s %s | %d | %v", method, path, statusCode, latency)
	}
}

// placeholderMetrics is a placeholder for Prometheus metrics endpoint
func placeholderMetrics(c *gin.Context) {
	c.String(200, "# Prometheus metrics will be added in Week 4\n")
}
