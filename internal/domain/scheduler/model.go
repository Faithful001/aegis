package scheduler

import (
	"errors"
)

var (
	// ErrNoWorkersAvailable is returned when no healthy worker with available capacity is present.
	ErrNoWorkersAvailable = errors.New("no healthy workers available")

	// ErrNoWorkersForModel is returned when no registered worker supports the requested LLM model.
	ErrNoWorkersForModel = errors.New("no workers support requested model")

	// ErrInvalidScheduleRequest is returned when the schedule request parameters are invalid.
	ErrInvalidScheduleRequest = errors.New("invalid schedule request")
)

// ScheduleRequest contains parameters needed by the scheduler to select a worker.
type ScheduleRequest struct {
	Model           string
	EstimatedTokens int
	Priority        int
	RetryCount      int
}

// ScheduleResult contains the outcome of a worker selection.
type ScheduleResult struct {
	WorkerID string
	Address  string
	Score    float64
}
