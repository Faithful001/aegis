package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
)

// MeteringConsumer subscribes to usage events on Kafka and delegates recording to UsageService.
type MeteringConsumer struct {
	eventConsumer kafka.EventConsumer
	service       *UsageService
	logger        *slog.Logger
}

// NewMeteringConsumer returns a new MeteringConsumer.
func NewMeteringConsumer(eventConsumer kafka.EventConsumer, service *UsageService, logger *slog.Logger) *MeteringConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &MeteringConsumer{
		eventConsumer: eventConsumer,
		service:       service,
		logger:        logger,
	}
}

// Start subscribes the metering consumer to TopicUsageEvents.
func (c *MeteringConsumer) Start(ctx context.Context) error {
	if c.eventConsumer == nil {
		return fmt.Errorf("event consumer is nil")
	}

	err := c.eventConsumer.Subscribe(ctx, kafka.TopicUsageEvents, c.handleUsageEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe metering consumer to topic %s: %w", kafka.TopicUsageEvents, err)
	}

	c.logger.Info("Metering consumer started successfully", "topic", kafka.TopicUsageEvents)
	return nil
}

func (c *MeteringConsumer) handleUsageEvent(ctx context.Context, topic string, key string, payload []byte) error {
	var evt kafka.UsageEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		c.logger.Error("Failed to unmarshal usage event JSON payload", "topic", topic, "error", err)
		return err
	}

	_, err := c.service.RecordUsageEvent(ctx, evt)
	if err != nil {
		c.logger.Error("Failed to record usage event in metering consumer",
			"event_id", evt.EventID,
			"request_id", evt.RequestID,
			"error", err,
		)
		return err
	}

	return nil
}
