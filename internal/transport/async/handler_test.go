package async

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

func TestWithTaskLogger(t *testing.T) {
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
				logger: mocklog.NewMockLogger(nil),
			},
			want: mocklog.NewMockLogger(nil),
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
			handler := new(baseTaskHandler)
			err := WithTaskLogger(tt.args.logger)(handler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, handler.logger)
		})
	}
}

func TestWithTaskTracer(t *testing.T) {
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
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: mocktrace.NewMockTracer(nil),
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
			handler := new(baseTaskHandler)
			err := WithTaskTracer(tt.args.tracer)(handler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, handler.tracer)
		})
	}
}

func TestWithTaskCustomFieldService(t *testing.T) {
	type args struct {
		customFieldService service.CustomFieldService
	}
	tests := []struct {
		name    string
		args    args
		want    service.CustomFieldService
		wantErr error
	}{
		{
			name: "create new option with custom field service",
			args: args{
				customFieldService: mocksvc.NewMockCustomFieldService(nil),
			},
			want: mocksvc.NewMockCustomFieldService(nil),
		},
		{
			name: "create new option with nil custom field service",
			args: args{
				customFieldService: nil,
			},
			wantErr: ErrNoCustomFieldService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := new(baseTaskHandler)
			err := WithTaskCustomFieldService(tt.args.customFieldService)(handler)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, handler.customFieldService)
		})
	}
}
