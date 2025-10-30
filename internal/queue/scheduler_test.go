package queue

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestWithSchedulerConfig(t *testing.T) {
	type args struct {
		config *config.WorkerConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *Scheduler
		wantErr error
	}{
		{
			name: "create new option with config",
			args: args{
				config: new(config.WorkerConfig),
			},
			want: &Scheduler{
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
			scheduler := new(Scheduler)
			err := WithSchedulerConfig(tt.args.config)(scheduler)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, scheduler)
			}
		})
	}
}

func TestWithSchedulerLogger(t *testing.T) {
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
				logger: mock.NewMockLogger(nil),
			},
			want: mock.NewMockLogger(nil),
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
			scheduler := new(Scheduler)
			err := WithSchedulerLogger(tt.args.logger)(scheduler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, scheduler.logger)
		})
	}
}

func TestWithSchedulerTracer(t *testing.T) {
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
				tracer: mock.NewMockTracer(nil),
			},
			want: mock.NewMockTracer(nil),
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
			scheduler := new(Scheduler)
			err := WithSchedulerTracer(tt.args.tracer)(scheduler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, scheduler.tracer)
		})
	}
}
