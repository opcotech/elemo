package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/assert"
)

func TestNewRoleService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    RoleService
		wantErr error
	}{
		{
			name: "new role service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithRoleRepository(mock.NewRoleRepository(nil)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &roleService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					roleRepo:          mock.NewRoleRepository(nil),
					userRepo:          new(mock.UserRepository),
					permissionService: mock.NewPermissionService(nil),
					licenseService:    mock.NewMockLicenseService(nil),
				},
			},
		},
		{
			name: "new role service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
					WithRoleRepository(mock.NewRoleRepository(nil)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new role service with no role repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoRoleRepository,
		},
		{
			name: "new role service with no user repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithRoleRepository(mock.NewRoleRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoUserRepository,
		},
		{
			name: "new role service with no permission service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithRoleRepository(mock.NewRoleRepository(nil)),
					WithUserRepository(new(mock.UserRepository)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new role service with no license service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithRoleRepository(mock.NewRoleRepository(nil)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewRoleService(tt.args.opts...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, role *model.Role) *baseService
	}
	type args struct {
		ctx       context.Context
		owner     model.ID
		belongsTo model.ID
		role      *model.Role
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, role *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Create(ctx, owner, belongsTo, role).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role:      testModel.NewRole(),
			},
		},
		{
			name: "create new role with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, role *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Create(ctx, owner, belongsTo, role).Return(assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role:      testModel.NewRole(),
			},
			wantErr: assert.AnError,
		},
		{
			name: "create new role license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role:      testModel.NewRole(),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create new role invalid role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role:      &model.Role{},
			},
			wantErr: ErrRoleCreate,
		},
		{
			name: "create new role quota exceeded",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role:      testModel.NewRole(),
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "create new role with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role:      testModel.NewRole(),
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.role),
			}
			err := s.Create(tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.role)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *model.Role) *baseService
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Role
		wantErr error
	}{
		{
			name: "get role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, id, belongsTo).Return(role, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindRead).Return(true)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: testModel.NewRole(),
		},
		{
			name: "get role with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, id, belongsTo).Return(role, assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindRead).Return(true)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.ID{},
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrRoleGet,
		},
		{
			name: "get role with no role permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindRead).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get role with no related permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, _ *model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindRead).Return(true)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.want),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleService_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseService
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		offset    int
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Role
		wantErr error
	}{
		{
			name: "get roles belongs to",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetAllBelongsTo", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(roles, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				offset:    0,
				limit:     10,
			},
			want: []*model.Role{
				testModel.NewRole(),
				testModel.NewRole(),
			},
		},
		{
			name: "get roles belongs to with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetAllBelongsTo", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(roles, assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				offset:    0,
				limit:     10,
			},
			wantErr: assert.AnError,
		},
		{
			name: "get roles belongs to with invalid role id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetAllBelongsTo", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.ID{},
				offset:    0,
				limit:     10,
			},
			wantErr: ErrRoleGetBelongsTo,
		},
		{
			name: "get roles belongs to with no permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, _ int, _ []*model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetAllBelongsTo", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				offset:    0,
				limit:     10,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get roles belongs to with invalid pagination offset",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetAllBelongsTo", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				offset:    -1,
				limit:     10,
			},
			wantErr: ErrInvalidPaginationParams,
		},
		{
			name: "get roles belongs to with invalid pagination limit",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*model.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetAllBelongsTo", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				offset:    0,
				limit:     0,
			},
			wantErr: ErrInvalidPaginationParams,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := s.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleService_Update(t *testing.T) {
	type fields struct {
		baseService *baseService
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
		patch     map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Role
		wantErr error
	}{
		/*{
			name: "update role",
		},
		{
			name:    "update role with error",
			wantErr: assert.AnError,
		},
		{
			name:    "update role with expired license",
			wantErr: license.ErrLicenseExpired,
		},
		{
			name:    "update role with invalid role id",
			wantErr: ErrRoleUpdate,
		},
		{
			name:    "update role with no permissions",
			wantErr: ErrNoPermission,
		},*/
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService,
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.args.patch)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleService_GetMembers(t *testing.T) {
	type fields struct {
		baseService *baseService
	}
	type args struct {
		ctx       context.Context
		roleID    model.ID
		belongsTo model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.User
		wantErr error
	}{}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService,
			}
			got, err := s.GetMembers(tt.args.ctx, tt.args.roleID, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleService_AddMember(t *testing.T) {
	type fields struct {
		baseService *baseService
	}
	type args struct {
		ctx       context.Context
		roleID    model.ID
		belongsTo model.ID
		memberID  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		/*{
			name: "add member to role",
		},
		{
			name:    "add member to role with error",
			wantErr: assert.AnError,
		},
		{
			name:    "add member to role with expired license",
			wantErr: license.ErrLicenseExpired,
		},
		{
			name:    "add member to role with invalid member id",
			wantErr: ErrRoleAddMember,
		},
		{
			name:    "add member to role with invalid role id",
			wantErr: ErrRoleAddMember,
		},
		{
			name:    "add member to role with no permissions",
			wantErr: ErrNoPermission,
		},*/
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService,
			}
			err := s.AddMember(tt.args.ctx, tt.args.roleID, tt.args.memberID, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_RemoveMember(t *testing.T) {
	type fields struct {
		baseService *baseService
	}
	type args struct {
		ctx       context.Context
		roleID    model.ID
		belongsTo model.ID
		memberID  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		/*{
			name: "remove member from role",
		},
		{
			name:    "remove member from role with error",
			wantErr: assert.AnError,
		},
		{
			name:    "remove member from role with expired license",
			wantErr: license.ErrLicenseExpired,
		},
		{
			name:    "remove member from role with invalid member id",
			wantErr: ErrRoleRemoveMember,
		},
		{
			name:    "remove member from role with invalid role id",
			wantErr: ErrRoleRemoveMember,
		},
		{
			name:    "remove member from role with no permissions",
			wantErr: ErrNoPermission,
		},*/
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService,
			}
			err := s.RemoveMember(tt.args.ctx, tt.args.roleID, tt.args.memberID, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_Delete(t *testing.T) {
	type fields struct {
		baseService *baseService
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		/*{
			name: "delete role",
		},
		{
			name:    "delete role with error",
			wantErr: assert.AnError,
		},
		{
			name:    "delete role with expired license",
			wantErr: license.ErrLicenseExpired,
		},
		{
			name:    "delete role with invalid role id",
			wantErr: ErrRoleUpdate,
		},
		{
			name:    "delete role with no permissions",
			wantErr: ErrNoPermission,
		},*/
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService,
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_AddPermission(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, targetID model.ID, kind model.PermissionKind) *baseService
	}
	type args struct {
		ctx         context.Context
		roleID      model.ID
		belongsToID model.ID
		targetID    model.ID
		kind        model.PermissionKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add permission to role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, _ model.ID, _ model.PermissionKind) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/AddPermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().Create(ctx, gomock.Any()).Return(nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
				targetID:    model.MustNewID(model.ResourceTypeDocument),
				kind:        model.PermissionKindRead,
			},
		},
		{
			name: "add permission with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _, _ model.ID, _ model.PermissionKind) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/AddPermission", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
				targetID:    model.MustNewID(model.ResourceTypeDocument),
				kind:        model.PermissionKindRead,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "add permission with invalid role ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _, _ model.ID, _ model.PermissionKind) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/AddPermission", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewNilID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
				targetID:    model.MustNewID(model.ResourceTypeDocument),
				kind:        model.PermissionKindRead,
			},
			wantErr: ErrRoleAddPermission,
		},
		{
			name: "add permission with no write permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, _ model.ID, _ model.PermissionKind) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/AddPermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
				targetID:    model.MustNewID(model.ResourceTypeDocument),
				kind:        model.PermissionKindRead,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "add permission with role not found",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, _ model.ID, _ model.PermissionKind) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/AddPermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(nil, assert.AnError)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
				targetID:    model.MustNewID(model.ResourceTypeDocument),
				kind:        model.PermissionKindRead,
			},
			wantErr: ErrRoleAddPermission,
		},
		{
			name: "add permission with permission service error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, _ model.ID, _ model.PermissionKind) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/AddPermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().Create(ctx, gomock.Any()).Return(assert.AnError)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
				targetID:    model.MustNewID(model.ResourceTypeDocument),
				kind:        model.PermissionKindRead,
			},
			wantErr: ErrRoleAddPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.roleID, tt.args.belongsToID, tt.args.targetID, tt.args.kind),
			}
			err := s.AddPermission(tt.args.ctx, tt.args.roleID, tt.args.belongsToID, tt.args.targetID, tt.args.kind)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_RemovePermission(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, permissionID model.ID) *baseService
	}
	type args struct {
		ctx          context.Context
		roleID       model.ID
		belongsToID  model.ID
		permissionID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove permission from role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, permissionID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/RemovePermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite).Return(true)
					perm := testModel.NewPermission(roleID, model.MustNewID(model.ResourceTypeDocument), model.PermissionKindRead)
					permSvc.EXPECT().Get(ctx, permissionID).Return(perm, nil)
					permSvc.EXPECT().Delete(ctx, permissionID).Return(nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:       model.MustNewID(model.ResourceTypeRole),
				belongsToID:  model.MustNewID(model.ResourceTypeOrganization),
				permissionID: model.MustNewID(model.ResourceTypePermission),
			},
		},
		{
			name: "remove permission with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/RemovePermission", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:       model.MustNewID(model.ResourceTypeRole),
				belongsToID:  model.MustNewID(model.ResourceTypeOrganization),
				permissionID: model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "remove permission with invalid permission ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/RemovePermission", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:       model.MustNewID(model.ResourceTypeRole),
				belongsToID:  model.MustNewID(model.ResourceTypeOrganization),
				permissionID: model.MustNewNilID(model.ResourceTypePermission),
			},
			wantErr: ErrRoleRemovePermission,
		},
		{
			name: "remove permission with no write permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/RemovePermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:       model.MustNewID(model.ResourceTypeRole),
				belongsToID:  model.MustNewID(model.ResourceTypeOrganization),
				permissionID: model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "remove permission with permission not belonging to role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID, permissionID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/RemovePermission", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite).Return(true)
					// Permission belongs to different role
					perm := testModel.NewPermission(model.MustNewID(model.ResourceTypeRole), model.MustNewID(model.ResourceTypeDocument), model.PermissionKindRead)
					permSvc.EXPECT().Get(ctx, permissionID).Return(perm, nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:       model.MustNewID(model.ResourceTypeRole),
				belongsToID:  model.MustNewID(model.ResourceTypeOrganization),
				permissionID: model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.roleID, tt.args.belongsToID, tt.args.permissionID),
			}
			err := s.RemovePermission(tt.args.ctx, tt.args.roleID, tt.args.belongsToID, tt.args.permissionID)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_GetPermissions(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID model.ID) *baseService
	}
	type args struct {
		ctx         context.Context
		roleID      model.ID
		belongsToID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permissions for role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetPermissions", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindRead).Return(true)
					perms := []*model.Permission{
						testModel.NewPermission(roleID, model.MustNewID(model.ResourceTypeDocument), model.PermissionKindRead),
					}
					permSvc.EXPECT().GetBySubject(ctx, roleID).Return(perms, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Permission{
				testModel.NewPermission(model.MustNewID(model.ResourceTypeRole), model.MustNewID(model.ResourceTypeDocument), model.PermissionKindRead),
			},
		},
		{
			name: "get permissions with invalid role ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetPermissions", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          mock.NewRoleRepository(ctrl),
						userRepo:          new(mock.UserRepository),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewNilID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrRoleGetPermissions,
		},
		{
			name: "get permissions with no read permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, roleID, belongsToID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/GetPermissions", gomock.Len(0)).Return(ctx, span)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, belongsToID).Return(testModel.NewRole(), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, belongsToID, model.PermissionKindRead).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          new(mock.UserRepository),
						permissionService: permSvc,
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roleID:      model.MustNewID(model.ResourceTypeRole),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.roleID, tt.args.belongsToID),
			}
			got, err := s.GetPermissions(tt.args.ctx, tt.args.roleID, tt.args.belongsToID)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
