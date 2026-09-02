package service_test

import (
	"context"
	"testing"
	"time"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"go.uber.org/mock/gomock"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

type licenseServiceDeps struct {
	logger            log.Logger
	tracer            tracing.Tracer
	permissionService service.PermissionService
}

func newLicenseServiceForTest(deps licenseServiceDeps, repo repository.LicenseRepository, lic *license.License) service.LicenseService {
	if repo == nil {
		repo = mockrepo.NewMockLicenseRepository(nil)
	}
	if deps.permissionService == nil {
		deps.permissionService = mocksvc.NewMockPermissionService(nil)
	}
	if lic == nil {
		lic = new(license.License)
	}
	var opts []service.Option
	if deps.logger != nil {
		opts = append(opts, service.WithLogger(deps.logger))
	}
	if deps.tracer != nil {
		opts = append(opts, service.WithTracer(deps.tracer))
	}
	svc, err := service.NewLicenseService(lic, repo, deps.permissionService, opts...)
	if err != nil {
		panic(err)
	}
	return svc
}

func TestNewLicenseService(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (service.LicenseService, error)
		wantErr error
	}{
		{
			name: "new license service",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(new(license.License), mockrepo.NewMockLicenseRepository(nil), mocksvc.NewMockPermissionService(nil), service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
		},
		{
			name: "new license service with no license",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(nil, mockrepo.NewMockLicenseRepository(nil), mocksvc.NewMockPermissionService(nil), service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: license.ErrNoLicense,
		},
		{
			name: "new license service with no license repository",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(new(license.License), nil, mocksvc.NewMockPermissionService(nil), service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: repository.ErrNoLicenseRepository,
		},
		{
			name: "new license service with no permission service",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(new(license.License), mockrepo.NewMockLicenseRepository(nil), nil, service.WithLogger(mocklog.NewMockLogger(nil)), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new license service with invalid options",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(new(license.License), mockrepo.NewMockLicenseRepository(nil), mocksvc.NewMockPermissionService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new license service with no logger",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(new(license.License), mockrepo.NewMockLicenseRepository(nil), mocksvc.NewMockPermissionService(nil), service.WithTracer(mocktrace.NewMockTracer(nil)))
			},
		},
		{
			name: "new license service with no tracer",
			build: func() (service.LicenseService, error) {
				return service.NewLicenseService(new(license.License), mockrepo.NewMockLicenseRepository(nil), mocksvc.NewMockPermissionService(nil), service.WithLogger(mocklog.NewMockLogger(nil)))
			},
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

func TestLicenseService_Expired(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				licenseRepo: mockrepo.NewMockLicenseRepository(nil),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				licenseRepo: mockrepo.NewMockLicenseRepository(nil),
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
			s := newLicenseServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx), tt.fields.licenseRepo, tt.fields.license)
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
		baseService func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/HasFeature", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				licenseRepo: mockrepo.NewMockLicenseRepository(nil),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/HasFeature", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				licenseRepo: mockrepo.NewMockLicenseRepository(nil),
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
			s := newLicenseServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx), tt.fields.licenseRepo, tt.fields.license)
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
		baseService    func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps
		newLicenseRepo func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository
		license        *license.License
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().DocumentCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().NamespaceCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().ActiveOrganizationCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().ProjectCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().RoleCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().ActiveUserCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(_ *gomock.Controller, _ context.Context) repository.LicenseRepository {
					return mockrepo.NewMockLicenseRepository(nil)
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
			wantErr: service.ErrQuotaInvalid,
		},
		{
			name: "quota exceeds threshold",
			args: args{
				ctx:   context.Background(),
				quota: license.QuotaUsers,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().ActiveUserCount(ctx).Return(1, nil).Times(1)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/WithinThreshold", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
					repo.EXPECT().ActiveUserCount(ctx).Return(0, assert.AnError).Times(1)
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
			wantErr: service.ErrQuotaUsageGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := newLicenseServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx), tt.fields.newLicenseRepo(ctrl, tt.args.ctx), tt.fields.license)
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
		baseService    func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps
		newLicenseRepo func(ctrl *gomock.Controller, ctx context.Context) repository.LicenseRepository
		license        *license.License
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/GetLicense", gomock.Len(0)).Return(ctx, span)

					permissionSvc := mocksvc.NewMockPermissionService(ctrl)
					permissionSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(true, nil)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permissionSvc,
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, _ context.Context) repository.LicenseRepository {
					repo := mockrepo.NewMockLicenseRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/GetLicense", gomock.Len(0)).Return(ctx, span)

					permissionSvc := mocksvc.NewMockPermissionService(ctrl)
					permissionSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(false, nil)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permissionSvc,
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, _ context.Context) repository.LicenseRepository {
					return mockrepo.NewMockLicenseRepository(ctrl)
				},
				license: nil,
			},
			wantErr: service.ErrNoPermission,
			want:    license.License{},
		},
		{
			name: "get license has system role error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/GetLicense", gomock.Len(0)).Return(ctx, span)

					permissionSvc := mocksvc.NewMockPermissionService(ctrl)
					permissionSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(false, nil)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permissionSvc,
					}
				},
				newLicenseRepo: func(ctrl *gomock.Controller, _ context.Context) repository.LicenseRepository {
					return mockrepo.NewMockLicenseRepository(ctrl)
				},
				license: nil,
			},
			wantErr: service.ErrLicenseGet,
			want:    license.License{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := newLicenseServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx), tt.fields.newLicenseRepo(ctrl, tt.args.ctx), tt.fields.license)
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
		baseService func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Ping", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				licenseRepo: mockrepo.NewMockLicenseRepository(nil),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) licenseServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.licenseService/Ping", gomock.Len(0)).Return(ctx, span).Times(1)
					tracer.EXPECT().Start(ctx, "service.licenseService/Expired", gomock.Len(0)).Return(ctx, span).Times(1)

					return licenseServiceDeps{
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: mocksvc.NewMockPermissionService(nil),
					}
				},
				licenseRepo: mockrepo.NewMockLicenseRepository(nil),
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
			s := newLicenseServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx), tt.fields.licenseRepo, tt.fields.license)
			err := s.Ping(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
