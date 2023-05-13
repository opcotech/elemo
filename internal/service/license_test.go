package service

import (
	"context"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/mock"
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
				repo: new(mock.LicenseRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithPermissionRepository(new(mock.PermissionRepository)),
				},
			},
			want: &licenseService{
				baseService: &baseService{
					logger:         new(mock.Logger),
					tracer:         new(mock.Tracer),
					permissionRepo: new(mock.PermissionRepository),
				},
				licenseRepo: new(mock.LicenseRepository),
				license:     new(license.License),
			},
		},
		{
			name: "new license service with no license",
			args: args{
				l:    nil,
				repo: new(mock.LicenseRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithPermissionRepository(new(mock.PermissionRepository)),
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
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithPermissionRepository(new(mock.PermissionRepository)),
				},
			},
			wantErr: repository.ErrNoLicenseRepository,
		},
		{
			name: "new license service with no permission repository",
			args: args{
				l:    new(license.License),
				repo: new(mock.LicenseRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoPermissionRepository,
		},
		{
			name: "new license service with invalid options",
			args: args{
				l:    new(license.License),
				repo: new(mock.LicenseRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithPermissionRepository(nil),
				},
			},
			wantErr: ErrNoPermissionRepository,
		},
		{
			name: "new license service with no logger",
			args: args{
				l:    new(license.License),
				repo: new(mock.LicenseRepository),
				opts: []Option{
					WithTracer(new(mock.Tracer)),
					WithPermissionRepository(new(mock.PermissionRepository)),
				},
			},
			want: &licenseService{
				baseService: &baseService{
					logger:         log.DefaultLogger(),
					tracer:         new(mock.Tracer),
					permissionRepo: new(mock.PermissionRepository),
				},
				licenseRepo: new(mock.LicenseRepository),
				license:     new(license.License),
			},
		},
		{
			name: "new license service with no tracer",
			args: args{
				l:    new(license.License),
				repo: new(mock.LicenseRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithPermissionRepository(new(mock.PermissionRepository)),
				},
			},
			want: &licenseService{
				baseService: &baseService{
					logger:         new(mock.Logger),
					tracer:         tracing.NoopTracer(),
					permissionRepo: new(mock.PermissionRepository),
				},
				licenseRepo: new(mock.LicenseRepository),
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
		baseService func(ctx context.Context) *baseService
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/Expired", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/Expired", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
			s := &licenseService{
				baseService: tt.fields.baseService(tt.args.ctx),
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
		baseService func(ctx context.Context) *baseService
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/HasFeature", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
			want: true,
		},
		{
			name: "license has no feature",
			args: args{
				ctx:     context.Background(),
				feature: license.Feature("no-feature"),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/HasFeature", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
			s := &licenseService{
				baseService: tt.fields.baseService(tt.args.ctx),
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
		baseService func(ctx context.Context) *baseService
		licenseRepo func(ctx context.Context) repository.LicenseRepository
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("DocumentCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("NamespaceCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("ActiveOrganizationCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("ProjectCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("RoleCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("ActiveUserCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					return new(mock.LicenseRepository)
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("ActiveUserCount", ctx).Return(1, nil)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/WithinThreshold", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: func(ctx context.Context) repository.LicenseRepository {
					repo := new(mock.LicenseRepository)
					repo.On("ActiveUserCount", ctx).Return(0, assert.AnError)
					return repo
				},
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
			s := &licenseService{
				baseService: tt.fields.baseService(tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo(tt.args.ctx),
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
		Organization: testutil.GenerateRandomString(10),
		Quotas:       license.DefaultQuotas,
		Features:     license.DefaultFeatures,
		ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
	}

	type args struct {
		ctx context.Context
	}
	type fields struct {
		baseService func(ctx context.Context) *baseService
		licenseRepo repository.LicenseRepository
		license     *license.License
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    license.License
		wantErr error
	}{
		{
			name: "get license success",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/GetLicense", []trace.SpanStartOption(nil)).Return(ctx, span)

					permissionRepo := new(mock.PermissionRepository)
					permissionRepo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
						model.SystemRoleSupport,
					}).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: permissionRepo,
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license:     expectedLicense,
			},
			want: *expectedLicense,
		},
		{
			name: "get license no context user",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/GetLicense", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license:     expectedLicense,
			},
			want:    license.License{},
			wantErr: ErrNoUser,
		},
		{
			name: "get license context user no permission",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/GetLicense", []trace.SpanStartOption(nil)).Return(ctx, span)

					permissionRepo := new(mock.PermissionRepository)
					permissionRepo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
						model.SystemRoleSupport,
					}).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: permissionRepo,
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license:     expectedLicense,
			},
			want:    license.License{},
			wantErr: ErrNoPermission,
		},
		{
			name: "get license has system role error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/GetLicense", []trace.SpanStartOption(nil)).Return(ctx, span)

					permissionRepo := new(mock.PermissionRepository)
					permissionRepo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
						model.SystemRoleSupport,
					}).Return(false, assert.AnError)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: permissionRepo,
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license:     expectedLicense,
			},
			want:    license.License{},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &licenseService{
				baseService: tt.fields.baseService(tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo,
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
		baseService func(ctx context.Context) *baseService
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
			name: "license ping success",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/Ping", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.licenseService/Expired", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
					Quotas:       license.DefaultQuotas,
					Features:     license.DefaultFeatures,
					ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
				},
			},
		},
		{
			name: "license ping with expired license",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.licenseService/Ping", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.licenseService/Expired", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						permissionRepo: new(mock.PermissionRepository),
					}
				},
				licenseRepo: new(mock.LicenseRepository),
				license: &license.License{
					ID:           xid.NilID(),
					Email:        testutil.GenerateEmail(10),
					Organization: testutil.GenerateRandomString(10),
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
			s := &licenseService{
				baseService: tt.fields.baseService(tt.args.ctx),
				licenseRepo: tt.fields.licenseRepo,
				license:     tt.fields.license,
			}
			err := s.Ping(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
