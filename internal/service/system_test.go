package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewSystemService(t *testing.T) {
	type args struct {
		resources map[model.HealthCheckComponent]Pingable
		version   *model.VersionInfo
		opts      []Option
	}
	tests := []struct {
		name    string
		args    args
		want    SystemService
		wantErr error
	}{
		{
			name: "new system service",
			args: args{
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: new(mock.PingableResource),
				},
				version: &model.VersionInfo{
					Version: "1.0.0",
				},
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &systemService{
				baseService: &baseService{
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: new(mock.PingableResource),
				},
			},
		},
		{
			name: "new system service with nil resources",
			args: args{
				resources: nil,
				version: &model.VersionInfo{
					Version: "1.0.0",
				},
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoResources,
		},
		{
			name: "new system service with nil version",
			args: args{
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: new(mock.PingableResource),
				},
				version: nil,
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoVersionInfo,
		},
		{
			name: "new system service with invalid options",
			args: args{
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: new(mock.PingableResource),
				},
				version: &model.VersionInfo{},
				opts: []Option{
					WithLogger(nil),
				},
			},
			wantErr: ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewSystemService(tt.args.resources, tt.args.version, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_systemService_GetHeartbeat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	span := new(mock.Span)
	span.On("End", []trace.SpanEndOption(nil)).Return()

	tracer := new(mock.Tracer)
	tracer.On("Start", ctx, "service.systemService/GetHeartbeat", []trace.SpanStartOption(nil)).Return(ctx, span)

	s := &systemService{
		baseService: &baseService{
			tracer: tracer,
		},
	}

	assert.NoError(t, s.GetHeartbeat(ctx))
}

func Test_systemService_GetVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	span := new(mock.Span)
	span.On("End", []trace.SpanEndOption(nil)).Return()

	tracer := new(mock.Tracer)
	tracer.On("Start", ctx, "service.systemService/GetVersion", []trace.SpanStartOption(nil)).Return(ctx, span)

	s := &systemService{
		baseService: &baseService{
			tracer: tracer,
		},
		versionInfo: &model.VersionInfo{
			Version:   "version",
			Commit:    "commit",
			Date:      "date",
			GoVersion: "go version",
		},
	}

	got := s.GetVersion(ctx)
	assert.Equal(t, s.versionInfo, got)
}

func Test_systemService_GetHealth(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context) *baseService
		versionInfo *model.VersionInfo
		resources   func(ctx context.Context) map[model.HealthCheckComponent]Pingable
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[model.HealthCheckComponent]model.HealthStatus
		wantErr error
	}{
		{
			name: "get health",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()
					span.On("AddEvent", fmt.Sprintf("Check %s health", model.HealthCheckComponentGraphDB)).Return()
					span.On("AddEvent", fmt.Sprintf("Check %s health", model.HealthCheckComponentRelationalDB)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.systemService/GetHealth", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("SpanFromContext", ctx).Return(span).Twice()

					return &baseService{
						tracer: tracer,
					}
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: func(ctx context.Context) map[model.HealthCheckComponent]Pingable {
					resource := new(mock.PingableResource)
					resource.On("Ping", ctx).Return(nil).Twice()

					return map[model.HealthCheckComponent]Pingable{
						model.HealthCheckComponentGraphDB:      resource,
						model.HealthCheckComponentRelationalDB: resource,
					}
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: map[model.HealthCheckComponent]model.HealthStatus{
				model.HealthCheckComponentGraphDB:      model.HealthStatusHealthy,
				model.HealthCheckComponentRelationalDB: model.HealthStatusHealthy,
			},
		},
		{
			name: "get health with error",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()
					span.On("AddEvent", fmt.Sprintf("Check %s health", model.HealthCheckComponentGraphDB)).Return()
					span.On("AddEvent", fmt.Sprintf("Check %s health", model.HealthCheckComponentRelationalDB)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.systemService/GetHealth", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("SpanFromContext", ctx).Return(span).Twice()

					return &baseService{
						tracer: tracer,
					}
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: func(ctx context.Context) map[model.HealthCheckComponent]Pingable {
					resource := new(mock.PingableResource)
					resource.On("Ping", ctx).Return(errors.New("error")).Twice()

					return map[model.HealthCheckComponent]Pingable{
						model.HealthCheckComponentGraphDB:      resource,
						model.HealthCheckComponentRelationalDB: resource,
					}
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: map[model.HealthCheckComponent]model.HealthStatus{
				model.HealthCheckComponentGraphDB:      model.HealthStatusUnhealthy,
				model.HealthCheckComponentRelationalDB: model.HealthStatusUnhealthy,
			},
			wantErr: ErrSystemHealthCheck,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &systemService{
				baseService: tt.fields.baseService(tt.args.ctx),
				versionInfo: tt.fields.versionInfo,
				resources:   tt.fields.resources(tt.args.ctx),
			}
			got, err := s.GetHealth(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
