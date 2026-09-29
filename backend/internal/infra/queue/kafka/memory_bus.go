package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// MemoryEventBus implements both EventProducer and EventConsumer
// using in-memory channels. Used in tests and offline fallback modes.
type MemoryEventBus struct {
	mu        sync.RWMutex
	handlers  map[string][]EventHandler
	published []publishedRecord
	closed    bool
}

type publishedRecord struct {
	Topic   string
	Key     string
	Payload []byte
}

// NewMemoryEventBus constructs an initialised MemoryEventBus.
func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (m *MemoryEventBus) Publish(ctx context.Context, topic string, key string, evt Event) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return fmt.Errorf("memory event bus is closed")
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	m.published = append(m.published, publishedRecord{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	})

	handlers := make([]EventHandler, len(m.handlers[topic]))
	copy(handlers, m.handlers[topic])
	m.mu.Unlock()

	// Dispatch to subscribers asynchronously
	for _, h := range handlers {
		handler := h
		go func() {
			_ = handler(ctx, topic, key, payload)
		}()
	}

	return nil
}

func (m *MemoryEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return fmt.Errorf("memory event bus is closed")
	}
	m.handlers[topic] = append(m.handlers[topic], handler)
	return nil
}

func (m *MemoryEventBus) GetPublished() []publishedRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]publishedRecord, len(m.published))
	copy(result, m.published)
	return result
}

func (m *MemoryEventBus) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
