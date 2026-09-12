Aegis Backend Implementation Prompt

You are building Aegis, a production-grade distributed AI inference backend designed to demonstrate scalable backend infrastructure, distributed systems, reliability, observability, usage metering, and fault-tolerant billing.

Project description

Distributed AI inference system with high-performance scheduling, rate limiting, streaming, usage metering, and fault-tolerant billing.

This is not a chatbot, RAG application, or simple LLM API wrapper.

The goal is to build the infrastructure behind an AI platform.

The system should demonstrate the kinds of backend engineering problems encountered when operating an AI inference platform at scale.

The implementation must prioritize:

correctness
clean architecture
Domain-Driven Design
distributed systems
concurrency
reliability
fault tolerance
observability
security
testability
clear separation of responsibilities

Do not add complexity simply for the sake of making the project look sophisticated.

Every component must have a clear responsibility.

1. Core architecture

Use a Domain-Driven Design architecture.

The Go backend is the primary API/control plane.

Python is used for inference workers because the Python AI ecosystem and LangChain are being used for model interaction.

The final architecture should look approximately like this:

                         ┌──────────────┐
                         │    Client    │
                         └──────┬───────┘
                                │
                              HTTP
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
                    │ Scheduler           │
                    │ Worker Registry    │
                    │ Streaming          │
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
          Metering         Billing        Analytics
              │               │
              └───────┬───────┘
                      ▼
                 PostgreSQL


                         Redis
              ┌──────────┼──────────┐
              ▼          ▼          ▼
          Rate Limits  Worker     Locks/
                       Registry   Ephemeral State

2. Communication model

There are two primary communication mechanisms between components.

Go → Python

Use gRPC for inference communication.

The Go backend should communicate with Python inference workers through a strongly typed gRPC contract.

Use gRPC for:

inference requests
streaming generated tokens
inference completion
cancellation
deadlines
worker communication where appropriate

Do NOT use REST between Go and Python for the primary inference path.

The public API remains HTTP.

Therefore:

Client
│
│ HTTP
▼
Go
│
│ gRPC
▼
Python
│
│ LangChain
▼
Mistral
Asynchronous communication

Use Kafka for asynchronous events.

Kafka should be used for:

inference job events where appropriate
usage events
billing events
metering events
analytics events
other durable asynchronous workflows

Do NOT use Kafka to stream individual generated tokens to the client.

Token streaming should use:

Python → gRPC → Go → SSE → Client

Kafka is for asynchronous event processing.

3. Repository structure

Use the following structure:

aegis/
│
├── cmd/
│ └── aegis/
│ └── main.go
│
├── internal/
│ │
│ ├── domain/
│ │ ├── auth/
│ │ │   └── dto/
│ │ ├── user/
│ │ │   └── dto/
│ │ ├── organization/
│ │ │   └── dto/
│ │ ├── project/
│ │ │   └── dto/
│ │ ├── inference/
│ │ │   └── dto/
│ │ ├── admission/
│ │ │   └── dto/
│ │ ├── scheduler/
│ │ │   └── dto/
│ │ ├── worker/
│ │ │   └── dto/
│ │ ├── usage/
│ │ │   └── dto/
│ │ ├── pricing/
│ │ │   └── dto/
│ │ ├── billing/
│ │ │   └── dto/
│ │ └── ledger/
│ │     └── dto/
│ │
│ ├── infra/
│ │ ├── db/
│ │ ├── redis/
│ │ ├── kafka/
│ │ ├── middleware/
│ │ ├── outbox/
│ │ └── observability/
│ │
│ └── router/
│ └── router.go
│
├── worker/
│ └── inference/
│ ├── main.py
│ ├── app/
│ ├── chains/
│ ├── models/
│ ├── clients/
│ └── streaming/
│
├── proto/
│ └── inference/
│ └── inference.proto
│
├── migrations/
│
├── tests/
│
├── deploy/
│ └── kubernetes/
│
├── scripts/
│
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
└── README.md

Preserve the existing DDD structure if parts of it are already implemented.

Do not unnecessarily rewrite existing working code.

4. Important cmd architecture decision

Initially, the Go system must have one executable:

cmd/aegis/main.go

Do NOT create:

cmd/gateway/
cmd/scheduler/
cmd/metering/
cmd/billing/

as separate Go applications initially.

The Go backend should run as one process containing:

HTTP API
application services
domain logic
scheduler
admission control
usage processing
billing logic
worker registry
infrastructure integrations

The reason is that these are logical components, not necessarily separate deployable services.

The architecture should allow them to be separated into independent processes later without moving their domain logic.

The Python inference worker is separate because it is a separate runtime and workload.

5. Domain-Driven Design

The internal/domain directory is the core of the application.

Do NOT organize the project around generic technical layers such as:

controllers/
services/
repositories/
models/
utils/

Instead organize around business domains.

For example:

internal/domain/billing/

contains billing concepts and business rules.

internal/domain/scheduler/

contains scheduling concepts and rules.

internal/domain/usage/

contains usage concepts and rules.

Each domain may contain:

entities
value objects
domain services
domain events
repository interfaces
domain errors
business rules

The domain must not depend on:

PostgreSQL
Redis
Kafka
HTTP
Gin
Echo
Fiber
infrastructure implementations 6. Application layer

The application layer coordinates use cases.

For example:

HTTP request
↓
router
↓
application/inference
↓
domain/inference
↓
domain/admission
↓
domain/scheduler

Application services may:

coordinate domain objects
execute use cases
manage transactions
call repository interfaces
publish application events
coordinate multiple domains

Do not place business rules inside HTTP handlers.

7. Infrastructure layer

Infrastructure contains implementations of external dependencies.

Examples:

internal/infra/db/
internal/infra/redis/
internal/infra/kafka/
internal/infra/repositories/
internal/infra/outbox/
internal/infra/observability/

Infrastructure may contain:

PostgreSQL implementations
Redis clients
Kafka producers/consumers
repository implementations
OpenTelemetry
Prometheus
external API clients
hashing implementations

The domain must not import infrastructure.

8. Dependency direction

Maintain this dependency direction:

Router
↓
Application
↓
Domain
↑
Infrastructure

Infrastructure implements interfaces required by inner layers.

Do not reverse this relationship.

9. Go technology

Use:

Go 1.24+
standard net/http
PostgreSQL
Redis
Kafka
gRPC
Protocol Buffers
OpenTelemetry
Prometheus
Docker
Kubernetes

Prefer the standard library where practical.

Avoid unnecessary frameworks.

10. Python technology

Use:

Python 3.12+
LangChain
Mistral API/model integration
asyncio
gRPC Python libraries

Use LangChain.

Do NOT use Pydantic AI.

LangChain should be responsible for model interaction and inference orchestration.

The Python service must not contain business logic belonging to:

billing
scheduling
rate limiting
organization quotas
admission control 11. User domain

Implement users with:

id
email
name
status
created_at
updated_at

Keep user rules in:

internal/domain/user/ 12. Organization domain

Organizations represent tenants.

Organizations own:

users
projects
API keys
usage
billing accounts

Implement strict tenant isolation.

Every request must resolve to the appropriate organization.

A user or API key belonging to one organization must never access another organization's resources.

13. Project domain

Projects belong to organizations.

Projects own:

API keys
inference requests
usage
quotas
configuration

The authentication chain should be:

API Key
↓
Project
↓
Organization 14. Authentication

Implement API-key authentication.

Never store API keys in plaintext.

Store:

key_hash
key_prefix

When creating an API key:

generate cryptographically secure random data
construct the API key
return it once
hash it
store the hash
store a non-secret prefix

Authentication middleware should resolve the API key into an authenticated principal.

15. HTTP API

Implement:

POST /v1/chat/completions

Use an OpenAI-compatible request structure.

Example:

{
"model": "mistral-small",
"messages": [
{
"role": "user",
"content": "Explain distributed systems."
}
],
"max_tokens": 500,
"temperature": 0.7,
"stream": true
}

Support:

model
messages
max_tokens
temperature
stream

Validate all requests.

16. Request lifecycle

The request lifecycle should be:

Client
↓
HTTP Router
↓
Authentication
↓
Authorization
↓
Validation
↓
Idempotency
↓
Distributed Rate Limit
↓
Token Estimation
↓
Admission Control
↓
Scheduler
↓
Worker Selection
↓
gRPC
↓
Python Worker
↓
LangChain
↓
Mistral

For streaming:

Mistral
↓
LangChain
↓
Python Worker
↓
gRPC stream
↓
Go
↓
SSE
↓
Client 17. Request IDs

Generate a unique request ID for every inference request.

Example:

req_01K...

Propagate it through:

HTTP
→ Go
→ scheduler
→ gRPC
→ worker
→ usage
→ billing

Use it for:

logging
tracing
debugging
idempotency
correlating usage and billing 18. Idempotency

Support:

Idempotency-Key: <key>

The same idempotency key must not result in duplicate processing.

Detect:

same key + same request

and return the previous result where appropriate.

Reject:

same key + different request

Use database constraints to guarantee correctness under concurrency.

19. Distributed rate limiting

Use Redis as the shared rate-limiting store.

Rate limits should work correctly if multiple Go backend instances are running.

Support limits at:

API key
Project
Organization

Return:

429 Too Many Requests

The limiter must be atomic.

Do not rely on an in-memory limiter as the source of truth.

20. Admission control

Create:

internal/domain/admission/

The admission controller determines whether a request should:

ACCEPT
QUEUE
REJECT

Consider:

rate limits
quotas
estimated input tokens
max output tokens
model
worker capacity
queue depth
current concurrency
priority

Prevent unbounded queues.

Implement backpressure.

21. Scheduler

Create:

internal/domain/scheduler/

The scheduler selects the appropriate inference worker.

Do not use simple round-robin scheduling as the final algorithm.

Consider:

supported model
worker health
active requests
worker capacity
queue depth
estimated token requirements
request priority
retry count

Document the scheduling strategy.

22. Worker registry

Create:

internal/domain/worker/

Track:

worker_id
status
supported_models
active_requests
max_concurrency
last_heartbeat

Worker states:

STARTING
READY
DRAINING
UNHEALTHY
OFFLINE

Workers must send heartbeats.

Use Redis for the distributed worker registry and heartbeat state.

If a worker stops sending heartbeats:

READY
↓
heartbeat timeout
↓
UNHEALTHY

The scheduler must stop sending new work to unhealthy workers.

23. gRPC contract

Create:

proto/inference/inference.proto

Define a strongly typed inference service.

For example:

service InferenceService {
rpc Generate(GenerateRequest)
returns (stream GenerateResponse);
}

The request should contain information such as:

request_id
job_id
model
messages
max_tokens
temperature

The response should support:

request_id
content
done
input_tokens
output_tokens
finish_reason
error

Generate Go and Python gRPC code from the protobuf definition.

Do not manually maintain duplicate request/response structures in both languages.

The protobuf contract is the source of truth.

24. Python inference worker

The Python worker should:

register with the Go backend
send heartbeats
advertise supported models
accept inference requests through gRPC
invoke LangChain
communicate with Mistral
stream generated tokens
report token usage
handle cancellation
handle timeouts
report failures
gracefully drain during shutdown

Workers are long-running processes.

Do not create a new Python process for every request.

Multiple worker instances should be able to run simultaneously.

25. LangChain

Use LangChain inside the Python inference worker.

The conceptual flow is:

gRPC Request
↓
Inference Worker
↓
LangChain
↓
Mistral
↓
Streaming Output

Keep LangChain-specific implementation details inside the Python worker.

Do not leak LangChain concepts into the Go domain model.

26. Streaming

When:

{
"stream": true
}

the system must stream tokens.

Use:

Python
↓
gRPC streaming
↓
Go
↓
Server-Sent Events
↓
Client

Do not buffer the entire model response.

Handle:

client disconnect
worker failure
cancellation
timeouts
partial responses

When the client disconnects, propagate cancellation to the Python worker where possible.

27. Inference job model

Represent inference jobs explicitly.

A job should contain:

job_id
request_id
organization_id
project_id
model
messages
max_tokens
temperature
priority
created_at
retry_count

Do not put secrets into jobs.

28. Usage domain

Create:

internal/domain/usage/

Usage represents actual model consumption.

Track:

request_id
organization_id
project_id
model
input_tokens
output_tokens
total_tokens
duration
status
created_at

Usage records should become immutable after finalization.

29. Usage metering

The Python inference worker is the component that knows the actual model token usage.

At inference completion it should produce a usage event containing:

event_id
request_id
organization_id
project_id
model
input_tokens
output_tokens
total_tokens
duration_ms
status

Publish the event asynchronously.

Preferred architecture:

Python Worker
│
│ UsageEvent
▼
Kafka
│
▼
Go Metering Consumer
│
▼
PostgreSQL

The Go metering consumer records durable usage.

Do not make PostgreSQL or Redis the mechanism by which Python waits for billing to complete.

Inference should finish independently of billing.

30. Usage event idempotency

Every usage event must have a unique ID.

If Kafka delivers:

evt_123
evt_123

the usage system must record it only once.

Enforce this with database constraints.

Do not rely solely on:

if alreadyProcessed(eventID) {
return
}

because concurrent consumers can race.

Use a unique database constraint and transaction.

31. Pricing domain

Create:

internal/domain/pricing/

Pricing must be independent from usage.

Store model pricing such as:

model
input_price_per_million_tokens
output_price_per_million_tokens
effective_from
effective_to
version

Do not hardcode pricing inside billing code.

Historical charges must retain the pricing information/version used to calculate them.

32. Billing domain

Create:

internal/domain/billing/

Billing converts usage into monetary charges.

Billing must be:

asynchronous
idempotent
transactional
auditable
fault tolerant

Do not implement billing as:

balance = balance - cost

The ledger is the accounting source of truth.

33. Ledger domain

Create:

internal/domain/ledger/

Implement double-entry accounting.

Core concepts:

Account
LedgerTransaction
LedgerEntry
Money

Every ledger transaction must satisfy:

sum(debits) == sum(credits)

For example:

Customer Account
DEBIT $0.0124

Platform Revenue
CREDIT $0.0124

Use PostgreSQL transactions.

Use row-level locking where necessary.

34. Billing idempotency

Use the usage event ID as an idempotency reference.

Processing:

evt_123

multiple times must produce exactly one billing transaction.

Enforce this with database constraints.

35. Transactional outbox

Implement the transactional outbox pattern.

When state and an event must be committed together:

BEGIN

update domain state

insert outbox event

COMMIT

Then a background publisher sends the event to Kafka.

Create:

outbox_events

with:

id
event_type
aggregate_id
payload
created_at
published_at
attempts
last_error

The outbox publisher should retry failed publications.

36. Kafka

Use Kafka for asynchronous processing.

Define topics clearly.

For example:

aegis.inference.events
aegis.usage.events
aegis.billing.events
aegis.dlq

Use explicit event schemas.

Document:

topic purpose
producer
consumer
partitioning strategy
retry strategy
delivery semantics
idempotency strategy 37. Redis

Use Redis for distributed ephemeral state.

Appropriate uses:

rate limiting
worker heartbeats
worker registry
distributed locks
short-lived cache
temporary state

Do not use Redis as the source of truth for:

billing
ledger
historical usage
financial balances

PostgreSQL is the durable source of truth.

38. PostgreSQL

Use PostgreSQL for:

users
organizations
projects
API keys
inference records
usage
pricing
billing accounts
ledger transactions
ledger entries
idempotency records
outbox events

Use migrations.

Add appropriate:

indexes
foreign keys
unique constraints
check constraints

Use the database to enforce important invariants.

39. Retry strategy

Implement retries for transient failures.

Use:

exponential backoff
jitter
maximum retry count

Distinguish:

client errors
transient errors
timeouts
cancellations
permanent errors

Do not retry permanent failures indefinitely.

40. Dead-letter queues

Messages that exceed their retry limit must be moved to a DLQ.

Document:

retry count
backoff strategy
permanent failure criteria
DLQ topic
investigation/recovery procedure 41. Failure handling

The system must explicitly handle:

Worker crash
Worker
↓
processing
↓
crash
↓
heartbeat expires
↓
UNHEALTHY
↓
scheduler stops assigning jobs

Handle unfinished jobs according to their processing state.

Do not blindly duplicate partially streamed responses.

Redis failure

Define the rate-limiting behavior when Redis is unavailable.

Do not accidentally bypass security controls silently.

Kafka failure

Events must not be silently lost.

Use the transactional outbox where durable state and event publication must be coordinated.

PostgreSQL failure

Do not acknowledge durable processing until the database transaction succeeds.

Client disconnect

Cancel downstream inference where possible.

Worker timeout

Terminate/cancel the inference appropriately and record the correct status.

42. Observability

Implement:

structured JSON logging
OpenTelemetry
Prometheus metrics
distributed tracing

Track metrics such as:

http_requests_total
http_request_duration_seconds

inference_requests_total
inference_duration_seconds
inference_errors_total

scheduler_queue_depth
scheduler_jobs_total
scheduler_job_failures_total

worker_active_jobs
worker_capacity
worker_heartbeat_age

usage_events_total
usage_processing_failures_total

billing_transactions_total
billing_failures_total

rate_limit_rejections_total

Propagate:

request_id
job_id
trace_id

Never log:

API keys
passwords
secrets
database credentials 43. Health endpoints

Implement:

GET /health
GET /ready

/health determines whether the process is alive.

/ready determines whether it can safely accept traffic.

44. Graceful shutdown

Handle:

SIGTERM
SIGINT

Shutdown sequence:

stop accepting new requests
↓
stop admission
↓
stop scheduling new work
↓
drain active work
↓
flush important events
↓
close Kafka
↓
close Redis
↓
close PostgreSQL
↓
exit

Every goroutine must have a controlled lifecycle.

Avoid goroutine leaks.

45. Security

Implement:

secure API key generation
API key hashing
authentication
authorization
tenant isolation
request size limits
validation
parameterized SQL
secret management
secure gRPC communication where appropriate

Never expose internal infrastructure publicly.

46. Testing

Write unit tests for:

authentication
authorization
admission
rate limiting
scheduler
worker selection
usage calculation
pricing
billing
ledger
idempotency

Write integration tests for:

PostgreSQL
Redis
Kafka
gRPC

Write end-to-end tests covering:

HTTP
↓
Auth
↓
Admission
↓
Scheduler
↓
gRPC
↓
Python
↓
LangChain
↓
Mistral
↓
Streaming
↓
Usage
↓
Billing
↓
Ledger

Run:

go test ./...
go test -race ./...
go vet ./...

Python:

pytest 47. Concurrency testing

Test:

100+ concurrent inference requests
duplicate idempotency keys
duplicate usage events
concurrent billing
concurrent ledger transactions
multiple workers
multiple Go backend instances
concurrent Redis operations

Use the Go race detector.

48. Load testing

Create load tests that measure:

requests/sec
p50 latency
p95 latency
p99 latency
error rate
queue depth
worker utilization

Test scaling:

1 worker
2 workers
4 workers

Document actual measured results.

Never fabricate performance numbers.

49. Docker Compose

Provide:

Go backend
Python inference worker
PostgreSQL
Redis
Kafka
Prometheus
Grafana

The development environment should start with:

docker compose up

Provide:

.env.example

Never commit secrets.

50. Kubernetes

After the local system works, provide Kubernetes manifests.

Deploy:

aegis
inference-worker

and required infrastructure.

Include:

Deployments
Services
ConfigMaps
Secrets
readiness probes
liveness probes
resource requests/limits
HorizontalPodAutoscaler where appropriate

The important scaling demonstration is that:

Go backend replicas

and:

Python inference worker replicas

can scale independently.

51. Development phases

Implement incrementally.

Do not attempt to generate an enormous codebase without validating each stage.

Phase 1

DDD foundation.

Implement:

cmd
domain
application
infra
router
configuration
logging

Ensure the project builds.

Phase 2

Users, organizations, projects and API keys.

Phase 3

Authentication and authorization.

Phase 4

Basic gRPC inference.

Go
↓
gRPC
↓
Python
↓
LangChain
↓
Mistral
Phase 5

Streaming.

Mistral
↓
Python
↓
gRPC
↓
Go
↓
SSE
Phase 6

Redis-based distributed rate limiting.

Phase 7

Admission control.

Phase 8

Worker registry and heartbeats.

Phase 9

Scheduler and capacity-aware worker selection.

Phase 10

Kafka event infrastructure.

Phase 11

Usage metering.

Phase 12

Pricing.

Phase 13

Billing.

Phase 14

Double-entry ledger.

Phase 15

Transactional outbox.

Phase 16

Retries and DLQs.

Phase 17

Observability.

Phase 18

Failure testing.

Phase 19

Load testing.

Phase 20

Docker and Kubernetes.

52. Final request lifecycle

The final system should support:

                         CLIENT
                           │
                           │ HTTP
                           ▼
                     ┌───────────┐
                     │   Router  │
                     └─────┬─────┘
                           │
                     Authentication
                           │
                     Authorization
                           │
                       Validation
                           │
                       Idempotency
                           │
                     Rate Limiting
                           │
                   Token Estimation
                           │
                  Admission Control
                           │
                       Scheduler
                           │
                    Worker Selection
                           │
                           │ gRPC
                           ▼
                  ┌─────────────────┐
                  │ Python Worker   │
                  │                 │
                  │ LangChain       │
                  └────────┬────────┘
                           │
                           ▼
                         Mistral
                           │
                           │ token stream
                           ▼
                     Python Worker
                           │
                           │ gRPC stream
                           ▼
                      Go Backend
                           │
                           │ SSE
                           ▼
                         CLIENT


                   ASYNCHRONOUS PATH

Python Worker
│
│ UsageEvent
▼
Kafka
│
▼
Metering
│
▼
PostgreSQL
│
▼
Pricing
│
▼
Billing
│
▼
Double-Entry
Ledger 53. Definition of done

Aegis is complete when:

DDD is implemented
domain packages live under internal/domain
application services coordinate use cases
infrastructure is isolated under internal/infra
the Go backend has one primary entrypoint
Python workers run as independent processes
Go and Python communicate through gRPC
LangChain is used
Pydantic AI is not used
Mistral integration works
streaming works
cancellation works
API-key authentication works
tenant isolation works
rate limiting works across multiple Go instances
admission control works
scheduling works
worker heartbeats work
unhealthy workers are removed from scheduling
Kafka events work
usage metering works
usage processing is idempotent
model pricing is versioned
billing is idempotent
double-entry accounting works
ledger transactions always balance
duplicate charges cannot occur
transactional outbox works
retries work
DLQs work
PostgreSQL is the durable source of truth
Redis handles distributed ephemeral state
OpenTelemetry tracing works
Prometheus metrics work
structured logging works
graceful shutdown works
unit tests exist
integration tests exist
end-to-end tests exist
concurrency tests exist
failure scenarios are tested
race detector passes
Docker Compose works
Kubernetes manifests exist
README documents architecture and design decisions 54. Coding-agent rules

Before writing code:

Inspect the existing repository.
Preserve the existing structure where it is sound.
Do not unnecessarily rewrite existing code.
Identify what has already been implemented.
Produce a concise implementation plan.
Implement one phase at a time.
Run tests after every meaningful phase.
Fix compilation/test failures before continuing.
Keep domain logic independent of infrastructure.
Do not put business logic in HTTP handlers.
Do not put business logic in main.go.
Do not create unnecessary microservices.
Do not create separate Go executables for scheduler, billing, metering, etc. initially.
Do not use Pydantic AI.
Use LangChain for the Python inference layer.
Use gRPC for Go ↔ Python inference communication.
Use Kafka for asynchronous events.
Do not use Kafka for token streaming.
Use Redis for distributed ephemeral state.
Use PostgreSQL as the durable source of truth.
Use database constraints to enforce important invariants.
Make asynchronous consumers idempotent.
Do not fabricate benchmark results.
Do not leave fake implementations disguised as production functionality.
Avoid unnecessary abstractions.
Keep the number of components reasonable.
Document important architectural decisions.
Prefer simple, idiomatic implementations over clever ones.
Ensure the project can be explained clearly in a technical interview.

The goal is not to create the largest possible codebase.

The goal is to create a coherent distributed AI inference system that demonstrates strong backend engineering, distributed systems, reliability, correctness, and production thinking.
