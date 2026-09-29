package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// HeartbeatMonitorConfig controls the monitoring loop behaviour.
type HeartbeatMonitorConfig struct {
	// CheckInterval is how often the monitor scans for stale workers.
	CheckInterval time.Duration

	// HeartbeatTimeout is the maximum allowed age of a heartbeat before the
	// worker is transitioned from READY → UNHEALTHY.
	HeartbeatTimeout time.Duration

	// OfflineTimeout is the maximum allowed time a worker can stay UNHEALTHY
	// before it is transitioned to OFFLINE.
	OfflineTimeout time.Duration
}

func DefaultHeartbeatMonitorConfig() HeartbeatMonitorConfig {
	return HeartbeatMonitorConfig{
		CheckInterval:    10 * time.Second,
		HeartbeatTimeout: 30 * time.Second,
		OfflineTimeout:   90 * time.Second,
	}
}

// WorkerService orchestrates worker lifecycle operations on top of the registry.
// It owns the heartbeat monitor goroutine.
type WorkerService struct {
	registry WorkerRegistry
	cfg      HeartbeatMonitorConfig
	logger   *slog.Logger

	mu      sync.Mutex
	stopped bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewWorkerService creates a WorkerService using the provided registry.
func NewWorkerService(registry WorkerRegistry, cfg HeartbeatMonitorConfig, logger *slog.Logger) *WorkerService {
	if logger == nil {
		logger = slog.Default()
	}
	return &WorkerService{
		registry: registry,
		cfg:      cfg,
		logger:   logger,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Register adds a worker to the registry or refreshes an existing record.
func (s *WorkerService) Register(ctx context.Context, payload RegistrationPayload) error {
	if payload.WorkerID == "" {
		return fmt.Errorf("worker registration requires a non-empty worker_id")
	}
	if payload.Address == "" {
		return fmt.Errorf("worker registration requires a non-empty address")
	}
	if payload.Status == "" {
		payload.Status = StatusStarting
	}
	if payload.MaxConcurrency <= 0 {
		payload.MaxConcurrency = 10
	}
	if err := s.registry.Register(ctx, payload); err != nil {
		return fmt.Errorf("failed to register worker %s: %w", payload.WorkerID, err)
	}
	s.logger.Info("worker registered", "worker_id", payload.WorkerID, "address", payload.Address,
		"models", payload.SupportedModels, "max_concurrency", payload.MaxConcurrency)
	return nil
}

// Heartbeat records a fresh heartbeat for an existing worker.
func (s *WorkerService) Heartbeat(ctx context.Context, payload HeartbeatPayload) error {
	if payload.WorkerID == "" {
		return fmt.Errorf("heartbeat requires a non-empty worker_id")
	}
	if err := s.registry.Heartbeat(ctx, payload); err != nil {
		return fmt.Errorf("heartbeat failed for worker %s: %w", payload.WorkerID, err)
	}
	return nil
}

// Deregister gracefully removes a worker from the registry.
func (s *WorkerService) Deregister(ctx context.Context, workerID string) error {
	if err := s.registry.Deregister(ctx, workerID); err != nil {
		return fmt.Errorf("failed to deregister worker %s: %w", workerID, err)
	}
	s.logger.Info("worker deregistered", "worker_id", workerID)
	return nil
}

// GetWorker returns the record for a specific worker.
func (s *WorkerService) GetWorker(ctx context.Context, workerID string) (*WorkerRecord, error) {
	return s.registry.Get(ctx, workerID)
}

// ListWorkers returns all registered workers.
func (s *WorkerService) ListWorkers(ctx context.Context) ([]*WorkerRecord, error) {
	return s.registry.List(ctx)
}

// ListHealthyWorkers returns workers in StatusReady state.
func (s *WorkerService) ListHealthyWorkers(ctx context.Context) ([]*WorkerRecord, error) {
	return s.registry.ListHealthy(ctx)
}

// ListWorkersForModel returns healthy workers that advertise support for model.
func (s *WorkerService) ListWorkersForModel(ctx context.Context, model string) ([]*WorkerRecord, error) {
	healthy, err := s.registry.ListHealthy(ctx)
	if err != nil {
		return nil, err
	}
	var matching []*WorkerRecord
	for _, w := range healthy {
		if w.SupportsModel(model) {
			matching = append(matching, w)
		}
	}
	return matching, nil
}

// StartHeartbeatMonitor begins the background goroutine that detects stale workers.
// It is non-blocking. Call StopHeartbeatMonitor to cleanly shut it down.
func (s *WorkerService) StartHeartbeatMonitor(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		s.logger.Warn("heartbeat monitor already stopped; cannot restart")
		return
	}

	go s.runMonitorLoop(ctx)
	s.logger.Info("heartbeat monitor started",
		"check_interval", s.cfg.CheckInterval,
		"heartbeat_timeout", s.cfg.HeartbeatTimeout,
		"offline_timeout", s.cfg.OfflineTimeout,
	)
}

// StopHeartbeatMonitor signals the monitor to stop and waits for it to finish.
func (s *WorkerService) StopHeartbeatMonitor() {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	close(s.stopCh)
	s.mu.Unlock()

	<-s.doneCh
	s.logger.Info("heartbeat monitor stopped")
}

// runMonitorLoop is the internal goroutine body.
func (s *WorkerService) runMonitorLoop(ctx context.Context) {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkWorkers(ctx)
		}
	}
}

// checkWorkers iterates all workers and transitions stale ones.
func (s *WorkerService) checkWorkers(ctx context.Context) {
	workers, err := s.registry.List(ctx)
	if err != nil {
		s.logger.Error("heartbeat monitor: failed to list workers", "error", err)
		return
	}

	now := time.Now().UTC()
	for _, w := range workers {
		age := now.Sub(w.LastHeartbeat)

		switch w.Status {
		case StatusReady, StatusStarting, StatusDraining:
			if age > s.cfg.HeartbeatTimeout {
				s.logger.Warn("worker missed heartbeat deadline; marking UNHEALTHY",
					"worker_id", w.WorkerID, "last_heartbeat_age", age)
				if err := s.registry.MarkUnhealthy(ctx, w.WorkerID); err != nil {
					s.logger.Error("failed to mark worker unhealthy", "worker_id", w.WorkerID, "error", err)
				}
			}

		case StatusUnhealthy:
			if age > s.cfg.OfflineTimeout {
				s.logger.Warn("unhealthy worker exceeded offline timeout; marking OFFLINE",
					"worker_id", w.WorkerID, "last_heartbeat_age", age)
				if err := s.registry.MarkOffline(ctx, w.WorkerID); err != nil {
					s.logger.Error("failed to mark worker offline", "worker_id", w.WorkerID, "error", err)
				}
			}

		case StatusOffline:
			// Nothing to do. offline workers stay until explicit deregistration.
		}
	}
}
