package asynq

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"

	testMock "github.com/opcotech/elemo/internal/testutil/mock"
)

func TestSetRateLimiter(t *testing.T) {
	t.Parallel()

	require.Nil(t, rateLimiter)

	// Set rate limiter.
	SetRateLimiter(1, 1)
	assert.Equal(t, rateLimiter.Limit(), rate.Limit(1))
	assert.Equal(t, rateLimiter.Burst(), 1)

	// Should not reconfigure rate limiter.
	SetRateLimiter(2, 2)
	assert.Equal(t, rateLimiter.Limit(), rate.Limit(1))
	assert.Equal(t, rateLimiter.Burst(), 1)

	rateLimiter = nil

	// Set rate limiter with 0 limit.
	SetRateLimiter(0, 0)
	assert.Equal(t, rateLimiter.Limit(), rate.Inf)
	assert.Equal(t, rateLimiter.Burst(), 0)
}

func TestWithMetricsExporter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	span := new(testMock.Span)
	span.On("End", []trace.SpanEndOption(nil)).Return()

	tracer := new(testMock.Tracer)
	tracer.On("Start", ctx, "transport.asynq.middleware/WithMetricsExporter", []trace.SpanStartOption(nil)).Return(ctx, span)

	assert.NoError(t,
		WithMetricsExporter(tracer)(asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			return nil
		})).ProcessTask(ctx, new(asynq.Task)),
	)

	assert.ErrorIs(t,
		WithMetricsExporter(tracer)(asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			return assert.AnError
		})).ProcessTask(ctx, new(asynq.Task)),
		assert.AnError,
	)
}

func TestWithRateLimiter(t *testing.T) {
	type args struct {
		tracer  trace.Tracer
		limiter RateLimiter
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "return handler if rate limiter is allowed",
			args: args{
				tracer: func() trace.Tracer {
					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", mock.Anything, "transport.asynq.middleware/WithRateLimiter", []trace.SpanStartOption(nil)).Return(context.Background(), span)

					return tracer
				}(),
				limiter: func() RateLimiter {
					limiter := new(testMock.RateLimiter)
					limiter.On("Allow").Return(true)
					return limiter
				}(),
			},
			wantErr: nil,
		},
		{
			name: "return error if rate limiter is not allowed",
			args: args{
				tracer: func() trace.Tracer {
					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", mock.Anything, "transport.asynq.middleware/WithRateLimiter", []trace.SpanStartOption(nil)).Return(context.Background(), span)

					return tracer
				}(),
				limiter: func() RateLimiter {
					limiter := new(testMock.RateLimiter)
					limiter.On("Allow").Return(false)
					return limiter
				}(),
			},
			wantErr: ErrRateLimitExceeded,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
				return nil
			})

			wrapped := WithRateLimiter(tt.args.tracer, tt.args.limiter)(handler)
			err := wrapped.ProcessTask(context.Background(), new(asynq.Task))

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
