package async

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestSetRateLimiter(t *testing.T) {
	original := rateLimiter
	defer func() {
		rateLimiter = original
	}()

	rateLimiter = nil
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

	span := new(mock.Span)
	span.On("End", []trace.SpanEndOption(nil)).Return()

	tracer := new(mock.Tracer)
	tracer.On("Start", ctx, "transport.asynq.middleware/WithMetricsExporter", []trace.SpanStartOption(nil)).Return(ctx, span)

	assert.NoError(t,
		WithMetricsExporter(tracer)(asynq.HandlerFunc(func(_ context.Context, _ *asynq.Task) error {
			return nil
		})).ProcessTask(ctx, new(asynq.Task)),
	)

	assert.ErrorIs(t,
		WithMetricsExporter(tracer)(asynq.HandlerFunc(func(_ context.Context, _ *asynq.Task) error {
			return assert.AnError
		})).ProcessTask(ctx, new(asynq.Task)),
		assert.AnError,
	)
}

func TestWithRateLimiter(t *testing.T) {
	type fields struct {
		limiter func(ctrl *gomock.Controller) RateLimiter
	}
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "return handler if rate limiter is allowed",
			fields: fields{
				limiter: func(ctrl *gomock.Controller) RateLimiter {
					limiter := mock.NewRateLimiter(ctrl)
					limiter.EXPECT().Allow().Return(true)
					return limiter
				},
			},
			args: args{
				tracer: func() tracing.Tracer {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", mock.Anything, "transport.asynq.middleware/WithRateLimiter", []trace.SpanStartOption(nil)).Return(context.Background(), span)

					return tracer
				}(),
			},
			wantErr: nil,
		},
		{
			name: "return error if rate limiter is not allowed",
			fields: fields{
				limiter: func(ctrl *gomock.Controller) RateLimiter {
					limiter := mock.NewRateLimiter(ctrl)
					limiter.EXPECT().Allow().Return(false)
					return limiter
				},
			},
			args: args{
				tracer: func() tracing.Tracer {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", mock.Anything, "transport.asynq.middleware/WithRateLimiter", []trace.SpanStartOption(nil)).Return(context.Background(), span)

					return tracer
				}(),
			},
			wantErr: ErrRateLimitExceeded,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := asynq.HandlerFunc(func(_ context.Context, _ *asynq.Task) error {
				return nil
			})

			wrapped := WithRateLimiter(tt.args.tracer, tt.fields.limiter(ctrl))(handler)
			err := wrapped.ProcessTask(context.Background(), new(asynq.Task))

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithErrorLogger(t *testing.T) {
	type fields struct {
		ctx    context.Context
		task   *asynq.Task
		logger func(ctx context.Context, task *asynq.Task) log.Logger
	}
	type args struct {
		tracer func(ctx context.Context) tracing.Tracer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "no error during processing",
			fields: fields{
				ctx:  context.Background(),
				task: asynq.NewTask("test:task", []byte("hello")),
				logger: func(_ context.Context, _ *asynq.Task) log.Logger {
					return new(mock.Logger)
				},
			},
			args: args{
				tracer: func(ctx context.Context) tracing.Tracer {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.middleware/WithErrorLogger", []trace.SpanStartOption(nil)).Return(ctx, span)

					return tracer
				},
			},
		},
		{
			name: "log error if error occurred during processing",
			fields: fields{
				ctx:  context.Background(),
				task: asynq.NewTask("test:task", []byte("hello")),
				logger: func(_ context.Context, task *asynq.Task) log.Logger {
					logger := new(mock.Logger)
					logger.On("Log", zap.ErrorLevel, assert.AnError.Error(), []zap.Field{
						log.WithKey(task.Type()),
						log.WithInput(string(task.Payload())),
						log.WithError(assert.AnError),
					}).Return()

					return logger
				},
			},
			args: args{
				tracer: func(ctx context.Context) tracing.Tracer {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.middleware/WithErrorLogger", []trace.SpanStartOption(nil)).Return(ctx, span)

					return tracer
				},
			},
			wantErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := asynq.HandlerFunc(func(_ context.Context, _ *asynq.Task) error {
				return tt.wantErr
			})

			ctx := tt.fields.ctx
			ctx = context.WithValue(ctx, pkg.CtxKeyLogger, tt.fields.logger(ctx, tt.fields.task))

			wrapped := WithErrorLogger(tt.args.tracer(ctx))(handler)
			err := wrapped.ProcessTask(ctx, tt.fields.task)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
