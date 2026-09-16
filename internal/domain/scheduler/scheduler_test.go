package scheduler_test

import (
	"context"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/scheduler"
	"github.com/Faithful001/aegis/internal/domain/worker"
	"github.com/Faithful001/aegis/internal/infra/workerregistry"
)

func TestWeightedScoreScheduler_SelectWorker(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid request with empty model", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()
		s := scheduler.NewWeightedScoreScheduler(reg, nil)

		res, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{Model: ""})
		if err != scheduler.ErrInvalidScheduleRequest {
			t.Fatalf("expected ErrInvalidScheduleRequest, got %v", err)
		}
		if res != nil {
			t.Fatalf("expected nil result, got %v", res)
		}
	})

	t.Run("empty registry returns ErrNoWorkersAvailable", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()
		s := scheduler.NewWeightedScoreScheduler(reg, nil)

		_, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{Model: "mistral-small"})
		if err != scheduler.ErrNoWorkersAvailable {
			t.Fatalf("expected ErrNoWorkersAvailable, got %v", err)
		}
	})

	t.Run("no workers supporting target model returns ErrNoWorkersForModel", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w1",
			Address:         "10.0.0.1:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"llama-3-8b"},
			MaxConcurrency:  10,
		})

		s := scheduler.NewWeightedScoreScheduler(reg, nil)
		_, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{Model: "mistral-small"})
		if err != scheduler.ErrNoWorkersForModel {
			t.Fatalf("expected ErrNoWorkersForModel, got %v", err)
		}
	})

	t.Run("unhealthy or zero capacity workers are excluded", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()
		// Unhealthy worker
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w-unhealthy",
			Address:         "10.0.0.1:50051",
			Status:          worker.StatusUnhealthy,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})
		// Full worker
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w-full",
			Address:         "10.0.0.2:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  5,
		})
		_ = reg.Heartbeat(ctx, worker.HeartbeatPayload{
			WorkerID:       "w-full",
			Status:         worker.StatusReady,
			ActiveRequests: 5,
			MaxConcurrency: 5,
		})

		s := scheduler.NewWeightedScoreScheduler(reg, nil)
		_, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{Model: "mistral-small"})
		if err != scheduler.ErrNoWorkersAvailable {
			t.Fatalf("expected ErrNoWorkersAvailable when candidates are unhealthy/full, got %v", err)
		}
	})

	t.Run("selects worker with higher free capacity", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()

		// Worker 1: 8/10 active (20% free)
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w1",
			Address:         "10.0.0.1:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})
		_ = reg.Heartbeat(ctx, worker.HeartbeatPayload{
			WorkerID:       "w1",
			Status:         worker.StatusReady,
			ActiveRequests: 8,
			MaxConcurrency: 10,
		})

		// Worker 2: 2/10 active (80% free)
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w2",
			Address:         "10.0.0.2:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})
		_ = reg.Heartbeat(ctx, worker.HeartbeatPayload{
			WorkerID:       "w2",
			Status:         worker.StatusReady,
			ActiveRequests: 2,
			MaxConcurrency: 10,
		})

		s := scheduler.NewWeightedScoreScheduler(reg, nil)
		res, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{Model: "mistral-small"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.WorkerID != "w2" {
			t.Fatalf("expected w2 (80%% free) to be selected, got %s", res.WorkerID)
		}
		if res.Address != "10.0.0.2:50051" {
			t.Fatalf("expected address 10.0.0.2:50051, got %s", res.Address)
		}
	})

	t.Run("retry count penalizes busier workers", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()

		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w-busy",
			Address:         "10.0.0.1:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})
		_ = reg.Heartbeat(ctx, worker.HeartbeatPayload{
			WorkerID:       "w-busy",
			Status:         worker.StatusReady,
			ActiveRequests: 5,
			MaxConcurrency: 10,
		})

		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "w-idle",
			Address:         "10.0.0.2:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})
		_ = reg.Heartbeat(ctx, worker.HeartbeatPayload{
			WorkerID:       "w-idle",
			Status:         worker.StatusReady,
			ActiveRequests: 1,
			MaxConcurrency: 10,
		})

		s := scheduler.NewWeightedScoreScheduler(reg, nil)
		res, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{
			Model:      "mistral-small",
			RetryCount: 3,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.WorkerID != "w-idle" {
			t.Fatalf("expected w-idle to be selected on retry, got %s", res.WorkerID)
		}
	})

	t.Run("tie breaking deterministically selects lower worker ID", func(t *testing.T) {
		reg := workerregistry.NewMemoryWorkerRegistry()

		// Identical capacity and state
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "worker-b",
			Address:         "10.0.0.2:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})
		_ = reg.Register(ctx, worker.RegistrationPayload{
			WorkerID:        "worker-a",
			Address:         "10.0.0.1:50051",
			Status:          worker.StatusReady,
			SupportedModels: []string{"mistral-small"},
			MaxConcurrency:  10,
		})

		s := scheduler.NewWeightedScoreScheduler(reg, nil)
		res, err := s.SelectWorker(ctx, scheduler.ScheduleRequest{Model: "mistral-small"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.WorkerID != "worker-a" {
			t.Fatalf("expected worker-a (tie break order), got %s", res.WorkerID)
		}
	})
}
