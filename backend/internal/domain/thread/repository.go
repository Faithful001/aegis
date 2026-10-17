package thread

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrThreadNotFound  = errors.New("chat thread not found")
	ErrMessageNotFound = errors.New("chat message not found")
)

type IThreadRepository interface {
	CreateThread(ctx context.Context, thread *Thread) error
	GetThreadByID(ctx context.Context, userID, threadID uuid.UUID) (*Thread, error)
	ListThreadsByUser(ctx context.Context, userID uuid.UUID) ([]*Thread, error)
	UpdateThread(ctx context.Context, thread *Thread) error
	DeleteThread(ctx context.Context, userID, threadID uuid.UUID) error

	CreateMessage(ctx context.Context, msg *Message) error
	ListMessagesByThread(ctx context.Context, threadID uuid.UUID) ([]*Message, error)
	AppendMessages(ctx context.Context, msgs []*Message) error
}

type ThreadRepository struct {
	db *gorm.DB
}

func NewThreadRepository(db *gorm.DB) IThreadRepository {
	return &ThreadRepository{db: db}
}

func (r *ThreadRepository) CreateThread(ctx context.Context, thread *Thread) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(thread).Error
}

func (r *ThreadRepository) GetThreadByID(ctx context.Context, userID, threadID uuid.UUID) (*Thread, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var t Thread
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ? AND user_id = ?", threadID, userID).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *ThreadRepository) ListThreadsByUser(ctx context.Context, userID uuid.UUID) ([]*Thread, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var threads []*Thread
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&threads).Error
	return threads, err
}

func (r *ThreadRepository) UpdateThread(ctx context.Context, thread *Thread) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Save(thread).Error
}

func (r *ThreadRepository) DeleteThread(ctx context.Context, userID, threadID uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", threadID, userID).
		Delete(&Thread{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrThreadNotFound
	}
	return nil
}

func (r *ThreadRepository) CreateMessage(ctx context.Context, msg *Message) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *ThreadRepository) ListMessagesByThread(ctx context.Context, threadID uuid.UUID) ([]*Message, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var messages []*Message
	err := r.db.WithContext(ctx).
		Where("thread_id = ?", threadID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *ThreadRepository) AppendMessages(ctx context.Context, msgs []*Message) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	if len(msgs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&msgs).Error
}
