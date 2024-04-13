package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
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
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithRoleRepository(new(mock.RoleRepository)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			want: &roleService{
				baseService: &baseService{
					logger:            new(mock.Logger),
					tracer:            new(mock.Tracer),
					roleRepo:          new(mock.RoleRepository),
					userRepo:          new(mock.UserRepository),
					permissionService: new(mock.PermissionService),
					licenseService:    new(mock.LicenseService),
				},
			},
		},
		{
			name: "new role service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(new(mock.Tracer)),
					WithRoleRepository(new(mock.RoleRepository)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new role service with no role repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoRoleRepository,
		},
		{
			name: "new role service with no user repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithRoleRepository(new(mock.RoleRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoUserRepository,
		},
		{
			name: "new role service with no permission service",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithRoleRepository(new(mock.RoleRepository)),
					WithUserRepository(new(mock.UserRepository)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new role service with no license service",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithRoleRepository(new(mock.RoleRepository)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(new(mock.PermissionService)),
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
		baseService func(ctx context.Context, owner, belongsTo model.ID, role *model.Role) *baseService
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
				baseService: func(ctx context.Context, owner, belongsTo model.ID, role *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					roleRepo := new(mock.RoleRepository)
					roleRepo.On("Create", ctx, owner, belongsTo, role).Return(nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, owner, belongsTo model.ID, role *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					roleRepo := new(mock.RoleRepository)
					roleRepo.On("Create", ctx, owner, belongsTo, role).Return(assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
						userRepo:          new(mock.UserRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
						userRepo:          new(mock.UserRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, _, belongsTo model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaRoles).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
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
				baseService: func(ctx context.Context, _, belongsTo model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
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
			s := &roleService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.role),
			}
			err := s.Create(tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.role)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, id, belongsTo model.ID, role *model.Role) *baseService
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
				baseService: func(ctx context.Context, id, belongsTo model.ID, role *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					roleRepo := new(mock.RoleRepository)
					roleRepo.On("Get", ctx, id, belongsTo).Return(role, nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id, belongsTo model.ID, role *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					roleRepo := new(mock.RoleRepository)
					roleRepo.On("Get", ctx, id, belongsTo).Return(role, assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
						userRepo:          new(mock.UserRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, id, belongsTo model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(false, nil)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
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
				baseService: func(ctx context.Context, id, belongsTo model.ID, _ *model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
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
			s := &roleService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.want),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleService_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseService
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
				baseService: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/GetAllBelongsTo", []trace.SpanStartOption(nil)).Return(ctx, span)

					roleRepo := new(mock.RoleRepository)
					roleRepo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(roles, nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/GetAllBelongsTo", []trace.SpanStartOption(nil)).Return(ctx, span)

					roleRepo := new(mock.RoleRepository)
					roleRepo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(roles, assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/GetAllBelongsTo", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
						userRepo:          new(mock.UserRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, belongsTo model.ID, _, _ int, _ []*model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/GetAllBelongsTo", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, belongsTo, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
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
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/GetAllBelongsTo", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
						userRepo:          new(mock.UserRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Role) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.roleService/GetAllBelongsTo", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						roleRepo:          new(mock.RoleRepository),
						userRepo:          new(mock.UserRepository),
						permissionService: new(mock.PermissionService),
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
			s := &roleService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
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
		},
		{
			name:    "update role with empty patch data",
			wantErr: ErrNoPatchData,
		},*/
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
			s := &roleService{
				baseService: tt.fields.baseService,
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
