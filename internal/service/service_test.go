package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
)

func TestWithLogger(t *testing.T) {
	tests := []struct {
		name    string
		logger  log.Logger
		wantErr bool
	}{
		{
			name:   "WithLogger sets the logger for the runtime",
			logger: mocklog.NewMockLogger(nil),
		},
		{
			name:    "WithLogger returns an error if no logger is provided",
			logger:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r runtime

			err := WithLogger(tt.logger)(&r)
			require.Equal(t, tt.wantErr, err != nil)

			if !tt.wantErr {
				assert.Equal(t, tt.logger, r.logger)
			}
		})
	}
}

func TestWithTracer(t *testing.T) {
	tests := []struct {
		name    string
		tracer  tracing.Tracer
		wantErr bool
	}{
		{
			name:   "WithTracer sets the tracer for the runtime",
			tracer: mocktrace.NewMockTracer(nil),
		},
		{
			name:    "WithTracer returns an error if no tracer is provided",
			tracer:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r runtime

			err := WithTracer(tt.tracer)(&r)
			require.Equal(t, tt.wantErr, err != nil)

			if !tt.wantErr {
				assert.Equal(t, tt.tracer, r.tracer)
			}
		})
	}
}

func Test_newRuntime(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		want    runtime
		wantErr error
	}{
		{
			name: "newRuntime returns a runtime with the provided options",
			opts: []Option{
				WithLogger(mocklog.NewMockLogger(nil)),
				WithTracer(mocktrace.NewMockTracer(nil)),
			},
			want: runtime{
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "newRuntime returns default logger if no logger is provided",
			opts: []Option{
				WithTracer(mocktrace.NewMockTracer(nil)),
			},
			want: runtime{
				logger: log.DefaultLogger(),
				tracer: mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "newRuntime returns default tracer if no tracer is provided",
			opts: []Option{
				WithLogger(mocklog.NewMockLogger(nil)),
			},
			want: runtime{
				logger: mocklog.NewMockLogger(nil),
				tracer: tracing.NoopTracer(),
			},
		},
		{
			name: "newRuntime returns error if nil logger is provided",
			opts: []Option{
				WithLogger(nil),
				WithTracer(mocktrace.NewMockTracer(nil)),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "newRuntime returns error if nil tracer is provided",
			opts: []Option{
				WithLogger(mocklog.NewMockLogger(nil)),
				WithTracer(nil),
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newRuntime(tt.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
