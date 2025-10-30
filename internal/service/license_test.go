package service

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLicenseService(t *testing.T) {
	type args struct {
		l    *license.License
		repo repository.LicenseRepository
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    LicenseService
		wantErr error
	}{
		{
			name: "new license service",
			args: args{
				l:    new(license.License),
				repo: mock.NewLicenseRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			want: &licenseService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					permissionService: mock.NewPermissionService(nil),
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license:     new(license.License),
			},
		},
		{
			name: "new license service with no license",
			args: args{
				l:    nil,
				repo: mock.NewLicenseRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			wantErr: license.ErrNoLicense,
		},
		{
			name: "new license service with no license repository",
			args: args{
				l:    new(license.License),
				repo: nil,
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			wantErr: repository.ErrNoLicenseRepository,
		},
		{
			name: "new license service with no permission service",
			args: args{
				l:    new(license.License),
				repo: mock.NewLicenseRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new license service with invalid options",
			args: args{
				l:    new(license.License),
				repo: mock.NewLicenseRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithPermissionService(nil),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new license service with no logger",
			args: args{
				l:    new(license.License),
				repo: mock.NewLicenseRepository(nil),
				opts: []Option{
					WithTracer(mock.NewMockTracer(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			want: &licenseService{
				baseService: &baseService{
					logger:            log.DefaultLogger(),
					tracer:            mock.NewMockTracer(nil),
					permissionService: mock.NewPermissionService(nil),
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license:     new(license.License),
			},
		},
		{
			name: "new license service with no tracer",
			args: args{
				l:    new(license.License),
				repo: mock.NewLicenseRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			want: &licenseService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            tracing.NoopTracer(),
					permissionService: mock.NewPermissionService(nil),
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license:     new(license.License),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewLicenseService(tt.args.l, tt.args.repo, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLicenseService_Expired(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
		licenseRepo repository.LicenseRepository
		license     *license.License
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    bool
		wantErr error
	}{
		{
			name: "license not expired",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: false,
		},
		{
			name: "license expired",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &licenseService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo,
				license:     tt.fields.license,
			}
			got, err := s.Expired(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLicenseService_HasFeature(t *testing.T) {
	type args struct {
		ctx     context.Context
		feature license.Feature
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
		licenseRepo repository.LicenseRepository
		license     *license.License
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    bool
		wantErr error
	}{
		{
			name: "license has feature",
			args: args{
				ctx:     context.Background(),
				feature: license.DefaultFeatures[0],
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/HasFeature", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "license does not have feature",
			args: args{
				ctx:     context.Background(),
				feature: license.Feature("no-such-feature"),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/HasFeature", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &licenseService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo,
				license:     tt.fields.license,
			}
			got, err := s.HasFeature(tt.args.ctx, tt.args.feature)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLicenseService_WithinThreshold(t *testing.T) {
	type args struct {
		ctx   context.Context
		quota license.Quota
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
		licenseRepo func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository
		license     *license.License
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    bool
		wantErr error
	}{
		{
			name: "document quota within threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaDocuments,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().DocumentCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "namespace quota within threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaNamespaces,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().NamespaceCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "organization quota within threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaOrganizations,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().ActiveOrganizationCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "project quota within threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaProjects,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().ProjectCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "role quota within threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaRoles,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().RoleCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "user quota within threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaUsers,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().ActiveUserCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "invalid quota type",
			args: args{
				ctx:   context.Background(),
				quota: license.Quota("invalid"),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(_ *gomock.Controller, _ context.Context) repository.LicenseRepository {
					return mock.NewLicenseRepository(nil)
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want:    false,
			wantErr: ErrQuotaInvalid,
		},
		{
			name: "quota exceeds threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaUsers,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().ActiveUserCount(ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas: map[license.Quota]uint32{
						license.QuotaUsers: 0,
					},
					Features:  license.DefaultFeatures,
					ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: false,
		},
		{
			name: "get quota count error",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaUsers,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					repo.EXPECT().ActiveUserCount(ctx).Return(0, assert.AnError)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas: map[license.Quota]uint32{
						license.QuotaUsers: 0,
					},
					Features:  license.DefaultFeatures,
					ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want:    false,
			wantErr: ErrQuotaUsageGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &licenseService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo(ctrl, tt.args.ctx),
				license:     tt.fields.license,
			}
			got, err := s.WithinThreshold(tt.args.ctx, tt.args.quota)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLicenseService_GetLicense(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	expectedLicense := &license.License{
		ID:           xid.NilID(),
		Email:        testutil.GenerateEmail(10),
		Organization: pkg.GenerateRandomString(10),
		Quotas:       license.DefaultQuotas,
		Features:     license.DefaultFeatures,
		ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
	}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
		licenseRepo func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository
		license     *license.License
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    license.License
		wantErr error
	}{
		{
			name: "get license",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/GetLicense", gomock.Len(0)).Return(ctx, span)

					permissionSvc := mock.NewPermissionService(ctrl)
					permissionSvc.EXPECT().CtxUserHasSystemRole(ctx, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
						model.SystemRoleSupport,
					}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permissionSvc,
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, _ context.Context) repository.LicenseRepository {
					repo := mock.NewLicenseRepository(ctrl)
					return repo
				},
				license: expectedLicense,
			},
			want: *expectedLicense,
		},
		{
			name: "get license context user no permission",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/GetLicense", gomock.Len(0)).Return(ctx, span)

					permissionSvc := mock.NewPermissionService(ctrl)
					permissionSvc.EXPECT().CtxUserHasSystemRole(ctx, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
						model.SystemRoleSupport,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permissionSvc,
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, _ context.Context) repository.LicenseRepository {
					return mock.NewLicenseRepository(ctrl)
				},
				license: nil,
			},
			wantErr: ErrNoPermission,
			want:    license.License{},
		},
		{
			name: "get license has system role error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/GetLicense", gomock.Len(0)).Return(ctx, span)

					permissionSvc := mock.NewPermissionService(ctrl)
					permissionSvc.EXPECT().CtxUserHasSystemRole(ctx, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
						model.SystemRoleSupport,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permissionSvc,
					}
				},
				licenseRepo: func(ctrl *gomock.Controller, _ context.Context) repository.LicenseRepository {
					return mock.NewLicenseRepository(ctrl)
				},
				license: nil,
			},
			wantErr: ErrLicenseGet,
			want:    license.License{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &licenseService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo(ctrl, tt.args.ctx),
				license:     tt.fields.license,
			}
			got, err := s.GetLicense(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLicenseService_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
		licenseRepo repository.LicenseRepository
		license     *license.License
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr error
	}{
		{
			name: "ping license valid",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Ping", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
		},
		{
			name: "ping license invalid",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Ping", gomock.Len(0)).Return(ctx, span).Times(1)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span).Times(1)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mock.NewPermissionService(nil),
					}
				},
				licenseRepo: mock.NewLicenseRepository(nil),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: pkg.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour),
				},
			},
			wantErr: license.ErrLicenseInvalid,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &licenseService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo,
				license:     tt.fields.license,
			}
			err := s.Ping(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
