package admission

import (
	"errors"
	"time"
)

var (
	ErrSystemOverloaded  = errors.New("system overloaded: request rejected by admission control")
	ErrMaxTokensExceeded = errors.New("requested token count exceeds maximum allowed limit")
	ErrQueueTimeout      = errors.New("queue timeout: request timed out waiting for worker slot")
)

type AdmissionDecision string

const (
	DecisionAccept AdmissionDecision = "ACCEPT"
	DecisionQueue  AdmissionDecision = "QUEUE"
	DecisionReject AdmissionDecision = "REJECT"
)

type AdmissionConfig struct {
	MaxActiveConcurrency int           // Max active requests processing concurrently across workers
	MaxQueueDepth        int           // Max requests allowed to wait in queue before REJECTing
	MaxQueueWaitTime     time.Duration // Max duration a request can wait in QUEUE before timing out
	MaxTokensPerRequest  int           // Max total allowed tokens (input + output) per request
}

func DefaultConfig() AdmissionConfig {
	return AdmissionConfig{
		MaxActiveConcurrency: 50,
		MaxQueueDepth:        100,
		MaxQueueWaitTime:     5 * time.Second,
		MaxTokensPerRequest:  8192,
	}
}

type AdmissionRequest struct {
	RequestID            string
	Model                string
	EstimatedInputTokens int
	MaxOutputTokens      int
	Priority             int
}

func (r *AdmissionRequest) TotalEstimatedTokens() int {
	return r.EstimatedInputTokens + r.MaxOutputTokens
}

type AdmissionResponse struct {
	Decision    AdmissionDecision
	Reason      string
	WaitTime    time.Duration
	TotalTokens int
}
