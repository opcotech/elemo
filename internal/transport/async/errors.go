package async

import "errors"

var (
	ErrNoEmailService       = errors.New("no email service set")             // no email service set
	ErrNoGraphDatabase      = errors.New("no graph database set")            // no graph database set
	ErrNoQueueClient        = errors.New("no queue client set")              // no queue client set
	ErrNoRateLimiter        = errors.New("no rate limiter set")              // no rate limiter set
	ErrNoSearchService      = errors.New("no search service set")            // no search service set
	ErrNoTaskHandler        = errors.New("no task handler set")              // no task handler set
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")              // rate limit exceeded
	ErrTaskPayloadUnmarshal = errors.New("failed to unmarshal task payload") // failed to unmarshal task payload
)
