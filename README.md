# Aegis

> A distributed AI inference harness with high-performance scheduling, rate limiting, and usage metering designed around a BYOK architecture.

Aegis is not a chatbot, a RAG application, or a simple LLM API wrapper.

It is a distributed infrastructure harness that simulates the engineering problems encountered when operating an AI inference platform at scale. Users supply their own API keys (BYOK, Bring Your Own Key) to interact with upstream LLM providers directly. There is no platform-level billing. Instead the platform focuses on **secure key management**, **admission control**, **distributed rate limiting**, **capacity-aware scheduling**, and **durable usage metering**.

---

## Table of Contents

- [Why Aegis?](#why-aegis)
- [Monorepo Structure](#monorepo-structure)
- [Tech Stack](#tech-stack)
  - [Frontend](#frontend)
  - [Backend](#backend)
- [High-Level Architecture](#high-level-architecture)
- [Component Breakdown](#component-breakdown)
  - [Frontend Dashboard](#frontend-dashboard)
  - [Go Control Plane](#go-control-plane)
  - [Python Inference Workers](#python-inference-workers)
  - [Kafka: Async Event Bus](#kafka--async-event-bus)
  - [Redis: Distributed Ephemeral State](#redis--distributed-ephemeral-state)
  - [PostgreSQL: Durable Source of Truth](#postgresql--durable-source-of-truth)
- [Request Lifecycle](#request-lifecycle)
  - [Synchronous (Streaming) Path](#synchronous-streaming-path)
  - [Asynchronous (Metering) Path](#asynchronous-metering-path)
- [Domain Model](#domain-model)
- [BYOK Model](#byok-model)
- [Scheduling Strategy](#scheduling-strategy)
- [Admission Control](#admission-control)
- [Distributed Rate Limiting](#distributed-rate-limiting)
- [Worker Registry & Heartbeats](#worker-registry--heartbeats)
- [Usage Metering & Idempotency](#usage-metering--idempotency)
- [Observability](#observability)
- [Security](#security)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Running Tests](#running-tests)

---

## Why Aegis?

Running an LLM-powered service in production is not primarily an AI problem: it is a **distributed systems** problem. The interesting engineering challenges are:

- How do you fairly limit and enforce quotas across multiple tenants without a centralised bottleneck?
- How do you route requests to healthy workers that have spare capacity for the requested model?
- How do you prevent a burst of requests from overwhelming your worker pool?
- How do you record token usage reliably even if a worker crashes mid-stream?
- How do you guarantee that usage events are processed exactly once even with Kafka redelivery?

Aegis is designed to surface and solve those problems in a clean, well-structured codebase. The LLM itself (Mistral via LangChain) is the *payload*: the interesting engineering is everything around it.

---

## Monorepo Structure

```
aegis/
├── frontend/     # React + TypeScript dashboard
├── backend/      # Go control plane + Python inference workers
└── start.sh      # Convenience script to start both services
```

---

## Tech Stack

### Frontend

| | Technology |
|---|---|
| Framework | React 18 + TypeScript |
| Build tool | Vite |
| Routing | TanStack Router |
| Data fetching | TanStack Query |
| UI primitives | Radix UI |
| Animations | Framer Motion |
| Styling | Tailwind CSS |
| HTTP client | Axios |
| Icons | Lucide React |
| Notifications | Sonner |

### Backend

| Layer | Technology |
|---|---|
| Control plane | Go 1.26+ |
| HTTP framework | Gin |
| Inference workers | Python 3.12+ |
| LLM orchestration | LangChain |
| LLM provider | Mistral (via BYOK API key) |
| Go to Python RPC | gRPC / Protocol Buffers |
| Token streaming | gRPC server-side streaming → SSE |
| Async event bus | Kafka (segmentio/kafka-go) |
| Distributed state | Redis (go-redis/v9) |
| Durable storage | PostgreSQL (GORM + pgx/v5) |
| Observability | OpenTelemetry + Prometheus |
| Containerisation | Docker / Docker Compose |
| Orchestration | Kubernetes |

---

## High-Level Architecture

```
                        ┌───────────────────┐
                        │  Browser / Client │
                        │  React Dashboard  │
                        └────────┬──────────┘
                                 │
                          HTTP / REST
                                 │
                                 ▼
                     ┌────────────────────┐
                     │    Go Backend      │
                     │                    │
                     │ Router             │
                     │ Authentication     │
                     │ Authorization      │
                     │ Rate Limiting      │
                     │ Admission Control  │
                     │ Scheduler          │
                     │ Worker Registry    │
                     │ Streaming (SSE)    │
                     └─────────┬──────────┘
                               │
                          gRPC streaming
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
        Python Worker    Python Worker    Python Worker
              │                │                │
              └────────────────┼────────────────┘
                               │
                           LangChain
                               │
                               ▼
                             Mistral


                          Async Events
                               │
                               ▼
                             Kafka
                               │
               ┌───────────────┼───────────────┐
               ▼               ▼               ▼
           Metering        Analytics          DLQ


                            Redis
               ┌────────────┼────────────┐
               ▼            ▼            ▼
          Rate Limits    Worker      Locks &
                         Registry    Ephemeral State


                          PostgreSQL
               ┌────────────┼────────────┐
               ▼            ▼            ▼
            Users        Usage        Outbox
            Orgs         Pricing      Events
            Projects     Inference
            API Keys     Records
```

---

## Component Breakdown

### Frontend Dashboard

The frontend (`frontend/`) is a React + TypeScript single-page application built with Vite. It serves as the management interface for the Aegis platform.

Pages and features:

| Page | Purpose |
|---|---|
| **Auth** | Login and registration |
| **Chat** | Interactive inference interface: send requests and receive streamed responses |
| **Projects** | Create and manage projects within an organisation |
| **Organizations** | Tenant management |
| **Analytics** | Usage charts and token consumption metrics per project and organisation |
| **Settings** | BYOK provider key management, API key management |
| **Profile** | User account settings |

The dashboard communicates with the Go backend over HTTP. Streaming responses from the inference endpoint are consumed via Server-Sent Events (SSE) and rendered token-by-token in the chat interface.

### Go Control Plane

The Go backend (`backend/cmd/aegis/main.go`) is the single process that runs as the API and control plane. It is intentionally a **monolith-first** design: all logical components (scheduler, admission, rate limiter, worker registry, metering consumer) live in one process so they can be cleanly separated into independent services later without moving domain logic.

Responsibilities:

- **HTTP API**: OpenAI-compatible endpoint (`POST /v1/chat/completions`) plus management, health, and metrics endpoints
- **Authentication**: API key resolution; keys are never stored in plaintext (hashed with a visible prefix only)
- **Authorization**: tenant isolation; a key belonging to one organisation can never access another's resources
- **Distributed rate limiting**: per API key, per project, and per organisation; enforced via Redis so it works across multiple backend replicas
- **Admission control**: decides whether to ACCEPT, QUEUE, or REJECT based on rate limits, quotas, token estimates, and worker capacity
- **Scheduler**: selects the best available worker using a capacity-aware strategy (not round-robin)
- **Worker registry**: tracks active Python workers and their health via heartbeats stored in Redis
- **gRPC client**: forwards admitted inference jobs to a selected Python worker
- **SSE streaming**: bridges the gRPC token stream from the worker back to the client
- **Kafka consumer**: consumes usage events published by workers and persists them durably to PostgreSQL

### Python Inference Workers

Each Python worker (`backend/worker/inference/`) is a long-running gRPC server process. Multiple workers can run simultaneously and scale independently from the Go backend.

Responsibilities:

- Register with the Go backend on startup and send periodic heartbeats
- Advertise which models they support
- Accept `Generate` RPC calls from Go
- Use **LangChain** to interact with the Mistral API using the user's BYOK key
- Stream generated tokens back over the gRPC connection
- Publish a `UsageEvent` to Kafka at inference completion (actual token counts from the model)
- Handle cancellation, timeouts, and graceful drain on shutdown

The Python worker has no knowledge of rate limiting, scheduling, quotas, or admission. Those are Go concerns.

### Kafka: Async Event Bus

Kafka decouples inference completion from usage persistence. The Python worker does not wait for PostgreSQL before returning. This keeps the hot inference path free of database latency.

| Topic | Producer | Consumer |
|---|---|---|
| `aegis.inference.events` | Go backend | Metering consumer |
| `aegis.usage.events` | Python worker | Go metering consumer |
| `aegis.dlq` | All consumers | Manual investigation |

Token streaming does **not** go through Kafka. Tokens flow: `Mistral → LangChain → gRPC stream → Go → SSE → Client`.

### Redis: Distributed Ephemeral State

Redis stores state that is fast to read/write, acceptable to lose on restart, and shared across multiple Go backend replicas.

| Key Namespace | Purpose |
|---|---|
| `rl:{api_key_id}` | Rate limit counters |
| `rl:{project_id}` | Project-level rate limits |
| `rl:{org_id}` | Organisation-level rate limits |
| `worker:{worker_id}` | Worker registry entry |
| `worker:{worker_id}:hb` | Heartbeat TTL key |
| `lock:{resource}` | Distributed locks |

Redis is **not** the source of truth for usage, pricing, or metering. PostgreSQL owns that.

### PostgreSQL: Durable Source of Truth

All durable state lives in PostgreSQL.

| Table | Purpose |
|---|---|
| `users` | User accounts |
| `organizations` | Tenants |
| `projects` | Workspaces within an organisation |
| `api_keys` | Hashed API keys with prefix |
| `inference_records` | One record per inference job |
| `usage_records` | Actual token consumption |
| `pricing` | Versioned model pricing |
| `idempotency_records` | Deduplication for API requests |
| `outbox_events` | Transactional outbox for Kafka |

---

## Request Lifecycle

### Synchronous (Streaming) Path

```
CLIENT (Browser / React Dashboard)
  │
  │ HTTP POST /v1/chat/completions
  ▼
Router
  │
Authentication          resolve API key → Project → Organisation
  │
Authorization           tenant isolation check
  │
Validation              request schema, field bounds
  │
Idempotency             check Idempotency-Key header
  │
Distributed Rate Limit  Redis atomic counter (per key / project / org)
  │
Token Estimation        estimate input token count
  │
Admission Control       ACCEPT / QUEUE / REJECT
  │                     considers: quotas, worker capacity, queue depth
Scheduler               select best available worker
  │                     considers: model support, health, active jobs,
  │                               capacity, queue depth, priority
  │
  │ gRPC Generate(request)
  ▼
Python Worker
  │
LangChain → Mistral API   (user BYOK key used here)
  │
token stream
  ▼
Python Worker
  │ gRPC stream (GenerateResponse chunks)
  ▼
Go Backend
  │ Server-Sent Events
  ▼
CLIENT (tokens rendered in real-time by React chat UI)
```

### Asynchronous (Metering) Path

```
Python Worker
  │
  │ UsageEvent {
  │   request_id, org_id, project_id,
  │   model, input_tokens, output_tokens,
  │   duration_ms, status
  │ }
  ▼
Kafka  (aegis.usage.events)
  │
  ▼
Go Metering Consumer
  │
  │ Idempotent upsert (unique DB constraint on event_id)
  ▼
PostgreSQL  (usage_records)
  │
  ▼
Analytics API  (consumed by React dashboard)
```

Inference completion is **independent** of metering. The client receives their full response before usage is persisted.

---

## Domain Model

Aegis uses Domain-Driven Design. Domain packages live under `backend/internal/domain/` and have **zero dependency on infrastructure** (no PostgreSQL, no Redis, no Kafka, no Gin).

```
backend/internal/domain/
├── auth/           Authentication domain, API key validation
├── user/           User entity, domain rules
├── organization/   Tenant entity, tenant isolation rules
├── project/        Project entity, quota and config
├── inference/      Inference job model, job lifecycle
├── admission/      Admission rules: ACCEPT / QUEUE / REJECT
├── scheduler/      Worker selection strategy
├── worker/         Worker entity, health state machine
├── provider/       BYOK provider key model
├── usage/          Usage record, token consumption
└── metrics/        Domain metric definitions
```

Dependency direction:

```
Router → Application → Domain
                          ↑
                    Infrastructure
```

Infrastructure (`backend/internal/infra/`) implements the interfaces defined by the domain. The domain never imports infra.

---

## BYOK Model

Aegis operates a **Bring Your Own Key** model. Users register their own Mistral API keys via the Settings page in the dashboard. When an inference request is admitted and dispatched to a Python worker, the worker retrieves the user's provider key (decrypted from secure storage) and uses it to call Mistral directly.

Implications:

- Aegis has **no billing responsibility**: the user pays Mistral directly for tokens consumed
- Aegis does **meter usage**: token counts are recorded per project and organisation for quota enforcement and analytics visible in the dashboard
- Provider keys are **never logged** and never included in inference job records
- Key rotation and revocation are supported at the project level via the dashboard

---

## Scheduling Strategy

The scheduler (`backend/internal/domain/scheduler/`) does not use round-robin. Worker selection scores each candidate on:

| Factor | Rationale |
|---|---|
| Supported model | Worker must support the requested model |
| Health status | Only `READY` workers are eligible |
| Active requests | Prefer workers with spare capacity |
| Max concurrency | Hard upper bound per worker |
| Queue depth | Avoid overloading a single worker |
| Estimated token load | Heavier jobs are routed to workers with more headroom |
| Request priority | Higher-priority jobs influence routing decisions |
| Retry count | Retried jobs avoid the worker that last failed them |

---

## Admission Control

The admission controller (`backend/internal/domain/admission/`) gates every request before it reaches the scheduler.

| Decision | Condition |
|---|---|
| `ACCEPT` | Rate limits clear, quota available, worker capacity available |
| `QUEUE` | Rate limits clear but no immediate worker capacity |
| `REJECT` | Rate limit exceeded, quota exhausted, or queue at maximum depth |

Queuing is bounded. The controller implements backpressure rather than allowing unbounded growth.

---

## Distributed Rate Limiting

Rate limits are enforced using **atomic Redis operations** (sliding window) at three levels:

1. **API key**: per-key RPM and TPM limits
2. **Project**: project-level aggregate limits
3. **Organisation**: tenant-level cap across all projects

Because limits are stored in Redis, they work correctly when **multiple Go backend instances** run behind a load balancer. In-memory limiters are not used as the source of truth.

On limit exceeded: `429 Too Many Requests` with `Retry-After` header.

On Redis unavailability: the system fails **closed**: it rejects rather than bypasses to avoid accidentally disabling security controls.

---

## Worker Registry & Heartbeats

Workers register themselves in Redis on startup, advertising their `worker_id`, `supported_models`, `max_concurrency`, and `status`. They send heartbeats at a fixed interval. If a worker stops sending heartbeats, the TTL expires and the registry transitions the worker to `UNHEALTHY`.

Worker state machine:

```
STARTING → READY → DRAINING → OFFLINE
                ↘
             UNHEALTHY
```

The scheduler never dispatches work to `UNHEALTHY` or `OFFLINE` workers. In-flight jobs on a crashed worker are recovered based on their processing state.

---

## Usage Metering & Idempotency

Usage events are emitted by the Python worker at the end of each inference job. The Go metering consumer processes them from Kafka and writes to PostgreSQL.

**Idempotency is guaranteed by a unique database constraint on `event_id`**, not by an application-level check. This prevents duplicate records even under concurrent Kafka redelivery.

The transactional outbox pattern is used wherever domain state and an event must be committed atomically:

```sql
BEGIN;
  INSERT INTO domain_table ...;
  INSERT INTO outbox_events (event_type, payload, ...) VALUES (...);
COMMIT;
```

A background publisher relays outbox events to Kafka with exponential backoff retries. Events are **never silently lost**.

---

## Observability

| Signal | Implementation |
|---|---|
| Structured logging | JSON with `request_id`, `job_id`, `trace_id` on every line |
| Distributed tracing | OpenTelemetry, propagated through HTTP → Go → gRPC → Python |
| Metrics | Prometheus, exposed at `GET /metrics` |

Key metrics tracked:

```
http_requests_total / http_request_duration_seconds
inference_requests_total / inference_duration_seconds / inference_errors_total
scheduler_queue_depth / scheduler_jobs_total
worker_active_jobs / worker_capacity / worker_heartbeat_age_seconds
usage_events_total / usage_processing_failures_total
rate_limit_rejections_total
```

Secrets (API keys, provider keys, passwords) are **never logged**.

---

## Security

- API keys are generated with cryptographically secure random data; only the hash and a non-secret prefix are stored
- BYOK provider keys are stored encrypted; the plaintext is only retrieved inside the Python worker at inference time
- Strict tenant isolation: a key from Organisation A can never access Organisation B's resources
- Parameterised SQL throughout; no string interpolation in queries
- Request size limits enforced at the router level
- Internal infrastructure (gRPC, Redis, Kafka) is not exposed publicly
- Secrets are managed via environment variables; never committed to the repository

---

## Getting Started

**Prerequisites:** Node.js 20+, Docker, Docker Compose, a Mistral API key.

### 1. Backend

```bash
cd backend
cp .env.example .env
# Fill in MISTRAL_API_KEY and other values

docker compose up
```

This starts the Go backend, Python inference worker(s), PostgreSQL, Redis, Kafka, Prometheus, and Grafana.

The API is available at `http://localhost:8080`.

### 2. Frontend

```bash
cd frontend
cp .env.example .env.local
# Set VITE_API_URL=http://localhost:8080

npm install
npm run dev
```

The dashboard is available at `http://localhost:5173`.

### Or: start everything at once

```bash
# From the repo root
./start.sh
```

---

## Environment Variables

### Backend: see [`backend/.env.example`](backend/.env.example)

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `KAFKA_BROKERS` | Comma-separated Kafka broker addresses |
| `GRPC_WORKER_ADDRESSES` | Comma-separated Python worker gRPC addresses |
| `JWT_SECRET` | Secret for signing JWT tokens |
| `ENCRYPTION_KEY` | Key used to encrypt stored BYOK provider keys |

### Frontend: see [`frontend/.env.example`](frontend/.env.example)

| Variable | Description |
|---|---|
| `VITE_API_URL` | Base URL of the Go backend API |

---

## Running Tests

**Backend:**

```bash
cd backend

# Unit and integration tests
go test ./...

# With race detector (required to pass)
go test -race ./...

# Vet
go vet ./...
```

**Python worker:**

```bash
cd backend/worker/inference
pytest
```

**Frontend:**

```bash
cd frontend
npm run lint
```

Load testing scripts are in `backend/scripts/`. Measured results are documented in `backend/test/load/results/`. No results are fabricated.
