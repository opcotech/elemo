package asynq

import (
	"context"
	"reflect"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestWithWorkerConfig(t *testing.T) {
	type args struct {
		config *config.WorkerConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *Worker
		wantErr error
	}{
		{
			name: "create new option with config",
			args: args{
				config: new(config.WorkerConfig),
			},
			want: &Worker{
				conf: new(config.WorkerConfig),
			},
		},
		{
			name: "create new option with nil config",
			args: args{
				config: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			worker := new(Worker)
			err := WithWorkerConfig(tt.args.config)(worker)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, worker)
			}
		})
	}
}

func TestWithWorkerTaskHandler(t *testing.T) {
	handler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return nil
	})

	type args struct {
		taskType TaskType
		handler  asynq.Handler
	}
	tests := []struct {
		name    string
		args    args
		want    *Worker
		wantErr error
	}{
		{
			name: "create new option with handler",
			args: args{
				taskType: TaskTypeSystemHealthCheck,
				handler:  handler,
			},
			want: &Worker{
				handlers: map[TaskType]asynq.Handler{
					TaskTypeSystemHealthCheck: handler,
				},
			},
		},
		{
			name: "create new option with nil handler",
			args: args{
				taskType: TaskTypeSystemHealthCheck,
				handler:  nil,
			},
			wantErr: ErrNoTaskHandler,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			worker := new(Worker)
			worker.handlers = make(map[TaskType]asynq.Handler)

			err := WithWorkerTaskHandler(tt.args.taskType, tt.args.handler)(worker)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				for k, v := range tt.want.handlers {
					assert.Equal(t, reflect.ValueOf(v).Pointer(), reflect.ValueOf(worker.handlers[k]).Pointer())
				}
			}
		})
	}
}

func TestWithWorkerLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			worker := new(Worker)
			err := WithWorkerLogger(tt.args.logger)(worker)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, worker.logger)
		})
	}
}

func TestWithWorkerTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			worker := new(Worker)
			err := WithWorkerTracer(tt.args.tracer)(worker)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, worker.tracer)
		})
	}
}
