package asynq

import "errors"

var (
	ErrNoRateLimiter        = errors.New("no rate limiter provided")         // no rate limiter provided
	ErrTaskPayloadUnmarshal = errors.New("failed to unmarshal task payload") // failed to unmarshal task payload
	ErrSendTask             = errors.New("failed to send task")              // failed to send task
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")              // rate limit exceeded
	ErrReceiveTask          = errors.New("failed to receive task")           // failed to receive task
)
