package workerregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Faithful001/aegis/internal/domain/worker"
	goredis "github.com/redis/go-redis/v9"
)

// Key schema:
//   aegis:workers:<workerID>   → JSON-encoded WorkerRecord (with TTL)
//   aegis:workers:index        → Redis Set of all known workerIDs

const (
	keyPrefix    = "aegis:workers:"
	indexKey     = "aegis:workers:index"
	// TTL for each worker key — slightly longer than the heartbeat timeout so
	// the monitor can still read stale records before Redis expires them.
	workerKeyTTL = 5 * time.Minute
)

func workerKey(workerID string) string {
	return keyPrefix + workerID
}

// RedisWorkerRegistry implements worker.WorkerRegistry using Redis.
type RedisWorkerRegistry struct {
	client *goredis.Client
}

// NewRedisWorkerRegistry constructs a registry backed by the given Redis client.
func NewRedisWorkerRegistry(client *goredis.Client) *RedisWorkerRegistry {
	return &RedisWorkerRegistry{client: client}
}

// Register upserts a WorkerRecord into Redis.
func (r *RedisWorkerRegistry) Register(ctx context.Context, payload worker.RegistrationPayload) error {
	now := time.Now().UTC()

	// Try to load existing record so we preserve RegisteredAt.
	existing, _ := r.Get(ctx, payload.WorkerID)

	var registeredAt time.Time
	if existing != nil {
		registeredAt = existing.RegisteredAt
	} else {
		registeredAt = now
	}

	record := &worker.WorkerRecord{
		WorkerID:        payload.WorkerID,
		Address:         payload.Address,
		Status:          payload.Status,
		SupportedModels: payload.SupportedModels,
		ActiveRequests:  0,
		MaxConcurrency:  payload.MaxConcurrency,
		LastHeartbeat:   now,
		RegisteredAt:    registeredAt,
	}

	if err := r.save(ctx, record); err != nil {
		return err
	}

	// Add to the global index set.
	if err := r.client.SAdd(ctx, indexKey, payload.WorkerID).Err(); err != nil {
		return fmt.Errorf("failed to add worker to index: %w", err)
	}

	return nil
}

// Heartbeat refreshes the worker's heartbeat timestamp and runtime counters.
func (r *RedisWorkerRegistry) Heartbeat(ctx context.Context, payload worker.HeartbeatPayload) error {
	record, err := r.Get(ctx, payload.WorkerID)
	if err != nil {
		return err
	}

	record.LastHeartbeat = time.Now().UTC()
	record.Status = payload.Status
	record.ActiveRequests = payload.ActiveRequests
	record.MaxConcurrency = payload.MaxConcurrency

	return r.save(ctx, record)
}

// Deregister removes the worker record and its index entry.
func (r *RedisWorkerRegistry) Deregister(ctx context.Context, workerID string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, workerKey(workerID))
	pipe.SRem(ctx, indexKey, workerID)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to deregister worker %s: %w", workerID, err)
	}
	return nil
}

// Get returns the record for a specific worker.
func (r *RedisWorkerRegistry) Get(ctx context.Context, workerID string) (*worker.WorkerRecord, error) {
	data, err := r.client.Get(ctx, workerKey(workerID)).Bytes()
	if err == goredis.Nil {
		return nil, worker.ErrWorkerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get worker %s from redis: %w", workerID, err)
	}

	var record worker.WorkerRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to decode worker record for %s: %w", workerID, err)
	}
	return &record, nil
}

// List returns all workers currently in the index.
func (r *RedisWorkerRegistry) List(ctx context.Context) ([]*worker.WorkerRecord, error) {
	workerIDs, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list worker IDs: %w", err)
	}

	records := make([]*worker.WorkerRecord, 0, len(workerIDs))
	for _, id := range workerIDs {
		rec, err := r.Get(ctx, id)
		if err == worker.ErrWorkerNotFound {
			// Key expired in Redis but index not yet cleaned — remove stale entry.
			_ = r.client.SRem(ctx, indexKey, id)
			continue
		}
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

// ListHealthy returns only workers in StatusReady state.
func (r *RedisWorkerRegistry) ListHealthy(ctx context.Context) ([]*worker.WorkerRecord, error) {
	all, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	healthy := make([]*worker.WorkerRecord, 0, len(all))
	for _, w := range all {
		if w.IsHealthy() {
			healthy = append(healthy, w)
		}
	}
	return healthy, nil
}

// MarkUnhealthy transitions a worker to StatusUnhealthy.
func (r *RedisWorkerRegistry) MarkUnhealthy(ctx context.Context, workerID string) error {
	return r.setStatus(ctx, workerID, worker.StatusUnhealthy)
}

// MarkOffline transitions a worker to StatusOffline.
func (r *RedisWorkerRegistry) MarkOffline(ctx context.Context, workerID string) error {
	return r.setStatus(ctx, workerID, worker.StatusOffline)
}

// setStatus is a helper that loads, mutates, and saves the worker status.
func (r *RedisWorkerRegistry) setStatus(ctx context.Context, workerID string, status worker.WorkerStatus) error {
	record, err := r.Get(ctx, workerID)
	if err != nil {
		return err
	}
	record.Status = status
	return r.save(ctx, record)
}

// save serialises a WorkerRecord and writes it to Redis with the standard TTL.
func (r *RedisWorkerRegistry) save(ctx context.Context, record *worker.WorkerRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to encode worker record: %w", err)
	}
	if err := r.client.Set(ctx, workerKey(record.WorkerID), data, workerKeyTTL).Err(); err != nil {
		return fmt.Errorf("failed to save worker record to redis: %w", err)
	}
	return nil
}
