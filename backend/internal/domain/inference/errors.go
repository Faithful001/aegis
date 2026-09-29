package inference

import "errors"

var (
	ErrModelNotSupported   = errors.New("requested model is not supported")
	ErrEmptyMessages       = errors.New("messages list cannot be empty")
	ErrInvalidTemperature  = errors.New("temperature must be between 0.0 and 2.0")
	ErrInvalidMaxTokens    = errors.New("max_tokens must be greater than 0")
	ErrWorkerUnavailable   = errors.New("no available inference worker")
	ErrInferenceFailed     = errors.New("inference generation failed")
	ErrStreamFailed        = errors.New("streaming inference generation failed")
	ErrContextCanceled     = errors.New("inference request was canceled")
)
