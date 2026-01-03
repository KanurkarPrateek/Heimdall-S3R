# RPC Load Balancer

A smart RPC load balancer for Solana that intelligently routes requests across multiple providers to optimize for cost, latency, and reliability.

## Features (Week 1 - MVP)

✅ **Multi-Provider Routing** (FR-1)
- Support for Helius, Alchemy, and QuickNode providers
- Round-robin routing strategy
- Automatic provider selection
- Request logging with provider information

🚧 **Coming in Week 2-4:**
- Health monitoring and automatic failover (FR-2, FR-3)
- Cost tracking and optimization (FR-4)
- Prometheus metrics and Grafana dashboards
- Circuit breaker pattern
- Docker Compose deployment

## Quick Start

### Prerequisites
- Go 1.21+ installed
- RPC provider API keys (Helius, Alchemy, or QuickNode)

### Setup

1. **Clone the repository**
```bash
cd /Users/prateekkanurkar/Documents/DeNova/rpc-load-balancer
```

2. **Create environment file**
```bash
cp .env.example .env
```

3. **Add your API keys to `.env`**
```bash
HELIUS_API_KEY=your_helius_api_key_here
ALCHEMY_API_KEY=your_alchemy_api_key_here
QUICKNODE_TOKEN=your_quicknode_token_here
```

4. **Load environment variables**
```bash
source .env
```

5. **Build and run**
```bash
go build -o rpc-load-balancer ./cmd/server
./rpc-load-balancer
```

The server will start on `http://localhost:8080`

## Usage

### Send RPC Request

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "getLatestBlockhash"
  }'
```

### Check Health

```bash
curl http://localhost:8080/health
```

### Test Round-Robin Distribution

```bash
# Send 10 requests and watch the logs
for i in {1..10}; do
  curl -X POST http://localhost:8080 \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"getHealth"}' &
done
```

You should see requests distributed evenly across configured providers in the logs.

## Configuration

Edit `config/config.yaml` to:
- Add/remove providers
- Change routing strategy (Week 2+)
- Adjust retry settings (Week 3+)
- Configure health check intervals (Week 2+)

## Project Structure

```
rpc-load-balancer/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── pkg/
│   ├── config/
│   │   └── config.go            # Configuration loader
│   ├── provider/
│   │   ├── provider.go          # Provider interface
│   │   ├── helius.go            # Helius implementation
│   │   ├── alchemy.go           # Alchemy implementation
│   │   └── quicknode.go         # QuickNode implementation
│   ├── pool/
│   │   └── pool.go              # Provider pool & routing
│   └── router/
│       └── handler.go           # HTTP request handler
├── config/
│   └── config.yaml              # Configuration file
├── .env.example                 # Environment variables template
├── .gitignore
├── go.mod
└── README.md
```

## Week 1 Acceptance Criteria

- [x] Support Helius, Alchemy, and QuickNode providers
- [x] Round-robin routing strategy
- [x] Log each routing decision (provider selected)
- [x] HTTP server accepting JSON-RPC requests
- [x] Basic error handling and validation
- [x] Configuration via YAML file
- [ ] Test with 100 req/s (manual testing pending)

## Development

### Run in Development Mode

```bash
# Enable debug logging
go run ./cmd/server
```

### Test Endpoints

```bash
# Test invalid JSON
curl -X POST http://localhost:8080 -d 'invalid-json'

# Test missing method
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1}'

# Test wrong JSON-RPC version
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"1.0","id":1,"method":"getHealth"}'
```

## Next Steps

### Week 2: Health Monitoring (FR-2)
- Redis integration for health cache
- Background health checking (every 5 seconds)
- Exclude unhealthy providers from rotation
- Prometheus health metrics

### Week 3: Automatic Failover (FR-3)
- Retry logic with exponential backoff
- Circuit breaker pattern
- Integration tests

### Week 4: Observability & Cost Tracking (FR-4)
- Full Prometheus instrumentation
- Grafana dashboard (5 panels)
- Cost tracking per provider
- Docker Compose deployment

## License

MIT

---

**Status**: ✅ Week 1 Complete - Multi-Provider Routing Working!
# Heimdall-S3R
