package notification

import (
	"context"
	"fmt"

	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/infra/email"
	"github.com/google/uuid"
)

type INotificationService interface {
	Create(ctx context.Context, userID uuid.UUID, notifType NotificationType, title string, message string) error
	ProcessNotification(ctx context.Context, userID uuid.UUID, notifType NotificationType, title string, message string) error
	SendToEmail(ctx context.Context, variant NotificationType, userID uuid.UUID, title string, message string) error
	List(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

type NotificationService struct {
	repo         INotificationRepository
	alerts       user.IAlertSettingsRepository
	emailService email.IEmailService
}

func NewNotificationService(
	repo INotificationRepository,
	alerts user.IAlertSettingsRepository,
	emailService email.IEmailService,
) *NotificationService {
	return &NotificationService{
		repo:         repo,
		alerts:       alerts,
		emailService: emailService,
	}
}

func (s *NotificationService) SetEmailService(emailService email.IEmailService) {
	s.emailService = emailService
}

// ProcessNotification checks user alert preferences, persists the in-app notification, and triggers email if configured.
func (s *NotificationService) ProcessNotification(
	ctx context.Context,
	userID uuid.UUID,
	notifType NotificationType,
	title string,
	message string,
) error {
	if s.alerts != nil {
		settings, err := s.alerts.GetByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to retrieve alert settings: %w", err)
		}
		if settings != nil && !s.IsAlertEnabled(settings, notifType) {
			return nil
		}
	}

	if err := s.Create(ctx, userID, notifType, title, message); err != nil {
		return err
	}

	_ = s.SendToEmail(ctx, notifType, userID, title, message)

	return nil
}

func (s *NotificationService) Create(
	ctx context.Context,
	userID uuid.UUID,
	notifType NotificationType,
	title string,
	message string,
) error {
	if notifType == "" {
		notifType = TypeFailedRequest
	}

	n := &Notification{
		UserID:  userID,
		Type:    string(notifType),
		Title:   title,
		Message: message,
	}

	return s.repo.Create(ctx, n)
}

func (s *NotificationService) IsAlertEnabled(settings *user.AlertSettings, notifType NotificationType) bool {
	if settings == nil {
		return true
	}

	switch notifType {
	case TypeBudgetAlert:
		return settings.IsBudgetAlertEnabled
	case TypeRateLimit:
		return settings.IsRateLimitAlertEnabled
	case TypeFailedRequest:
		return settings.IsFailedRequestAlertEnabled
	case TypeModifiedCredential:
		return settings.IsModifiedCredentialAlertEnabled
	case TypeWeeklyUsage:
		return settings.IsWeeklyUsageAlertEnabled
	default:
		return true
	}
}

func (s *NotificationService) SendToEmail(
	ctx context.Context,
	notifType NotificationType,
	userID uuid.UUID,
	title string,
	message string,
) error {
	if s.emailService == nil || s.alerts == nil {
		return nil
	}

	settings, err := s.alerts.GetByUserID(ctx, userID)
	if err != nil || settings == nil || settings.AlertEmail == "" {
		return nil
	}

	if !s.IsAlertEnabled(settings, notifType) {
		return nil
	}

	subject := fmt.Sprintf("[Aegis Alert] %s", title)
	htmlBody := formatAlertEmailHTML(title, message, notifType)

	return s.emailService.SendAlertEmail(ctx, settings.AlertEmail, subject, htmlBody)
}

func formatAlertEmailHTML(title, message string, notifType NotificationType) string {
	badgeColor := "#ef4444"
	badgeText := "Alert"

	switch notifType {
	case TypeBudgetAlert:
		badgeColor = "#f59e0b"
		badgeText = "Budget Alert"
	case TypeRateLimit:
		badgeColor = "#eab308"
		badgeText = "Rate Limit Warning"
	case TypeFailedRequest:
		badgeColor = "#ef4444"
		badgeText = "Failed Request"
	case TypeModifiedCredential:
		badgeColor = "#8b5cf6"
		badgeText = "Security Alert"
	case TypeWeeklyUsage:
		badgeColor = "#3b82f6"
		badgeText = "Weekly Usage"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#09090b;font-family:'Inter',sans-serif;color:#e4e4e7;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 20px;">
    <tr><td align="center">
      <table width="520" cellpadding="0" cellspacing="0" style="background:#18181b;border-radius:16px;border:1px solid #27272a;overflow:hidden;">
        <tr>
          <td style="padding:24px 32px;border-bottom:1px solid #27272a;">
            <div style="display:inline-flex;align-items:center;gap:12px;">
              <span style="font-size:18px;font-weight:700;letter-spacing:-0.5px;color:#fff;">Aegis</span>
            </div>
          </td>
        </tr>
        <tr>
          <td style="padding:32px;">
            <div style="display:inline-block;padding:4px 10px;border-radius:9999px;font-size:12px;font-weight:600;background:%s20;color:%s;border:1px solid %s40;margin-bottom:16px;">%s</div>
            <h2 style="margin:0 0 12px;font-size:20px;font-weight:700;color:#fff;">%s</h2>
            <p style="margin:0 0 24px;color:#a1a1aa;font-size:15px;line-height:1.6;">%s</p>
          </td>
        </tr>
        <tr>
          <td style="padding:16px 32px;border-top:1px solid #27272a;">
            <p style="margin:0;color:#71717a;font-size:12px;">Aegis Platform &bull; Sent to configured alert email</p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, badgeColor, badgeColor, badgeColor, badgeText, title, message)
}

func (s *NotificationService) List(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]Notification, error) {
	return s.repo.List(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkAsRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) error {
	return s.repo.MarkAsRead(ctx, userID, notificationID)
}

func (s *NotificationService) MarkAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) CountUnread(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *NotificationService) Delete(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) error {
	return s.repo.Delete(ctx, userID, notificationID)
}

func (s *NotificationService) DeleteAll(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return s.repo.DeleteAll(ctx, userID)
}
