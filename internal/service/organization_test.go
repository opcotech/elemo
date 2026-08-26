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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func mockNotificationServiceAllowCreate(ctrl *gomock.Controller) *mocksvc.MockNotificationService {
	n := mocksvc.NewMockNotificationService(ctrl)
	n.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&service.Notification{}, nil).AnyTimes()
	return n
}

func organizationsToRepository(orgs []*service.Organization) []*repository.Organization {
	out := make([]*repository.Organization, len(orgs))
	for i, o := range orgs {
		if o == nil {
			continue
		}
		out[i] = &repository.Organization{
			ID: o.ID, Slug: o.Slug, Name: o.Name, Email: o.Email, Logo: o.Logo, Website: o.Website,
			Status: o.Status, NamespaceCount: o.NamespaceCount, TeamCount: o.TeamCount, MemberCount: o.MemberCount,
			DocumentCount: o.DocumentCount,
			CreatedAt:     o.CreatedAt, UpdatedAt: o.UpdatedAt,
		}
	}
	return out
}

func organizationToRepository(o *service.Organization) *repository.Organization {
	if o == nil {
		return nil
	}
	return &repository.Organization{
		ID: o.ID, Slug: o.Slug, Name: o.Name, Email: o.Email, Logo: o.Logo, Website: o.Website,
		Status: o.Status, NamespaceCount: o.NamespaceCount, TeamCount: o.TeamCount, MemberCount: o.MemberCount,
		DocumentCount: o.DocumentCount,
		CreatedAt:     o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}

func TestNewOrganizationService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.OrganizationService, error)
		wantErr error
	}{
		{
			name: "new organization service",
			build: func(ctrl *gomock.Controller) (service.OrganizationService, error) {
				return service.NewOrganizationService(mockrepo.NewMockOrganizationRepository(nil), mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), mockrepo.NewMockRoleRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mocksvc.NewMockEmailService(nil), mocksvc.NewMockNotificationService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new organization service with invalid options",
			build: func(_ *gomock.Controller) (service.OrganizationService, error) {
				return service.NewOrganizationService(mockrepo.NewMockOrganizationRepository(nil), mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), mockrepo.NewMockRoleRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mocksvc.NewMockEmailService(nil), mocksvc.NewMockNotificationService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new organization service with no organization repository",
			build: func(ctrl *gomock.Controller) (service.OrganizationService, error) {
				return service.NewOrganizationService(nil, mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), mockrepo.NewMockRoleRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mocksvc.NewMockEmailService(nil), mocksvc.NewMockNotificationService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoOrganizationRepository,
		},
		{
			name: "new organization service with no permission service",
			build: func(ctrl *gomock.Controller) (service.OrganizationService, error) {
				return service.NewOrganizationService(mockrepo.NewMockOrganizationRepository(nil), mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), mockrepo.NewMockRoleRepository(nil), nil, mocksvc.NewMockLicenseService(nil), mocksvc.NewMockEmailService(nil), mocksvc.NewMockNotificationService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoPermissionService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := tt.build(ctrl)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

func TestOrganizationService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, opts service.CreateOrganizationOpts) service.OrganizationService
	}
	type args struct {
		ctx   context.Context
		owner model.ID
		opts  service.CreateOrganizationOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(testModel.NewRepositoryOrganization(), nil)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}, nil).AnyTimes()
					roleRepo.EXPECT().GetByKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.Role{ID: model.MustNewID(model.ResourceTypeRole)}, nil).AnyTimes()

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(true, nil)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							roleRepo,
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mockSearchIndex(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: service.CreateOrganizationOpts{
					Slug:    "acme-org",
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: service.CreateOrganizationOpts{
					Slug:    "acme-org",
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts:  service.CreateOrganizationOpts{},
			},
			wantErr: service.ErrOrganizationCreate,
		},
		{
			name: "create organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: service.CreateOrganizationOpts{
					Slug:    "acme-org",
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: service.ErrOrganizationCreate,
		},
		{
			name: "create organization out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: service.CreateOrganizationOpts{
					Slug:    "acme-org",
					Name:    "test-org",
					Email:   "org@example.com",
					Logo:    "https://www.gravatar.com/avatar",
					Website: "https://example.com/",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: service.ErrQuotaExceeded,
		},
		{
			name: "create organization with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: service.CreateOrganizationOpts{
					Slug:    "acme-org",
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateOrganizationOpts) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner: userID,
				opts: service.CreateOrganizationOpts{
					Slug:    "acme-org",
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.opts)
			_, err := s.Create(tt.args.ctx, tt.args.owner, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *service.Organization) service.OrganizationService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Organization
		wantErr error
	}{
		{
			name: "get organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(testModel.NewRepositoryOrganization(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: service.OrganizationFromRepository(testModel.NewRepositoryOrganization()),
		},
		{
			name: "get organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: service.ErrOrganizationGet,
		},
		{
			name: "get organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: service.ErrOrganizationGet,
		},
		{
			name: "get organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(nil, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: service.ErrOrganizationGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.want)
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

func TestOrganizationService_GetByRef(t *testing.T) {
	t.Parallel()

	repoOrg := testModel.NewRepositoryOrganization()
	want := service.OrganizationFromRepository(repoOrg)
	ctx := context.Background()

	t.Run("get organization by slug", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.organizationService/GetByRef", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "service.organizationService/Resolve", gomock.Len(0)).Return(ctx, span)

		organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
		organizationRepo.EXPECT().GetByRef(ctx, model.ID{}, repoOrg.Slug, repository.OrganizationDetailProjection()).Return(repoOrg, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, repoOrg.ID, model.ActionOrganizationRead).Return(true, nil)

		svc, err := service.NewOrganizationService(
			organizationRepo,
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			mockrepo.NewMockRoleRepository(ctrl),
			permSvc,
			mocksvc.NewMockLicenseService(ctrl),
			mocksvc.NewMockEmailService(ctrl),
			mocksvc.NewMockNotificationService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)

		got, err := svc.GetByRef(ctx, model.ID{}, repoOrg.Slug)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get organization by slug without permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.organizationService/GetByRef", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "service.organizationService/Resolve", gomock.Len(0)).Return(ctx, span)

		organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
		organizationRepo.EXPECT().GetByRef(ctx, model.ID{}, repoOrg.Slug, repository.OrganizationDetailProjection()).Return(repoOrg, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, repoOrg.ID, model.ActionOrganizationRead).Return(false, nil)

		svc, err := service.NewOrganizationService(
			organizationRepo,
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			mockrepo.NewMockRoleRepository(ctrl),
			permSvc,
			mocksvc.NewMockLicenseService(ctrl),
			mocksvc.NewMockEmailService(ctrl),
			mocksvc.NewMockNotificationService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)

		got, err := svc.GetByRef(ctx, model.ID{}, repoOrg.Slug)
		require.ErrorIs(t, err, service.ErrNoPermission)
		require.Nil(t, got)
	})
}

func TestOrganizationService_List(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, _, _ int, organizations []*service.Organization) service.OrganizationService
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
		want    []*service.Organization
		wantErr error
	}{
		{
			name: "get all organizations organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, organizations []*service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					userID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{
						UserID:     userID,
						Action:     model.ActionOrganizationRead,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionDesc,
						Projection: repository.OrganizationListProjection(),
					}).Return(repository.Page[*repository.Organization]{Items: organizationsToRepository(organizations)}, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				offset: 0,
				limit:  10,
			},
			want: []*service.Organization{
				service.OrganizationFromRepository(testModel.NewRepositoryOrganization()),
				service.OrganizationFromRepository(testModel.NewRepositoryOrganization()),
			},
		},
		{
			name: "get all organizations with invalid offset",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: -1,
				limit:  10,
			},
			wantErr: service.ErrOrganizationList,
		},
		{
			name: "get all organizations with invalid limit",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  -1,
			},
			wantErr: service.ErrOrganizationList,
		},
		{
			name: "get all organizations with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					userID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{
						UserID:     userID,
						Action:     model.ActionOrganizationRead,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionDesc,
						Projection: repository.OrganizationListProjection(),
					}).Return(repository.Page[*repository.Organization]{}, assert.AnError)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				offset: 0,
				limit:  10,
			},
			wantErr: service.ErrOrganizationList,
		},
		{
			name: "get all organizations with missing user ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: service.ErrOrganizationList,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.want)
			got, err := s.List(tt.args.ctx, service.CursorPage{Size: 10})
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateOrganizationOpts, organization *service.Organization) service.OrganizationService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts service.UpdateOrganizationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Organization
		wantErr error
	}{
		{
			name: "update organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateOrganizationOpts, organization *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(organizationToRepository(organization), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mockSearchIndex(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: service.UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			want: service.OrganizationFromRepository(testModel.NewRepositoryOrganization()),
		},
		{
			name: "update organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, otherOrganizationID),
				id:  organizationID,
				opts: service.UpdateOrganizationOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update organization with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				opts: service.UpdateOrganizationOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: service.ErrOrganizationUpdate,
		},
		{
			name: "update organization with empty patch",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, repository.ErrNotFound)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   organizationID,
				opts: service.UpdateOrganizationOpts{},
			},
			wantErr: service.ErrOrganizationUpdate,
		},
		{
			name: "update organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: service.UpdateOrganizationOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: service.ErrOrganizationUpdate,
		},
		{
			name: "update organization out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: service.UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			wantErr: service.ErrQuotaExceeded,
		},
		{
			name: "update organization with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: service.UpdateOrganizationOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.OrganizationStatusActive),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update organization with expired license error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateOrganizationOpts, _ *service.Organization) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				opts: service.UpdateOrganizationOpts{
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want)
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(testModel.NewRepositoryOrganization(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mockSearchDeleteByScope(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mockSearchDeleteByScope(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "force delete organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: true,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete organization with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.ID{},
				force: false,
			},
			wantErr: service.ErrOrganizationDelete,
		},
		{
			name: "soft delete organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: false,
			},
			wantErr: service.ErrOrganizationDelete,
		},
		{
			name: "force delete organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeOrganization),
				force: true,
			},
			wantErr: service.ErrOrganizationDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id)
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_AddMember(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) service.OrganizationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().AddMember(ctx, organization, userID).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, organization, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "add member to organization with permission error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "add member to organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.ID{},
				member:       userID,
			},
			wantErr: service.ErrOrganizationMemberAdd,
		},
		{
			name: "add member to organization with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       model.ID{},
			},
			wantErr: service.ErrOrganizationMemberAdd,
		},
		{
			name: "add member to organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().AddMember(ctx, organization, userID).Return(assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: service.ErrOrganizationMemberAdd,
		},
		{
			name: "add member to organization with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AddMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organization)
			err := s.AddMember(tt.args.ctx, tt.args.organization, tt.args.member)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_ListMembers(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, members []*repository.OrganizationMember, expected []*service.OrganizationMember) service.OrganizationService
	}
	type args struct {
		ctx            context.Context
		organizationID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*service.OrganizationMember
		wantErr error
	}{
		{
			name: "get members of organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, members []*repository.OrganizationMember, _ []*service.OrganizationMember) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/ListMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().ListMembers(ctx, organizationID, gomock.Any()).Return(repository.Page[*repository.OrganizationMember]{Items: members}, nil)

					permissionService := mocksvc.NewMockPermissionService(ctrl)
					// Mock permission check for the context user
					permissionService.EXPECT().CtxUserHas(ctx, organizationID, gomock.Any()).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permissionService,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func() []*service.OrganizationMember {
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
				expected1 := &service.OrganizationMember{
					ID: user1.ID, FirstName: user1.FirstName, LastName: user1.LastName, Email: user1.Email,
					Picture: picture1, Status: user1.Status, Roles: []string{},
				}
				// User2: has "Member" role -> should get "Admin" virtual role (since write permission)
				expected2 := &service.OrganizationMember{
					ID: user2.ID, FirstName: user2.FirstName, LastName: user2.LastName, Email: user2.Email,
					Picture: picture2, Status: user2.Status, Roles: []string{},
				}
				// User3: has "Admin", "Member" roles -> should get "Admin" virtual role (deduplicated)
				expected3 := &service.OrganizationMember{
					ID: user3.ID, FirstName: user3.FirstName, LastName: user3.LastName, Email: user3.Email,
					Picture: picture3, Status: user3.Status, Roles: []string{},
				}
				// User4: has no roles -> should get "Member" virtual role (since read permission)
				expected4 := &service.OrganizationMember{
					ID: user4.ID, FirstName: user4.FirstName, LastName: user4.LastName, Email: user4.Email,
					Picture: picture4, Status: user4.Status, Roles: []string{},
				}

				return []*service.OrganizationMember{expected1, expected2, expected3, expected4}
			}(),
		},
		{
			name: "get members of organization with invalid organization id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ []*repository.OrganizationMember, _ []*service.OrganizationMember) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/ListMembers", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.ID{},
			},
			wantErr: service.ErrOrganizationMembersGet,
		},
		{
			name: "get members of organization with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, _ []*repository.OrganizationMember, _ []*service.OrganizationMember) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/ListMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().ListMembers(ctx, organizationID, gomock.Any()).Return(repository.Page[*repository.OrganizationMember]{}, assert.AnError)

					permissionService := mocksvc.NewMockPermissionService(ctrl)
					permissionService.EXPECT().CtxUserHas(ctx, organizationID, gomock.Any()).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permissionService,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: service.ErrOrganizationMembersGet,
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organizationID, membersFromRepo, tt.want)
			page, err := s.ListMembers(tt.args.ctx, tt.args.organizationID, service.CursorPage{Size: 100})
			members := page.Items
			require.ErrorIs(t, err, tt.wantErr)

			if err == nil {
				require.Equal(t, len(tt.want), len(members))

				// Build lookup map for expected members
				expectedMap := make(map[model.ID]*service.OrganizationMember, len(tt.want))
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) service.OrganizationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(nil)
					organizationRepo.EXPECT().Get(ctx, organization, repository.OrganizationDetailProjection()).Return(&repository.Organization{Name: "org"}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, organization, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().ListByPrincipal(ctx, userID).Return([]*service.Grant{}, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					notificationSvc := mocksvc.NewMockNotificationService(ctrl)
					notificationSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&service.Notification{}, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							notificationSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "add member to organization with permission error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, model.MustNewNilID(model.ResourceTypeOrganization), gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "add member to organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.ID{},
				member:       userID,
			},
			wantErr: service.ErrOrganizationMemberRemove,
		},
		{
			name: "add member to organization with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       model.ID{},
			},
			wantErr: service.ErrOrganizationMemberRemove,
		},
		{
			name: "add member to organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, organization, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().ListByPrincipal(ctx, userID).Return([]*service.Grant{}, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							organizationRepo,
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				organization: model.MustNewNilID(model.ResourceTypeOrganization),
				member:       userID,
			},
			wantErr: service.ErrOrganizationMemberRemove,
		},
		{
			name: "add member to organization with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organization)
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, roleID model.ID) service.OrganizationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mocksvc.NewMockEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							emailService,
							mockNotificationServiceAllowCreate(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					// Use an email that will generate both firstName and lastName
					testEmail := "john.doe@example.com"

					user := testModel.NewUser()
					user.Email = testEmail
					user.Status = model.UserStatusPending

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, testEmail, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)
					userRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, opts repository.CreateUserOpts) (*repository.User, error) {
						user.ID = model.MustNewID(model.ResourceTypeUser)
						user.Status = model.UserStatusPending
						user.FirstName = opts.FirstName
						user.LastName = opts.LastName
						user.Email = opts.Email
						return user, nil
					})

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, gomock.Any()).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().Has(ctx, gomock.Any(), orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, gomock.Any(), model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mocksvc.NewMockEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							emailService,
							mockNotificationServiceAllowCreate(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mocksvc.NewMockEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							emailService,
							mockNotificationServiceAllowCreate(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, invalidOrgID model.ID, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after orgID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, invalidOrgID, gomock.Any()).Return(false, nil).AnyTimes()

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  model.MustNewNilID(model.ResourceTypeOrganization),
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: service.ErrOrganizationMemberInvite,
		},
		{
			name: "invite member with empty email",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "invite member when user already exists as member",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(true, nil)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: service.ErrOrganizationMemberAlreadyExists,
		},
		{
			name: "invite member with invalid user status",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusDeleted

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)
					// HasPermission is not called when user status is invalid - code returns early

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: service.ErrOrganizationMemberInvalidStatus,
		},
		{
			name: "invite member with email service error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewRepositoryOrganization()
					organization.ID = orgID

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().Has(ctx, user.ID, orgID, gomock.Any()).Return(false, nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.UserToken{}, nil)

					emailService := mocksvc.NewMockEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							emailService,
							mockNotificationServiceAllowCreate(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				email:  email,
				roleID: []model.ID{},
			},
			wantErr: service.ErrOrganizationMemberInvite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID, tt.args.email, func() model.ID {
				if len(tt.args.roleID) > 0 {
					return tt.args.roleID[0]
				}
				return model.MustNewNilID(model.ResourceTypeRole)
			}())
			err := s.InviteMember(tt.args.ctx, tt.args.orgID, service.InviteOrganizationMemberOpts{
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) service.OrganizationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusActive

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					// GetAll is only called for pending users, not active users

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, invalidOrgID, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after orgID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, invalidOrgID, gomock.Any()).Return(false, nil).AnyTimes()

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  model.MustNewNilID(model.ResourceTypeOrganization),
				userID: userID,
			},
			wantErr: service.ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation with invalid userID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after userID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false, nil).AnyTimes()

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: service.ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "revoke invitation with user not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  orgID,
				userID: userID,
			},
			wantErr: service.ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation and cleanup pending user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)
					userRepo.EXPECT().Delete(ctx, userID).Return(nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{
						UserID:     userID,
						Action:     model.ActionOrganizationRead,
						Page:       repository.CursorPage{Size: 1},
						Order:      repository.SortDirectionDesc,
						Projection: repository.OrganizationListProjection(),
					}).Return(repository.Page[*repository.Organization]{}, nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
					logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{
						UserID:     userID,
						Action:     model.ActionOrganizationRead,
						Page:       repository.CursorPage{Size: 1},
						Order:      repository.SortDirectionDesc,
						Projection: repository.OrganizationListProjection(),
					}).Return(repository.Page[*repository.Organization]{Items: []*repository.Organization{testModel.NewRepositoryOrganization()}}, nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID)
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, userID model.ID, token string, userPassword string, roleID model.ID) service.OrganizationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)
					userRepo.EXPECT().Update(ctx, userID, gomock.Any()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							roleRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        model.MustNewNilID(model.ResourceTypeOrganization),
				token:        "valid-token",
				userPassword: "",
			},
			wantErr: service.ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with empty token",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "",
				userPassword: "",
			},
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "accept invitation with invalid token format",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:          context.Background(),
				orgID:        orgID,
				token:        "invalid-token",
				userPassword: "",
			},
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "accept invitation with expired token",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			wantErr: service.ErrExpiredToken,
		},
		{
			name: "accept invitation with wrong organization ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, _ string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "accept invitation with user not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			wantErr: service.ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with invalid user status",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			wantErr: service.ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with pending user missing password",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			wantErr: service.ErrOrganizationInviteAccept,
		},
		{
			name: "accept invitation with token not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, userID model.ID, _ string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							mockrepo.NewMockOrganizationRepository(ctrl),
							mockrepo.NewMockUserRepository(ctrl),
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "accept invitation when user already member",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) service.OrganizationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
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

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID, repository.UserDetailProjection()).Return(user, nil)

					orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().GrantRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return func() service.OrganizationService {
						svc, err := service.NewOrganizationService(
							orgRepo,
							userRepo,
							userTokenRepo,
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockEmailService(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(logger),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID, userID, publicToken, tt.args.userPassword, roleID)
			err := s.AcceptInvitation(tt.args.ctx, tt.args.orgID, service.AcceptOrganizationInvitationOpts{Token: tt.args.token, Password: tt.args.userPassword})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_Create_SeedsAuth(t *testing.T) {
	t.Parallel()

	owner := model.MustNewID(model.ResourceTypeUser)
	org := testModel.NewRepositoryOrganization()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, owner)
	opts := service.CreateOrganizationOpts{
		Slug:    "acme-org",
		Name:    "test-org",
		Email:   "org@example.com",
		Logo:    "https://www.gravatar.com/avatar",
		Website: "https://example.com/",
		Status:  model.OrganizationStatusActive,
	}

	t.Run("creates role templates and grants owner admin plus org member", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

		organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
		organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(org, nil)

		roleRepo := mockrepo.NewMockRoleRepository(ctrl)
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

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(true, nil)
		permSvc.EXPECT().GrantRole(ctx, owner, org.ID, adminRole.ID).Return(nil)
		permSvc.EXPECT().GrantRole(ctx, org.ID, org.ID, memberRole.ID).Return(nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

		s := func() service.OrganizationService {
			svc, err := service.NewOrganizationService(
				organizationRepo,
				mockrepo.NewMockUserRepository(ctrl),
				mockrepo.NewMockUserTokenRepository(ctrl),
				roleRepo,
				permSvc,
				licenseSvc,
				mocksvc.NewMockEmailService(ctrl),
				mocksvc.NewMockNotificationService(ctrl),
				mockSearchIndex(ctrl),
				service.WithLogger(mocklog.NewMockLogger(ctrl)),
				service.WithTracer(tracer),
			)
			if err != nil {
				panic(err)
			}
			return svc
		}()
		got, err := s.Create(ctx, owner, opts)
		require.NoError(t, err)
		assert.Equal(t, org.ID, got.ID)
	})

	t.Run("seed failure wraps service.ErrOrganizationCreate", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

		organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
		organizationRepo.EXPECT().Create(ctx, gomock.Any()).Return(org, nil)

		roleRepo := mockrepo.NewMockRoleRepository(ctrl)
		roleRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(true, nil)

		s := func() service.OrganizationService {
			svc, err := service.NewOrganizationService(
				organizationRepo,
				mockrepo.NewMockUserRepository(ctrl),
				mockrepo.NewMockUserTokenRepository(ctrl),
				roleRepo,
				permSvc,
				licenseSvc,
				mocksvc.NewMockEmailService(ctrl),
				mocksvc.NewMockNotificationService(ctrl),
				mocksvc.NewMockSearchService(ctrl),
				service.WithLogger(mocklog.NewMockLogger(ctrl)),
				service.WithTracer(tracer),
			)
			if err != nil {
				panic(err)
			}
			return svc
		}()
		_, err := s.Create(ctx, owner, opts)
		require.ErrorIs(t, err, service.ErrOrganizationCreate)
	})
}

func TestOrganizationService_RemoveMember_DeletesOrgScopedGrants(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	otherOrgID := model.MustNewID(model.ResourceTypeOrganization)
	matchingGrant := &service.Grant{ID: model.MustNewID(model.ResourceTypePermission), Scope: orgID}
	foreignGrant := &service.Grant{ID: model.MustNewID(model.ResourceTypePermission), Scope: otherOrgID}
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	ctrl := gomock.NewController(t)
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

	organizationRepo := mockrepo.NewMockOrganizationRepository(ctrl)
	organizationRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
	organizationRepo.EXPECT().Get(ctx, orgID, repository.OrganizationDetailProjection()).Return(&repository.Organization{Name: "org"}, nil)

	permSvc := mocksvc.NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationMembersManage).Return(true, nil)
	permSvc.EXPECT().ListByPrincipal(ctx, userID).Return([]*service.Grant{matchingGrant, foreignGrant}, nil)
	permSvc.EXPECT().Delete(ctx, matchingGrant.ID).Return(nil)

	licenseSvc := mocksvc.NewMockLicenseService(ctrl)
	licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

	notificationSvc := mocksvc.NewMockNotificationService(ctrl)
	notificationSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&service.Notification{}, nil)

	s := func() service.OrganizationService {
		svc, err := service.NewOrganizationService(
			organizationRepo,
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			mockrepo.NewMockRoleRepository(ctrl),
			permSvc,
			licenseSvc,
			mocksvc.NewMockEmailService(ctrl),
			notificationSvc,
			mocksvc.NewMockSearchService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()
	require.NoError(t, s.RemoveMember(ctx, orgID, userID))
}
