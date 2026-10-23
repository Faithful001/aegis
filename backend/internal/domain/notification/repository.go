package notification

import (
	"context"

	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/google/uuid"
)

type INotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	List(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

type NotificationRepository struct {}

func NewNotificationRepository() INotificationRepository {
	return &NotificationRepository{}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *Notification) error {
	return db.DB.WithContext(ctx).Create(notification).Error
}

func (r *NotificationRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	) ([]Notification, error) {
	var notifications []Notification
	err := db.DB.WithContext(ctx).
				Where("user_id = ?", userID).
				Order("created_at DESC").
				Limit(limit).
				Offset(offset).
				Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) MarkAsRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) error {
	return db.DB.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND id = ?", userID, notificationID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return db.DB.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) CountUnread(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	var count int64
	err := db.DB.WithContext(ctx).Model(&Notification{}).
					Where("user_id = ? AND is_read = false", userID).
					Count(&count).Error
	return int(count), err
}

func (r *NotificationRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) error {
	return db.DB.WithContext(ctx).Model(&Notification{}).
					Where("user_id = ? AND id = ?", userID, notificationID).
					Delete(&Notification{}).Error
}

func (r *NotificationRepository) DeleteAll(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return db.DB.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ?", userID).
		Delete(&Notification{}).Error
}