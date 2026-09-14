package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/Faithful001/aegis/internal/domain/worker"
	"github.com/Faithful001/aegis/internal/infra/workerregistry"
)

// newTestService creates a WorkerService backed by an in-memory registry.
func newTestService(cfg worker.HeartbeatMonitorConfig) *worker.WorkerService {
	reg := workerregistry.NewMemoryWorkerRegistry()
	return worker.NewWorkerService(reg, cfg, nil)
}

func defaultPayload(id string) worker.RegistrationPayload {
	return worker.RegistrationPayload{
		WorkerID:        id,
		Address:         "localhost:50051",
		SupportedModels: []string{"mistral-small", "mistral-small-latest"},
		MaxConcurrency:  5,
		Status:          worker.StatusReady,
	}
}

// ---------------------------------------------------------------------------
// Registration tests
// ---------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	ctx := context.Background()

	if err := svc.Register(ctx, defaultPayload("w-1")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	rec, err := svc.GetWorker(ctx, "w-1")
	if err != nil {
		t.Fatalf("expected worker to exist after registration: %v", err)
	}
	if rec.WorkerID != "w-1" {
		t.Errorf("expected worker_id w-1, got %s", rec.WorkerID)
	}
	if rec.Status != worker.StatusReady {
		t.Errorf("expected status READY, got %s", rec.Status)
	}
}

func TestRegister_MissingWorkerID(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	err := svc.Register(context.Background(), worker.RegistrationPayload{Address: "localhost:50051"})
	if err == nil {
		t.Fatal("expected error for empty worker_id, got nil")
	}
}

func TestRegister_MissingAddress(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	err := svc.Register(context.Background(), worker.RegistrationPayload{WorkerID: "w-1"})
	if err == nil {
		t.Fatal("expected error for empty address, got nil")
	}
}

func TestRegister_Upsert(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	ctx := context.Background()

	_ = svc.Register(ctx, defaultPayload("w-1"))
	rec1, _ := svc.GetWorker(ctx, "w-1")
	registeredAt := rec1.RegisteredAt

	// Re-register with a different address.
	updated := defaultPayload("w-1")
	updated.Address = "localhost:50052"
	_ = svc.Register(ctx, updated)

	rec2, _ := svc.GetWorker(ctx, "w-1")
	if rec2.Address != "localhost:50052" {
		t.Errorf("expected updated address, got %s", rec2.Address)
	}
	// RegisteredAt should be preserved.
	if !rec2.RegisteredAt.Equal(registeredAt) {
		t.Errorf("RegisteredAt should not change on re-registration")
	}
}

// ---------------------------------------------------------------------------
// Heartbeat tests
// ---------------------------------------------------------------------------

func TestHeartbeat_UpdatesStatus(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	ctx := context.Background()

	_ = svc.Register(ctx, defaultPayload("w-2"))

	err := svc.Heartbeat(ctx, worker.HeartbeatPayload{
		WorkerID:       "w-2",
		Status:         worker.StatusReady,
		ActiveRequests: 3,
		MaxConcurrency: 5,
	})
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}

	rec, _ := svc.GetWorker(ctx, "w-2")
	if rec.ActiveRequests != 3 {
		t.Errorf("expected active_requests=3, got %d", rec.ActiveRequests)
	}
}

func TestHeartbeat_UnknownWorker(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	err := svc.Heartbeat(context.Background(), worker.HeartbeatPayload{WorkerID: "ghost"})
	if err == nil {
		t.Fatal("expected error heartbeating an unknown worker")
	}
}

// ---------------------------------------------------------------------------
// Deregistration tests
// ---------------------------------------------------------------------------

func TestDeregister_RemovesWorker(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	ctx := context.Background()

	_ = svc.Register(ctx, defaultPayload("w-3"))
	_ = svc.Deregister(ctx, "w-3")

	_, err := svc.GetWorker(ctx, "w-3")
	if err != worker.ErrWorkerNotFound {
		t.Errorf("expected ErrWorkerNotFound after deregistration, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Model filtering tests
// ---------------------------------------------------------------------------

func TestListWorkersForModel(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	ctx := context.Background()

	p1 := defaultPayload("w-4")
	p1.SupportedModels = []string{"mistral-small"}
	_ = svc.Register(ctx, p1)

	p2 := defaultPayload("w-5")
	p2.SupportedModels = []string{"gpt-4o"}
	_ = svc.Register(ctx, p2)

	matches, err := svc.ListWorkersForModel(ctx, "mistral-small")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 worker for mistral-small, got %d", len(matches))
	}
	if matches[0].WorkerID != "w-4" {
		t.Errorf("expected worker w-4, got %s", matches[0].WorkerID)
	}
}

func TestListWorkersForModel_NoMatch(t *testing.T) {
	svc := newTestService(worker.DefaultHeartbeatMonitorConfig())
	ctx := context.Background()
	_ = svc.Register(ctx, defaultPayload("w-6"))

	matches, err := svc.ListWorkersForModel(ctx, "unknown-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

// ---------------------------------------------------------------------------
// Heartbeat monitor tests
// ---------------------------------------------------------------------------

func TestHeartbeatMonitor_MarksUnhealthy(t *testing.T) {
	cfg := worker.HeartbeatMonitorConfig{
		CheckInterval:    20 * time.Millisecond,
		HeartbeatTimeout: 50 * time.Millisecond,
		OfflineTimeout:   5 * time.Second,
	}
	svc := newTestService(cfg)
	ctx := context.Background()

	_ = svc.Register(ctx, defaultPayload("w-7"))
	svc.StartHeartbeatMonitor(ctx)
	defer svc.StopHeartbeatMonitor()

	// Wait long enough for the monitor to fire and detect the stale worker.
	time.Sleep(200 * time.Millisecond)

	rec, err := svc.GetWorker(ctx, "w-7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Status != worker.StatusUnhealthy {
		t.Errorf("expected worker to be UNHEALTHY after missing heartbeat, got %s", rec.Status)
	}
}

func TestHeartbeatMonitor_MarksOffline(t *testing.T) {
	cfg := worker.HeartbeatMonitorConfig{
		CheckInterval:    20 * time.Millisecond,
		HeartbeatTimeout: 30 * time.Millisecond,
		OfflineTimeout:   50 * time.Millisecond,
	}
	svc := newTestService(cfg)
	ctx := context.Background()

	_ = svc.Register(ctx, defaultPayload("w-8"))
	svc.StartHeartbeatMonitor(ctx)
	defer svc.StopHeartbeatMonitor()

	// Wait long enough to pass both timeouts.
	time.Sleep(300 * time.Millisecond)

	rec, err := svc.GetWorker(ctx, "w-8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Status != worker.StatusOffline {
		t.Errorf("expected worker to be OFFLINE, got %s", rec.Status)
	}
}

func TestHeartbeatMonitor_StaysHealthyWithHeartbeats(t *testing.T) {
	cfg := worker.HeartbeatMonitorConfig{
		CheckInterval:    20 * time.Millisecond,
		HeartbeatTimeout: 80 * time.Millisecond,
		OfflineTimeout:   5 * time.Second,
	}
	svc := newTestService(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = svc.Register(ctx, defaultPayload("w-9"))
	svc.StartHeartbeatMonitor(ctx)
	defer svc.StopHeartbeatMonitor()

	// Send heartbeats every 30ms for 200ms — well within the timeout.
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()
		deadline := time.After(200 * time.Millisecond)
		for {
			select {
			case <-deadline:
				return
			case <-ticker.C:
				_ = svc.Heartbeat(ctx, worker.HeartbeatPayload{
					WorkerID: "w-9",
					Status:   worker.StatusReady,
				})
			}
		}
	}()
	<-done

	rec, err := svc.GetWorker(ctx, "w-9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Status != worker.StatusReady {
		t.Errorf("expected worker to remain READY with active heartbeats, got %s", rec.Status)
	}
}

// ---------------------------------------------------------------------------
// WorkerRecord helper method tests
// ---------------------------------------------------------------------------

func TestWorkerRecord_IsHealthy(t *testing.T) {
	cases := []struct {
		status worker.WorkerStatus
		want   bool
	}{
		{worker.StatusReady, true},
		{worker.StatusStarting, false},
		{worker.StatusDraining, false},
		{worker.StatusUnhealthy, false},
		{worker.StatusOffline, false},
	}
	for _, tc := range cases {
		w := &worker.WorkerRecord{Status: tc.status}
		if got := w.IsHealthy(); got != tc.want {
			t.Errorf("IsHealthy() for %s: want %v, got %v", tc.status, tc.want, got)
		}
	}
}

func TestWorkerRecord_HasCapacity(t *testing.T) {
	w := &worker.WorkerRecord{MaxConcurrency: 3, ActiveRequests: 2}
	if !w.HasCapacity() {
		t.Error("expected capacity available")
	}
	w.ActiveRequests = 3
	if w.HasCapacity() {
		t.Error("expected no capacity at max")
	}
}

func TestWorkerRecord_SupportsModel(t *testing.T) {
	w := &worker.WorkerRecord{SupportedModels: []string{"mistral-small", "mistral-small-latest"}}
	if !w.SupportsModel("mistral-small") {
		t.Error("expected model to be supported")
	}
	if w.SupportsModel("gpt-4o") {
		t.Error("expected model to be unsupported")
	}
}
