package usage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageRepository defines the persistence contract for usage records.
type IUsageRepository interface {
	Create(ctx context.Context, record *UsageRecord) error
	GetByEventID(ctx context.Context, eventID string) (*UsageRecord, error)
	GetByRequestID(ctx context.Context, reqID string) (*UsageRecord, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error)
	GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens int, err error)
}

// GORMUsageRepository implements UsageRepository using PostgreSQL via GORM.
type UsageRepository struct {
	db *gorm.DB
}

// NewGORMUsageRepository returns a new GORMUsageRepository.
func NewUsageRepository(db *gorm.DB) IUsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Create(ctx context.Context, record *UsageRecord) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	err := r.db.WithContext(ctx).Create(record).Error
	if err != nil {
		// Detect unique constraint violation on event_id
		if strings.Contains(err.Error(), "idx_usage_event_id") ||
			strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateUsageEvent
		}
		return err
	}
	return nil
}

func (r *UsageRepository) GetByEventID(ctx context.Context, eventID string) (*UsageRecord, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var rec UsageRecord
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUsageRecordNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (r *UsageRepository) GetByRequestID(ctx context.Context, reqID string) (*UsageRecord, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var rec UsageRecord
	err := r.db.WithContext(ctx).Where("request_id = ?", reqID).First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUsageRecordNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (r *UsageRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	if limit <= 0 {
		limit = 50
	}

	var records []*UsageRecord
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

func (r *UsageRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	if limit <= 0 {
		limit = 50
	}

	var records []*UsageRecord
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

func (r *UsageRepository) GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens int, err error) {
	if r.db == nil {
		return 0, 0, 0, errors.New("database connection is nil")
	}

	type Result struct {
		Input  int
		Output int
		Total  int
	}

	var res Result
	query := r.db.WithContext(ctx).Model(&UsageRecord{}).
		Select("COALESCE(SUM(input_tokens), 0) as input, COALESCE(SUM(output_tokens), 0) as output, COALESCE(SUM(total_tokens), 0) as total").
		Where("organization_id = ?", orgID)

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Scan(&res).Error; err != nil {
		return 0, 0, 0, err
	}

	return res.Input, res.Output, res.Total, nil
}
