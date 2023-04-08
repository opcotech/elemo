package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestWithLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    log.Logger
	}{
		{
			name: "WithLogger sets the logger for the baseService",
			args: args{
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
		},
		{
			name: "WithLogger returns an error if no logger is provided",
			args: args{
				logger: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithLogger(tt.args.logger)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.logger)
			}
		})
	}
}

func TestWithTracer(t *testing.T) {
	type args struct {
		tracer trace.Tracer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    trace.Tracer
	}{
		{
			name: "WithTracer sets the tracer for the baseService",
			args: args{
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
		},
		{
			name: "WithTracer returns an error if no tracer is provided",
			args: args{
				tracer: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithTracer(tt.args.tracer)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.tracer)
			}
		})
	}
}

func Test_newService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *baseService
		wantErr error
	}{
		{
			name: "newService returns a baseService with the provided options",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &baseService{
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "newService returns default logger if no logger is provided",
			args: args{
				opts: []Option{
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &baseService{
				logger: log.DefaultLogger(),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "newService returns default tracer if no tracer is provided",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
				},
			},
			want: &baseService{
				logger: new(mock.Logger),
				tracer: tracing.NoopTracer(),
			},
		},
		{
			name: "newService returns error if nil logger is provided",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "newService returns error if nil tracer is provided",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(nil),
				},
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newService(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
