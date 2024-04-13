package asynq

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

var (
	rateLimiter RateLimiter

	processedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_processed_tasks_total",
			Help: "The total number of processed tasks.",
		},
		[]string{"task_type"},
	)

	failedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_failed_tasks_total",
			Help: "The total number of times processing failed.",
		},
		[]string{"task_type"},
	)

	inProgressGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "worker_in_progress_tasks",
			Help: "The number of tasks currently being processed.",
		},
		[]string{"task_type"},
	)
)

// RateLimiter is an interface for rate limiter.
type RateLimiter interface {
	Limit() rate.Limit
	Burst() int
	TokensAt(t time.Time) float64
	Tokens() float64
	Allow() bool
	AllowN(t time.Time, n int) bool
}

// SetRateLimiter configures the rate limiter if not yet configured.
func SetRateLimiter(limit float64, burst int) {
	if rateLimiter != nil {
		return
	}

	rateLimit := rate.Limit(limit)

	if rateLimit == 0 {
		rateLimit = rate.Inf
	}

	rateLimiter = rate.NewLimiter(rateLimit, burst)
}

// WithMetricsExporter middleware exports prometheus metrics for asynq.
func WithMetricsExporter(tracer tracing.Tracer) func(next asynq.Handler) asynq.Handler {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			ctx, span := tracer.Start(ctx, "transport.asynq.middleware/WithMetricsExporter")
			defer span.End()

			defer processedCounter.WithLabelValues(t.Type()).Inc()
			defer inProgressGauge.WithLabelValues(t.Type()).Dec()

			inProgressGauge.WithLabelValues(t.Type()).Inc()
			if err := next.ProcessTask(ctx, t); err != nil {
				failedCounter.WithLabelValues(t.Type()).Inc()
				return err
			}

			return nil
		})
	}
}

// WithRateLimiter middleware limits the number of tasks processed per second.
func WithRateLimiter(tracer tracing.Tracer, r RateLimiter) func(next asynq.Handler) asynq.Handler {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			ctx, span := tracer.Start(ctx, "transport.asynq.middleware/WithRateLimiter")
			defer span.End()

			if !r.Allow() {
				return ErrRateLimitExceeded
			}
			return next.ProcessTask(ctx, t)
		})
	}
}

// WithErrorLogger logs task processing errors.
func WithErrorLogger(tracer tracing.Tracer) func(next asynq.Handler) asynq.Handler {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			ctx, span := tracer.Start(ctx, "transport.asynq.middleware/WithErrorLogger")
			defer span.End()

			err := next.ProcessTask(ctx, t)
			if err != nil {
				log.Error(ctx, err, log.WithKey(t.Type()), log.WithInput(string(t.Payload())))
				return err
			}

			return nil
		})
	}
}
