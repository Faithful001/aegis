package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer implements event.EventProducer using segmentio/kafka-go.
type KafkaProducer struct {
	brokers []string
	writers map[string]*kafka.Writer // topic -> *kafka.Writer
	logger  *slog.Logger
	mu      sync.Mutex
	closed  bool
}

// NewKafkaProducer constructs a KafkaProducer instance for the specified broker addresses.
func NewKafkaProducer(brokers []string, logger *slog.Logger) *KafkaProducer {
	if logger == nil {
		logger = slog.Default()
	}
	return &KafkaProducer{
		brokers: brokers,
		writers: make(map[string]*kafka.Writer),
		logger:  logger,
	}
}

func (p *KafkaProducer) getWriter(topic string) *kafka.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if w, ok := p.writers[topic]; ok {
		return w
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	p.writers[topic] = w
	return w
}

// Publish serializes and sends a domain event to the specified Kafka topic.
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, evt Event) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("kafka producer is closed")
	}
	p.mu.Unlock()

	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka event: %w", err)
	}

	writer := p.getWriter(topic)
	msg := kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  evt.GetTimestamp(),
	}

	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("Failed to publish message to Kafka",
			"topic", topic,
			"key", key,
			"error", err,
		)
		return fmt.Errorf("failed to write kafka message to topic %s: %w", topic, err)
	}

	p.logger.Debug("Successfully published event to Kafka topic",
		"topic", topic,
		"key", key,
		"event_id", evt.GetID(),
		"event_type", evt.GetType(),
		"timestamp", evt.GetTimestamp(),
	)

	return nil
}

// Close gracefully flushes and closes all topic writers.
func (p *KafkaProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	var lastErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Warn("Error closing kafka writer", "topic", topic, "error", err)
			lastErr = err
		}
	}
	return lastErr
}
