package asynq

import "errors"

var (
	ErrNoEmailService       = errors.New("no email service set")             // no email service set
	ErrNoRateLimiter        = errors.New("no rate limiter set")              // no rate limiter set
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")              // rate limit exceeded
	ErrReceiveTask          = errors.New("failed to receive task")           // failed to receive task
	ErrSendTask             = errors.New("failed to send task")              // failed to send task
	ErrTaskPayloadUnmarshal = errors.New("failed to unmarshal task payload") // failed to unmarshal task payload
)
