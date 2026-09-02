package service_test

import (
	"context"
	"testing"

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
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestNewRoleService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.RoleService, error)
		wantErr error
	}{
		{
			name: "new role service",
			build: func(ctrl *gomock.Controller) (service.RoleService, error) {
				return service.NewRoleService(mockrepo.NewMockRoleRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mockrepo.NewMockOrganizationRepository(nil), mocksvc.NewMockNotificationService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new role service with no role repository",
			build: func(ctrl *gomock.Controller) (service.RoleService, error) {
				return service.NewRoleService(nil, mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mockrepo.NewMockOrganizationRepository(nil), mocksvc.NewMockNotificationService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoRoleRepository,
		},
		{
			name: "new role service with no permission service",
			build: func(ctrl *gomock.Controller) (service.RoleService, error) {
				return service.NewRoleService(mockrepo.NewMockRoleRepository(nil), nil, mocksvc.NewMockLicenseService(nil), mockrepo.NewMockOrganizationRepository(nil), mocksvc.NewMockNotificationService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new role service with no license service",
			build: func(ctrl *gomock.Controller) (service.RoleService, error) {
				return service.NewRoleService(mockrepo.NewMockRoleRepository(nil), mocksvc.NewMockPermissionService(nil), nil, mockrepo.NewMockOrganizationRepository(nil), mocksvc.NewMockNotificationService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new role service with invalid options",
			build: func(_ *gomock.Controller) (service.RoleService, error) {
				return service.NewRoleService(mockrepo.NewMockRoleRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mockrepo.NewMockOrganizationRepository(nil), mocksvc.NewMockNotificationService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			got, err := tt.build(ctrl)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

func TestRoleService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, opts service.CreateRoleOpts) service.RoleService
	}
	type args struct {
		ctx       context.Context
		owner     model.ID
		belongsTo model.ID
		opts      service.CreateRoleOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, opts service.CreateRoleOpts) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Create(ctx, repository.CreateRoleOpts{
						Key: opts.Key, Name: opts.Name, Description: opts.Description, Actions: model.ActionStrings(opts.Actions), CreatedBy: owner, BelongsTo: belongsTo,
					}).Return(testModel.NewRepositoryRole(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							roleRepo,
							permSvc,
							licenseSvc,
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: service.CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
		},
		{
			name: "create new role with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, opts service.CreateRoleOpts) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Create(ctx, repository.CreateRoleOpts{
						Key: opts.Key, Name: opts.Name, Description: opts.Description, Actions: model.ActionStrings(opts.Actions), CreatedBy: owner, BelongsTo: belongsTo,
					}).Return(nil, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							roleRepo,
							permSvc,
							licenseSvc,
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: service.CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: assert.AnError,
		},
		{
			name: "create new role license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.CreateRoleOpts) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: service.CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create new role invalid role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.CreateRoleOpts) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts:      service.CreateRoleOpts{},
			},
			wantErr: service.ErrRoleCreate,
		},
		{
			name: "create new role quota exceeded",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ service.CreateRoleOpts) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(false, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: service.CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: service.ErrQuotaExceeded,
		},
		{
			name: "create new role with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ service.CreateRoleOpts) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							licenseSvc,
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: service.CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.opts)
			_, err := s.Create(tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.opts)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *repository.Role) service.RoleService
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		repoRole *repository.Role
		wantErr  error
	}{
		{
			name: "get role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, id, belongsTo, repository.RoleDetailProjection()).Return(role, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							roleRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(nil)),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			repoRole: testModel.NewRepositoryRole(),
		},
		{
			name: "get role with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, id, belongsTo, repository.RoleDetailProjection()).Return(role, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							roleRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(nil)),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: assert.AnError,
		},
		{
			name: "get role with invalid role id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(nil)),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.ID{},
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: service.ErrRoleGet,
		},
		{
			name: "get role with no role permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ *repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(nil)),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get role with no related permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ *repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(nil)),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.repoRole)
			got, err := s.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, service.RoleFromRepository(tt.repoRole), got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestRoleService_ListBelongsTo(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, page service.CursorPage, roles []*repository.Role) service.RoleService
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		page      service.CursorPage
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		repoRoles []*repository.Role
		wantErr   error
	}{
		{
			name: "get roles belongs to",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, page service.CursorPage, roles []*repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().ListBelongsTo(ctx, belongsTo, page, repository.RoleListProjection()).Return(repository.Page[*repository.Role]{Items: roles}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							roleRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      service.CursorPage{Size: 10},
			},
			repoRoles: []*repository.Role{
				testModel.NewRepositoryRole(),
				testModel.NewRepositoryRole(),
			},
		},
		{
			name: "get roles belongs to with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, page service.CursorPage, _ []*repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					roleRepo := mockrepo.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().ListBelongsTo(ctx, belongsTo, page, repository.RoleListProjection()).Return(repository.Page[*repository.Role]{}, assert.AnError)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							roleRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      service.CursorPage{Size: 10},
			},
			wantErr: assert.AnError,
		},
		{
			name: "get roles belongs to with invalid role id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.ID{},
				page:      service.CursorPage{Size: 10},
			},
			wantErr: service.ErrRoleGetBelongsTo,
		},
		{
			name: "get roles belongs to with no permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ service.CursorPage, _ []*repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false, nil)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get roles belongs to with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      service.CursorPage{Size: -1},
			},
			wantErr: service.ErrRoleGetBelongsTo,
		},
		{
			name: "get roles belongs to with oversized page",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Role) service.RoleService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					return func() service.RoleService {
						svc, err := service.NewRoleService(
							mockrepo.NewMockRoleRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mockrepo.NewMockOrganizationRepository(ctrl),
							mocksvc.NewMockNotificationService(ctrl),
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
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      service.CursorPage{Size: repository.MaxPageSize + 1},
			},
			wantErr: service.ErrRoleGetBelongsTo,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.page, tt.repoRoles)
			got, err := s.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.page)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, service.RolesFromRepository(tt.repoRoles), got.Items)
			} else {
				assert.Empty(t, got.Items)
			}
		})
	}
}

//nolint:revive // test factories take gomock.Controller first
func newRoleServiceForTest(ctrl *gomock.Controller, ctx context.Context, spanName string) (service.RoleService, *mockrepo.MockRoleRepository, *mocksvc.MockPermissionService, *mocksvc.MockLicenseService) {
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, spanName, gomock.Len(0)).Return(ctx, span)

	roleRepo := mockrepo.NewMockRoleRepository(ctrl)
	permSvc := mocksvc.NewMockPermissionService(ctrl)
	licenseSvc := mocksvc.NewMockLicenseService(ctrl)
	orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)
	orgRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.Organization{Name: "org"}, nil).AnyTimes()
	roleRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.Role{Name: "role"}, nil).AnyTimes()
	notificationSvc := mocksvc.NewMockNotificationService(ctrl)
	notificationSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&service.Notification{}, nil).AnyTimes()

	return func() service.RoleService {
		svc, err := service.NewRoleService(
			roleRepo,
			permSvc,
			licenseSvc,
			orgRepo,
			notificationSvc,
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}(), roleRepo, permSvc, licenseSvc
}

func TestRoleService_Update(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	repoRole := testModel.NewRepositoryRole()
	repoRole.Name = "updated-role"
	ctx := context.Background()
	opts := service.UpdateRoleOpts{Name: optional.Some("updated-role")}

	t.Run("update role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true, nil)
		roleRepo.EXPECT().Update(ctx, roleID, orgID, repository.UpdateRoleOpts{Name: opts.Name}).Return(repoRole, nil)

		got, err := s.Update(ctx, roleID, orgID, opts)
		require.NoError(t, err)
		assert.Equal(t, "updated-role", got.Name)
	})

	t.Run("update role clears actions", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Update")
		cleared := testModel.NewRepositoryRole()
		cleared.Actions = []string{}
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true, nil)
		roleRepo.EXPECT().Update(ctx, roleID, orgID, repository.UpdateRoleOpts{
			Actions: optional.Some([]string{}),
		}).Return(cleared, nil)

		got, err := s.Update(ctx, roleID, orgID, service.UpdateRoleOpts{
			Actions: optional.Null[[]model.Action](),
		})
		require.NoError(t, err)
		assert.Empty(t, got.Actions)
	})

	t.Run("update role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true, nil)
		roleRepo.EXPECT().Update(ctx, roleID, orgID, repository.UpdateRoleOpts{Name: opts.Name}).Return(nil, assert.AnError)

		_, err := s.Update(ctx, roleID, orgID, opts)
		require.ErrorIs(t, err, service.ErrRoleUpdate)
	})

	t.Run("update role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		_, err := s.Update(ctx, roleID, orgID, opts)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("update role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		_, err := s.Update(ctx, model.ID{}, orgID, opts)
		require.ErrorIs(t, err, service.ErrRoleUpdate)
	})

	t.Run("update role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(false, nil)

		_, err := s.Update(ctx, roleID, orgID, opts)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestRoleService_ListMembers(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	repoUser := testModel.NewRepositoryUser()
	page := service.CursorPage{Size: 10}
	ctx := context.Background()

	t.Run("list role members", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, _ := newRoleServiceForTest(ctrl, ctx, "service.roleService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		roleRepo.EXPECT().ListMembers(ctx, roleID, orgID, page).Return(repository.Page[*repository.User]{Items: []*repository.User{repoUser}}, nil)

		got, err := s.ListMembers(ctx, roleID, orgID, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoUser.ID, got.Items[0].ID)
	})

	t.Run("list role members with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, _ := newRoleServiceForTest(ctrl, ctx, "service.roleService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		roleRepo.EXPECT().ListMembers(ctx, roleID, orgID, page).Return(repository.Page[*repository.User]{}, assert.AnError)

		_, err := s.ListMembers(ctx, roleID, orgID, page)
		require.ErrorIs(t, err, service.ErrOrganizationMembersGet)
	})

	t.Run("list role members with invalid belongs-to id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, _ := newRoleServiceForTest(ctrl, ctx, "service.roleService/ListMembers")

		_, err := s.ListMembers(ctx, roleID, model.ID{}, page)
		require.ErrorIs(t, err, service.ErrRoleGetBelongsTo)
	})

	t.Run("list role members with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, _ := newRoleServiceForTest(ctrl, ctx, "service.roleService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		_, err := s.ListMembers(ctx, roleID, orgID, page)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestRoleService_AddMember(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	memberID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.Background()

	t.Run("add member to role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		roleRepo.EXPECT().AddMember(ctx, roleID, memberID, orgID).Return(nil)

		require.NoError(t, s.AddMember(ctx, roleID, memberID, orgID))
	})

	t.Run("add member to role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		roleRepo.EXPECT().AddMember(ctx, roleID, memberID, orgID).Return(assert.AnError)

		err := s.AddMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, service.ErrRoleAddMember)
	})

	t.Run("add member to role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := s.AddMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("add member to role with invalid member id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := s.AddMember(ctx, roleID, model.ID{}, orgID)
		require.ErrorIs(t, err, service.ErrRoleAddMember)
	})

	t.Run("add member to role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := s.AddMember(ctx, model.ID{}, memberID, orgID)
		require.ErrorIs(t, err, service.ErrRoleAddMember)
	})

	t.Run("add member to role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		err := s.AddMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestRoleService_RemoveMember(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	memberID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.Background()

	t.Run("remove member from role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		roleRepo.EXPECT().RemoveMember(ctx, roleID, memberID, orgID).Return(nil)

		require.NoError(t, s.RemoveMember(ctx, roleID, memberID, orgID))
	})

	t.Run("remove member from role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		roleRepo.EXPECT().RemoveMember(ctx, roleID, memberID, orgID).Return(assert.AnError)

		err := s.RemoveMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, service.ErrRoleRemoveMember)
	})

	t.Run("remove member from role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := s.RemoveMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("remove member from role with invalid member id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := s.RemoveMember(ctx, roleID, model.ID{}, orgID)
		require.ErrorIs(t, err, service.ErrRoleRemoveMember)
	})

	t.Run("remove member from role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := s.RemoveMember(ctx, model.ID{}, memberID, orgID)
		require.ErrorIs(t, err, service.ErrRoleRemoveMember)
	})

	t.Run("remove member from role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		err := s.RemoveMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestRoleService_Delete(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	ctx := context.Background()

	t.Run("delete role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true, nil)
		roleRepo.EXPECT().Delete(ctx, roleID, orgID).Return(nil)

		require.NoError(t, s.Delete(ctx, roleID, orgID))
	})

	t.Run("delete role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, roleRepo, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true, nil)
		roleRepo.EXPECT().Delete(ctx, roleID, orgID).Return(assert.AnError)

		err := s.Delete(ctx, roleID, orgID)
		require.ErrorIs(t, err, service.ErrRoleDelete)
	})

	t.Run("delete role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := s.Delete(ctx, roleID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("delete role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := s.Delete(ctx, model.ID{}, orgID)
		require.ErrorIs(t, err, service.ErrRoleDelete)
	})

	t.Run("delete role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newRoleServiceForTest(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(false, nil)

		err := s.Delete(ctx, roleID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}
