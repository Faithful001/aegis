package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer implements EventConsumer for subscribing to Kafka topics.
type KafkaConsumer struct {
	brokers []string
	groupID string
	readers []*kafka.Reader
	logger  *slog.Logger
	mu      sync.Mutex
	closed  bool
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(brokers []string, groupID string, logger *slog.Logger) *KafkaConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &KafkaConsumer{
		brokers: brokers,
		groupID: groupID,
		logger:  logger,
	}
}

// Subscribe attaches a handler to process messages for the specified topic.
func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("kafka consumer is closed")
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.brokers,
		GroupID:        c.groupID,
		Topic:          topic,
		MinBytes:       10,    // 10B
		MaxBytes:       10e6,  // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.FirstOffset,
	})
	c.readers = append(c.readers, r)
	c.mu.Unlock()

	c.logger.Info("Subscribed to Kafka topic reader group", "topic", topic, "group_id", c.groupID)

	go func() {
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Error("Kafka consumer reader error", "topic", topic, "error", err)
				return
			}

			if handlerErr := handler(ctx, topic, string(m.Key), m.Value); handlerErr != nil {
				c.logger.Error("Handler error processing Kafka event",
					"topic", topic,
					"key", string(m.Key),
					"error", handlerErr,
				)
			}
		}
	}()

	return nil
}

// Close gracefully stops all background reader loops.
func (c *KafkaConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	var lastErr error
	for _, r := range c.readers {
		if err := r.Close(); err != nil {
			c.logger.Warn("Error closing kafka reader", "error", err)
			lastErr = err
		}
	}
	return lastErr
}
