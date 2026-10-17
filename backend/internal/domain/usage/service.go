package usage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/google/uuid"
)

// UsageService manages model usage records and token metering.
type UsageService struct {
	repo   IUsageRepository
	logger *slog.Logger
}

// NewUsageService returns a new UsageService.
func NewUsageService(repo IUsageRepository, logger *slog.Logger) *UsageService {
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
func (s *UsageService) RecordUsageEvent(ctx context.Context, evt kafka.UsageEvent) (*UsageRecord, error) {
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

// modelPricing holds per-1M-token rates (USD) for known models.
// Rates sourced from official provider pricing pages (approximate as of late 2024).
// Falls back to the "default" entry for unknown models.
var modelPricing = map[string][2]float64{
	// OpenAI
	"gpt-4o":              {2.50, 10.00},
	"gpt-4o-mini":         {0.15, 0.60},
	"gpt-4-turbo":         {10.00, 30.00},
	"gpt-4":               {30.00, 60.00},
	"gpt-3.5-turbo":       {0.50, 1.50},
	"o1":                  {15.00, 60.00},
	"o1-mini":             {3.00, 12.00},
	// Anthropic
	"claude-3-5-sonnet":   {3.00, 15.00},
	"claude-3-5-haiku":    {0.80, 4.00},
	"claude-3-opus":       {15.00, 75.00},
	"claude-3-sonnet":     {3.00, 15.00},
	"claude-3-haiku":      {0.25, 1.25},
	// Google
	"gemini-1.5-pro":      {1.25, 5.00},
	"gemini-1.5-flash":    {0.075, 0.30},
	"gemini-2.0-flash":    {0.10, 0.40},
	"gemini-1.0-pro":      {0.50, 1.50},
	// Mistral
	"mistral-large":       {2.00, 6.00},
	"mistral-small":       {0.20, 0.60},
	"mistral-nemo":        {0.13, 0.13},
	"codestral":           {0.20, 0.60},
	// Meta (via OpenRouter)
	"llama-3.1-405b":      {3.00, 3.00},
	"llama-3.1-70b":       {0.88, 0.88},
	"llama-3.1-8b":        {0.18, 0.18},
	// Fallback
	"default":             {1.00, 3.00},
}

// lookupRate returns the [inputRate, outputRate] per 1M tokens for a given model name.
// Matches on longest prefix, falling back to "default".
func lookupRate(model string) (inputPer1M, outputPer1M float64) {
	best := ""
	for key := range modelPricing {
		if key == "default" {
			continue
		}
		if len(key) > len(best) && len(model) >= len(key) && model[:len(key)] == key {
			best = key
		}
	}
	if best == "" {
		r := modelPricing["default"]
		return r[0], r[1]
	}
	r := modelPricing[best]
	return r[0], r[1]
}

// UsageSummary is the rich aggregate returned by GetTotalUsage.
type UsageSummary struct {
	InputTokens   int
	OutputTokens  int
	TotalTokens   int
	TotalRequests int
	TotalCostUSD  float64
}

// GetTotalUsage returns aggregate token counts, request count, and estimated cost
// for an organization within an optional time range.
func (s *UsageService) GetTotalUsage(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (*UsageSummary, error) {
	inTok, outTok, totTok, reqs, err := s.repo.GetTotalUsage(ctx, orgID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Compute cost from per-model breakdown.
	breakdown, err := s.repo.GetModelBreakdown(ctx, orgID, startTime, endTime)
	if err != nil {
		// Non-fatal: log and continue with zero cost.
		s.logger.Warn("GetModelBreakdown failed, cost will be 0", "error", err)
	}

	var totalCostUSD float64
	for _, row := range breakdown {
		inRate, outRate := lookupRate(row.Model)
		totalCostUSD += float64(row.InputTokens) / 1_000_000.0 * inRate
		totalCostUSD += float64(row.OutputTokens) / 1_000_000.0 * outRate
	}

	return &UsageSummary{
		InputTokens:   inTok,
		OutputTokens:  outTok,
		TotalTokens:   totTok,
		TotalRequests: reqs,
		TotalCostUSD:  totalCostUSD,
	}, nil
}

// ListByOrganization retrieves usage records for an organization with pagination.
func (s *UsageService) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	return s.repo.ListByOrganization(ctx, orgID, limit, offset)
}

// ListByProject retrieves usage records for a project with pagination.
func (s *UsageService) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*UsageRecord, error) {
	return s.repo.ListByProject(ctx, projectID, limit, offset)
}

// GetTokenActivity returns bucketed daily/weekly/cumulative token totals for a GitHub-style heatmap.
// If since is zero it defaults to 12 months ago. granularity must be "daily", "weekly", or "cumulative".
func (s *UsageService) GetTokenActivity(ctx context.Context, orgID uuid.UUID, granularity string, since time.Time) (*TokenActivityResult, error) {
	if since.IsZero() {
		since = time.Now().UTC().AddDate(-1, 0, 0)
	}

	switch granularity {
	case "weekly", "cumulative", "daily":
	default:
		granularity = "daily"
	}

	buckets, err := s.repo.GetTokenActivity(ctx, orgID, granularity, since)
	if err != nil {
		return nil, fmt.Errorf("GetTokenActivity: %w", err)
	}

	// Compute max for UI normalisation.
	var maxTokens int
	for _, b := range buckets {
		if b.Tokens > maxTokens {
			maxTokens = b.Tokens
		}
	}

	return &TokenActivityResult{
		Granularity: granularity,
		Buckets:     buckets,
		MaxTokens:   maxTokens,
	}, nil
}

// TokenActivityResult is the service-layer return type for the heatmap query.
type TokenActivityResult struct {
	Granularity string
	Buckets     []TokenActivityBucket
	MaxTokens   int
}
