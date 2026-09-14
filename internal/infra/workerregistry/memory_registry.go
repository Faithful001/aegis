package workerregistry

import (
	"context"
	"sync"
	"time"

	"github.com/Faithful001/aegis/internal/domain/worker"
)

// MemoryWorkerRegistry is a thread-safe in-memory implementation of
// worker.WorkerRegistry. Used in tests and when Redis is unavailable.
type MemoryWorkerRegistry struct {
	mu      sync.RWMutex
	records map[string]*worker.WorkerRecord
}

// NewMemoryWorkerRegistry returns an initialised in-memory registry.
func NewMemoryWorkerRegistry() *MemoryWorkerRegistry {
	return &MemoryWorkerRegistry{
		records: make(map[string]*worker.WorkerRecord),
	}
}

func (m *MemoryWorkerRegistry) Register(_ context.Context, payload worker.RegistrationPayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	existing, ok := m.records[payload.WorkerID]

	registeredAt := now
	if ok {
		registeredAt = existing.RegisteredAt
	}

	m.records[payload.WorkerID] = &worker.WorkerRecord{
		WorkerID:        payload.WorkerID,
		Address:         payload.Address,
		Status:          payload.Status,
		SupportedModels: payload.SupportedModels,
		MaxConcurrency:  payload.MaxConcurrency,
		ActiveRequests:  0,
		LastHeartbeat:   now,
		RegisteredAt:    registeredAt,
	}
	return nil
}

func (m *MemoryWorkerRegistry) Heartbeat(_ context.Context, payload worker.HeartbeatPayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.records[payload.WorkerID]
	if !ok {
		return worker.ErrWorkerNotFound
	}
	rec.LastHeartbeat = time.Now().UTC()
	rec.Status = payload.Status
	rec.ActiveRequests = payload.ActiveRequests
	rec.MaxConcurrency = payload.MaxConcurrency
	return nil
}

func (m *MemoryWorkerRegistry) Deregister(_ context.Context, workerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, workerID)
	return nil
}

func (m *MemoryWorkerRegistry) Get(_ context.Context, workerID string) (*worker.WorkerRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.records[workerID]
	if !ok {
		return nil, worker.ErrWorkerNotFound
	}
	// Return a copy to prevent external mutation.
	copy := *rec
	return &copy, nil
}

func (m *MemoryWorkerRegistry) List(_ context.Context) ([]*worker.WorkerRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*worker.WorkerRecord, 0, len(m.records))
	for _, rec := range m.records {
		copy := *rec
		result = append(result, &copy)
	}
	return result, nil
}

func (m *MemoryWorkerRegistry) ListHealthy(ctx context.Context) ([]*worker.WorkerRecord, error) {
	all, err := m.List(ctx)
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

func (m *MemoryWorkerRegistry) MarkUnhealthy(_ context.Context, workerID string) error {
	return m.setStatus(workerID, worker.StatusUnhealthy)
}

func (m *MemoryWorkerRegistry) MarkOffline(_ context.Context, workerID string) error {
	return m.setStatus(workerID, worker.StatusOffline)
}

func (m *MemoryWorkerRegistry) setStatus(workerID string, status worker.WorkerStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[workerID]
	if !ok {
		return worker.ErrWorkerNotFound
	}
	rec.Status = status
	return nil
}
