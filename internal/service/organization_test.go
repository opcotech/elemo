package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
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
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithOrganizationRepository(new(mock.OrganizationRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			want: &organizationService{
				baseService: &baseService{
					logger:            new(mock.Logger),
					tracer:            new(mock.Tracer),
					userRepo:          new(mock.UserRepository),
					organizationRepo:  new(mock.OrganizationRepository),
					permissionService: new(mock.PermissionService),
					licenseService:    new(mock.LicenseService),
				},
			},
		},
		{
			name: "new organization service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithOrganizationRepository(new(mock.OrganizationRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new organization service with no organization repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoOrganizationRepository,
		},
		{
			name: "new organization service with no permission repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithOrganizationRepository(new(mock.OrganizationRepository)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new organization service with no license service",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithOrganizationRepository(new(mock.OrganizationRepository)),
					WithPermissionService(new(mock.PermissionService)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
		{
			name: "new organization service with no user repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithOrganizationRepository(new(mock.OrganizationRepository)),
					WithPermissionService(new(mock.PermissionService)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoUserRepository,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewOrganizationService(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctx context.Context, organization *model.Organization) *baseService
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Create", ctx, userID, organization).Return(nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindCreate,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindCreate,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindCreate,
					}).Return(false, assert.AnError)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Create", ctx, userID, organization).Return(assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindCreate,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindCreate,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.organization),
			}
			err := s.Create(tt.args.ctx, tt.args.owner, tt.args.organization)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, id model.ID, organization *model.Organization) *baseService
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
				baseService: func(ctx context.Context, id model.ID, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Get", ctx, id).Return(organization, nil)

					return &baseService{
						logger:           new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, id model.ID, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Get", ctx, id).Return(nil, assert.AnError)

					return &baseService{
						logger:           new(mock.Logger),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationService_GetAll(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseService
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
				baseService: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("GetAll", ctx, offset, limit).Return(organizations, nil)

					return &baseService{
						logger:           new(mock.Logger),
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
				baseService: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, limit, offset int, organizations []*model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, offset, limit int, organization []*model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("GetAll", ctx, offset, limit).Return(nil, assert.AnError)

					return &baseService{
						logger:           new(mock.Logger),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
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
		baseService func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Update", ctx, id, patch).Return(organization, nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Update", ctx, id, patch).Return(organization, nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Update", ctx, id, patch).Return(nil, assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, errors.New("test error"))

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
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
		baseService func(ctx context.Context, id model.ID) *baseService
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					patch := map[string]any{
						"status": model.OrganizationStatusDeleted.String(),
					}

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Update", ctx, id, patch).Return(new(model.Organization), nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Delete", ctx, id).Return(nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					patch := map[string]any{
						"status": model.OrganizationStatusDeleted.String(),
					}

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Update", ctx, id, patch).Return(new(model.Organization), nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Delete", ctx, id).Return(nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					patch := map[string]any{
						"status": model.OrganizationStatusDeleted.String(),
					}

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.organizationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Update", ctx, id, patch).Return(nil, assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Delete", ctx, id).Return(assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id),
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_AddMember(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctx context.Context, organization model.ID) *baseService
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("AddMember", ctx, organization, userID).Return(nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false, assert.AnError)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: permSvc,
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: permSvc,
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("AddMember", ctx, organization, userID).Return(assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/AddMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.organization),
			}
			err := s.AddMember(tt.args.ctx, tt.args.organization, tt.args.member)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationService_GetMembers(t *testing.T) {
	type fields struct {
		baseService  func(ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService
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
				baseService: func(ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetMembers", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					for i, userID := range organization.Members {
						userRepo.On("Get", ctx, userID).Return(members[i], nil)
					}

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Get", ctx, organizationID).Return(organization, nil)

					return &baseService{
						logger:           new(mock.Logger),
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
				baseService: func(ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetMembers", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: new(mock.OrganizationRepository),
						userRepo:         new(mock.UserRepository),
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
				baseService: func(ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetMembers", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Get", ctx, organizationID).Return(nil, assert.AnError)

					return &baseService{
						logger:           new(mock.Logger),
						tracer:           tracer,
						organizationRepo: organizationRepo,
						userRepo:         new(mock.UserRepository),
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
				baseService: func(ctx context.Context, organizationID model.ID, organization *model.Organization, members []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/GetMembers", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Get", ctx, organization.Members[0]).Return(nil, assert.AnError)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("Get", ctx, organizationID).Return(organization, nil)

					return &baseService{
						logger:           new(mock.Logger),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.organizationID, tt.fields.organization, tt.want),
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
		baseService func(ctx context.Context, organization model.ID) *baseService
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("RemoveMember", ctx, organization, userID).Return(nil)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false, assert.AnError)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: permSvc,
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: permSvc,
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					organizationRepo := new(mock.OrganizationRepository)
					organizationRepo.On("RemoveMember", ctx, organization, userID).Return(assert.AnError)

					permSvc := new(mock.PermissionService)
					permSvc.On("CtxUserHasPermission", ctx, model.MustNewNilID(model.ResourceTypeOrganization), []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaOrganizations).Return(true, nil)

					return &baseService{
						logger:            new(mock.Logger),
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
				baseService: func(ctx context.Context, organization model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.organizationService/RemoveMember", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            new(mock.Logger),
						tracer:            tracer,
						organizationRepo:  new(mock.OrganizationRepository),
						permissionService: new(mock.PermissionService),
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
			s := &organizationService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.organization),
			}
			err := s.RemoveMember(tt.args.ctx, tt.args.organization, tt.args.member)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
