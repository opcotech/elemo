package service

import (
	"context"
	"slices"
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

					userID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().GetAll(ctx, userID, offset, limit).Return(organizations, nil)

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

					userID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().GetAll(ctx, userID, offset, limit).Return(nil, assert.AnError)

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
		baseService func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, members []*model.OrganizationMember, expected []*model.OrganizationMember) *baseService
	}
	type args struct {
		ctx            context.Context
		organizationID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.OrganizationMember
		wantErr error
	}{
		{
			name: "get members of organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, members []*model.OrganizationMember, expected []*model.OrganizationMember) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().GetMembers(ctx, organizationID).Return(members, nil)

					permissionService := mock.NewPermissionService(ctrl)
					// Mock permission fetching for each member
					// Create a map of member ID to expected roles for easier lookup
					expectedRolesMap := make(map[model.ID][]string)
					for _, expectedMember := range expected {
						expectedRolesMap[expectedMember.ID] = expectedMember.Roles
					}

					// Set up permissions for each member in the repository
					for _, member := range members {
						permissions := []*model.Permission{}
						// Set up permissions based on expected roles
						if expectedRoles, ok := expectedRolesMap[member.ID]; ok {
							// Check if virtual roles are present in expected roles
							hasOwnerRole := false
							hasAdminRole := false
							hasMemberRole := false
							for _, role := range expectedRoles {
								switch role {
								case "Owner":
									hasOwnerRole = true
								case "Admin":
									hasAdminRole = true
								case "Member":
									hasMemberRole = true
								}
							}
							// Set up permissions to match expected virtual roles
							// Priority: owner > admin > member
							switch {
							case hasOwnerRole:
								// Owner: needs PermissionKindAll OR (Read + Write + Delete)
								permissions = append(permissions, &model.Permission{
									Kind: model.PermissionKindAll,
								})
							case hasAdminRole:
								// Admin: needs Write permission
								permissions = append(permissions, &model.Permission{
									Kind: model.PermissionKindWrite,
								})
							case hasMemberRole:
								// Member: needs ONLY Read permission (no Write, no Delete)
								permissions = append(permissions, &model.Permission{
									Kind: model.PermissionKindRead,
								})
							}
						}
						permissionService.EXPECT().GetBySubjectAndTarget(ctx, member.ID, organizationID).Return(permissions, nil)
					}

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
			want: func() []*model.OrganizationMember {
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
				expected1, _ := model.NewOrganizationMember(user1.ID, user1.FirstName, user1.LastName, user1.Email, picture1, user1.Status, []string{"Owner"})
				// User2: has "Member" role -> should get "Admin" virtual role (since write permission)
				expected2, _ := model.NewOrganizationMember(user2.ID, user2.FirstName, user2.LastName, user2.Email, picture2, user2.Status, []string{"Admin", "Member"})
				// User3: has "Admin", "Member" roles -> should get "Admin" virtual role (deduplicated)
				expected3, _ := model.NewOrganizationMember(user3.ID, user3.FirstName, user3.LastName, user3.Email, picture3, user3.Status, []string{"Admin", "Member"})
				// User4: has no roles -> should get "Member" virtual role (since read permission)
				expected4, _ := model.NewOrganizationMember(user4.ID, user4.FirstName, user4.LastName, user4.Email, picture4, user4.Status, []string{"Member"})

				return []*model.OrganizationMember{expected1, expected2, expected3, expected4}
			}(),
		},
		{
			name: "get members of organization with invalid organization id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ []*model.OrganizationMember, _ []*model.OrganizationMember) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: mock.NewOrganizationRepository(ctrl),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organizationID model.ID, _ []*model.OrganizationMember, _ []*model.OrganizationMember) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/GetMembers", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().GetMembers(ctx, organizationID).Return(nil, assert.AnError)

					return &baseService{
						logger:           mock.NewMockLogger(ctrl),
						tracer:           tracer,
						organizationRepo: organizationRepo,
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
				hasAdminVirtual := slices.Contains(roles, "Admin")

				actualRoles := make([]string, 0)
				for _, role := range roles {
					// Owner and Admin are always virtual
					if role == "Owner" || role == "Admin" {
						continue
					}
					// Member is actual only if Admin virtual role is also present
					if role == "Member" {
						if hasAdminVirtual {
							actualRoles = append(actualRoles, role)
						}
						continue
					}
					// All other roles are actual
					actualRoles = append(actualRoles, role)
				}
				return actualRoles
			}

			// Prepare members from repository (without virtual roles)
			var membersFromRepo []*model.OrganizationMember
			if len(tt.want) > 0 {
				membersFromRepo = make([]*model.OrganizationMember, len(tt.want))
				for i, expected := range tt.want {
					actualRoles := extractActualRoles(expected.Roles)
					member, err := model.NewOrganizationMember(
						expected.ID,
						expected.FirstName,
						expected.LastName,
						expected.Email,
						expected.Picture,
						expected.Status,
						actualRoles,
					)
					require.NoError(t, err, "failed to create OrganizationMember for test")
					require.NotZero(t, member.ID, "member ID should not be zero")
					membersFromRepo[i] = member
				}
			}

			s := &organizationService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.organizationID, membersFromRepo, tt.want),
			}
			members, err := s.GetMembers(tt.args.ctx, tt.args.organizationID)
			require.ErrorIs(t, err, tt.wantErr)

			if err == nil {
				require.Equal(t, len(tt.want), len(members))

				// Build lookup map for expected members
				expectedMap := make(map[model.ID]*model.OrganizationMember, len(tt.want))
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
