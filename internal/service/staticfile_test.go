package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewStaticFileService(t *testing.T) {
	type args struct {
		repo repository.StaticFileRepository
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    StaticFileService
		wantErr error
	}{
		{
			name: "new static file service",
			args: args{
				repo: repository.NewMockStaticFileRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &staticFileService{
				baseService: &baseService{
					logger:         mock.NewMockLogger(nil),
					tracer:         mock.NewMockTracer(nil),
					licenseService: mock.NewMockLicenseService(nil),
				},
				staticFileRepo: repository.NewMockStaticFileRepository(nil),
			},
		},
		{
			name: "new static file service with invalid options",
			args: args{
				repo: repository.NewMockStaticFileRepository(nil),
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new static file service with no static file repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoStaticFileRepository,
		},
		{
			name: "new static file service with no license service",
			args: args{
				repo: repository.NewMockStaticFileRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewStaticFileService(tt.args.repo, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
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
		baseService    func(ctrl *gomock.Controller, ctx context.Context) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
				data: []byte("file-content"),
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "create static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
				data: []byte("file-content"),
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "create static file with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Create(ctx, "/assets/logo.png", data).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("file-content"),
			},
			wantErr: ErrStaticFileCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &staticFileService{
				baseService:    tt.fields.baseService(ctrl, tt.args.ctx),
				staticFileRepo: tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path, tt.args.data),
			}
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
		baseService    func(ctrl *gomock.Controller, ctx context.Context) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "get static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "get static file with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Get(ctx, "/assets/logo.png").Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			wantErr: ErrStaticFileGet,
		},
		{
			name: "get static file not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
			s := &staticFileService{
				baseService:    tt.fields.baseService(ctrl, tt.args.ctx),
				staticFileRepo: tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path),
			}
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
		baseService    func(ctrl *gomock.Controller, ctx context.Context) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
				data: []byte("updated-content"),
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "update static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
				data: []byte("updated-content"),
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "update static file with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ []byte) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string, data []byte) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Update(ctx, "/assets/logo.png", data).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
				data: []byte("updated-content"),
			},
			wantErr: ErrStaticFileUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &staticFileService{
				baseService:    tt.fields.baseService(ctrl, tt.args.ctx),
				staticFileRepo: tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path, tt.args.data),
			}
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
		baseService    func(ctrl *gomock.Controller, ctx context.Context) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "delete static file with absolute path",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/etc/passwd",
			},
			wantErr: ErrStaticFileInvalidPath,
		},
		{
			name: "delete static file with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, _ context.Context, _ string) repository.StaticFileRepository {
					return repository.NewMockStaticFileRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.staticFileService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
				staticFileRepo: func(ctrl *gomock.Controller, ctx context.Context, _ string) repository.StaticFileRepository {
					repo := repository.NewMockStaticFileRepository(ctrl)
					repo.EXPECT().Delete(ctx, "/assets/logo.png").Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "assets/logo.png",
			},
			wantErr: ErrStaticFileDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &staticFileService{
				baseService:    tt.fields.baseService(ctrl, tt.args.ctx),
				staticFileRepo: tt.fields.staticFileRepo(ctrl, tt.args.ctx, tt.args.path),
			}
			err := s.Delete(tt.args.ctx, tt.args.path)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
