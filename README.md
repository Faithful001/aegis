# Aegis

> A distributed harness for running AI inference reliably under load. Handles scheduling, admission control, rate limiting, and durable usage metering, and it’s built around BYOK so customers keep control of their own model credentials.

Aegis is not just a chatbot, a RAG application, or a simple LLM API wrapper.

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

Aegis is designed to surface and solve those problems in a clean, well-structured codebase. The LLM itself is the _payload_: the interesting engineering is everything around it.

---

## Monorepo Structure

```
aegis/
├── frontend/     # React + TypeScript dashboard
├── backend/      # Go control plane + Python inference workers
├── Makefile      # make dev / dev-backend / dev-frontend
└── start.sh      # Convenience script to start backend and frontend
```

---

## Tech Stack

### Frontend

|               | Technology                   |
| ------------- | ---------------------------- |
| Framework     | React 18 + TypeScript        |
| Build tool    | Vite                         |
| Routing       | TanStack Router              |
| Data fetching | TanStack Query               |
| UI primitives | Radix UI                     |
| Animations    | Framer Motion                |
| Styling       | Tailwind CSS                 |
| HTTP client   | Axios                        |
| Icons         | Phosphor Icons               |
| Markdown      | React Markdown + remark-gfm  |
| Notifications | Sonner                       |
| Deployment    | Vercel (`vercel.json`)       |

### Backend

| Layer             | Technology                                                          |
| ----------------- | ------------------------------------------------------------------- |
| Control plane     | Go 1.26+                                                            |
| HTTP framework    | Gin                                                                 |
| Authentication    | JWT (`golang-jwt/v5`) + bcrypt; hashed API keys                     |
| OAuth sign-in     | Google and GitHub (OAuth 2.0)                                       |
| Inference workers | Python 3.12+                                                        |
| LLM orchestration | LangChain (Python worker) + provider adapters in Go                 |
| LLM providers     | OpenAI, Anthropic, Google Gemini, Mistral AI, OpenRouter (BYOK)     |
| Go to Python RPC  | gRPC / Protocol Buffers                                             |
| Token streaming   | gRPC server-side streaming → SSE                                    |
| Async event bus   | Kafka (segmentio/kafka-go)                                          |
| Distributed state | Redis (go-redis/v9)                                                 |
| Durable storage   | PostgreSQL (GORM + pgx/v5)                                          |
| Email             | Resend (`resend-go/v4`) — OTP and notification emails               |
| Configuration     | `godotenv` + environment variables                                  |
| Observability     | Prometheus metrics + Grafana dashboards; JSON logs via `log/slog`   |
| Containerisation  | Docker Compose (Postgres, Kafka, Redis, Prometheus, Grafana)        |

> Kubernetes manifests are not part of the repository yet; local development uses the Docker Compose stack in `backend/docker-compose.yaml`.

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
      OpenAI │ Anthropic │ Google Gemini │ Mistral │ OpenRouter


                          Async Events
                               │
                               ▼
                             Kafka
                               │
               ┌───────────────┼───────────────┐
               ▼               ▼               ▼
           Metering         Email        Notifications


                            Redis
                    ┌────────┴────────┐
                    ▼                 ▼
              Rate Limits       Worker Registry


                          PostgreSQL
               ┌────────────┼────────────┐
               ▼            ▼            ▼
            Users        Usage        Provider
            Orgs         Records      Credentials
            Projects     Sessions     Threads &
            API Keys                  Messages
```

---

## Component Breakdown

### Frontend Dashboard

The frontend (`frontend/`) is a React + TypeScript single-page application built with Vite. It serves as the management interface for the Aegis platform.

Pages and features:

| Page                                        | Purpose                                                                 |
| ------------------------------------------- | ----------------------------------------------------------------------- |
| **Home** (`/`)                              | Landing page mounted inside the app shell                               |
| **Auth** (`/auth/*`)                        | Login, registration, email verification, OTP, password reset, OAuth callback |
| **Chat** (`/chat`, `/chat/$id`, `/c/$id`)   | Interactive inference interface with persisted, streamed conversations  |
| **Analytics** (`/analytics`)                | Token usage charts and activity heatmap                                 |
| **Settings** (`/settings?tab=…`)            | Profile, security, BYOK providers/models, notifications, connected accounts |
| **Profile** (`/profile`)                    | User account settings                                                   |

The dashboard communicates with the Go backend over HTTP. Streaming responses from the inference endpoint are consumed via Server-Sent Events (SSE) and rendered token-by-token in the chat interface.

### Go Control Plane

The Go backend (`backend/cmd/aegis/main.go`) is the single process that runs as the API and control plane. It is intentionally a **monolith-first** design: all logical components (scheduler, admission, rate limiter, worker registry, metering consumer) live in one process so they can be cleanly separated into independent services later without moving domain logic.

Responsibilities:

- **HTTP API**: OpenAI-compatible endpoints (`POST /v1/chat/completions`, `GET /v1/models`) plus the JWT-protected management API under `/api/v1` and health/metrics endpoints
- **Authentication**: JWT access and refresh tokens with bcrypt password hashes and revocable sessions. The API key model and `APIKeyAuthMiddleware` exist in the codebase but are not wired into the router yet
- **Authorization**: tenant isolation; a token belonging to one user/organisation can never access another's resources
- **Rate limiting**: Redis-backed sliding window, enforced per user (the limiter also supports API key, project, and organisation tiers)
- **Admission control**: in-process concurrency and request-size gate (ACCEPT / QUEUE / REJECT). Implemented but not wired into the router yet — see [Admission Control](#admission-control)
- **Scheduler**: selects the best available worker using a capacity-aware strategy (not round-robin)
- **Worker registry**: tracks active Python workers and their health via heartbeats stored in Redis
- **gRPC client**: forwards admitted inference jobs to a selected Python worker
- **SSE streaming**: bridges the gRPC token stream from the worker back to the client
- **Kafka consumer**: consumes usage, email, and notification events and persists/propagates them

### Python Inference Workers

Each Python worker (`backend/worker/inference/`) is a long-running gRPC server process. Multiple workers can run simultaneously and scale independently from the Go backend.

Responsibilities:

- Serve a `HealthCheck` RPC (registration and heartbeat reporting are not implemented yet — see [Worker Registry & Heartbeats](#worker-registry--heartbeats))
- Advertise which models they support
- Accept `Generate` RPC calls from Go
- Use **LangChain** to interact with supported LLM providers using the user's BYOK keys.
- Stream generated tokens back over the gRPC connection
- Report token usage counts so the Go control plane can publish metering events to Kafka
- Handle cancellation, timeouts, and graceful drain on shutdown

The Python worker has no knowledge of rate limiting, scheduling, quotas, or admission. Those are Go concerns.

### Kafka: Async Event Bus

Kafka decouples inference completion from usage persistence. The Python worker does not wait for PostgreSQL before returning. This keeps the hot inference path free of database latency.

| Topic                            | Producer                          | Consumer                  |
| -------------------------------- | --------------------------------- | ------------------------- |
| `aegis.events.usage`             | Go inference service              | Usage metering consumer   |
| `aegis.events.email`             | Go auth service                   | Email consumer (Resend)   |
| `aegis.events.failed_requests`   | Go inference service              | Notification consumer     |
| `aegis.events.budget_alerts`     | Notification scheduler            | Notification consumer     |
| `aegis.events.weekly_usage`      | Notification scheduler            | Notification consumer     |
| `aegis.events.rate_limits`       | Rate limit middleware             | Notification consumer     |
| `aegis.events.modified_credentials` | Auth / provider services       | Notification consumer     |
| `aegis.events.notifications`     | Any service                       | Notification consumer     |
| `aegis.events.jobs`              | —                                 | Reserved, not yet consumed |

Token streaming does **not** go through Kafka. Tokens flow: `LLM Provider → LangChain (worker) or Go provider gateway → gRPC/SSE → Client`.

### Redis: Distributed Ephemeral State

Redis stores state that is fast to read/write, acceptable to lose on restart, and shared across multiple Go backend replicas.

| Key Namespace                    | Purpose                                  |
| -------------------------------- | ---------------------------------------- |
| `ratelimit:{tier}:{id}:rpm`      | Sliding-window RPM counters              |
| `ratelimit:{tier}:{id}:tpm`      | Sliding-window TPM counters              |
| `aegis:workers:{worker_id}`      | Worker registry entry (5-minute TTL)     |
| `aegis:workers:index`            | Set of known worker IDs                  |

Redis is **not** the source of truth for usage or metering. PostgreSQL owns that.

### PostgreSQL: Durable Source of Truth

All durable state lives in PostgreSQL.

| Table                            | Purpose                                              |
| -------------------------------- | ---------------------------------------------------- |
| `users`                          | User accounts, password hash, Google/GitHub IDs      |
| `organizations`                  | Tenants                                              |
| `organization_members`           | Organisation membership                              |
| `projects`                       | Workspaces within an organisation                    |
| `api_keys`                       | Hashed API keys with a non-secret prefix             |
| `provider_credentials`           | AES-256-GCM encrypted BYOK provider keys             |
| `user_models`                    | Per-user model visibility toggles                    |
| `chat_threads` / `chat_messages` | Chat history persistence                             |
| `usage_records`                  | Actual token consumption (unique `event_id`)         |
| `user_sessions`                  | Refresh-token sessions, revocable per device         |
| `blacklisted_tokens`             | Revoked access tokens                                |
| `otp_records`                    | Email verification and password-reset codes          |
| `alert_settings`                 | Budget alert preferences                             |
| `notifications`                  | In-app notifications                                 |

Schema is created at startup via GORM `AutoMigrate` (`backend/cmd/aegis/main.go`). Separate `pricing`, `idempotency_records`, and `outbox_events` tables are not implemented yet — pricing is currently a hardcoded map in `backend/internal/domain/usage/`.

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
  │                     considers: request size, local concurrency,
  │                               queue depth (not wired yet)
Scheduler               select best available worker
  │                     considers: model support, health, active jobs,
  │                               capacity, queue depth, priority
  │
  │ gRPC Generate(request)
  ▼
Python Worker
  │
LangChain → Selected LLM Provider API (user's BYOK key used here)
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
Go Inference Service (and Python Worker for the gRPC path)
  │
  │ UsageEvent {
  │   request_id, org_id, project_id,
  │   model, input_tokens, output_tokens,
  │   duration_ms, status
  │ }
  ▼
Kafka  (aegis.events.usage)
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

> As implemented today, the Go control plane can also stream from a provider directly through its provider gateway (`backend/internal/infra/provider/`), bypassing the Python worker. The gRPC worker path shown above is used when a worker is healthy and scheduled for the requested model.

---

## Domain Model

Aegis follows a layered, domain-oriented layout. Logical domains live under `backend/internal/domain/`; each package currently contains its entity/model, repository, service, controller, and (where relevant) Kafka consumer together.

> Note: the domain packages still import infrastructure libraries (GORM for models/repositories, Gin for controllers, Redis/Kafka for consumers). The dependency rule below is the target, not a current invariant — extracting repositories/controllers into `infra`/`application` layers is an ongoing refactor.

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
├── thread/         Chat threads and message persistence
├── notification/   Budget alerts and notification delivery
├── otp/            Email verification / password-reset codes
├── token/          Access and refresh token definitions
└── metrics/        Domain metric definitions
```

Dependency direction:

```
Router → Application → Domain
                          ↑
                    Infrastructure
```

Interfaces are declared alongside their domain types, and `backend/internal/infra/` provides concrete adapters (Postgres, Redis, Kafka, provider gateways, crypto, email). Several domain packages also contain their own GORM repositories and Gin controllers today.

---

## BYOK Model

Aegis operates a **Bring Your Own Key** model. Users register their own LLM provider API keys via the Settings page in the dashboard. When an inference request is admitted, the Go control plane decrypts the stored provider key (AES-256-GCM) and the provider gateway in `backend/internal/infra/provider/` calls the upstream LLM directly, streaming tokens back to the client. The gRPC worker path remains available as a fallback when a worker is registered and scheduled.

Implications:

- Aegis has **no billing responsibility**: the user pays the LLM provider directly for tokens consumed
- Aegis does **meter usage**: token counts are recorded per project and organisation for quota enforcement and analytics visible in the dashboard
- Provider keys are **never logged** and never included in inference job records
- Key rotation and revocation are supported per provider credential via the Settings page

---

## Scheduling Strategy

The scheduler (`backend/internal/domain/scheduler/`) does not use round-robin. Worker selection scores each candidate on:

| Factor               | Rationale                                             |
| -------------------- | ----------------------------------------------------- |
| Supported model      | Worker must support the requested model               |
| Health status        | Only `READY` workers are eligible                     |
| Active requests      | Prefer workers with spare capacity                    |
| Max concurrency      | Hard upper bound per worker                           |
| Queue depth          | Avoid overloading a single worker                     |
| Estimated token load | Heavier jobs are routed to workers with more headroom |
| Request priority     | Higher-priority jobs influence routing decisions      |
| Retry count          | Retried jobs avoid the worker that last failed them   |

---

## Admission Control

The admission controller (`backend/internal/domain/admission/`) is an **in-process backpressure gate**: it decides whether a request can start, wait briefly for a slot, or be rejected.

Under BYOK it is **not** a billing or quota mechanism — the user pays the provider directly — and it does not protect a shared worker pool, because inference currently runs through the Go provider gateway. Its only real purpose is protecting the control-plane process (goroutines, sockets, DB pool, Kafka writes) from stream pile-ups, and stopping one user from occupying all local capacity.

### Behaviour

`AdmissionService.Evaluate` considers only request size and local concurrency — rate limiting is a separate middleware that runs before it:

| Decision | Condition (defaults)                                                  |
| -------- | --------------------------------------------------------------------- |
| `ACCEPT` | estimated input + output tokens ≤ 8,192 **and** active requests < 50  |
| `QUEUE`  | tokens within budget but 50 requests active; queue not yet at 100     |
| `REJECT` | token estimate > 8,192, or queue already at 100                       |

`AcquireSlot` increments the active counter on `ACCEPT`, or waits up to 5 s for a slot on `QUEUE` before returning `ErrQueueTimeout`. `ReleaseSlot` decrements the counter and wakes one waiter. Token estimates are a `len(prompt)/4` heuristic with a 500-token default output. The middleware returns `503` with `Retry-After: 5` on rejection and adds `X-Admission-Wait-Ms` to requests that had to wait.

### Status and limitations

> `main.go` never constructs an `AdmissionService` and does not pass one to the router, so the middleware is **dormant** — no request is gated today.

If it is wired up as-is, note that:

- State is per-process and in-memory, so N replicas allow N × 50 concurrent requests rather than 50.
- Counters are global, not per tenant, so one user's burst can occupy the whole pool.
- The 8,192-token cap is fixed with no configuration path and will reject legitimate long-context requests.
- A 5-second queue is marginal value for LLM calls that hold a slot for tens of seconds.
- The middleware's `400` branch (`ErrMaxTokensExceeded`) is unreachable: `AcquireSlot` returns `ErrSystemOverloaded` for every rejection.

### Fit with BYOK

The natural admission unit in a BYOK system is **per user / per provider key**, because upstream provider rate limits are per key. A global cap answers "the platform is full"; a per-user cap answers "your key is saturated", which is the more useful signal for the caller.

Rate limiting already covers request rate per user; admission should cover **concurrency**. If this is finished, the sensible shape is a global safety semaphore plus a per-user concurrency limit (optionally Redis-backed so it holds across replicas), rejecting immediately with `429`/`503` rather than queuing. Until then it can stay dormant.

---

## Distributed Rate Limiting

Rate limits are enforced by an **atomic Redis Lua script** implementing a sliding window. The limiter supports four tiers:

1. **User** — enforced by the router today
2. **API key** — per-key RPM and TPM limits
3. **Project** — project-level aggregate limits
4. **Organisation** — tenant-level cap across all projects

Defaults are 600 RPM and 100,000 TPM. Because limits are stored in Redis, they work correctly when **multiple Go backend instances** run behind a load balancer. The in-memory limiter is only used for tests/fallback, not as the source of truth.

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

> Status: the Redis registry and heartbeat monitor run in the control plane, but nothing currently calls `WorkerService.Register` (there is no registration endpoint yet) and the Python worker does not send heartbeats. Until that is wired up, inference goes through the direct provider gateway.

---

## Usage Metering & Idempotency

Usage events are emitted when an inference job completes — by the Python worker for the gRPC path, and by the Go inference service for the direct provider path. The Go metering consumer processes them from Kafka and writes to PostgreSQL.

**Idempotency is guaranteed by a unique database constraint on `event_id`**, not by an application-level check. This prevents duplicate records even under concurrent Kafka redelivery.

Events are currently published to Kafka directly from the Go inference service. The transactional outbox pattern (`outbox_events` table plus a background relay) is planned but not implemented yet; delivery retries are handled by the Kafka producer.

---

## Observability

| Signal              | Implementation                                                        |
| ------------------- | --------------------------------------------------------------------- |
| Structured logging  | JSON or text via `log/slog`, with `request_id`, `job_id`, `trace_id` context fields |
| Metrics             | Prometheus, exposed at `GET /metrics`                                 |
| Dashboards          | Grafana, provisioned from `backend/deploy/grafana/`                   |
| Health probes       | `GET /health` (liveness) and `GET /ready` (Postgres + Redis checks)   |

> OpenTelemetry tracing is not wired in yet; trace IDs are propagated through request context and logs only.

Key metrics tracked (all prefixed `aegis_`):

```
http_requests_total / http_request_duration_seconds / http_in_flight_requests
inference_requests_total / inference_duration_seconds / inference_tokens_total
ratelimit_hits_total
worker_active_count
```

Secrets (API keys, provider keys, passwords) are **never logged**.

---

## Security

- API keys are generated with cryptographically secure random data; only the hash and a non-secret prefix are stored
- BYOK provider keys are stored encrypted with AES-256-GCM (key derived from `JWT_SECRET`) and decrypted only inside the Go control plane at inference time
- Strict tenant isolation: a token or key from one user/organisation can never access another's resources
- Parameterised SQL throughout; no string interpolation in queries
- Internal infrastructure (gRPC, Redis, Kafka) is not exposed publicly
- Secrets are managed via environment variables; never committed to the repository

---

## Getting Started

**Prerequisites:** Node.js 20+, Go 1.26+, Python 3.12+, Docker, and Docker Compose. Users can configure their LLM provider and API key through the frontend settings page.

### 1. Infrastructure

```bash
cd backend
cp .env.example .env

docker compose up -d
```

This starts PostgreSQL, Kafka, Redis, Prometheus, and Grafana. The Go control plane and Python workers are **not** part of the Compose file; run them separately below.

### 2. Go control plane

```bash
cd backend
# The Vite dev proxy targets 127.0.0.1:8080, so set PORT to match
PORT=8080 go run ./cmd/aegis
# or, with hot reload (air installed):
PORT=8080 air
```

The API is available at `http://localhost:8080` (`PORT` defaults to `5000` in `internal/infra/config/config.go`).

### 3. Python inference worker (optional)

```bash
cd backend/worker/inference
pip install -r ../requirements.txt
python main.py
```

The worker listens on `localhost:50051` by default; point the backend at it with `INFERENCE_WORKER_ADDR`. If no worker is reachable, the backend falls back to its built-in mock worker client and the direct provider gateway.

### 4. Frontend

```bash
cd frontend
# Set VITE_API_BASE_URL in .env if you are not using the Vite dev proxy
npm install
npm run dev
```

The dashboard is available at `http://localhost:5173`.

### Or: start backend and frontend at once

```bash
# From the repo root (infrastructure must already be running)
./start.sh

# Or via the Makefile, which uses npx concurrently
make dev
```

---

## Environment Variables

### Backend: see [`backend/.env.example`](backend/.env.example)

| Variable                                                 | Description                                                        |
| -------------------------------------------------------- | ------------------------------------------------------------------ |
| `ENVIRONMENT`                                            | `development` / `production`                                       |
| `PORT`                                                   | HTTP port for the Go control plane (default `5000`)                |
| `LOG_LEVEL`, `LOG_FORMAT`                                | Log level and `json` / `text` format                               |
| `DATABASE_URL`                                           | PostgreSQL connection string                                       |
| `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`               | Redis connection (`REDIS_URL` is also supported)                   |
| `KAFKA_BROKERS`, `KAFKA_GROUP_ID`                        | Kafka broker(s) and consumer group                                 |
| `INFERENCE_WORKER_ADDR`                                  | Python worker gRPC address (default `localhost:50051`)             |
| `JWT_SECRET`                                             | Signs JWTs **and** derives the AES-256-GCM key for BYOK secrets    |
| `JWT_ISSUER`, `ACCESS_TOKEN_EXPIRY`, `REFRESH_TOKEN_EXPIRY` | Token issuer and lifetimes                                      |
| `RESEND_API_KEY`, `RESEND_FROM_ADDR`                     | Resend credentials for OTP / notification email                    |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URI` | Google OAuth 2.0 credentials                              |
| `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URI` | GitHub OAuth 2.0 credentials                              |
| `WORKER_ID`, `GRPC_HOST`, `GRPC_PORT`, `GRPC_MAX_WORKERS` | Python inference worker settings (read by the worker process)     |

> There is no separate `ENCRYPTION_KEY`: BYOK provider keys are encrypted with a key derived from `JWT_SECRET`. Rotating `JWT_SECRET` makes existing stored provider keys unreadable.

### Frontend: see [`frontend/.env`](frontend/.env)

| Variable            | Description                                                                 |
| ------------------- | --------------------------------------------------------------------------- |
| `VITE_API_BASE_URL` | Base URL of the Go API, including the `/api/v1` suffix (defaults to `/api/v1` for the Vite dev proxy) |

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

Load testing scripts live in `backend/scripts/` and `backend/test/load-test.js`. No performance results are published in the repository.
