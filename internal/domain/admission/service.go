package admission

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	cfg               AdmissionConfig
	activeConcurrency int64
	queueDepth        int64
	slotChan          chan struct{}
	mu                sync.Mutex
}

func NewService(cfg AdmissionConfig) *Service {
	if cfg.MaxActiveConcurrency <= 0 {
		cfg.MaxActiveConcurrency = 50
	}
	if cfg.MaxQueueDepth <= 0 {
		cfg.MaxQueueDepth = 100
	}
	if cfg.MaxQueueWaitTime <= 0 {
		cfg.MaxQueueWaitTime = 5 * time.Second
	}
	if cfg.MaxTokensPerRequest <= 0 {
		cfg.MaxTokensPerRequest = 8192
	}

	return &Service{
		cfg:      cfg,
		slotChan: make(chan struct{}, cfg.MaxQueueDepth),
	}
}

// EstimateTokens calculates estimated input tokens from character count (1 token ≈ 4 chars)
func EstimateTokens(prompt string, maxOutput int) (inputTokens, totalTokens int) {
	charCount := len(prompt)
	inputTokens = charCount / 4
	if inputTokens == 0 && charCount > 0 {
		inputTokens = 1
	}
	if maxOutput <= 0 {
		maxOutput = 500
	}
	return inputTokens, inputTokens + maxOutput
}

// GetStats returns current concurrency and queue statistics
func (s *Service) GetStats() (active int64, queued int64) {
	return atomic.LoadInt64(&s.activeConcurrency), atomic.LoadInt64(&s.queueDepth)
}

// Evaluate determines whether a request can be accepted immediately, queued, or rejected
func (s *Service) Evaluate(req AdmissionRequest) AdmissionResponse {
	totalTokens := req.TotalEstimatedTokens()
	if totalTokens > s.cfg.MaxTokensPerRequest {
		return AdmissionResponse{
			Decision:    DecisionReject,
			Reason:      fmt.Sprintf("requested total tokens (%d) exceeds max limit (%d)", totalTokens, s.cfg.MaxTokensPerRequest),
			TotalTokens: totalTokens,
		}
	}

	active := atomic.LoadInt64(&s.activeConcurrency)
	if active < int64(s.cfg.MaxActiveConcurrency) {
		return AdmissionResponse{
			Decision:    DecisionAccept,
			Reason:      "system capacity available",
			TotalTokens: totalTokens,
		}
	}

	queued := atomic.LoadInt64(&s.queueDepth)
	if queued < int64(s.cfg.MaxQueueDepth) {
		return AdmissionResponse{
			Decision:    DecisionQueue,
			Reason:      "workers at capacity, request queued",
			TotalTokens: totalTokens,
		}
	}

	return AdmissionResponse{
		Decision:    DecisionReject,
		Reason:      fmt.Sprintf("system overloaded: queue depth (%d) reached max capacity (%d)", queued, s.cfg.MaxQueueDepth),
		TotalTokens: totalTokens,
	}
}

// AcquireSlot handles admission evaluation and queue waiting logic
func (s *Service) AcquireSlot(ctx context.Context, req AdmissionRequest) (*AdmissionResponse, error) {
	resp := s.Evaluate(req)

	switch resp.Decision {
	case DecisionReject:
		return &resp, ErrSystemOverloaded

	case DecisionAccept:
		atomic.AddInt64(&s.activeConcurrency, 1)
		return &resp, nil

	case DecisionQueue:
		atomic.AddInt64(&s.queueDepth, 1)
		defer atomic.AddInt64(&s.queueDepth, -1)

		startTime := time.Now()
		timer := time.NewTimer(s.cfg.MaxQueueWaitTime)
		defer timer.Stop()

		for {
			// Fast path check if active capacity opened up
			s.mu.Lock()
			active := atomic.LoadInt64(&s.activeConcurrency)
			if active < int64(s.cfg.MaxActiveConcurrency) {
				atomic.AddInt64(&s.activeConcurrency, 1)
				s.mu.Unlock()
				resp.Decision = DecisionAccept
				resp.WaitTime = time.Since(startTime)
				resp.Reason = fmt.Sprintf("accepted after waiting %v in queue", resp.WaitTime)
				return &resp, nil
			}
			s.mu.Unlock()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timer.C:
				resp.Decision = DecisionReject
				resp.Reason = fmt.Sprintf("queue wait timeout after %v", s.cfg.MaxQueueWaitTime)
				return &resp, ErrQueueTimeout
			case <-s.slotChan:
				// Slot became available, re-loop to check capacity
			}
		}
	}

	return &resp, nil
}

// ReleaseSlot notifies waiting requests and decrements active concurrency counter
func (s *Service) ReleaseSlot() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if atomic.LoadInt64(&s.activeConcurrency) > 0 {
		atomic.AddInt64(&s.activeConcurrency, -1)
	}

	// Signal one queued waiter if any
	select {
	case s.slotChan <- struct{}{}:
	default:
	}
}
