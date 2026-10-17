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
	// GetTotalUsage returns aggregate token counts and request count for an org.
	GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens, totalRequests int, err error)
	// GetModelBreakdown returns per-model input/output token totals; used for cost estimation.
	GetModelBreakdown(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) ([]ModelUsageRow, error)
	// GetTokenActivity returns bucketed token totals for the heatmap visualisation.
	// granularity must be one of: "daily", "weekly", "cumulative".
	GetTokenActivity(ctx context.Context, orgID uuid.UUID, granularity string, since time.Time) ([]TokenActivityBucket, error)
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

func (r *UsageRepository) GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens, totalRequests int, err error) {
	if r.db == nil {
		return 0, 0, 0, 0, errors.New("database connection is nil")
	}

	type Result struct {
		Input    int
		Output   int
		Total    int
		Requests int
	}

	var res Result
	query := r.db.WithContext(ctx).Model(&UsageRecord{}).
		Select("COALESCE(SUM(input_tokens), 0) as input, COALESCE(SUM(output_tokens), 0) as output, COALESCE(SUM(total_tokens), 0) as total, COUNT(*) as requests").
		Where("organization_id = ?", orgID)

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Scan(&res).Error; err != nil {
		return 0, 0, 0, 0, err
	}

	return res.Input, res.Output, res.Total, res.Requests, nil
}

// ModelUsageRow holds per-model token totals used for cost estimation.
type ModelUsageRow struct {
	Model        string
	InputTokens  int
	OutputTokens int
}

// GetModelBreakdown returns per-model input/output token totals within an optional time range.
func (r *UsageRepository) GetModelBreakdown(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) ([]ModelUsageRow, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var rows []ModelUsageRow
	query := r.db.WithContext(ctx).Model(&UsageRecord{}).
		Select("model, COALESCE(SUM(input_tokens), 0) as input_tokens, COALESCE(SUM(output_tokens), 0) as output_tokens").
		Where("organization_id = ?", orgID).
		Group("model")

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// TokenActivityBucket is the local scan target matching the repository result row.
type TokenActivityBucket struct {
	Date     string
	Tokens   int
	Requests int
}

// GetTokenActivity aggregates token usage into daily, weekly, or cumulative buckets.
func (r *UsageRepository) GetTokenActivity(
	ctx context.Context,
	orgID uuid.UUID,
	granularity string,
	since time.Time,
) ([]TokenActivityBucket, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	type row struct {
		Bucket   string
		Tokens   int
		Requests int
	}

	var rows []row

	switch granularity {
	case "weekly":
		err := r.db.WithContext(ctx).Raw(`
			SELECT
				TO_CHAR(DATE_TRUNC('week', created_at), 'IYYY-"W"IW') AS bucket,
				COALESCE(SUM(total_tokens), 0)                        AS tokens,
				COUNT(*)                                               AS requests
			FROM usage_records
			WHERE organization_id = ? AND created_at >= ?
			GROUP BY 1
			ORDER BY 1 ASC
		`, orgID, since).Scan(&rows).Error
		if err != nil {
			return nil, err
		}

	case "cumulative":
		err := r.db.WithContext(ctx).Raw(`
			SELECT
				TO_CHAR(day_series.day, 'YYYY-MM-DD') AS bucket,
				COALESCE(SUM(COALESCE(d.daily_tokens, 0))   OVER (ORDER BY day_series.day), 0) AS tokens,
				COALESCE(SUM(COALESCE(d.daily_requests, 0)) OVER (ORDER BY day_series.day), 0) AS requests
			FROM (
				SELECT generate_series(DATE_TRUNC('day', ?::timestamptz), NOW(), '1 day'::interval) AS day
			) day_series
			LEFT JOIN (
				SELECT
					DATE_TRUNC('day', created_at) AS day,
					SUM(total_tokens)             AS daily_tokens,
					COUNT(*)                      AS daily_requests
				FROM usage_records
				WHERE organization_id = ? AND created_at >= ?
				GROUP BY 1
			) d ON d.day = day_series.day
			ORDER BY day_series.day ASC
		`, since, orgID, since).Scan(&rows).Error
		if err != nil {
			return nil, err
		}

	default: // "daily"
		err := r.db.WithContext(ctx).Raw(`
			SELECT
				TO_CHAR(day_series.day, 'YYYY-MM-DD')       AS bucket,
				COALESCE(d.daily_tokens, 0)                AS tokens,
				COALESCE(d.daily_requests, 0)              AS requests
			FROM (
				SELECT generate_series(DATE_TRUNC('day', ?::timestamptz), NOW(), '1 day'::interval) AS day
			) day_series
			LEFT JOIN (
				SELECT
					DATE_TRUNC('day', created_at) AS day,
					SUM(total_tokens)             AS daily_tokens,
					COUNT(*)                      AS daily_requests
				FROM usage_records
				WHERE organization_id = ? AND created_at >= ?
				GROUP BY 1
			) d ON d.day = day_series.day
			ORDER BY day_series.day ASC
		`, since, orgID, since).Scan(&rows).Error
		if err != nil {
			return nil, err
		}
	}

	buckets := make([]TokenActivityBucket, len(rows))
	for i, r := range rows {
		buckets[i] = TokenActivityBucket{Date: r.Bucket, Tokens: r.Tokens, Requests: r.Requests}
	}
	return buckets, nil
}
