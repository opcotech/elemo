package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
	msvc "github.com/opcotech/elemo/internal/testutil/mock/service"
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

func TestWithSystemService(t *testing.T) {
	type args struct {
		s SystemService
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    SystemService
	}{
		{
			name: "WithSystemService sets the system service for the baseService",
			args: args{
				s: new(msvc.MockSystemService),
			},
			want: new(msvc.MockSystemService),
		},
		{
			name: "WithSystemService returns an error if no system service is provided",
			args: args{
				s: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithSystemService(tt.args.s)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.systemService)
			}
		})
	}
}
