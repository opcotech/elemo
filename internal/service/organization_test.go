package service

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func organizationsToRepository(orgs []*Organization) []*repository.Organization {
	out := make([]*repository.Organization, len(orgs))
	for i, o := range orgs {
		if o == nil {
			continue
		}
		out[i] = &repository.Organization{
			ID: o.ID, Name: o.Name, Email: o.Email, Logo: o.Logo, Website: o.Website,
			Status: o.Status, NamespaceCount: o.NamespaceCount, TeamCount: o.TeamCount, MemberCount: o.MemberCount,
			DocumentCount: o.DocumentCount,
			CreatedAt:     o.CreatedAt, UpdatedAt: o.UpdatedAt,
		}
	}
	return out
}

func organizationToRepository(o *Organization) *repository.Organization {
	if o == nil {
		return nil
	}
	return &repository.Organization{
		ID: o.ID, Name: o.Name, Email: o.Email, Logo: o.Logo, Website: o.Website,
		Status: o.Status, NamespaceCount: o.NamespaceCount, TeamCount: o.TeamCount, MemberCount: o.MemberCount,
		DocumentCount: o.DocumentCount,
		CreatedAt:     o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}

func TestNewOrganizationService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    OrganizationService
		wantErr error
	}{
		{
			name: "new organization service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithOrganizationRepository(repository.NewMockOrganizationRepository(nil)),
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
					WithEmailService(mock.NewEmailService(nil)),
				},
			},
			want: &organizationService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					userRepo:          repository.NewMockUserRepository(nil),
					organizationRepo:  repository.NewMockOrganizationRepository(nil),
					roleRepo:          repository.NewMockRoleRepository(nil),
					userTokenRepo:     repository.NewMockUserTokenRepository(nil),
					permissionService: NewMockPermissionService(nil),
					licenseService:    mock.NewMockLicenseService(nil),
					emailService:      mock.NewEmailService(nil),
				},
			},
		},
		{
			name: "new organization service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithOrganizationRepository(repository.NewMockOrganizationRepository(nil)),
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
					WithEmailService(mock.NewEmailService(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new organization service with no organization repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoOrganizationRepository,
		},
		{
			name: "new organization service with no permission repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithOrganizationRepository(repository.NewMockOrganizationRepository(nil)),
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
					WithEmailService(mock.NewEmailService(nil)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new organization service with no license service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithOrganizationRepository(repository.NewMockOrganizationRepository(nil)),
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithEmailService(mock.NewEmailService(nil)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
		{
			name: "new organization service with no user repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithOrganizationRepository(repository.NewMockOrganizationRepository(nil)),
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoUserRepository,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := NewOrganizationService(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, opts CreateOrganizationOpts) *baseService
	}
	type args struct {
		ctx   context.Context
		owner model.ID
		opts  CreateOrganizationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(testModel.NewRepositoryOrganization(), nil)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}, nil).AnyTimes()
					roleRepo.EXPECT().GetByKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}, nil).AnyTimes()

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(true)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						roleRepo:          roleRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: CreateOrganizationOpts{
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
		},
		{
			name: "create organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: CreateOrganizationOpts{
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
						licenseService:   licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts:  CreateOrganizationOpts{},
			},
			wantErr: ErrOrganizationCreate,
		},
		{
			name: "create organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: CreateOrganizationOpts{
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: ErrOrganizationCreate,
		},
		{
			name: "create organization out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: CreateOrganizationOpts{
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "create organization with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: CreateOrganizationOpts{
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create organization with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateOrganizationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: CreateOrganizationOpts{
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := s.Create(tt.args.ctx, tt.args.owner, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *Organization) *baseService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Organization
		wantErr error
	}{
		{
			name: "get organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(testModel.NewRepositoryOrganization(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: organizationFromRepository(testModel.NewRepositoryOrganization()),
		},
		{
			name: "get organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrOrganizationGet,
		},
		{
			name: "get organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: ErrOrganizationGet,
		},
		{
			name: "get organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrOrganizationGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
			}
		})
	}
}

func TestOrganizationService_GetAll(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, _, _ int, organizations []*Organization) *baseService
	}
	type args struct {
		ctx    context.Context
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Organization
		wantErr error
	}{
		{
			name: "get all organizations organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, organizations []*Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					userID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().List(ctx, userID, gomock.Any(), repository.OrganizationListProjection()).Return(repository.Page[*repository.Organization]{Items: organizationsToRepository(organizations)}, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				offset: 0,
				limit:  10,
			},
			want: []*Organization{
				organizationFromRepository(testModel.NewRepositoryOrganization()),
				organizationFromRepository(testModel.NewRepositoryOrganization()),
			},
		},
		{
			name: "get all organizations with invalid offset",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: -1,
				limit:  10,
			},
			wantErr: ErrOrganizationGetAll,
		},
		{
			name: "get all organizations with invalid limit",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  -1,
			},
			wantErr: ErrOrganizationGetAll,
		},
		{
			name: "get all organizations with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					userID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().List(ctx, userID, gomock.Any(), repository.OrganizationListProjection()).Return(repository.Page[*repository.Organization]{}, assert.AnError)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrOrganizationGetAll,
		},
		{
			name: "get all organizations with missing user ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrOrganizationGetAll,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := s.List(tt.args.ctx, CursorPage{Size: 10})
			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				require.Equal(t, tt.want, got.Items)
			}
		})
	}
}

func TestOrganizationService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	organizationID := model.MustNewID(model.ResourceTypeOrganization)
	otherOrganizationID := model.MustNewID(model.ResourceTypeOrganization)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateOrganizationOpts, organization *Organization) *baseService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateOrganizationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Organization
		wantErr error
	}{
		{
			name: "update organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateOrganizationOpts, organization *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(organizationToRepository(organization), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			want: organizationFromRepository(testModel.NewRepositoryOrganization()),
		},
		{
			name: "update organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, otherOrganizationID),
				id:  organizationID,
				opts: UpdateOrganizationOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update organization with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
						licenseService:   licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				opts: UpdateOrganizationOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: ErrOrganizationUpdate,
		},
		{
			name: "update organization with empty patch",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, repository.ErrNotFound)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  orgRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   organizationID,
				opts: UpdateOrganizationOpts{},
			},
			wantErr: ErrOrganizationUpdate,
		},
		{
			name: "update organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: UpdateOrganizationOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: ErrOrganizationUpdate,
		},
		{
			name: "update organization out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "update organization with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update organization with expired license error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateOrganizationOpts, _ *Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
			}
		})
	}
}

func TestOrganizationService_Delete(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		force bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "soft delete organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(testModel.NewRepositoryOrganization(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
		},
		{
			name: "force delete organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: true,
			},
		},
		{
			name: "delete organization license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete organization license error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "soft delete organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "force delete organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: true,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete organization with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.ID{},
				force: false,
			},
			wantErr: ErrOrganizationDelete,
		},
		{
			name: "soft delete organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
			wantErr: ErrOrganizationDelete,
		},
		{
			name: "force delete organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: true,
			},
			wantErr: ErrOrganizationDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id),
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_AddMember(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService
	}
	type args struct {
		ctx          context.Context
		organization model.ID
		member       model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add member to organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().AddMember(ctx, organization, userID).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, organization, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
		},
		{
			name: "add member to organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "add member to organization with permission error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "add member to organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.ID{},
				member:       userID,
			},
			wantErr: ErrOrganizationMemberAdd,
		},
		{
			name: "add member to organization with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       model.ID{},
			},
			wantErr: ErrOrganizationMemberAdd,
		},
		{
			name: "add member to organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().AddMember(ctx, organization, userID).Return(assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: ErrOrganizationMemberAdd,
		},
		{
			name: "add member to organization with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organization),
			}
			err := s.AddMember(tt.args.ctx, tt.args.organization, tt.args.member)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_ListMembers(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, members []*repository.OrganizationMember, expected []*OrganizationMember) *baseService
	}
	type args struct {
		ctx            context.Context
		organizationID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*OrganizationMember
		wantErr error
	}{
		{
			name: "get members of organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, members []*repository.OrganizationMember, _ []*OrganizationMember) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/ListMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().ListMembers(ctx, organizationID, gomock.Any()).Return(repository.Page[*repository.OrganizationMember]{Items: members}, nil)

					permissionService := NewMockPermissionService(ctrl)
					// Mock permission check for the context user
					permissionService.EXPECT().CtxUserHas(ctx, organizationID, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permissionService,
					}
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func() []*OrganizationMember {
				user1 := testModel.NewUser()
				user1.ID = model.MustNewID(model.ResourceTypeUser)
				user2 := testModel.NewUser()
				user2.ID = model.MustNewID(model.ResourceTypeUser)
				user3 := testModel.NewUser()
				user3.ID = model.MustNewID(model.ResourceTypeUser)
				user4 := testModel.NewUser()
				user4.ID = model.MustNewID(model.ResourceTypeUser)

				picture1 := func() *string {
					if user1.Picture == "" {
						return nil
					}
					p := user1.Picture
					return &p
				}()
				picture2 := func() *string {
					if user2.Picture == "" {
						return nil
					}
					p := user2.Picture
					return &p
				}()
				picture3 := func() *string {
					if user3.Picture == "" {
						return nil
					}
					p := user3.Picture
					return &p
				}()
				picture4 := func() *string {
					if user4.Picture == "" {
						return nil
					}
					p := user4.Picture
					return &p
				}()

				// Expected results with combined virtual and actual roles
				// User1: has "Owner" permission -> should get "Owner" virtual role
				expected1 := &OrganizationMember{
					ID: user1.ID, FirstName: user1.FirstName, LastName: user1.LastName, Email: user1.Email,
					Picture: picture1, Status: user1.Status, Roles: []string{},
				}
				// User2: has "Member" role -> should get "Admin" virtual role (since write permission)
				expected2 := &OrganizationMember{
					ID: user2.ID, FirstName: user2.FirstName, LastName: user2.LastName, Email: user2.Email,
					Picture: picture2, Status: user2.Status, Roles: []string{},
				}
				// User3: has "Admin", "Member" roles -> should get "Admin" virtual role (deduplicated)
				expected3 := &OrganizationMember{
					ID: user3.ID, FirstName: user3.FirstName, LastName: user3.LastName, Email: user3.Email,
					Picture: picture3, Status: user3.Status, Roles: []string{},
				}
				// User4: has no roles -> should get "Member" virtual role (since read permission)
				expected4 := &OrganizationMember{
					ID: user4.ID, FirstName: user4.FirstName, LastName: user4.LastName, Email: user4.Email,
					Picture: picture4, Status: user4.Status, Roles: []string{},
				}

				return []*OrganizationMember{expected1, expected2, expected3, expected4}
			}(),
		},
		{
			name: "get members of organization with invalid organization id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ []*repository.OrganizationMember, _ []*OrganizationMember) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/ListMembers", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: repository.NewMockOrganizationRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.ID{},
			},
			wantErr: ErrOrganizationMembersGet,
		},
		{
			name: "get members of organization with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, _ []*repository.OrganizationMember, _ []*OrganizationMember) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/ListMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().ListMembers(ctx, organizationID, gomock.Any()).Return(repository.Page[*repository.OrganizationMember]{}, assert.AnError)

					permissionService := NewMockPermissionService(ctrl)
					permissionService.EXPECT().CtxUserHas(ctx, organizationID, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permissionService,
					}
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrOrganizationMembersGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Helper function to extract actual roles from expected roles.
			// Virtual roles ("Owner", "Admin", "Member") are computed from permissions.
			// "Member" is only actual if "Admin" is also present (meaning user has
			// both Admin virtual role from write permission AND Member actual role).
			extractActualRoles := func(roles []string) []string {
				return roles
			}

			// Prepare members from repository (without virtual roles)
			var membersFromRepo []*repository.OrganizationMember
			if len(tt.want) > 0 {
				membersFromRepo = make([]*repository.OrganizationMember, len(tt.want))
				for i, expected := range tt.want {
					actualRoles := extractActualRoles(expected.Roles)
					member := &repository.OrganizationMember{
						ID: expected.ID, FirstName: expected.FirstName, LastName: expected.LastName,
						Email: expected.Email, Picture: expected.Picture, Status: expected.Status, Roles: actualRoles,
					}
					membersFromRepo[i] = member
				}
			}

			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organizationID, membersFromRepo, tt.want),
			}
			page, err := s.ListMembers(tt.args.ctx, tt.args.organizationID, CursorPage{Size: 100})
			members := page.Items
			require.ErrorIs(t, err, tt.wantErr)

			if err == nil {
				require.Equal(t, len(tt.want), len(members))

				// Build lookup map for expected members
				expectedMap := make(map[model.ID]*OrganizationMember, len(tt.want))
				for _, expected := range tt.want {
					expectedMap[expected.ID] = expected
				}

				// Verify each member matches expected values
				for _, member := range members {
					expected, ok := expectedMap[member.ID]
					require.True(t, ok, "member with ID %s not found in expected results", member.ID)

					require.Equal(t, expected.ID, member.ID)
					require.Equal(t, expected.FirstName, member.FirstName, "FirstName mismatch for member %s", member.ID)
					require.Equal(t, expected.LastName, member.LastName, "LastName mismatch for member %s", member.ID)
					require.Equal(t, expected.Email, member.Email, "Email mismatch for member %s", member.ID)
					if expected.Picture == nil {
						require.Nil(t, member.Picture)
					} else {
						require.Equal(t, *expected.Picture, *member.Picture)
					}
					require.Equal(t, expected.Status, member.Status)
					require.ElementsMatch(t, expected.Roles, member.Roles, "roles mismatch for member %s: expected %v, got %v", member.ID, expected.Roles, member.Roles)
				}
			}
		})
	}
}

func TestOrganizationService_RemoveMember(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService
	}
	type args struct {
		ctx          context.Context
		organization model.ID
		member       model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove member from organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, organization, gomock.Any()).Return(true)
					permSvc.EXPECT().ListByPrincipal(ctx, userID).Return([]*Grant{}, nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
		},
		{
			name: "add member to organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "add member to organization with permission error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "add member to organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.ID{},
				member:       userID,
			},
			wantErr: ErrOrganizationMemberRemove,
		},
		{
			name: "add member to organization with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       model.ID{},
			},
			wantErr: ErrOrganizationMemberRemove,
		},
		{
			name: "add member to organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := repository.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, organization, gomock.Any()).Return(true)
					permSvc.EXPECT().ListByPrincipal(ctx, userID).Return([]*Grant{}, nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  organizationRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: ErrOrganizationMemberRemove,
		},
		{
			name: "add member to organization with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organization),
			}
			err := s.RemoveMember(tt.args.ctx, tt.args.organization, tt.args.member)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_InviteMember(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	email := "test@example.com"

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, roleID model.ID) *baseService
	}
	type args struct {
		ctx    context.Context
		orgID  model.ID
		email  string
		roleID []model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "invite member to organization with existing user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						emailService:      emailService,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
		},
		{
			name: "invite member to organization with new pending user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					// Use an email that will generate both firstName and lastName
					testEmail := "john.doe@example.com"

					user := testModel.NewUser()
					user.Email = testEmail
					user.Status = model.UserStatusPending

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, testEmail, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)
					userRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, opts repository.CreateUserOpts) (*repository.User, error) {
						user.ID = model.MustNewID(model.ResourceTypeUser)
						user.Status = model.UserStatusPending
						user.FirstName = opts.FirstName
						user.LastName = opts.LastName
						user.Email = opts.Email
						return user, nil
					})

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, gomock.Any()).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
					permSvc.EXPECT().Has(ctx, gomock.Any(), orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, gomock.Any(), model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						emailService:      emailService,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  "john.doe@example.com", // Use email that generates both firstName and lastName
				roleID: []model.ID{},
			},
		},
		{
			name: "invite member to organization with roleID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						emailService:      emailService,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{roleID},
			},
		},
		{
			name: "invite member with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "invite member with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, invalidOrgID model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after orgID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, invalidOrgID, gomock.Any()).Return(false).AnyTimes()

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  model.MustNewNilID(model.ResourceTypeOrganization),
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: ErrOrganizationMemberInvite,
		},
		{
			name: "invite member with empty email",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  "",
				roleID: []model.ID{},
			},
			wantErr: model.ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "invite member with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "invite member when user already exists as member",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(true, nil)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: ErrOrganizationMemberAlreadyExists,
		},
		{
			name: "invite member with invalid user status",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusDeleted

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
					// HasPermission is not called when user status is invalid - code returns early

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: ErrOrganizationMemberInvalidStatus,
		},
		{
			name: "invite member with email service error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						emailService:      emailService,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: ErrOrganizationMemberInvite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID, tt.args.email, func() model.ID {
					if len(tt.args.roleID) > 0 {
						return tt.args.roleID[0]
					}
					return model.MustNewNilID(model.ResourceTypeRole)
				}()),
			}
			err := s.InviteMember(tt.args.ctx, tt.args.orgID, InviteOrganizationMemberOpts{
				Email: tt.args.email,
				RoleID: func() model.ID {
					if len(tt.args.roleID) > 0 {
						return tt.args.roleID[0]
					}
					return model.ID{}
				}(),
			})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_RevokeInvitation(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *baseService
	}
	type args struct {
		ctx    context.Context
		orgID  model.ID
		userID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "revoke invitation successfully",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusActive

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					// GetAll is only called for pending users, not active users

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
		},
		{
			name: "revoke invitation with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "revoke invitation with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, invalidOrgID, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after orgID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, invalidOrgID, gomock.Any()).Return(false).AnyTimes()

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  model.MustNewNilID(model.ResourceTypeOrganization),
				userID: userID,
			},
			wantErr: ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation with invalid userID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after userID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false).AnyTimes()

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "revoke invitation with user not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  repository.NewMockOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
			wantErr: ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation and cleanup pending user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)
					userRepo.EXPECT().Delete(ctx, userID).Return(nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().List(ctx, userID, repository.CursorPage{Size: 1}, repository.OrganizationListProjection()).Return(repository.Page[*repository.Organization]{}, nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
					logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
		},
		{
			name: "revoke invitation with pending user in multiple organizations",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().List(ctx, userID, repository.CursorPage{Size: 1}, repository.OrganizationListProjection()).Return(repository.Page[*repository.Organization]{Items: []*repository.Organization{testModel.NewRepositoryOrganization()}}, nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
			}
			err := s.RevokeInvitation(tt.args.ctx, tt.args.orgID, tt.args.userID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_AcceptInvitation(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, userID model.ID, token string, userPassword string, roleID model.ID) *baseService
	}
	type args struct {
		ctx          context.Context
		orgID        model.ID
		token        string
		userPassword string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "accept invitation with pending user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					// Extract secret from the public token passed in
					// The token parameter contains the public token, we need to extract the secret from it
					_, secret, _ := auth.SplitToken(token)
					// Hash the secret to match what's stored in userToken
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  user.Email,
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)
					userRepo.EXPECT().Update(ctx, userID, gomock.Any()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "",
				userPassword: "password123",
			},
		},
		{
			name: "accept invitation with active user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					// Extract secret from the public token passed in
					// The token parameter contains the public token, we need to extract the secret from it
					_, secret, _ := auth.SplitToken(token)
					// Hash the secret to match what's stored in userToken
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  user.Email,
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "",
				userPassword: "",
			},
		},
		{
			name: "accept invitation with roleID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  user.Email,
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					roleRepo := repository.NewMockRoleRepository(ctrl)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						roleRepo:          roleRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "",
				userPassword: "",
			},
		},
		{
			name: "accept invitation with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        model.MustNewNilID(model.ResourceTypeOrganization),
				token:        "valid-token",
				userPassword: "",
			},
			wantErr: ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with empty token",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "",
				userPassword: "",
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "accept invitation with invalid token format",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "invalid-token",
				userPassword: "",
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "accept invitation with expired token",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  "test@example.com",
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now
					// Set CreatedAt to be older than deadline
					oldTime := time.Now().Add(-8 * 24 * time.Hour)
					userToken.CreatedAt = &oldTime

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					return &baseService{
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						userTokenRepo: userTokenRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					tokenData := map[string]any{
						"organization_id": orgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
			wantErr: ErrExpiredToken,
		},
		{
			name: "accept invitation with wrong organization ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, _ string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					wrongOrgID := model.MustNewID(model.ResourceTypeOrganization)
					tokenData := map[string]any{
						"organization_id": wrongOrgID.String(),
						"user_id":         userID.String(),
					}
					_, secretToken, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  "test@example.com",
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					return &baseService{
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						userTokenRepo: userTokenRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					wrongOrgID := model.MustNewID(model.ResourceTypeOrganization)
					tokenData := map[string]any{
						"organization_id": wrongOrgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "accept invitation with user not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  "test@example.com",
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)

					return &baseService{
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						userRepo:      userRepo,
						userTokenRepo: userTokenRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					tokenData := map[string]any{
						"organization_id": orgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
			wantErr: ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with invalid user status",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusDeleted

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  user.Email,
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					return &baseService{
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						userRepo:      userRepo,
						userTokenRepo: userTokenRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					tokenData := map[string]any{
						"organization_id": orgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
			wantErr: ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with pending user missing password",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  user.Email,
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					return &baseService{
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						userRepo:      userRepo,
						userTokenRepo: userTokenRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					tokenData := map[string]any{
						"organization_id": orgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
			wantErr: ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with token not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, _ string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)

					return &baseService{
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						userTokenRepo: userTokenRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					tokenData := map[string]any{
						"organization_id": orgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "accept invitation when user already member",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken := &repository.UserToken{
						ID:      model.MustNewID(model.ResourceTypeUserToken),
						UserID:  userID,
						SentTo:  user.Email,
						Token:   secretToken,
						Context: model.UserTokenContextInvite,
					}
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := repository.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					userTokenRepo := repository.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					logger := mock.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return &baseService{
						logger:            logger,
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  orgRepo,
						userTokenRepo:     userTokenRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				token: func() string {
					tokenData := map[string]any{
						"organization_id": orgID.String(),
						"user_id":         userID.String(),
					}
					publicToken, _, _ := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
					return publicToken
				}(),
				userPassword: "",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Generate token if needed for args - this must happen before baseService is called
			// so the token can be used in both the args and the baseService mocks
			var publicToken string
			if tt.args.token == "" && tt.wantErr == nil {
				tokenData := map[string]any{
					"organization_id": tt.args.orgID.String(),
					"user_id":         userID.String(),
				}
				if roleID != model.MustNewNilID(model.ResourceTypeRole) {
					tokenData["role_id"] = roleID.String()
				}
				var err error
				publicToken, _, err = auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
				require.NoError(t, err)
				tt.args.token = publicToken
			} else if tt.args.token != "" {
				publicToken = tt.args.token
			}

			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID, userID, publicToken, tt.args.userPassword, roleID),
			}
			err := s.AcceptInvitation(tt.args.ctx, tt.args.orgID, AcceptOrganizationInvitationOpts{Token: tt.args.token, Password: tt.args.userPassword})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_Create_SeedsAuth(t *testing.T) {
	t.Parallel()

	owner := model.MustNewID(model.ResourceTypeUser)
	org := testModel.NewRepositoryOrganization()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, owner)
	opts := CreateOrganizationOpts{
		Name:    "test-org",
		Email:   "org@example.com",
		Logo:    "https://www.gravatar.com/avatar",
		Website: "https://example.com/",
		Status:  model.OrganizationStatusActive,
	}

	t.Run("creates role templates and grants owner admin plus org member", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

		organizationRepo := repository.NewMockOrganizationRepository(ctrl)
		organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(org, nil)

		roleRepo := repository.NewMockRoleRepository(ctrl)
		for _, tmpl := range model.RoleTemplates {
			roleRepo.EXPECT().Create(ctx, repository.CreateRoleOpts{
				Key:         tmpl.Key,
				Name:        tmpl.Name,
				Description: tmpl.Description,
				Actions:     tmpl.ActionStrings(),
				CreatedBy:   owner,
				BelongsTo:   org.ID,
			}).Return(&repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}, nil)
		}

		adminRole := &repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}
		memberRole := &repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}
		roleRepo.EXPECT().GetByKey(ctx, org.ID, model.RoleKeyOrgAdmin).Return(adminRole, nil)
		roleRepo.EXPECT().GetByKey(ctx, org.ID, model.RoleKeyOrgMember).Return(memberRole, nil)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(true)
		permSvc.EXPECT().GrantRole(ctx, owner, org.ID, adminRole.ID).Return(nil)
		permSvc.EXPECT().GrantRole(ctx, org.ID, org.ID, memberRole.ID).Return(nil)

		licenseSvc := mock.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

		s := &organizationService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			organizationRepo:  organizationRepo,
			roleRepo:          roleRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		}}
		got, err := s.Create(ctx, owner, opts)
		require.NoError(t, err)
		assert.Equal(t, org.ID, got.ID)
	})

	t.Run("seed failure wraps ErrOrganizationCreate", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

		organizationRepo := repository.NewMockOrganizationRepository(ctrl)
		organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(org, nil)

		roleRepo := repository.NewMockRoleRepository(ctrl)
		roleRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(true)

		licenseSvc := mock.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

		s := &organizationService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			organizationRepo:  organizationRepo,
			roleRepo:          roleRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		}}
		_, err := s.Create(ctx, owner, opts)
		require.ErrorIs(t, err, ErrOrganizationCreate)
	})
}

func TestOrganizationService_RemoveMember_DeletesOrgScopedGrants(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	otherOrgID := model.MustNewID(model.ResourceTypeOrganization)
	matchingGrant := &Grant{ID: model.MustNewID(model.ResourceTypePermission), Scope: orgID}
	foreignGrant := &Grant{ID: model.MustNewID(model.ResourceTypePermission), Scope: otherOrgID}
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	ctrl := gomock.NewController(t)
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

	organizationRepo := repository.NewMockOrganizationRepository(ctrl)
	organizationRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)

	permSvc := NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationMembersManage).Return(true)
	permSvc.EXPECT().ListByPrincipal(ctx, userID).Return([]*Grant{matchingGrant, foreignGrant}, nil)
	permSvc.EXPECT().Delete(ctx, matchingGrant.ID).Return(nil)

	licenseSvc := mock.NewMockLicenseService(ctrl)
	licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

	s := &organizationService{baseService: &baseService{
		logger:            mock.NewMockLogger(ctrl),
		tracer:            tracer,
		organizationRepo:  organizationRepo,
		permissionService: permSvc,
		licenseService:    licenseSvc,
	}}
	require.NoError(t, s.RemoveMember(ctx, orgID, userID))
}
