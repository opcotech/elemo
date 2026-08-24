package service_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
)

func TestNewStaticFileService(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (service.StaticFileService, error)
		wantErr error
	}{
		{
			name: "new static file service",
			build: func() (service.StaticFileService, error) {
				return service.NewStaticFileService(mockrepo.NewMockStaticFileRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
		},
		{
			name: "new static file service with invalid options",
			build: func() (service.StaticFileService, error) {
				return service.NewStaticFileService(mockrepo.NewMockStaticFileRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(nil), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new static file service with no static file repository",
			build: func() (service.StaticFileService, error) {
				return service.NewStaticFileService(nil, mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: service.ErrNoStaticFileRepository,
		},
		{
			name: "new static file service with no license service",
			build: func() (service.StaticFileService, error) {
				return service.NewStaticFileService(mockrepo.NewMockStaticFileRepository(nil), nil, service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: service.ErrNoLicenseService,
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

func TestStaticFileService_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
		data []byte
	}
	type fields struct {
		baseService    func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService
		staticFileRepo func(ctrl *gomock.Controller, ctx context.Context, path string, data []byte) repository.StaticFileRepository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create static file",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Create(ctx, "/assets/logo.png", data).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("file-content"),
			},
		},
		{
			name: "create static file with cleaned path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Create(ctx, "/bar.txt", data).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "foo/../bar.txt",
				data: []byte("file-content"),
			},
		},
		{
			name: "create static file with empty data",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Create(ctx, "/assets/logo.png", data).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte{},
			},
		},
		{
			name: "create static file with nil data",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Create(ctx, "/assets/logo.png", data).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: nil,
			},
		},
		{
			name: "create static file with empty path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
				data: []byte("file-content"),
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "create static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
				data: []byte("file-content"),
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "create static file with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("file-content"),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create static file with license service error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("file-content"),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create static file with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Create(ctx, "/assets/logo.png", data).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("file-content"),
			},
			wantErr: service.ErrStaticFileCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx)
			service.SetStaticFileServiceRepo(s, tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path, tt.args.data))
			err := s.Create(tt.args.ctx, tt.args.path, tt.args.data)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestStaticFileService_Get(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}
	type fields struct {
		baseService    func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService
		staticFileRepo func(ctrl *gomock.Controller, ctx context.Context, path string) repository.StaticFileRepository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "get static file",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Get(ctx, "/assets/logo.png").Return([]byte("file-content"), nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			want: []byte("file-content"),
		},
		{
			name: "get static file with cleaned path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Get(ctx, "/bar.txt").Return([]byte("file-content"), nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "foo/../bar.txt",
			},
			want: []byte("file-content"),
		},
		{
			name: "get static file with empty path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "get static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "get static file with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Get(ctx, "/assets/logo.png").Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			wantErr: service.ErrStaticFileGet,
		},
		{
			name: "get static file not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Get(ctx, "/missing.txt").Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "missing.txt",
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx)
			service.SetStaticFileServiceRepo(s, tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path))
			got, err := s.Get(tt.args.ctx, tt.args.path)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStaticFileService_Update(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
		data []byte
	}
	type fields struct {
		baseService    func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService
		staticFileRepo func(ctrl *gomock.Controller, ctx context.Context, path string, data []byte) repository.StaticFileRepository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "update static file",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Update(ctx, "/assets/logo.png", data).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("updated-content"),
			},
		},
		{
			name: "update static file with empty data",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Update(ctx, "/assets/logo.png", data).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte{},
			},
		},
		{
			name: "update static file with empty path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
				data: []byte("updated-content"),
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "update static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
				data: []byte("updated-content"),
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "update static file with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("updated-content"),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update static file with license service error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("updated-content"),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update static file with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Update(ctx, "/assets/logo.png", data).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("updated-content"),
			},
			wantErr: service.ErrStaticFileUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx)
			service.SetStaticFileServiceRepo(s, tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path, tt.args.data))
			err := s.Update(tt.args.ctx, tt.args.path, tt.args.data)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestStaticFileService_Delete(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}
	type fields struct {
		baseService    func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService
		staticFileRepo func(ctrl *gomock.Controller, ctx context.Context, path string) repository.StaticFileRepository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete static file",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Delete(ctx, "/assets/logo.png").Return(nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
		},
		{
			name: "delete static file with empty path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "delete static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
			},
			wantErr: service.ErrStaticFileInvalidPath,
		},
		{
			name: "delete static file with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete static file with license service error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return mockrepo.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete static file with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.StaticFileService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.StaticFileService {
						svc, err := service.NewStaticFileService(
							mockrepo.NewMockStaticFileRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := mockrepo.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Delete(ctx, "/assets/logo.png").Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			wantErr: service.ErrStaticFileDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx)
			service.SetStaticFileServiceRepo(s, tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path))
			err := s.Delete(tt.args.ctx, tt.args.path)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
