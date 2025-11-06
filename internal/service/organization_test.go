package service

import (
	"context"
	"slices"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
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
					WithUserTokenRepository(mock.NewUserTokenRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
					WithEmailService(mock.NewEmailService(nil)),
				},
			},
			want: &organizationService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					userRepo:          mock.NewUserRepository(nil),
					organizationRepo:  mock.NewOrganizationRepository(nil),
					userTokenRepo:     mock.NewUserTokenRepository(nil),
					permissionService: mock.NewPermissionService(nil),
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
					WithUserRepository(mock.NewUserRepository(nil)),
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithUserTokenRepository(mock.NewUserTokenRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
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
					WithUserTokenRepository(mock.NewUserTokenRepository(nil)),
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
					WithUserRepository(mock.NewUserRepository(nil)),
					WithOrganizationRepository(mock.NewOrganizationRepository(nil)),
					WithUserTokenRepository(mock.NewUserTokenRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Organization) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)

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
					permSvc.EXPECT().CtxUserHasPermission(ctx, organization, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().Create(ctx, gomock.Any()).Return(nil)

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
					// Mock permission check for the context user
					permissionService.EXPECT().CtxUserHasPermission(ctx, organizationID, model.PermissionKindRead).Return(true)
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

					permissionService := mock.NewPermissionService(ctrl)
					permissionService.EXPECT().CtxUserHasPermission(ctx, organizationID, model.PermissionKindRead).Return(true)

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
			name: "remove member from organization",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RemoveMember", gomock.Len(0)).Return(ctx, span)

					organizationRepo := mock.NewOrganizationRepository(ctrl)
					organizationRepo.EXPECT().RemoveMember(ctx, organization, userID).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, organization, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().GetBySubjectAndTarget(ctx, userID, organization).Return([]*model.Permission{}, nil)

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
					permSvc.EXPECT().CtxUserHasPermission(ctx, organization, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().GetBySubjectAndTarget(ctx, userID, organization).Return([]*model.Permission{}, nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewOrganization()
					organization.ID = orgID

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().HasPermission(ctx, user.ID, orgID, model.PermissionKindRead).Return(false, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, organization, user, gomock.Any()).Return(nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					// Use an email that will generate both firstName and lastName
					testEmail := "john.doe@example.com"

					user := testModel.NewUser()
					user.Email = testEmail
					user.Status = model.UserStatusPending

					organization := testModel.NewOrganization()
					organization.ID = orgID

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, testEmail).Return(nil, repository.ErrNotFound)
					userRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, u *model.User) error {
						user.ID = u.ID
						user.Status = model.UserStatusPending
						user.FirstName = u.FirstName
						user.LastName = u.LastName
						user.Email = u.Email
						return nil
					})

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, gomock.Any()).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().HasPermission(ctx, gomock.Any(), orgID, model.PermissionKindRead).Return(false, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, gomock.Any(), model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, organization, gomock.Any(), gomock.Any()).Return(nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, email string, roleID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.Email = email
					user.Status = model.UserStatusActive

					organization := testModel.NewOrganization()
					organization.ID = orgID

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().HasPermission(ctx, user.ID, orgID, model.PermissionKindRead).Return(false, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, organization, user, gomock.Any()).Return(nil)

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, invalidOrgID, model.PermissionKindWrite).Return(false).AnyTimes()

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
			wantErr: ErrInvalidEmail,
		},
		{
			name: "invite member with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/InviteMember", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(false)

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

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().HasPermission(ctx, user.ID, orgID, model.PermissionKindRead).Return(true, nil)

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)
					// HasPermission is not called when user status is invalid - code returns early

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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

					organization := testModel.NewOrganization()
					organization.ID = orgID

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddInvitation(ctx, orgID, user.ID).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)
					permSvc.EXPECT().HasPermission(ctx, user.ID, orgID, model.PermissionKindRead).Return(false, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, user.ID, model.UserTokenContextInvite).Return(nil, repository.ErrNotFound)
					userTokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

					emailService := mock.NewEmailService(ctrl)
					emailService.EXPECT().SendOrganizationInvitationEmail(ctx, organization, user, gomock.Any()).Return(assert.AnError)

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
			err := s.InviteMember(tt.args.ctx, tt.args.orgID, tt.args.email, tt.args.roleID...)
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

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					// GetAll is only called for pending users, not active users

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)

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
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
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
					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, invalidOrgID, model.PermissionKindWrite).Return(false).AnyTimes()

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
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:  model.MustNewNilID(model.ResourceTypeOrganization),
				userID: userID,
			},
			wantErr: ErrOrganizationInviteRevoke,
		},
		{
			name: "revoke invitation with invalid userID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, invalidUserID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/RevokeInvitation", gomock.Len(0)).Return(ctx, span)

					// Permission check happens after userID validation, but if validation passes (nil ID might pass),
					// we need to expect the permission call
					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(false).AnyTimes()

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

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(false)

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

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(nil, repository.ErrNotFound)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						organizationRepo:  mock.NewOrganizationRepository(ctrl),
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

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)
					userRepo.EXPECT().Delete(ctx, userID).Return(nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().GetAll(ctx, userID, 0, 1).Return([]*model.Organization{}, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)

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

					otherOrg := testModel.NewOrganization()

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().RemoveMember(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().GetAll(ctx, userID, 0, 1).Return([]*model.Organization{otherOrg}, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite).Return(true)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, userPassword string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusPending

					organization := testModel.NewOrganization()
					organization.ID = orgID

					// Extract secret from the public token passed in
					// The token parameter contains the public token, we need to extract the secret from it
					_, secret, _ := auth.SplitToken(token)
					// Hash the secret to match what's stored in userToken
					secretToken := auth.HashPassword(secret)

					userToken, _ := model.NewUserToken(userID, user.Email, secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)
					userRepo.EXPECT().Update(ctx, userID, gomock.Any()).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().Create(ctx, gomock.Any()).Return(nil)

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

					organization := testModel.NewOrganization()
					organization.ID = orgID

					// Extract secret from the public token passed in
					// The token parameter contains the public token, we need to extract the secret from it
					_, secret, _ := auth.SplitToken(token)
					// Hash the secret to match what's stored in userToken
					secretToken := auth.HashPassword(secret)

					userToken, _ := model.NewUserToken(userID, user.Email, secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().Create(ctx, gomock.Any()).Return(nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, roleID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					user := testModel.NewUser()
					user.ID = userID
					user.Status = model.UserStatusActive

					organization := testModel.NewOrganization()
					organization.ID = orgID

					role := testModel.NewRole()
					role.ID = roleID

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken, _ := model.NewUserToken(userID, user.Email, secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)
					orgRepo.EXPECT().AddMember(ctx, orgID, userID).Return(nil)

					roleRepo := mock.NewRoleRepository(ctrl)
					roleRepo.EXPECT().Get(ctx, roleID, orgID).Return(role, nil)
					roleRepo.EXPECT().AddMember(ctx, roleID, userID, orgID).Return(nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)
					userTokenRepo.EXPECT().Delete(ctx, userID, model.UserTokenContextInvite).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().Create(ctx, gomock.Any()).Return(nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID, _ string, _ string, _ model.ID) *baseService {
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID, _ string, _ string, _ model.ID) *baseService {
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken, _ := model.NewUserToken(userID, "test@example.com", secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now
					// Set CreatedAt to be older than deadline
					oldTime := time.Now().Add(-8 * 24 * time.Hour)
					userToken.CreatedAt = &oldTime

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
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

					userToken, _ := model.NewUserToken(userID, "test@example.com", secretToken, model.UserTokenContextInvite)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken, _ := model.NewUserToken(userID, "test@example.com", secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(nil, repository.ErrNotFound)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
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

					userToken, _ := model.NewUserToken(userID, user.Email, secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
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

					userToken, _ := model.NewUserToken(userID, user.Email, secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
					userTokenRepo.EXPECT().Get(ctx, userID, model.UserTokenContextInvite).Return(userToken, nil)

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID, token string, _ string, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
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

					organization := testModel.NewOrganization()
					organization.ID = orgID
					organization.Members = []model.ID{userID}

					// Extract secret from the public token passed in
					_, secret, _ := auth.SplitToken(token)
					secretToken := auth.HashPassword(secret)

					userToken, _ := model.NewUserToken(userID, user.Email, secretToken, model.UserTokenContextInvite)
					now := time.Now()
					userToken.CreatedAt = &now

					userRepo := mock.NewUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, userID).Return(user, nil)

					orgRepo := mock.NewOrganizationRepository(ctrl)
					orgRepo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					orgRepo.EXPECT().Get(ctx, orgID).Return(organization, nil)

					userTokenRepo := mock.NewUserTokenRepository(ctrl)
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
						permissionService: mock.NewPermissionService(ctrl),
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
			err := s.AcceptInvitation(tt.args.ctx, tt.args.orgID, tt.args.token, tt.args.userPassword)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
