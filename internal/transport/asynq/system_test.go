package asynq

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewSystemHealthCheckTask(t *testing.T) {
	tests := []struct {
		name    string
		want    *asynq.Task
		wantErr error
	}{
		{
			name: "create new task",
			want: asynq.NewTask(TaskTypeSystemHealthCheck.String(),
				[]byte(`{"message":"healthy"}`),
				asynq.Timeout(5*time.Second),
				asynq.Retention(5*time.Second)),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSystemHealthCheckTask()
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSystemHealthCheckTaskHandler(t *testing.T) {
	type args struct {
		opts []TaskOption
	}
	tests := []struct {
		name    string
		args    args
		want    *SystemHealthCheckTaskHandler
		wantErr error
	}{
		{
			name: "create new task handler",
			args: args{
				opts: []TaskOption{
					WithTaskLogger(new(mock.Logger)),
					WithTaskTracer(new(mock.Tracer)),
				},
			},
			want: &SystemHealthCheckTaskHandler{
				baseTaskHandler: &baseTaskHandler{
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
			},
		},
		{
			name: "create new task handler with invalid option",
			args: args{
				opts: []TaskOption{
					WithTaskLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSystemHealthCheckTaskHandler(tt.args.opts...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSystemHealthCheckTaskHandler_ProcessTask(t *testing.T) {
	type fields struct {
		baseTaskHandler func(ctx context.Context, task *asynq.Task) *baseTaskHandler
	}
	type args struct {
		ctx  context.Context
		task *asynq.Task
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "process task",
			fields: fields{
				baseTaskHandler: func(ctx context.Context, task *asynq.Task) *baseTaskHandler {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.SystemHealthCheckTaskHandler/ProcessTask", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseTaskHandler{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					task, _ := NewSystemHealthCheckTask()
					return task
				}(),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &SystemHealthCheckTaskHandler{
				baseTaskHandler: tt.fields.baseTaskHandler(tt.args.ctx, tt.args.task),
			}
			err := h.ProcessTask(tt.args.ctx, tt.args.task)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
