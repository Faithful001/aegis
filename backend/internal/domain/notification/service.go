package notification

import (
	"context"

	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/google/uuid"
)

type INotificationService interface {
	Create(ctx context.Context, userID uuid.UUID, notification *Notification) error
	List(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

type NotificationService struct {
	repo   INotificationRepository
	alerts user.IAlertSettingsRepository
}

func NewNotificationService(
	repo INotificationRepository,
	alerts user.IAlertSettingsRepository,
) *NotificationService {
	return &NotificationService{
		repo:   repo,
		alerts: alerts,
	}
}

func (s *NotificationService) Create(
	ctx context.Context,
	userID uuid.UUID,
	title string,
	message string,
) error {
	n := &Notification{
		UserID: userID,
		Title:  title,
		Message: message,
	}

	return s.repo.Create(ctx, n)
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
