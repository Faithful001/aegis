package usage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Faithful001/aegis/internal/infra/events"
	"github.com/google/uuid"
)

// UsageService manages model usage records and token metering.
type UsageService struct {
	repo   UsageRepository
	logger *slog.Logger
}

// NewUsageService returns a new UsageService.
func NewUsageService(repo UsageRepository, logger *slog.Logger) *UsageService {
	if logger == nil {
		logger = slog.Default()
	}
	return &UsageService{
		repo:   repo,
		logger: logger,
	}
}

// RecordUsageEvent processes an incoming domain UsageEvent and persists an immutable UsageRecord.
// Enforces database-level idempotency by gracefully returning on duplicate event IDs.
func (s *UsageService) RecordUsageEvent(ctx context.Context, evt events.UsageEvent) (*UsageRecord, error) {
	rec, err := NewUsageRecord(
		evt.EventID,
		evt.RequestID,
		evt.OrganizationID,
		evt.ProjectID,
		evt.Model,
		evt.InputTokens,
		evt.OutputTokens,
		evt.DurationMS,
		evt.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid usage event: %w", err)
	}

	err = s.repo.Create(ctx, rec)
	if err != nil {
		if errors.Is(err, ErrDuplicateUsageEvent) {
			s.logger.Warn("Duplicate usage event ignored (idempotent)",
				"event_id", evt.EventID,
				"request_id", evt.RequestID,
			)
			// Idempotent: return existing record or newly constructed record without error
			existing, getErr := s.repo.GetByEventID(ctx, evt.EventID)
			if getErr == nil {
				return existing, nil
			}
			return rec, nil
		}
		s.logger.Error("Failed to persist usage record", "error", err, "event_id", evt.EventID)
		return nil, err
	}

	s.logger.Info("Usage record persisted successfully",
		"event_id", rec.EventID,
		"request_id", rec.RequestID,
		"total_tokens", rec.TotalTokens,
		"org_id", rec.OrganizationID,
	)

	return rec, nil
}

// GetTotalUsage returns total token consumption for an organization within an optional time range.
func (s *UsageService) GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (totalInput, totalOutput, totalTokens int, err error) {
	return s.repo.GetTotalUsage(ctx, orgID, startTime, endTime)
}

// ListByOrganization retrieves usage records for an organization with pagination.
func (s *UsageService) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	return s.repo.ListByOrganization(ctx, orgID, limit, offset)
}

// ListByProject retrieves usage records for a project with pagination.
func (s *UsageService) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	return s.repo.ListByProject(ctx, projectID, limit, offset)
}
