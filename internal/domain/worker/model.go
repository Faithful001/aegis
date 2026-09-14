package worker

import (
	"errors"
	"time"
)

var (
	ErrWorkerNotFound  = errors.New("worker not found in registry")
	ErrWorkerUnhealthy = errors.New("worker is unhealthy or offline")
	ErrNoWorkersReady  = errors.New("no healthy workers available")
)

type WorkerStatus string

const (
	// StatusStarting means the worker process has started but is not yet ready.
	StatusStarting WorkerStatus = "STARTING"
	// StatusReady means the worker is healthy and accepting inference requests.
	StatusReady WorkerStatus = "READY"
	// StatusDraining means the worker is finishing existing work before shutting down.
	StatusDraining WorkerStatus = "DRAINING"
	// StatusUnhealthy means the worker missed too many heartbeats; no new work is sent.
	StatusUnhealthy WorkerStatus = "UNHEALTHY"
	// StatusOffline means the worker has been deregistered or timed out completely.
	StatusOffline WorkerStatus = "OFFLINE"
)

// WorkerRecord is the authoritative in-memory representation of a registered worker.
type WorkerRecord struct {
	// WorkerID uniquely identifies the worker process (e.g. "worker-<uuid>").
	WorkerID string `json:"worker_id"`

	// Address is the host:port of the worker's gRPC server.
	Address string `json:"address"`

	// Status is the current lifecycle state of the worker.
	Status WorkerStatus `json:"status"`

	// SupportedModels is the list of model identifiers this worker can serve.
	SupportedModels []string `json:"supported_models"`

	// ActiveRequests is the current in-flight request count reported by the worker.
	ActiveRequests int `json:"active_requests"`

	// MaxConcurrency is the max number of concurrent requests the worker supports.
	MaxConcurrency int `json:"max_concurrency"`

	// LastHeartbeat is the UTC timestamp of the most recent heartbeat.
	LastHeartbeat time.Time `json:"last_heartbeat"`

	// RegisteredAt is the UTC timestamp when the worker first registered.
	RegisteredAt time.Time `json:"registered_at"`
}

// IsHealthy returns true if the worker can accept new inference requests.
func (w *WorkerRecord) IsHealthy() bool {
	return w.Status == StatusReady
}

// HasCapacity returns true if the worker has room for at least one more request.
func (w *WorkerRecord) HasCapacity() bool {
	if w.MaxConcurrency <= 0 {
		return true // unlimited
	}
	return w.ActiveRequests < w.MaxConcurrency
}

// SupportsModel returns true if the worker advertises support for the given model.
func (w *WorkerRecord) SupportsModel(model string) bool {
	for _, m := range w.SupportedModels {
		if m == model {
			return true
		}
	}
	return false
}

// HeartbeatPayload carries the data published by a worker on every heartbeat tick.
type HeartbeatPayload struct {
	WorkerID       string       `json:"worker_id"`
	Status         WorkerStatus `json:"status"`
	ActiveRequests int          `json:"active_requests"`
	MaxConcurrency int          `json:"max_concurrency"`
}

// RegistrationPayload carries the data sent by a worker during initial registration.
type RegistrationPayload struct {
	WorkerID        string       `json:"worker_id"`
	Address         string       `json:"address"`
	SupportedModels []string     `json:"supported_models"`
	MaxConcurrency  int          `json:"max_concurrency"`
	Status          WorkerStatus `json:"status"`
}
