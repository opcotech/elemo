package service_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
)

func TestNewSystemService(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (service.SystemService, error)
		wantErr error
	}{
		{
			name: "new system service",
			build: func() (service.SystemService, error) {
				return service.NewSystemService(map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(nil)}, &model.VersionInfo{Version: "1.0.0"}, service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
		},
		{
			name: "new system service with nil resources",
			build: func() (service.SystemService, error) {
				return service.NewSystemService(nil, &model.VersionInfo{Version: "1.0.0"}, service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: service.ErrNoResources,
		},
		{
			name: "new system service with nil version",
			build: func() (service.SystemService, error) {
				return service.NewSystemService(map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(nil)}, nil, service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: service.ErrNoVersionInfo,
		},
		{
			name: "new system service with invalid options",
			build: func() (service.SystemService, error) {
				return service.NewSystemService(map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(nil)}, &model.VersionInfo{Version: "1.0.0"}, service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.build()
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

func Test_systemService_GetHeartbeat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.systemService/GetHeartbeat", gomock.Len(0)).Return(ctx, span)

	s := func() service.SystemService {
		svc, err := service.NewSystemService(
			map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(ctrl)},
			&model.VersionInfo{Version: "test"},
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()

	assert.NoError(t, s.GetHeartbeat(ctx))
}

func Test_systemService_GetVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.systemService/GetVersion", gomock.Len(0)).Return(ctx, span)

	s := func() service.SystemService {
		svc, err := service.NewSystemService(
			map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(ctrl)},
			&model.VersionInfo{
				Version:   "version",
				Commit:    "commit",
				Date:      "date",
				GoVersion: "go version",
			},
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()

	got := s.GetVersion(ctx)
	assert.Equal(t, service.SystemServiceVersion(s), got)
}

func Test_systemService_GetHealth(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) service.SystemService
		versionInfo *model.VersionInfo
		resources   func(ctx context.Context, ctrl *gomock.Controller) map[model.HealthCheckComponent]service.Pingable
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.SystemService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.systemService/GetHealth", gomock.Len(0)).Return(ctx, span)

					return func() service.SystemService {
						svc, err := service.NewSystemService(
							map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(ctrl)},
							&model.VersionInfo{Version: "test"},
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: func(ctx context.Context, ctrl *gomock.Controller) map[model.HealthCheckComponent]service.Pingable {
					resource := mocksvc.NewMockPingable(ctrl)
					resource.EXPECT().Ping(ctx).Return(nil).Times(4)

					return map[model.HealthCheckComponent]service.Pingable{
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.SystemService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.systemService/GetHealth", gomock.Len(0)).Return(ctx, span)

					return func() service.SystemService {
						svc, err := service.NewSystemService(
							map[model.HealthCheckComponent]service.Pingable{model.HealthCheckComponentGraphDB: mocksvc.NewMockPingable(ctrl)},
							&model.VersionInfo{Version: "test"},
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				versionInfo: &model.VersionInfo{
					Version: "1.0.0",
				},
				resources: func(ctx context.Context, ctrl *gomock.Controller) map[model.HealthCheckComponent]service.Pingable {
					resource := mocksvc.NewMockPingable(ctrl)
					resource.EXPECT().Ping(ctx).Return(assert.AnError).Times(4)

					return map[model.HealthCheckComponent]service.Pingable{
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
			wantErr: service.ErrSystemHealthCheck,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx)
			service.SetSystemServiceState(s, tt.fields.versionInfo, tt.fields.resources(tt.args.ctx, ctrl))
			got, err := s.GetHealth(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
