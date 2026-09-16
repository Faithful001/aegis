package usage_test

import (
	"context"
	"testing"
	"time"

	"github.com/Faithful001/aegis/internal/domain/usage"
	"github.com/Faithful001/aegis/internal/infra/events"
	"github.com/google/uuid"
)

func TestUsageService_RecordUsageEvent_Idempotency(t *testing.T) {
	ctx := context.Background()
	repo := usage.NewMemoryUsageRepository()
	service := usage.NewUsageService(repo, nil)

	orgID := uuid.New()
	projectID := uuid.New()

	evt := events.NewUsageEvent(
		"evt_idempotent_1",
		"req_123",
		orgID,
		projectID,
		"mistral-small",
		100,
		50,
		250,
		"COMPLETED",
	)

	// First record call
	rec1, err := service.RecordUsageEvent(ctx, evt)
	if err != nil {
		t.Fatalf("first RecordUsageEvent call failed: %v", err)
	}
	if rec1.TotalTokens != 150 {
		t.Errorf("expected 150 total tokens, got %d", rec1.TotalTokens)
	}

	// Duplicate record call (same EventID)
	rec2, err := service.RecordUsageEvent(ctx, evt)
	if err != nil {
		t.Fatalf("duplicate RecordUsageEvent call should succeed idempotently, got error: %v", err)
	}
	if rec2.EventID != "evt_idempotent_1" {
		t.Errorf("expected EventID evt_idempotent_1, got %s", rec2.EventID)
	}

	// Verify only 1 record exists for org
	records, err := service.ListByOrganization(ctx, orgID, 10, 0)
	if err != nil {
		t.Fatalf("failed to list by org: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record in repository due to idempotency, got %d", len(records))
	}
}

func TestUsageService_GetTotalUsage(t *testing.T) {
	ctx := context.Background()
	repo := usage.NewMemoryUsageRepository()
	service := usage.NewUsageService(repo, nil)

	orgID := uuid.New()
	projectID := uuid.New()

	evt1 := events.NewUsageEvent("evt_1", "req_1", orgID, projectID, "mistral-small", 10, 20, 100, "COMPLETED")
	evt2 := events.NewUsageEvent("evt_2", "req_2", orgID, projectID, "mistral-small", 30, 40, 200, "COMPLETED")

	_, _ = service.RecordUsageEvent(ctx, evt1)
	_, _ = service.RecordUsageEvent(ctx, evt2)

	totalInput, totalOutput, totalTokens, err := service.GetTotalUsage(ctx, orgID, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("failed to get total usage: %v", err)
	}

	if totalInput != 40 {
		t.Errorf("expected total input 40, got %d", totalInput)
	}
	if totalOutput != 60 {
		t.Errorf("expected total output 60, got %d", totalOutput)
	}
	if totalTokens != 100 {
		t.Errorf("expected total tokens 100, got %d", totalTokens)
	}
}

func TestMeteringConsumer_HandleUsageEvent(t *testing.T) {
	ctx := context.Background()
	repo := usage.NewMemoryUsageRepository()
	service := usage.NewUsageService(repo, nil)
	bus := events.NewMemoryEventBus()

	consumer := usage.NewMeteringConsumer(bus, service, nil)
	err := consumer.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start metering consumer: %v", err)
	}

	orgID := uuid.New()
	projectID := uuid.New()
	sentEvt := events.NewUsageEvent("evt_consumer_1", "req_consumer_1", orgID, projectID, "mistral-small", 50, 50, 300, "COMPLETED")

	err = bus.Publish(ctx, events.TopicUsageEvents, sentEvt.RequestID, sentEvt)
	if err != nil {
		t.Fatalf("failed to publish to memory bus: %v", err)
	}

	// Wait briefly for goroutine async dispatch
	time.Sleep(50 * time.Millisecond)

	rec, err := repo.GetByEventID(ctx, "evt_consumer_1")
	if err != nil {
		t.Fatalf("expected record in repository after consumer handling, got %v", err)
	}
	if rec.TotalTokens != 100 {
		t.Errorf("expected total tokens 100, got %d", rec.TotalTokens)
	}
}
