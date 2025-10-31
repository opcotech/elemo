package async

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End().Return().Times(2)

	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "transport.asynq.middleware/WithMetricsExporter").Return(ctx, span).Times(2)

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
		tracer func(ctrl *gomock.Controller) tracing.Tracer
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
				tracer: func(ctrl *gomock.Controller) tracing.Tracer {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "transport.asynq.middleware/WithRateLimiter").Return(context.Background(), span)

					return tracer
				},
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
				tracer: func(ctrl *gomock.Controller) tracing.Tracer {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "transport.asynq.middleware/WithRateLimiter").Return(context.Background(), span)

					return tracer
				},
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

			wrapped := WithRateLimiter(tt.args.tracer(ctrl), tt.fields.limiter(ctrl))(handler)
			err := wrapped.ProcessTask(context.Background(), new(asynq.Task))

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithErrorLogger(t *testing.T) {
	type fields struct {
		ctx    context.Context
		task   *asynq.Task
		logger func(ctx context.Context, task *asynq.Task, ctrl *gomock.Controller) log.Logger
	}
	type args struct {
		tracer func(ctx context.Context, ctrl *gomock.Controller) tracing.Tracer
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
				logger: func(_ context.Context, _ *asynq.Task, _ *gomock.Controller) log.Logger {
					return mock.NewMockLogger(nil)
				},
			},
			args: args{
				tracer: func(ctx context.Context, ctrl *gomock.Controller) tracing.Tracer {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "transport.asynq.middleware/WithErrorLogger").Return(ctx, span)

					return tracer
				},
			},
		},
		{
			name: "log error if error occurred during processing",
			fields: fields{
				ctx:  context.Background(),
				task: asynq.NewTask("test:task", []byte("hello")),
				logger: func(_ context.Context, _ *asynq.Task, ctrl *gomock.Controller) log.Logger {
					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Log(gomock.Any(), log.LevelError, assert.AnError.Error(), gomock.Any()).Return()
					return logger
				},
			},
			args: args{
				tracer: func(ctx context.Context, ctrl *gomock.Controller) tracing.Tracer {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "transport.asynq.middleware/WithErrorLogger").Return(ctx, span)

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := asynq.HandlerFunc(func(_ context.Context, _ *asynq.Task) error {
				return tt.wantErr
			})

			ctx := tt.fields.ctx
			ctx = context.WithValue(ctx, pkg.CtxKeyLogger, tt.fields.logger(ctx, tt.fields.task, ctrl))

			wrapped := WithErrorLogger(tt.args.tracer(ctx, ctrl))(handler)
			err := wrapped.ProcessTask(ctx, tt.fields.task)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
