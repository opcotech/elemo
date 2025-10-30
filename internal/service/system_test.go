package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					model.HealthCheckComponentGraphDB: mock.NewPingableResource(nil),
				},
				version: &model.VersionInfo{
					Version: "1.0.0",
				},
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			want: &systemService{
				baseService: &baseService{
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: mock.NewPingableResource(nil),
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
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: ErrNoResources,
		},
		{
			name: "new system service with nil version",
			args: args{
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: mock.NewPingableResource(nil),
				},
				version: nil,
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: ErrNoVersionInfo,
		},
		{
			name: "new system service with invalid options",
			args: args{
				resources: map[model.HealthCheckComponent]Pingable{
					model.HealthCheckComponentGraphDB: mock.NewPingableResource(nil),
				},
				version: &model.VersionInfo{},
				opts: []Option{
					WithLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.systemService/GetHeartbeat", gomock.Len(0)).Return(ctx, span)

	s := &systemService{
		baseService: &baseService{
			tracer: tracer,
		},
	}

	assert.NoError(t, s.GetHeartbeat(ctx))
}

func Test_systemService_GetVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.systemService/GetVersion", gomock.Len(0)).Return(ctx, span)

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
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
		versionInfo *model.VersionInfo
		resources   func(ctx context.Context, ctrl *gomock.Controller) map[model.HealthCheckComponent]Pingable
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.systemService/GetHealth", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						tracer: tracer,
					}
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: func(ctx context.Context, ctrl *gomock.Controller) map[model.HealthCheckComponent]Pingable {
					resource := mock.NewPingableResource(ctrl)
					resource.EXPECT().Ping(ctx).Return(nil).Times(4)

					return map[model.HealthCheckComponent]Pingable{
						model.HealthCheckComponentGraphDB:      resource,
						model.HealthCheckComponentRelationalDB: resource,
						model.HealthCheckComponentLicense:      resource,
						model.HealthCheckComponentMessageQueue: resource,
					}
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: map[model.HealthCheckComponent]model.HealthStatus{
				model.HealthCheckComponentGraphDB:      model.HealthStatusHealthy,
				model.HealthCheckComponentRelationalDB: model.HealthStatusHealthy,
				model.HealthCheckComponentLicense:      model.HealthStatusHealthy,
				model.HealthCheckComponentMessageQueue: model.HealthStatusHealthy,
			},
		},
		{
			name: "get health with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.systemService/GetHealth", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						tracer: tracer,
					}
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: func(ctx context.Context, ctrl *gomock.Controller) map[model.HealthCheckComponent]Pingable {
					resource := mock.NewPingableResource(ctrl)
					resource.EXPECT().Ping(ctx).Return(assert.AnError).Times(4)

					return map[model.HealthCheckComponent]Pingable{
						model.HealthCheckComponentGraphDB:      resource,
						model.HealthCheckComponentRelationalDB: resource,
						model.HealthCheckComponentLicense:      resource,
						model.HealthCheckComponentMessageQueue: resource,
					}
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: map[model.HealthCheckComponent]model.HealthStatus{
				model.HealthCheckComponentGraphDB:      model.HealthStatusUnhealthy,
				model.HealthCheckComponentRelationalDB: model.HealthStatusUnhealthy,
				model.HealthCheckComponentLicense:      model.HealthStatusUnhealthy,
				model.HealthCheckComponentMessageQueue: model.HealthStatusUnhealthy,
			},
			wantErr: ErrSystemHealthCheck,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &systemService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				versionInfo: tt.fields.versionInfo,
				resources:   tt.fields.resources(tt.args.ctx, ctrl),
			}
			got, err := s.GetHealth(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
