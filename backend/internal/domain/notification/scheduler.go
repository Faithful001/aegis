package notification

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Faithful001/aegis/internal/domain/usage"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
)

type NotificationSchedulerConfig struct {
	BudgetCheckInterval time.Duration
	WeeklyCheckInterval time.Duration
}

func DefaultSchedulerConfig() NotificationSchedulerConfig {
	return NotificationSchedulerConfig{
		BudgetCheckInterval: 1 * time.Hour,
		WeeklyCheckInterval: 6 * time.Hour,
	}
}

// NotificationScheduler runs background loops for scheduled alerts (weekly digests, budget threshold scans).
type NotificationScheduler struct {
	alertsRepo          user.IAlertSettingsRepository
	usageService        *usage.UsageService
	eventProducer       kafka.EventProducer
	notificationService *NotificationService
	logger              *slog.Logger
	cfg                 NotificationSchedulerConfig
	stopCh              chan struct{}
	wg                  sync.WaitGroup
	mu                  sync.Mutex
	running             bool
}

func NewNotificationScheduler(
	alertsRepo user.IAlertSettingsRepository,
	usageService *usage.UsageService,
	eventProducer kafka.EventProducer,
	notificationService *NotificationService,
	logger *slog.Logger,
	cfg ...NotificationSchedulerConfig,
) *NotificationScheduler {
	if logger == nil {
		logger = slog.Default()
	}

	config := DefaultSchedulerConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return &NotificationScheduler{
		alertsRepo:          alertsRepo,
		usageService:        usageService,
		eventProducer:       eventProducer,
		notificationService: notificationService,
		logger:              logger,
		cfg:                 config,
		stopCh:              make(chan struct{}),
	}
}

// Start launches the background goroutines for periodic checks.
func (s *NotificationScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting notification scheduler",
		"budget_interval", s.cfg.BudgetCheckInterval,
		"weekly_interval", s.cfg.WeeklyCheckInterval,
	)

	s.wg.Add(2)
	go s.runBudgetCheckLoop(ctx)
	go s.runWeeklyUsageLoop(ctx)
}

// Stop gracefully terminates the background tickers.
func (s *NotificationScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("Notification scheduler stopped")
}

func (s *NotificationScheduler) runBudgetCheckLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.BudgetCheckInterval)
	defer ticker.Stop()

	// Initial check upon start
	s.CheckBudgetAlerts(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.CheckBudgetAlerts(ctx)
		}
	}
}

func (s *NotificationScheduler) runWeeklyUsageLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.WeeklyCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			now := time.Now().UTC()
			// Run on Sundays
			if now.Weekday() == time.Sunday {
				s.CheckWeeklyUsage(ctx)
			}
		}
	}
}

// CheckBudgetAlerts scans all users with budget alerts enabled, compares month-to-date spend against threshold,
// and emits BudgetAlertEvent if threshold is breached.
func (s *NotificationScheduler) CheckBudgetAlerts(ctx context.Context) {
	if s.alertsRepo == nil || s.usageService == nil {
		return
	}

	settingsList, err := s.alertsRepo.ListBudgetAlertEnabled(ctx)
	if err != nil {
		s.logger.Error("Failed to list budget-alert-enabled users", "error", err)
		return
	}

	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	for _, setting := range settingsList {
		// If an alert was already sent within this billing month, skip to prevent spamming
		if setting.LastAlertSentAt != nil && setting.LastAlertSentAt.After(startOfMonth) {
			continue
		}

		if setting.MonthlyBudgetUSD <= 0 {
			continue
		}

		summary, err := s.usageService.GetTotalUsage(ctx, setting.UserID, startOfMonth, now)
		if err != nil {
			s.logger.Warn("Failed to compute monthly usage for user", "user_id", setting.UserID, "error", err)
			continue
		}

		if setting.IsBudgetExceeded(summary.TotalCostUSD) {
			percentReached := setting.PercentOfBudgetUsed(summary.TotalCostUSD)
			title := fmt.Sprintf("Budget Alert: %d%% of monthly limit reached", percentReached)
			message := fmt.Sprintf(
				"Your API usage this month has reached $%.2f of your configured $%.2f monthly budget (%d%%). Threshold limit is %d%% ($%.2f).",
				summary.TotalCostUSD,
				setting.MonthlyBudgetUSD,
				percentReached,
				setting.ThresholdPercentage,
				setting.ThresholdUSD(),
			)

			evt := NewBudgetAlertEvent(
				setting.UserID,
				setting.UserID,
				summary.TotalCostUSD,
				setting.MonthlyBudgetUSD,
				percentReached,
				title,
				message,
			)

			s.logger.Info("Budget threshold exceeded, dispatching alert",
				"user_id", setting.UserID,
				"spend", summary.TotalCostUSD,
				"budget", setting.MonthlyBudgetUSD,
				"percentage", percentReached,
			)

			// Publish to Kafka, or fallback to direct processing
			if s.eventProducer != nil {
				if err := s.eventProducer.Publish(ctx, kafka.TopicBudgetAlertEvents, setting.UserID.String(), evt); err != nil {
					s.logger.Error("Failed to publish budget alert event to Kafka", "error", err)
				}
			} else if s.notificationService != nil {
				_ = s.notificationService.ProcessNotification(ctx, setting.UserID, TypeBudgetAlert, title, message)
			}

			// Record last alert time
			_ = s.alertsRepo.UpdateLastAlertSentAt(ctx, setting.UserID, now)
		}
	}
}

// CheckWeeklyUsage aggregates the previous 7 days of token and cost data for users with weekly usage enabled.
func (s *NotificationScheduler) CheckWeeklyUsage(ctx context.Context) {
	if s.alertsRepo == nil || s.usageService == nil {
		return
	}

	settingsList, err := s.alertsRepo.ListWeeklyUsageAlertEnabled(ctx)
	if err != nil {
		s.logger.Error("Failed to list weekly-usage-enabled users", "error", err)
		return
	}

	now := time.Now().UTC()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	for _, setting := range settingsList {
		summary, err := s.usageService.GetTotalUsage(ctx, setting.UserID, sevenDaysAgo, now)
		if err != nil {
			s.logger.Warn("Failed to compute weekly usage for user", "user_id", setting.UserID, "error", err)
			continue
		}

		title := "Your Aegis Weekly Usage Summary"
		message := fmt.Sprintf(
			"In the past 7 days, you made %d API requests consuming %d tokens with an estimated cost of $%.2f.",
			summary.TotalRequests,
			summary.TotalTokens,
			summary.TotalCostUSD,
		)

		evt := NewWeeklyUsageEvent(
			setting.UserID,
			setting.UserID,
			int64(summary.TotalRequests),
			int64(summary.TotalTokens),
			summary.TotalCostUSD,
			title,
			message,
		)

		s.logger.Info("Dispatching weekly usage digest",
			"user_id", setting.UserID,
			"requests", summary.TotalRequests,
			"tokens", summary.TotalTokens,
			"cost", summary.TotalCostUSD,
		)

		if s.eventProducer != nil {
			if err := s.eventProducer.Publish(ctx, kafka.TopicWeeklyUsageEvents, setting.UserID.String(), evt); err != nil {
				s.logger.Error("Failed to publish weekly usage event to Kafka", "error", err)
			}
		} else if s.notificationService != nil {
			_ = s.notificationService.ProcessNotification(ctx, setting.UserID, TypeWeeklyUsage, title, message)
		}
	}
}
