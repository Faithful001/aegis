package worker

import "context"

type WorkerRegistry interface {
	// Register persists a new worker record. Calling Register again with the
	// same WorkerID updates the existing record.
	Register(ctx context.Context, payload RegistrationPayload) error

	// Heartbeat refreshes the worker's last-seen timestamp and updates its
	// mutable runtime counters (status, active_requests).
	Heartbeat(ctx context.Context, payload HeartbeatPayload) error

	// Deregister removes the worker record from the registry.
	Deregister(ctx context.Context, workerID string) error

	// Get returns the record for a specific worker, or ErrWorkerNotFound.
	Get(ctx context.Context, workerID string) (*WorkerRecord, error)

	// List returns all currently registered worker records.
	List(ctx context.Context) ([]*WorkerRecord, error)

	// ListHealthy returns only workers whose status is StatusReady.
	ListHealthy(ctx context.Context) ([]*WorkerRecord, error)

	// MarkUnhealthy transitions a worker to StatusUnhealthy.
	// Called by the heartbeat monitor when a worker misses its deadline.
	MarkUnhealthy(ctx context.Context, workerID string) error

	// MarkOffline transitions a worker to StatusOffline.
	// Called after a worker has been unhealthy for too long.
	MarkOffline(ctx context.Context, workerID string) error
}
