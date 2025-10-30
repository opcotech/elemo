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
	"github.com/stretchr/testify/require"
)

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
					WithUserRepository(mock.NewUserRepository(nil)),
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &organizationService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					userRepo:          mock.NewUserRepository(nil),
					organizationRepo:  mock.NewOrganizationRepository(nil),
					permissionService: mock.NewPermissionService(nil),
					licenseService:    mock.NewMockLicenseService(nil),
				},
			},
		},
		{
			name: "new organization service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
					WithUserRepository(mock.NewUserRepository(nil)),
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
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
					WithUserRepository(mock.NewUserRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
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
					WithUserRepository(mock.NewUserRepository(nil)),
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
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
					WithUserRepository(mock.NewUserRepository(nil)),
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
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
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, organization *model.Organization) *baseService
	}
	type args struct {
		ctx          context.Context
		owner        model.ID
		organization *model.Organization
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Create(ctx, userID, organization).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindCreate}).Return(true)

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
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
			},
		},
		{
			name: "create organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindCreate}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create organization with permission error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindCreate}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
						licenseService:   licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: &model.Organization{},
			},
			wantErr: ErrOrganizationCreate,
		},
		{
			name: "create organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Create(ctx, userID, organization).Return(assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindCreate}).Return(true)

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
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
			},
			wantErr: ErrOrganizationCreate,
		},
		{
			name: "create organization out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindCreate}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "create organization with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create organization with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:          context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				owner:        userID,
				organization: testModel.NewOrganization(),
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
			err := s.Create(tt.args.ctx, tt.args.owner, tt.args.organization)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Organization
		wantErr error
	}{
		{
			name: "get organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, id).Return(organization, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: testModel.NewOrganization(),
		},
		{
			name: "get organization with invalid organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Get", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, id).Return(nil, assert.AnError)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
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
			require.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationService_GetAll(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) *baseService
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
		want    []*model.Organization
		wantErr error
	}{
		{
			name: "get all organizations organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetAll", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().GetAll(ctx, offset, limit).Return(organizations, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*model.Organization{
				testModel.NewOrganization(),
				testModel.NewOrganization(),
			},
		},
		{
			name: "get all organizations with invalid offset",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetAll", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().GetAll(ctx, offset, limit).Return(nil, assert.AnError)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
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
			got, err := s.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	organizationID := model.MustNewID(model.ResourceTypeOrganization)
	otherOrganizationID := model.MustNewID(model.ResourceTypeOrganization)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		patch map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Organization
		wantErr error
	}{
		{
			name: "update organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, patch).Return(organization, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

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
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.OrganizationStatusActive.String(),
				},
			},
			want: testModel.NewOrganization(),
		},
		{
			name: "update organization with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false)

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
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update organization with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
						licenseService:   licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrOrganizationUpdate,
		},
		{
			name: "update organization with empty patch",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    organizationID,
				patch: map[string]any{},
			},
			wantErr: ErrOrganizationUpdate,
		},
		{
			name: "update organization with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, patch).Return(nil, assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

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
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrOrganizationUpdate,
		},
		{
			name: "update organization out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaOrganizations).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.OrganizationStatusActive.String(),
				},
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "update organization with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.OrganizationStatusActive.String(),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update organization with expired license error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  organizationID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.OrganizationStatusActive.String(),
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
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
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
					patch := map[string]any{
						"status": model.OrganizationStatusDeleted.String(),
					}

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, patch).Return(new(model.Organization), nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

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

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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

					organizationRepo := mock.NewOrganizationRepository(ctrl)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(false)

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

					organizationRepo := mock.NewOrganizationRepository(ctrl)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(false)

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
					patch := map[string]any{
						"status": model.OrganizationStatusDeleted.String(),
					}

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Delete", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Update(ctx, id, patch).Return(nil, assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

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

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

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

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().AddMember(ctx, organization, userID).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(true)

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

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().AddMember(ctx, organization, userID).Return(assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(true)

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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

func TestOrganizationService_GetMembers(t *testing.T) {
	type fields struct {
		baseService  func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService
		organization *model.Organization
	}
	type args struct {
		ctx            context.Context
		organizationID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.User
		wantErr error
	}{
		{
			name: "get members of organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					userRepo := mock.NewUserRepository(ctrl)
					for i, userID := range organization.Members {
						userRepo.EXPECT().Get(ctx, userID).Return(members[i], nil)
					}

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, organizationID).Return(organization, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
						userRepo:         userRepo,
					}
				},
				organization: &model.Organization{
					Members: []model.ID{
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
					},
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.User{
				testModel.NewUser(),
				testModel.NewUser(),
				testModel.NewUser(),
				testModel.NewUser(),
			},
		},
		{
			name: "get members of organization with invalid organization id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization, _ []*model.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
						userRepo:         mock.NewUserRepository(ctrl),
					}
				},
				organization: &model.Organization{
					Members: []model.ID{
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
					},
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.ID{},
			},
			wantErr: ErrOrganizationMembersGet,
		},
		{
			name: "get members of organization with organization get error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, _ *model.Organization, _ []*model.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, organizationID).Return(nil, assert.AnError)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
						userRepo:         mock.NewUserRepository(nil),
					}
				},
				organization: &model.Organization{
					Members: []model.ID{
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
					},
				},
			},
			args: args{
				ctx:            context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				organizationID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrOrganizationMembersGet,
		},
		{
			name: "get members of organization with user get error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, organization *model.Organization, _ []*model.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, organization.Members[0]).Return(nil, assert.AnError)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().Get(ctx, organizationID).Return(organization, nil)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
						userRepo:         userRepo,
					}
				},
				organization: &model.Organization{
					Members: []model.ID{
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
						model.MustNewID(model.ResourceTypeUser),
					},
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
			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organizationID, tt.fields.organization, tt.want),
			}
			members, err := s.GetMembers(tt.args.ctx, tt.args.organizationID)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, members)
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
			name: "add member to organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(true)

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

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(assert.AnError)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{model.PermissionKindWrite}).Return(true)

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
