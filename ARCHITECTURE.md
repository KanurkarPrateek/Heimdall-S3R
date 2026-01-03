# RPC Load Balancer - Architecture Guide

**Version:** 1.0  
**Date:** January 3, 2026  
**Purpose:** Comprehensive architectural overview with diagrams

---

## 1. What Is It & What's Its Use?

### System Overview

The **RPC Load Balancer** is a smart proxy that sits between your Solana application and multiple RPC providers, intelligently routing requests to optimize for cost, latency, and reliability.

```mermaid
graph TB
    subgraph "Your Infrastructure"
        APP[Solana dApp/Service]
        LB[RPC Load Balancer<br/>★ Smart Routing<br/>★ Health Monitoring<br/>★ Cost Tracking]
    end
    
    subgraph "RPC Providers"
        H[Helius<br/>$0.0001/req]
        A[Alchemy<br/>$0.00012/req]
        Q[QuickNode<br/>$0.00015/req]
    end
    
    subgraph "Observability"
        G[Grafana Dashboard<br/>★ Live Metrics<br/>★ Cost Tracking<br/>★ Health Status]
    end
    
    APP -->|JSON-RPC Request| LB
    LB -->|Routes to Healthy| H
    LB -->|Failover| A
    LB -->|Backup| Q
    LB -->|Metrics| G
    
    style LB fill:#4CAF50,stroke:#2E7D32,stroke-width:3px,color:#fff
    style APP fill:#2196F3,stroke:#1565C0,stroke-width:2px,color:#fff
    style G fill:#FF9800,stroke:#E65100,stroke-width:2px,color:#fff
```

### Core Value Propositions

| Feature | Benefit | Example |
|---------|---------|---------|
| **Multi-Provider** | No vendor lock-in | Switch from Helius to Alchemy instantly |
| **Auto-Failover** | Zero downtime | Helius outage? Requests auto-route to Alchemy |
| **Cost Optimization** | 30-50% savings | Route to cheapest available provider |
| **Observability** | Full visibility | See exactly where your RPC budget goes |

---

## 2. How Users Will Use It

### User Journey: DevOps Engineer Deployment

```mermaid
sequenceDiagram
    actor DevOps as DevOps Engineer
    participant Laptop as Local Machine
    participant Docker as Docker Compose
    participant LB as Load Balancer
    participant Grafana as Grafana Dashboard
    participant App as Solana dApp
    
    Note over DevOps,App: Day 1: Setup (30 minutes)
    
    DevOps->>Laptop: 1. Clone repository
    DevOps->>Laptop: 2. Create .env file with API keys
    DevOps->>Docker: 3. docker-compose up -d
    Docker->>LB: Start Load Balancer
    Docker->>Grafana: Start Grafana + Prometheus
    
    DevOps->>Grafana: 4. Open http://localhost:3000
    Grafana-->>DevOps: Show dashboard (no data yet)
    
    Note over DevOps,App: Day 1: Testing
    
    DevOps->>App: 5. Update app config to use Load Balancer
    Note right of App: Change RPC URL to<br/>http://localhost:8080
    
    DevOps->>LB: 6. Send test request
    LB->>Helius: Route to Helius
    Helius-->>LB: Response
    LB-->>DevOps: Success ✓
    
    DevOps->>Grafana: 7. Check dashboard
    Grafana-->>DevOps: Show metrics updating!
    
    Note over DevOps,App: Day 2+: Production Use
    
    App->>LB: Normal traffic
    LB->>Helius: 50% of requests
    LB->>Alchemy: 50% of requests
    
    Note over DevOps: Monitor costs weekly<br/>Receive alerts if provider down
```

### User Flow: Handling Provider Outage

```mermaid
stateDiagram-v2
    [*] --> NormalOperation
    
    NormalOperation: All Providers Healthy
    note right of NormalOperation
        Helius: ✓ Healthy
        Alchemy: ✓ Healthy
        Traffic: 50/50 split
    end note
    
    NormalOperation --> DetectFailure: Helius goes down
    
    DetectFailure: Health Monitor Detects Issue
    note right of DetectFailure
        3 consecutive failures
        ~10 seconds to detect
    end note
    
    DetectFailure --> AutoFailover: Mark Helius as DOWN
    
    AutoFailover: Route All Traffic to Alchemy
    note right of AutoFailover
        100% to Alchemy
        No user impact!
        Alert sent to Slack
    end note
    
    AutoFailover --> Monitoring: Continue health checks
    
    Monitoring: Wait for Helius Recovery
    note right of Monitoring
        Probe every 5 seconds
        Circuit breaker ready
    end note
    
    Monitoring --> GradualRecovery: Helius back online
    
    GradualRecovery: Circuit Breaker Closes
    note right of GradualRecovery
        Resume 50/50 split
        Normal operation restored
    end note
    
    GradualRecovery --> NormalOperation
```

### Daily Operations

```mermaid
graph LR
    A[Morning: Check Dashboard] --> B{Costs Normal?}
    B -->|Yes| C[Continue Monitoring]
    B -->|No| D[Investigate Spike]
    
    C --> E[Weekly Review]
    D --> F[Adjust Provider Mix]
    F --> E
    
    E --> G{Need Changes?}
    G -->|Yes| H[Update config.yaml]
    G -->|No| I[Done]
    
    H --> J[docker-compose restart]
    J --> I
    
    style A fill:#4CAF50,color:#fff
    style E fill:#2196F3,color:#fff
    style I fill:#FF9800,color:#fff
```

---

## 3. What Problem It Solves (Architecturally)

### Problem 1: Single Point of Failure

**Before (No Load Balancer):**

```mermaid
graph LR
    A[Your dApp] -->|All Traffic| B[Helius RPC]
    
    B -.->|Outage| X[❌ Your App Down]
    
    style B fill:#f44336,color:#fff
    style X fill:#D32F2F,color:#fff
```

**Issue**: If Helius goes down → your entire application is unavailable

---

**After (With Load Balancer):**

```mermaid
graph TB
    A[Your dApp] -->|All Traffic| LB[Load Balancer]
    
    LB -->|Primary| H[Helius RPC]
    LB -->|Failover| AL[Alchemy RPC]
    LB -->|Backup| Q[QuickNode RPC]
    
    H -.->|Outage| X[❌ Helius Down]
    X -.->|Auto-Reroute| AL
    AL -->|✓| A
    
    style LB fill:#4CAF50,color:#fff
    style AL fill:#2196F3,color:#fff
    style X fill:#f44336,color:#fff
```

**Solution**: Automatic failover to healthy providers → zero downtime

---

### Problem 2: Cost Opacity & Waste

**Before:**

```mermaid
graph TD
    A[Get Bill: $15,000] --> B{Why So High?}
    B --> C[Manual Log Analysis]
    C --> D[Excel Spreadsheets]
    D --> E[Hours of Work]
    E --> F[Still Not Sure!]
    
    style A fill:#f44336,color:#fff
    style F fill:#f44336,color:#fff
```

---

**After:**

```mermaid
graph TB
    subgraph "Real-Time Cost Dashboard"
        A[Grafana Panel]
        B[Helius: $8,500<br/>12M requests]
        C[Alchemy: $5,200<br/>8M requests]
        D[Total: $13,700<br/>Saved $1,300!]
    end
    
    A --> B
    A --> C
    A --> D
    
    E[Engineer] -->|Glance at Dashboard| A
    E --> F[Instant Insight!]
    
    style D fill:#4CAF50,color:#fff
    style F fill:#4CAF50,color:#fff
```

**Solution**: Real-time cost visibility + automatic routing to cheaper providers

---

### Problem 3: Manual Provider Management

**Before (Manual):**

```mermaid
sequenceDiagram
    participant Engineer
    participant App
    participant Helius
    
    Helius->>Engineer: Maintenance window notice
    Engineer->>App: 1. Update RPC URL to Alchemy
    Engineer->>App: 2. Deploy new config
    Note over Engineer,App: 2 hours of work
    
    Helius->>Engineer: Maintenance complete
    Engineer->>App: 3. Revert RPC URL to Helius
    Engineer->>App: 4. Deploy again
    Note over Engineer,App: Another 2 hours
    
    Note right of Engineer: 4 hours wasted<br/>Manual, error-prone
```

---

**After (Automated):**

```mermaid
sequenceDiagram
    participant Engineer
    participant LB as Load Balancer
    participant Helius
    participant Alchemy
    
    Helius->>LB: Starts returning errors
    LB->>LB: Detect unhealthy (10 sec)
    LB->>Alchemy: Auto-route traffic
    LB->>Engineer: Slack alert: "Helius degraded"
    
    Note over Engineer: Aware but no action needed
    
    Helius->>LB: Recovery detected
    LB->>Helius: Resume normal traffic
    LB->>Engineer: Slack: "Helius recovered"
    
    Note right of Engineer: 0 hours work<br/>Automatic, reliable
```

**Solution**: Automated health monitoring and failover → no manual intervention

---

## 4. How We're Going to Build It - Implementation Architecture

### Layered Architecture

```mermaid
graph TB
    subgraph "Layer 1: HTTP Interface"
        HTTP[Gin HTTP Server<br/>Port 8080]
        Handler[Request Handler]
    end
    
    subgraph "Layer 2: Routing Logic"
        Pool[Provider Pool<br/>Round-Robin]
        Retry[Retry Handler<br/>Exponential Backoff]
        CB[Circuit Breaker]
    end
    
    subgraph "Layer 3: Provider Abstraction"
        P1[Helius Provider]
        P2[Alchemy Provider]
        P3[QuickNode Provider]
    end
    
    subgraph "Layer 4: Health Monitor"
        HM[Health Monitor<br/>Background Goroutine]
        Redis[(Redis Cache<br/>5s TTL)]
    end
    
    subgraph "Layer 5: Observability"
        Prom[Prometheus<br/>Metrics]
        Graf[Grafana<br/>Dashboard]
    end
    
    HTTP --> Handler
    Handler --> Retry
    Retry --> CB
    CB --> Pool
    Pool --> P1
    Pool --> P2
    Pool --> P3
    
    HM -->|Probe| P1
    HM -->|Probe| P2
    HM -->|Probe| P3
    HM -->|Update| Redis
    Pool -->|Check Health| Redis
    
    Handler -->|Record| Prom
    Prom --> Graf
    
    style HTTP fill:#2196F3,color:#fff
    style Pool fill:#4CAF50,color:#fff
    style HM fill:#FF9800,color:#fff
    style Prom fill:#9C27B0,color:#fff
```

---

### Data Flow: Single Request Journey

```mermaid
sequenceDiagram
    participant Client
    participant Handler as Request Handler
    participant Pool as Provider Pool
    participant Redis
    participant Circuit as Circuit Breaker
    participant Provider as Helius RPC
    participant Prom as Prometheus
    
    Client->>Handler: POST /<br/>{"method":"getLatestBlockhash"}
    
    Handler->>Pool: Get next provider
    Pool->>Redis: Check health status
    Redis-->>Pool: Helius: healthy ✓
    Pool-->>Handler: Use Helius
    
    Handler->>Circuit: Execute request
    Circuit->>Provider: Forward RPC call
    Provider-->>Circuit: Response
    Circuit-->>Handler: Success
    
    Handler->>Prom: Record metrics:<br/>• latency: 120ms<br/>• status: success<br/>• cost: $0.0001
    
    Handler-->>Client: Return response
    
    Note over Handler,Prom: Total overhead: ~5ms
```

---

### Component Interaction Map

```mermaid
graph TB
    subgraph "Core Components"
        Main[main.go<br/>Entry Point]
        Config[config.yaml<br/>Configuration]
    end
    
    subgraph "pkg/router"
        H[handler.go<br/>HTTP Endpoints]
        R[retry.go<br/>Retry Logic]
    end
    
    subgraph "pkg/pool"
        Pool[pool.go<br/>Provider Selection]
    end
    
    subgraph "pkg/provider"
        PI[provider.go<br/>Interface]
        PH[helius.go]
        PA[alchemy.go]
        PQ[quicknode.go]
    end
    
    subgraph "pkg/health"
        HM[monitor.go<br/>Health Checks]
    end
    
    subgraph "pkg/metrics"
        PM[prometheus.go<br/>Metrics]
    end
    
    Main --> Config
    Main --> H
    Main --> HM
    
    H --> R
    R --> Pool
    Pool --> PI
    PI --> PH
    PI --> PA
    PI --> PQ
    
    HM --> PI
    HM --> Redis[(Redis)]
    Pool --> Redis
    
    H --> PM
    PM --> Prom[(Prometheus)]
    
    style Main fill:#2196F3,color:#fff
    style Pool fill:#4CAF50,color:#fff
    style HM fill:#FF9800,color:#fff
```

---

### Deployment Architecture

```mermaid
graph TB
    subgraph "Docker Compose Stack"
        subgraph "Container 1: Load Balancer"
            LB[Go Application<br/>Port 8080]
            LBM[Prometheus Endpoint<br/>Port 8080/metrics]
        end
        
        subgraph "Container 2: Redis"
            R[Redis Server<br/>Port 6379]
        end
        
        subgraph "Container 3: Prometheus"
            P[Prometheus<br/>Port 9090]
        end
        
        subgraph "Container 4: Grafana"
            G[Grafana<br/>Port 3000]
        end
    end
    
    subgraph "External Services"
        H[Helius API]
        A[Alchemy API]
        Q[QuickNode API]
    end
    
    LB --> R
    LBM --> P
    P --> G
    
    LB --> H
    LB --> A
    LB --> Q
    
    Client[Your Application] --> LB
    User[You] --> G
    
    style LB fill:#4CAF50,color:#fff
    style R fill:#DC382D,color:#fff
    style P fill:#E6522C,color:#fff
    style G fill:#F46800,color:#fff
```

---

### Build Timeline & Phases

```mermaid
gantt
    title 4-Week Implementation Timeline
    dateFormat  YYYY-MM-DD
    section Week 1
    Project Setup           :w1-1, 2026-01-06, 1d
    Provider Interface      :w1-2, after w1-1, 2d
    Provider Pool           :w1-3, after w1-2, 2d
    HTTP Server            :w1-4, after w1-3, 2d
    
    section Week 2
    Redis Integration      :w2-1, 2026-01-13, 2d
    Health Monitor         :w2-2, after w2-1, 3d
    Health Metrics         :w2-3, after w2-2, 2d
    
    section Week 3
    Retry Logic            :w3-1, 2026-01-20, 3d
    Circuit Breaker        :w3-2, after w3-1, 2d
    Integration Tests      :w3-3, after w3-2, 2d
    
    section Week 4
    Cost Tracking          :w4-1, 2026-01-27, 2d
    Grafana Dashboard      :w4-2, after w4-1, 2d
    Docker Compose         :w4-3, after w4-2, 2d
    Load Testing          :w4-4, after w4-3, 1d
```

---

### Technology Stack Position

```mermaid
graph TB
    subgraph "Application Layer"
        A1[Your Solana dApp<br/>TypeScript/Rust]
        A2[Your Backend API<br/>Node.js/Python/Go]
    end
    
    subgraph "RPC Load Balancer Layer ⭐"
        LB[Load Balancer<br/>Go + Gin + Redis]
        style LB fill:#4CAF50,color:#fff,stroke:#2E7D32,stroke-width:4px
    end
    
    subgraph "RPC Provider Layer"
        P1[Helius]
        P2[Alchemy]
        P3[QuickNode]
    end
    
    subgraph "Blockchain Layer"
        BC[Solana Mainnet<br/>Validators]
    end
    
    A1 --> LB
    A2 --> LB
    LB --> P1
    LB --> P2
    LB --> P3
    P1 --> BC
    P2 --> BC
    P3 --> BC
    
    Note[★ Our Load Balancer sits here<br/>Acting as intelligent middleware]
    Note -.-> LB
```

---

## Key Architectural Decisions

### 1. Why Go?
- **High Performance**: Handles 10,000+ concurrent requests
- **Simple Concurrency**: Goroutines for health monitoring
- **Single Binary**: Easy deployment (no dependencies)
- **Fast Startup**: Sub-second cold start

### 2. Why Redis?
- **Speed**: Sub-millisecond health lookups
- **TTL Support**: Auto-expire stale health data (5s)
- **Atomic Operations**: Thread-safe health updates
- **Simple**: Single node sufficient for MVP

### 3. Why Round-Robin (not Least-Latency)?
- **Simplicity**: Easy to implement and debug
- **Predictability**: Clear cost distribution
- **Good Enough**: MVP doesn't need optimization
- **Future**: Can upgrade to least-latency in Phase 2

### 4. Why Docker Compose (not Kubernetes)?
- **Simplicity**: MVP doesn't need orchestration
- **Fast Setup**: `docker-compose up` in 30 seconds
- **Local Dev**: Easy testing on laptop
- **Future**: Can migrate to K8s in Phase 2

---

## Architecture Patterns Used

| Pattern | Where | Why |
|---------|-------|-----|
| **Proxy** | Load Balancer | Forward requests transparently |
| **Round-Robin** | Provider Pool | Simple load distribution |
| **Circuit Breaker** | Retry Handler | Prevent cascade failures |
| **Health Check** | Monitor | Active probing for liveness |
| **Observer** | Prometheus | Metrics collection |
| **Strategy** | Routing | Pluggable routing algorithms |

---

## Scalability Considerations

### MVP (Single Instance)
- **Throughput**: 1,000 req/s
- **Latency**: p95 < 50ms
- **Deployment**: Single Docker container

### Phase 2 (Horizontal Scale)
```mermaid
graph TB
    LB1[Load Balancer 1]
    LB2[Load Balancer 2]
    LB3[Load Balancer 3]
    
    NGX[Nginx Load Balancer] --> LB1
    NGX --> LB2
    NGX --> LB3
    
    Redis[(Redis Cluster)]
    
    LB1 --> Redis
    LB2 --> Redis
    LB3 --> Redis
    
    LB1 --> Providers
    LB2 --> Providers
    LB3 --> Providers
```
- **Throughput**: 10,000+ req/s
- **High Availability**: No single point of failure
- **Deployment**: Kubernetes with HPA

---

## Security Architecture

```mermaid
graph TB
    subgraph "External"
        Client[Client Application]
    end
    
    subgraph "Load Balancer Security"
        TLS[TLS/HTTPS<br/>Optional]
        RL[Rate Limiter<br/>1000 req/s per IP]
        Val[Request Validator<br/>JSON-RPC Schema]
    end
    
    subgraph "Secrets Management"
        Env[.env File<br/>API Keys]
        K8s[K8s Secrets<br/>Phase 2]
    end
    
    subgraph "Audit"
        Logs[Request Logs<br/>Sanitized]
        Metrics[Prometheus<br/>Aggregates Only]
    end
    
    Client --> TLS
    TLS --> RL
    RL --> Val
    
    Val --> Handler[Request Handler]
    Handler --> Env
    
    Handler --> Logs
    Handler --> Metrics
    
    style TLS fill:#4CAF50,color:#fff
    style Env fill:#FF9800,color:#fff
```

---

## Summary: Why This Architecture?

✅ **Simple**: Easy to understand and debug  
✅ **Reliable**: Auto-failover, circuit breakers, health checks  
✅ **Observable**: Full visibility into costs and performance  
✅ **Scalable**: Start small, grow to Kubernetes  
✅ **Maintainable**: Clean separation of concerns  
✅ **Cost-Effective**: Optimize RPC spend automatically  

---

**Ready to build? Let's start with Week 1!** 🚀
