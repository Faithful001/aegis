package usage

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageRepository defines the persistence contract for usage records.
type UsageRepository interface {
	Create(ctx context.Context, record *UsageRecord) error
	GetByEventID(ctx context.Context, eventID string) (*UsageRecord, error)
	GetByRequestID(ctx context.Context, reqID string) (*UsageRecord, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error)
	GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens int, err error)
}

// GORMUsageRepository implements UsageRepository using PostgreSQL via GORM.
type GORMUsageRepository struct {
	db *gorm.DB
}

// NewGORMUsageRepository returns a new GORMUsageRepository.
func NewGORMUsageRepository(db *gorm.DB) *GORMUsageRepository {
	return &GORMUsageRepository{db: db}
}

func (r *GORMUsageRepository) Create(ctx context.Context, record *UsageRecord) error {
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

func (r *GORMUsageRepository) GetByEventID(ctx context.Context, eventID string) (*UsageRecord, error) {
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

func (r *GORMUsageRepository) GetByRequestID(ctx context.Context, reqID string) (*UsageRecord, error) {
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

func (r *GORMUsageRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
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

func (r *GORMUsageRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
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

func (r *GORMUsageRepository) GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens int, err error) {
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

// MemoryUsageRepository provides an in-memory implementation of UsageRepository.
type MemoryUsageRepository struct {
	mu            sync.RWMutex
	records       []*UsageRecord
	byEventID     map[string]*UsageRecord
	byRequestID   map[string]*UsageRecord
}

// NewMemoryUsageRepository returns an initialised MemoryUsageRepository.
func NewMemoryUsageRepository() *MemoryUsageRepository {
	return &MemoryUsageRepository{
		records:     make([]*UsageRecord, 0),
		byEventID:   make(map[string]*UsageRecord),
		byRequestID: make(map[string]*UsageRecord),
	}
}

func (m *MemoryUsageRepository) Create(_ context.Context, record *UsageRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.byEventID[record.EventID]; exists {
		return ErrDuplicateUsageEvent
	}

	copyRec := *record
	m.records = append(m.records, &copyRec)
	m.byEventID[record.EventID] = &copyRec
	m.byRequestID[record.RequestID] = &copyRec
	return nil
}

func (m *MemoryUsageRepository) GetByEventID(_ context.Context, eventID string) (*UsageRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, ok := m.byEventID[eventID]
	if !ok {
		return nil, ErrUsageRecordNotFound
	}
	copyRec := *rec
	return &copyRec, nil
}

func (m *MemoryUsageRepository) GetByRequestID(_ context.Context, reqID string) (*UsageRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, ok := m.byRequestID[reqID]
	if !ok {
		return nil, ErrUsageRecordNotFound
	}
	copyRec := *rec
	return &copyRec, nil
}

func (m *MemoryUsageRepository) ListByOrganization(_ context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*UsageRecord
	for _, rec := range m.records {
		if rec.OrganizationID == orgID {
			copyRec := *rec
			matched = append(matched, &copyRec)
		}
	}

	if offset >= len(matched) {
		return []*UsageRecord{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(matched) {
		end = len(matched)
	}
	return matched[offset:end], nil
}

func (m *MemoryUsageRepository) ListByProject(_ context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*UsageRecord
	for _, rec := range m.records {
		if rec.ProjectID == projectID {
			copyRec := *rec
			matched = append(matched, &copyRec)
		}
	}

	if offset >= len(matched) {
		return []*UsageRecord{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(matched) {
		end = len(matched)
	}
	return matched[offset:end], nil
}

func (m *MemoryUsageRepository) GetTotalUsage(_ context.Context, orgID uuid.UUID, startTime, endTime time.Time) (int, int, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalInput, totalOutput, totalTokens int
	for _, rec := range m.records {
		if rec.OrganizationID == orgID {
			if !startTime.IsZero() && rec.CreatedAt.Before(startTime) {
				continue
			}
			if !endTime.IsZero() && rec.CreatedAt.After(endTime) {
				continue
			}
			totalInput += rec.InputTokens
			totalOutput += rec.OutputTokens
			totalTokens += rec.TotalTokens
		}
	}
	return totalInput, totalOutput, totalTokens, nil
}
