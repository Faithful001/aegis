package admission

import (
	"context"
	"testing"
	"time"
)

func TestAdmissionService_Accept(t *testing.T) {
	cfg := AdmissionConfig{
		MaxActiveConcurrency: 2,
		MaxQueueDepth:        5,
		MaxQueueWaitTime:     100 * time.Millisecond,
		MaxTokensPerRequest:  1000,
	}
	svc := NewService(cfg)

	req := AdmissionRequest{
		RequestID:            "req_01",
		EstimatedInputTokens: 10,
		MaxOutputTokens:      50,
	}

	resp, err := svc.AcquireSlot(context.Background(), req)
	if err != nil {
		t.Fatalf("expected slot acquisition to succeed, got %v", err)
	}
	if resp.Decision != DecisionAccept {
		t.Errorf("expected ACCEPT decision, got %s", resp.Decision)
	}

	active, queued := svc.GetStats()
	if active != 1 {
		t.Errorf("expected 1 active concurrency, got %d", active)
	}
	if queued != 0 {
		t.Errorf("expected 0 queued, got %d", queued)
	}

	svc.ReleaseSlot()
	activeAfter, _ := svc.GetStats()
	if activeAfter != 0 {
		t.Errorf("expected 0 active after release, got %d", activeAfter)
	}
}

func TestAdmissionService_RejectMaxTokensExceeded(t *testing.T) {
	cfg := AdmissionConfig{
		MaxActiveConcurrency: 5,
		MaxQueueDepth:        5,
		MaxQueueWaitTime:     100 * time.Millisecond,
		MaxTokensPerRequest:  100,
	}
	svc := NewService(cfg)

	req := AdmissionRequest{
		RequestID:            "req_oversized",
		EstimatedInputTokens: 50,
		MaxOutputTokens:      100, // Total 150 > 100
	}

	_, err := svc.AcquireSlot(context.Background(), req)
	if err == nil {
		t.Errorf("expected error for oversized token request, got nil")
	}
}

func TestAdmissionService_QueueAndAcquire(t *testing.T) {
	cfg := AdmissionConfig{
		MaxActiveConcurrency: 1,
		MaxQueueDepth:        5,
		MaxQueueWaitTime:     500 * time.Millisecond,
		MaxTokensPerRequest:  1000,
	}
	svc := NewService(cfg)

	req1 := AdmissionRequest{RequestID: "req_1", EstimatedInputTokens: 10, MaxOutputTokens: 10}
	req2 := AdmissionRequest{RequestID: "req_2", EstimatedInputTokens: 10, MaxOutputTokens: 10}

	// Slot 1 acquired immediately
	resp1, err := svc.AcquireSlot(context.Background(), req1)
	if err != nil || resp1.Decision != DecisionAccept {
		t.Fatalf("req1 failed: %v", err)
	}

	// Slot 2 queues in background and succeeds when req1 releases
	doneChan := make(chan bool)
	go func() {
		resp2, err := svc.AcquireSlot(context.Background(), req2)
		if err == nil && resp2.Decision == DecisionAccept {
			doneChan <- true
		} else {
			doneChan <- false
		}
	}()

	time.Sleep(50 * time.Millisecond)
	svc.ReleaseSlot() // Frees slot for req2

	select {
	case success := <-doneChan:
		if !success {
			t.Errorf("expected queued req2 to acquire slot successfully")
		}
	case <-time.After(1 * time.Second):
		t.Errorf("queued request timed out waiting for release signal")
	}
}

func TestAdmissionService_QueueTimeout(t *testing.T) {
	cfg := AdmissionConfig{
		MaxActiveConcurrency: 1,
		MaxQueueDepth:        2,
		MaxQueueWaitTime:     50 * time.Millisecond,
		MaxTokensPerRequest:  1000,
	}
	svc := NewService(cfg)

	req1 := AdmissionRequest{RequestID: "req_1", EstimatedInputTokens: 10, MaxOutputTokens: 10}
	req2 := AdmissionRequest{RequestID: "req_2", EstimatedInputTokens: 10, MaxOutputTokens: 10}

	_, _ = svc.AcquireSlot(context.Background(), req1) // Occupy full capacity

	_, err := svc.AcquireSlot(context.Background(), req2) // Should time out
	if err != ErrQueueTimeout {
		t.Errorf("expected ErrQueueTimeout, got %v", err)
	}
}
