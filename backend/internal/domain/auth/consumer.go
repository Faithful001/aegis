package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Faithful001/aegis/internal/domain/otp"
	"github.com/Faithful001/aegis/internal/infra/email"
	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
)

// EmailConsumer subscribes to asynchronous email events on Kafka and delivers them via IEmailService.
type EmailConsumer struct {
	eventConsumer kafka.EventConsumer
	emailService  email.IEmailService
	logger        *slog.Logger
}

// NewEmailConsumer creates a new EmailConsumer instance.
func NewEmailConsumer(eventConsumer kafka.EventConsumer, emailService email.IEmailService, logger *slog.Logger) *EmailConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &EmailConsumer{
		eventConsumer: eventConsumer,
		emailService:  emailService,
		logger:        logger,
	}
}

// Start begins consuming email events from TopicEmailEvents.
func (c *EmailConsumer) Start(ctx context.Context) error {
	if c.eventConsumer == nil {
		return fmt.Errorf("event consumer is nil")
	}
	if c.emailService == nil {
		return fmt.Errorf("email service is nil")
	}

	err := c.eventConsumer.Subscribe(ctx, kafka.TopicEmailEvents, c.handleEmailEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe email consumer to topic %s: %w", kafka.TopicEmailEvents, err)
	}

	c.logger.Info("Email consumer started successfully", "topic", kafka.TopicEmailEvents)
	return nil
}

func (c *EmailConsumer) handleEmailEvent(ctx context.Context, topic string, key string, payload []byte) error {
	var evt kafka.EmailEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		c.logger.Error("Failed to unmarshal email event payload", "topic", topic, "error", err)
		return err
	}

	c.logger.Info("Consuming email event from Kafka",
		"event_id", evt.EventID,
		"to_email", evt.ToEmail,
		"purpose", evt.Purpose,
	)

	err := c.emailService.SendOTPEmail(ctx, evt.ToEmail, evt.Name, evt.OTPCode, otp.OTPPurpose(evt.Purpose))
	if err != nil {
		c.logger.Error("Failed to deliver OTP email in Kafka consumer",
			"event_id", evt.EventID,
			"to_email", evt.ToEmail,
			"error", err,
		)
		return err
	}

	c.logger.Info("Successfully sent email via Kafka consumer",
		"event_id", evt.EventID,
		"to_email", evt.ToEmail,
	)
	return nil
}
