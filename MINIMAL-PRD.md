# Minimal Product Requirements
## RPC Load Balancer - Core Features Only

**Version:** 1.0 Minimal  
**Timeline:** 2-3 weeks  
**Goal:** Ship the absolute minimum to get value

---

## Problem We're Solving

Solana developers need:
1. **No single point of failure** - one provider goes down = app goes down
2. **Cost visibility** - "Why is my RPC bill so high?"
3. **Simple setup** - should take <1 hour, not days

**That's it.** Everything else is nice-to-have.

---

## Core Features (Week 1-3)

### 1. Multi-Provider Routing ⭐⭐⭐
**What**: Send requests to 2 providers (Helius + Alchemy)

**How**: Simple round-robin (50/50 split)

**Done when**:
- [ ] Configure 2 providers in YAML
- [ ] Requests alternate between them
- [ ] Works with 100 req/s

---

### 2. Basic Health Checks ⭐⭐⭐
**What**: Don't send requests to dead providers

**How**: Ping `getHealth` every 30 seconds

**Done when**:
- [ ] Unhealthy provider is skipped
- [ ] Automatically recovers when it comes back
- [ ] Takes <30 seconds to detect failure

---

### 3. Simple Observability ⭐⭐
**What**: See what's happening (console logs + basic metrics)

**How**: 
- Print request logs to stdout
- Single Grafana panel: "Requests per provider"

**Done when**:
- [ ] Can see which provider handled each request
- [ ] Can see if one provider is getting more traffic
- [ ] Grafana dashboard exists (1 panel is fine)

---

## What We're NOT Building

Delete these from scope:

- ❌ Cost tracking (can add later)
- ❌ Retry logic (just fail fast for now)
- ❌ Circuit breakers (YAGNI)
- ❌ Multiple routing strategies (round-robin is enough)
- ❌ Advanced dashboards (1 panel is fine)
- ❌ Rate limiting (providers handle this)
- ❌ Request caching (optimization for later)
- ❌ ML/AI anything (Phase 2)
- ❌ Kubernetes (Docker Compose only)

---

## Success Criteria

**Before shipping, we must have**:

1. ✅ Routes to 2 providers (Helius, Alchemy)
2. ✅ Skips unhealthy providers automatically
3. ✅ Handles 100 req/s without crashing
4. ✅ Deploy with `docker-compose up`
5. ✅ Basic Grafana dashboard showing traffic

**Nice to have** (only if easy):
- Handles 1000 req/s
- Detects failure in <10 seconds
- Pretty error messages

---

## Minimal Tech Stack

```yaml
Language: Go
Web Framework: net/http (standard library - skip Gin for now)
Health Storage: In-memory map (skip Redis)
Metrics: Prometheus
Dashboard: Grafana (1 panel)
Deployment: Docker Compose
Config: YAML file
```

**Why so minimal?**
- No external dependencies = faster to build
- Can refactor to Redis later if needed
- Standard library is plenty for MVP

---

## 3-Week Timeline

### Week 1: Routing Only
**Goal**: Proxy requests to 2 providers

**Tasks**:
- [ ] HTTP server accepts JSON-RPC
- [ ] Forward to Helius URL
- [ ] Forward to Alchemy URL  
- [ ] Round-robin between them
- [ ] Log each request

**Test**: `curl` sending 10 requests → see 5 to Helius, 5 to Alchemy

---

### Week 2: Health Checks
**Goal**: Skip dead providers

**Tasks**:
- [ ] Background goroutine pings providers every 30s
- [ ] Store health in memory (`map[string]bool`)
- [ ] Router checks health before forwarding
- [ ] Auto-recover when provider comes back

**Test**: Kill Helius API → all requests go to Alchemy → restore → traffic resumes

---

### Week 3: Deploy + Observe
**Goal**: Ship it

**Tasks**:
- [ ] Prometheus metrics endpoint
- [ ] Grafana dashboard (single panel: requests by provider)
- [ ] Dockerfile
- [ ] docker-compose.yml (balancer + grafana + prometheus)
- [ ] README with setup instructions

**Test**: Fresh laptop → `docker-compose up` → working in <30 min

---

## Minimal Architecture

```
┌─────────┐
│  Client │
└────┬────┘
     │ JSON-RPC
     ▼
┌──────────────────┐
│  Load Balancer   │
│  (Go HTTP Server)│
├──────────────────┤
│  Health Map      │  ← in-memory
│  (Helius: ✓)    │
│  (Alchemy: ✓)   │
└─────┬──────┬─────┘
      │      │
      ▼      ▼
   Helius  Alchemy
```

**That's it.** No Redis, no database, no complexity.

---

## File Structure (Minimal)

```
rpc-load-balancer/
├── main.go                  # Everything in one file for now
├── config.yaml              # Provider URLs
├── docker-compose.yml       # Balancer + Grafana + Prometheus
├── Dockerfile
├── grafana-dashboard.json   # Single panel
└── README.md
```

**Start with 1 file.** Refactor to packages later if needed.

---

## Configuration (Minimal)

```yaml
# config.yaml
providers:
  - name: helius
    url: https://mainnet.helius-rpc.com/?api-key=YOUR_KEY
  
  - name: alchemy
    url: https://solana-mainnet.g.alchemy.com/v2/YOUR_KEY

server:
  port: 8080

health:
  check_interval: 30s
```

**No fancy features.** Just the basics.

---

## Code Sketch (Single File)

```go
// main.go
package main

import (
    "net/http"
    "sync"
    "time"
)

// In-memory health tracking
var (
    health = make(map[string]bool)
    healthMu sync.RWMutex
    currentProvider = 0
)

// Providers from config
var providers = []Provider{
    {Name: "helius", URL: "https://..."},
    {Name: "alchemy", URL: "https://..."},
}

func main() {
    // Start health checker
    go healthCheckLoop()
    
    // Start HTTP server
    http.HandleFunc("/", handleRPC)
    http.HandleFunc("/metrics", handleMetrics)  // Prometheus
    http.ListenAndServe(":8080", nil)
}

func handleRPC(w http.ResponseWriter, r *http.Request) {
    // 1. Get next healthy provider (round-robin)
    provider := getNextHealthyProvider()
    
    // 2. Forward request
    resp := forwardRequest(provider, r.Body)
    
    // 3. Return response
    w.Write(resp)
    
    // 4. Log
    log.Printf("Request → %s", provider.Name)
}

func healthCheckLoop() {
    for {
        for _, p := range providers {
            healthy := checkHealth(p)
            healthMu.Lock()
            health[p.Name] = healthy
            healthMu.Unlock()
        }
        time.Sleep(30 * time.Second)
    }
}
```

**~200 lines total.** That's the whole MVP.

---

## Testing (Minimal)

### Manual Test
```bash
# 1. Start stack
docker-compose up

# 2. Send requests
for i in {1..10}; do
  curl -X POST http://localhost:8080 \
    -d '{"jsonrpc":"2.0","id":1,"method":"getHealth"}'
done

# 3. Check logs
# Should see: 5 → Helius, 5 → Alchemy

# 4. Check Grafana
open http://localhost:3000
# Should see: bar chart with 2 bars
```

### Failover Test
```bash
# 1. Invalidate Helius key (make it fail)
# 2. Send requests
# 3. Should see: 10 → Alchemy (all traffic failed over)
```

**No fancy testing framework.** Just manual verification.

---

## Definition of Done

Before we call this "shipped":

- [ ] Two people can deploy it successfully
- [ ] Handles provider failure gracefully  
- [ ] Grafana dashboard shows data
- [ ] README explains setup in <10 steps
- [ ] Works on localhost

**That's it.**

---

## What Happens After?

Once this works, we can add (in order of value):

1. **Cost tracking** - add counter for spend
2. **Retry logic** - 3 attempts before failing
3. **Better routing** - least-latency instead of round-robin
4. **Redis** - persistent health state
5. **More providers** - QuickNode, etc.
6. **Kubernetes** - scale beyond 1 instance

But **don't add these until core is proven**.

---

## Open Questions

| Question | Answer | Status |
|----------|--------|--------|
| Do we need retry logic in v1? | No, fail fast | ✅ Decided |
| Should we use Redis? | No, in-memory is fine | ✅ Decided |
| Need cost tracking? | No, add later | ✅ Decided |
| How many providers? | 2 (Helius + Alchemy) | ✅ Decided |

---

## Ship Checklist

Week 3 final tasks:

- [ ] Code compiles and runs
- [ ] Docker Compose starts without errors
- [ ] Grafana dashboard loads
- [ ] Test with 2 fresh users (not you)
- [ ] README has clear setup steps
- [ ] Push to GitHub
- [ ] **Ship it** 🚀

---

**Remember: Perfect is the enemy of shipped. Let's build the minimum and iterate.**
