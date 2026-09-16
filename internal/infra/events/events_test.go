package events_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/Faithful001/aegis/internal/infra/events"
	"github.com/google/uuid"
)

func TestMemoryEventBus_PubSub(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bus := events.NewMemoryEventBus()
	defer func() { _ = bus.Close() }()

	orgID := uuid.New()
	projectID := uuid.New()

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedUsage events.UsageEvent
	var receivedKey string

	err := bus.Subscribe(ctx, events.TopicUsageEvents, func(ctx context.Context, topic string, key string, payload []byte) error {
		defer wg.Done()
		receivedKey = key
		return json.Unmarshal(payload, &receivedUsage)
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	sentEvt := events.NewUsageEvent(
		"evt_test_123",
		"req_test_123",
		orgID,
		projectID,
		"mistral-small",
		15,
		25,
		120,
		"COMPLETED",
	)

	err = bus.Publish(ctx, events.TopicUsageEvents, sentEvt.RequestID, sentEvt)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	wg.Wait()

	if receivedUsage.EventID != "evt_test_123" {
		t.Errorf("expected EventID evt_test_123, got %s", receivedUsage.EventID)
	}
	if receivedUsage.RequestID != "req_test_123" {
		t.Errorf("expected RequestID req_test_123, got %s", receivedUsage.RequestID)
	}
	if receivedUsage.TotalTokens != 40 {
		t.Errorf("expected TotalTokens 40, got %d", receivedUsage.TotalTokens)
	}
	if receivedKey != "req_test_123" {
		t.Errorf("expected partition key req_test_123, got %s", receivedKey)
	}

	published := bus.GetPublished()
	if len(published) != 1 {
		t.Fatalf("expected 1 published record in memory log, got %d", len(published))
	}
}

func TestKafkaProducer_ConstructAndClose(t *testing.T) {
	producer := events.NewKafkaProducer([]string{"localhost:9092"}, nil)
	if producer == nil {
		t.Fatalf("expected non-nil producer")
	}

	err := producer.Close()
	if err != nil {
		t.Fatalf("unexpected producer close error: %v", err)
	}
}
