package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &roleService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					roleRepo:          repository.NewMockRoleRepository(nil),
					userRepo:          repository.NewMockUserRepository(nil),
					permissionService: NewMockPermissionService(nil),
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
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
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
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
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
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
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
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
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
					WithRoleRepository(repository.NewMockRoleRepository(nil)),
					WithUserRepository(repository.NewMockUserRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, opts CreateRoleOpts) *baseService
	}
	type args struct {
		ctx       context.Context
		owner     model.ID
		belongsTo model.ID
		opts      CreateRoleOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, opts CreateRoleOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Create(ctx, repository.CreateRoleOpts{
						Key: opts.Key, Name: opts.Name, Description: opts.Description, Actions: model.ActionStrings(opts.Actions), CreatedBy: owner, BelongsTo: belongsTo,
					}).Return(testModel.NewRepositoryRole(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
		},
		{
			name: "create new role with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, owner, belongsTo model.ID, opts CreateRoleOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Create(ctx, repository.CreateRoleOpts{
						Key: opts.Key, Name: opts.Name, Description: opts.Description, Actions: model.ActionStrings(opts.Actions), CreatedBy: owner, BelongsTo: belongsTo,
					}).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: assert.AnError,
		},
		{
			name: "create new role license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ CreateRoleOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create new role invalid role",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ CreateRoleOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts:      CreateRoleOpts{},
			},
			wantErr: ErrRoleCreate,
		},
		{
			name: "create new role quota exceeded",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ CreateRoleOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaRoles).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "create new role with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ CreateRoleOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:     userID,
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: CreateRoleOpts{
					Name:        "test-role",
					Description: "test description",
				},
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
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.opts),
			}
			_, err := s.Create(tt.args.ctx, tt.args.owner, tt.args.belongsTo, tt.args.opts)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *repository.Role) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, id, belongsTo, repository.RoleDetailProjection()).Return(role, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, id, belongsTo, repository.RoleDetailProjection()).Return(role, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          repository.NewMockUserRepository(nil),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: NewMockPermissionService(ctrl),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ *repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, belongsTo model.ID, _ *repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(nil),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
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
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.repoRole),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, roleFromRepository(tt.repoRole), got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestRoleService_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, page CursorPage, roles []*repository.Role) *baseService
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		page      CursorPage
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, page CursorPage, roles []*repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().ListBelongsTo(ctx, belongsTo, page, repository.RoleListProjection()).Return(repository.Page[*repository.Role]{Items: roles}, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      CursorPage{Size: 10},
			},
			repoRoles: []*repository.Role{
				testModel.NewRepositoryRole(),
				testModel.NewRepositoryRole(),
			},
		},
		{
			name: "get roles belongs to with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, page CursorPage, _ []*repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					roleRepo := repository.NewMockRoleRepository(ctrl)
					roleRepo.EXPECT().ListBelongsTo(ctx, belongsTo, page, repository.RoleListProjection()).Return(repository.Page[*repository.Role]{}, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          roleRepo,
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      CursorPage{Size: 10},
			},
			wantErr: assert.AnError,
		},
		{
			name: "get roles belongs to with invalid role id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ CursorPage, _ []*repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: NewMockPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.ID{},
				page:      CursorPage{Size: 10},
			},
			wantErr: ErrRoleGetBelongsTo,
		},
		{
			name: "get roles belongs to with no permissions",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ CursorPage, _ []*repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      CursorPage{Size: 10},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get roles belongs to with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ CursorPage, _ []*repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: NewMockPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      CursorPage{Size: -1},
			},
			wantErr: ErrRoleGetBelongsTo,
		},
		{
			name: "get roles belongs to with oversized page",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ CursorPage, _ []*repository.Role) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.roleService/ListBelongsTo", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						roleRepo:          repository.NewMockRoleRepository(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						permissionService: NewMockPermissionService(ctrl),
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				page:      CursorPage{Size: repository.MaxPageSize + 1},
			},
			wantErr: ErrRoleGetBelongsTo,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &roleService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.page, tt.repoRoles),
			}
			got, err := s.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.page)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, rolesFromRepository(tt.repoRoles), got.Items)
			} else {
				assert.Empty(t, got.Items)
			}
		})
	}
}

//nolint:revive // test factories take gomock.Controller first
func newRoleServiceTestBase(ctrl *gomock.Controller, ctx context.Context, spanName string) (*baseService, *repository.MockRoleRepository, *MockPermissionService, *mock.MockLicenseService) {
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, spanName, gomock.Len(0)).Return(ctx, span)

	roleRepo := repository.NewMockRoleRepository(ctrl)
	permSvc := NewMockPermissionService(ctrl)
	licenseSvc := mock.NewMockLicenseService(ctrl)

	return &baseService{
		logger:            mock.NewMockLogger(ctrl),
		tracer:            tracer,
		roleRepo:          roleRepo,
		userRepo:          repository.NewMockUserRepository(ctrl),
		permissionService: permSvc,
		licenseService:    licenseSvc,
	}, roleRepo, permSvc, licenseSvc
}

func TestRoleService_Update(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	repoRole := testModel.NewRepositoryRole()
	repoRole.Name = "updated-role"
	ctx := context.Background()
	opts := UpdateRoleOpts{Name: optional.Some("updated-role")}

	t.Run("update role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true)
		roleRepo.EXPECT().Update(ctx, roleID, orgID, repository.UpdateRoleOpts{Name: opts.Name}).Return(repoRole, nil)

		got, err := (&roleService{baseService: base}).Update(ctx, roleID, orgID, opts)
		require.NoError(t, err)
		assert.Equal(t, "updated-role", got.Name)
	})

	t.Run("update role clears actions", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Update")
		cleared := testModel.NewRepositoryRole()
		cleared.Actions = []string{}
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true)
		roleRepo.EXPECT().Update(ctx, roleID, orgID, repository.UpdateRoleOpts{
			Actions: optional.Some([]string{}),
		}).Return(cleared, nil)

		got, err := (&roleService{baseService: base}).Update(ctx, roleID, orgID, UpdateRoleOpts{
			Actions: optional.Null[[]model.Action](),
		})
		require.NoError(t, err)
		assert.Empty(t, got.Actions)
	})

	t.Run("update role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true)
		roleRepo.EXPECT().Update(ctx, roleID, orgID, repository.UpdateRoleOpts{Name: opts.Name}).Return(nil, assert.AnError)

		_, err := (&roleService{baseService: base}).Update(ctx, roleID, orgID, opts)
		require.ErrorIs(t, err, ErrRoleUpdate)
	})

	t.Run("update role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		_, err := (&roleService{baseService: base}).Update(ctx, roleID, orgID, opts)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("update role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		_, err := (&roleService{baseService: base}).Update(ctx, model.ID{}, orgID, opts)
		require.ErrorIs(t, err, ErrRoleUpdate)
	})

	t.Run("update role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(false)

		_, err := (&roleService{baseService: base}).Update(ctx, roleID, orgID, opts)
		require.ErrorIs(t, err, ErrNoPermission)
	})
}

func TestRoleService_ListMembers(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	repoUser := testModel.NewRepositoryUser()
	page := CursorPage{Size: 10}
	ctx := context.Background()

	t.Run("list role members", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, _ := newRoleServiceTestBase(ctrl, ctx, "service.roleService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		roleRepo.EXPECT().ListMembers(ctx, roleID, orgID, page).Return(repository.Page[*repository.User]{Items: []*repository.User{repoUser}}, nil)

		got, err := (&roleService{baseService: base}).ListMembers(ctx, roleID, orgID, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoUser.ID, got.Items[0].ID)
	})

	t.Run("list role members with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, _ := newRoleServiceTestBase(ctrl, ctx, "service.roleService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		roleRepo.EXPECT().ListMembers(ctx, roleID, orgID, page).Return(repository.Page[*repository.User]{}, assert.AnError)

		_, err := (&roleService{baseService: base}).ListMembers(ctx, roleID, orgID, page)
		require.ErrorIs(t, err, ErrOrganizationMembersGet)
	})

	t.Run("list role members with invalid belongs-to id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, _ := newRoleServiceTestBase(ctrl, ctx, "service.roleService/ListMembers")

		_, err := (&roleService{baseService: base}).ListMembers(ctx, roleID, model.ID{}, page)
		require.ErrorIs(t, err, ErrRoleGetBelongsTo)
	})

	t.Run("list role members with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, _ := newRoleServiceTestBase(ctrl, ctx, "service.roleService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		_, err := (&roleService{baseService: base}).ListMembers(ctx, roleID, orgID, page)
		require.ErrorIs(t, err, ErrNoPermission)
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
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		roleRepo.EXPECT().AddMember(ctx, roleID, memberID, orgID).Return(nil)

		require.NoError(t, (&roleService{baseService: base}).AddMember(ctx, roleID, memberID, orgID))
	})

	t.Run("add member to role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		roleRepo.EXPECT().AddMember(ctx, roleID, memberID, orgID).Return(assert.AnError)

		err := (&roleService{baseService: base}).AddMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, ErrRoleAddMember)
	})

	t.Run("add member to role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := (&roleService{baseService: base}).AddMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("add member to role with invalid member id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := (&roleService{baseService: base}).AddMember(ctx, roleID, model.ID{}, orgID)
		require.ErrorIs(t, err, ErrRoleAddMember)
	})

	t.Run("add member to role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := (&roleService{baseService: base}).AddMember(ctx, model.ID{}, memberID, orgID)
		require.ErrorIs(t, err, ErrRoleAddMember)
	})

	t.Run("add member to role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		err := (&roleService{baseService: base}).AddMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
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
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		roleRepo.EXPECT().RemoveMember(ctx, roleID, memberID, orgID).Return(nil)

		require.NoError(t, (&roleService{baseService: base}).RemoveMember(ctx, roleID, memberID, orgID))
	})

	t.Run("remove member from role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		roleRepo.EXPECT().RemoveMember(ctx, roleID, memberID, orgID).Return(assert.AnError)

		err := (&roleService{baseService: base}).RemoveMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, ErrRoleRemoveMember)
	})

	t.Run("remove member from role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := (&roleService{baseService: base}).RemoveMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("remove member from role with invalid member id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := (&roleService{baseService: base}).RemoveMember(ctx, roleID, model.ID{}, orgID)
		require.ErrorIs(t, err, ErrRoleRemoveMember)
	})

	t.Run("remove member from role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := (&roleService{baseService: base}).RemoveMember(ctx, model.ID{}, memberID, orgID)
		require.ErrorIs(t, err, ErrRoleRemoveMember)
	})

	t.Run("remove member from role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		err := (&roleService{baseService: base}).RemoveMember(ctx, roleID, memberID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
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
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true)
		roleRepo.EXPECT().Delete(ctx, roleID, orgID).Return(nil)

		require.NoError(t, (&roleService{baseService: base}).Delete(ctx, roleID, orgID))
	})

	t.Run("delete role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, roleRepo, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(true)
		roleRepo.EXPECT().Delete(ctx, roleID, orgID).Return(assert.AnError)

		err := (&roleService{baseService: base}).Delete(ctx, roleID, orgID)
		require.ErrorIs(t, err, ErrRoleDelete)
	})

	t.Run("delete role with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := (&roleService{baseService: base}).Delete(ctx, roleID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("delete role with invalid role id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		err := (&roleService{baseService: base}).Delete(ctx, model.ID{}, orgID)
		require.ErrorIs(t, err, ErrRoleDelete)
	})

	t.Run("delete role with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newRoleServiceTestBase(ctrl, ctx, "service.roleService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionRoleManage).Return(false)

		err := (&roleService{baseService: base}).Delete(ctx, roleID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
	})
}
