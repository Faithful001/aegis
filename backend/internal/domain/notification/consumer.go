package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
)

type NotificationConsumer struct {
	eventConsumer       kafka.EventConsumer
	notificationService *NotificationService
	logger              *slog.Logger
}

func NewNotificationConsumer(
	eventConsumer kafka.EventConsumer,
	notificationService *NotificationService,
	logger *slog.Logger,
) *NotificationConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &NotificationConsumer{
		eventConsumer:       eventConsumer,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Start registers subscriptions for all notification event topics on Kafka.
func (c *NotificationConsumer) Start(ctx context.Context) error {
	if c.eventConsumer == nil {
		return fmt.Errorf("event consumer is nil")
	}

	subscriptions := map[string]kafka.EventHandler{
		kafka.TopicFailedRequestEvents:      c.handleFailedRequestEvent,
		kafka.TopicBudgetAlertEvents:        c.handleBudgetAlertEvent,
		kafka.TopicRateLimitEvents:          c.handleRateLimitEvent,
		kafka.TopicModifiedCredentialEvents: c.handleModifiedCredentialEvent,
		kafka.TopicWeeklyUsageEvents:        c.handleWeeklyUsageEvent,
		kafka.TopicNotificationEvents:       c.handleGenericNotificationEvent,
	}

	for topic, handler := range subscriptions {
		if err := c.eventConsumer.Subscribe(ctx, topic, handler); err != nil {
			c.logger.Error("Failed to subscribe notification consumer to topic", "topic", topic, "error", err)
			return fmt.Errorf("failed to subscribe notification consumer to topic %s: %w", topic, err)
		}
		c.logger.Info("Subscribed notification consumer to topic", "topic", topic)
	}

	c.logger.Info("Notification consumer started successfully across all event topics")
	return nil
}

func (c *NotificationConsumer) handleFailedRequestEvent(ctx context.Context, topic, key string, value []byte) error {
	var event FailedRequestEvent
	if err := json.Unmarshal(value, &event); err != nil {
		c.logger.Error("Failed to unmarshal failed request event", "error", err)
		return fmt.Errorf("failed to unmarshal failed request event: %w", err)
	}

	title := event.Title
	if title == "" {
		title = "Request Execution Failed"
	}

	c.logger.Info("Processing failed request event", "user_id", event.UserID, "model", event.Model)
	return c.notificationService.ProcessNotification(ctx, event.UserID, TypeFailedRequest, title, event.Message)
}

func (c *NotificationConsumer) handleBudgetAlertEvent(ctx context.Context, topic, key string, value []byte) error {
	var event BudgetAlertEvent
	if err := json.Unmarshal(value, &event); err != nil {
		c.logger.Error("Failed to unmarshal budget alert event", "error", err)
		return fmt.Errorf("failed to unmarshal budget alert event: %w", err)
	}

	title := event.Title
	if title == "" {
		title = fmt.Sprintf("Budget Alert: %d%% of limit reached", event.Percentage)
	}

	c.logger.Info("Processing budget alert event", "user_id", event.UserID, "current_spend", event.CurrentSpend)
	return c.notificationService.ProcessNotification(ctx, event.UserID, TypeBudgetAlert, title, event.Message)
}

func (c *NotificationConsumer) handleRateLimitEvent(ctx context.Context, topic, key string, value []byte) error {
	var event RateLimitEvent
	if err := json.Unmarshal(value, &event); err != nil {
		c.logger.Error("Failed to unmarshal rate limit event", "error", err)
		return fmt.Errorf("failed to unmarshal rate limit event: %w", err)
	}

	title := event.Title
	if title == "" {
		title = "Rate Limit Warning"
	}

	c.logger.Info("Processing rate limit event", "user_id", event.UserID, "endpoint", event.Endpoint)
	return c.notificationService.ProcessNotification(ctx, event.UserID, TypeRateLimit, title, event.Message)
}

func (c *NotificationConsumer) handleModifiedCredentialEvent(ctx context.Context, topic, key string, value []byte) error {
	var event ModifiedCredentialEvent
	if err := json.Unmarshal(value, &event); err != nil {
		c.logger.Error("Failed to unmarshal modified credential event", "error", err)
		return fmt.Errorf("failed to unmarshal modified credential event: %w", err)
	}

	title := event.Title
	if title == "" {
		title = "Security Alert: Credentials Modified"
	}

	c.logger.Info("Processing modified credential event", "user_id", event.UserID, "action", event.Action)
	return c.notificationService.ProcessNotification(ctx, event.UserID, TypeModifiedCredential, title, event.Message)
}

func (c *NotificationConsumer) handleWeeklyUsageEvent(ctx context.Context, topic, key string, value []byte) error {
	var event WeeklyUsageEvent
	if err := json.Unmarshal(value, &event); err != nil {
		c.logger.Error("Failed to unmarshal weekly usage event", "error", err)
		return fmt.Errorf("failed to unmarshal weekly usage event: %w", err)
	}

	title := event.Title
	if title == "" {
		title = "Your Aegis Weekly Usage Summary"
	}

	c.logger.Info("Processing weekly usage event", "user_id", event.UserID, "total_requests", event.TotalRequests)
	return c.notificationService.ProcessNotification(ctx, event.UserID, TypeWeeklyUsage, title, event.Message)
}

func (c *NotificationConsumer) handleGenericNotificationEvent(ctx context.Context, topic, key string, value []byte) error {
	var event GenericNotificationEvent
	if err := json.Unmarshal(value, &event); err != nil {
		c.logger.Error("Failed to unmarshal generic notification event", "error", err)
		return fmt.Errorf("failed to unmarshal generic notification event: %w", err)
	}

	c.logger.Info("Processing generic notification event", "user_id", event.UserID, "type", event.Type)
	return c.notificationService.ProcessNotification(ctx, event.UserID, event.Type, event.Title, event.Message)
}