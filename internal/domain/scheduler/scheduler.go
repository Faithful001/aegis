package scheduler

import (
	"context"
	"log/slog"
	"math"
	"sort"

	"github.com/Faithful001/aegis/internal/domain/worker"
)

// Scheduler defines the contract for selecting an inference worker.
type Scheduler interface {
	SelectWorker(ctx context.Context, req ScheduleRequest) (*ScheduleResult, error)
}

// WeightedScoreScheduler implements Scheduler using a multi-factor capacity scoring algorithm.
type WeightedScoreScheduler struct {
	registry worker.WorkerRegistry
	logger   *slog.Logger
}

// NewWeightedScoreScheduler constructs a new WeightedScoreScheduler.
func NewWeightedScoreScheduler(registry worker.WorkerRegistry, logger *slog.Logger) *WeightedScoreScheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &WeightedScoreScheduler{
		registry: registry,
		logger:   logger,
	}
}

// SelectWorker filters healthy workers supporting the target model and selects the highest scoring node.
func (s *WeightedScoreScheduler) SelectWorker(ctx context.Context, req ScheduleRequest) (*ScheduleResult, error) {
	if req.Model == "" {
		return nil, ErrInvalidScheduleRequest
	}

	allWorkers, err := s.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	if len(allWorkers) == 0 {
		return nil, ErrNoWorkersAvailable
	}

	// Check if any worker supports the requested model (even if currently unhealthy/full)
	var modelSupported bool
	for _, w := range allWorkers {
		if w.SupportsModel(req.Model) {
			modelSupported = true
			break
		}
	}
	if !modelSupported {
		return nil, ErrNoWorkersForModel
	}

	// Filter eligible candidate workers
	var candidates []*worker.WorkerRecord
	for _, w := range allWorkers {
		if w.IsHealthy() && w.HasCapacity() && w.SupportsModel(req.Model) {
			candidates = append(candidates, w)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoWorkersAvailable
	}

	// Score candidates
	type candidateScore struct {
		worker *worker.WorkerRecord
		score  float64
	}

	scores := make([]candidateScore, 0, len(candidates))
	for _, w := range candidates {
		sc := s.calculateScore(w, req)
		scores = append(scores, candidateScore{worker: w, score: sc})
	}

	// Rank candidate workers by score descending, tie-break by WorkerID ascending
	sort.Slice(scores, func(i, j int) bool {
		diff := scores[i].score - scores[j].score
		if math.Abs(diff) > 1e-6 {
			return scores[i].score > scores[j].score
		}
		return scores[i].worker.WorkerID < scores[j].worker.WorkerID
	})

	best := scores[0]
	s.logger.Debug("Selected worker for inference job",
		"worker_id", best.worker.WorkerID,
		"address", best.worker.Address,
		"score", best.score,
		"model", req.Model,
	)

	return &ScheduleResult{
		WorkerID: best.worker.WorkerID,
		Address:  best.worker.Address,
		Score:    best.score,
	}, nil
}

func (s *WeightedScoreScheduler) calculateScore(w *worker.WorkerRecord, req ScheduleRequest) float64 {
	if w.MaxConcurrency <= 0 {
		return 0.0
	}

	freeCapacity := w.MaxConcurrency - w.ActiveRequests
	if freeCapacity < 0 {
		freeCapacity = 0
	}

	freeRatio := float64(freeCapacity) / float64(w.MaxConcurrency)
	baseScore := freeRatio * 60.0

	priorityBoost := 0.0
	if req.Priority > 0 {
		priorityBoost = float64(req.Priority) * freeRatio * 5.0
	}

	retryPenalty := 0.0
	if req.RetryCount > 0 {
		retryPenalty = float64(req.RetryCount) * (1.0 - freeRatio) * 20.0
	}

	score := baseScore + priorityBoost - retryPenalty
	return score
}
