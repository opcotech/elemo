package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewPermissionService(t *testing.T) {
	type args struct {
		permissionRepo repository.PermissionRepository
		opts           []Option
	}
	tests := []struct {
		name    string
		args    args
		want    PermissionService
		wantErr error
	}{
		{
			name: "new permission service",
			args: args{
				permissionRepo: new(mock.PermissionRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &permissionService{
				baseService: &baseService{
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				permissionRepo: new(mock.PermissionRepository),
			},
		},
		{
			name: "new permission service with nil permission repository",
			args: args{
				permissionRepo: nil,
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoPermissionRepository,
		},
		{
			name: "new permission service with nil logger",
			args: args{
				permissionRepo: new(mock.PermissionRepository),
				opts: []Option{
					WithLogger(nil),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new permission service with nil tracer",
			args: args{
				permissionRepo: new(mock.PermissionRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(nil),
				},
			},
			wantErr: tracing.ErrNoTracer,
		},
		{
			name: "new permission service with missing logger",
			args: args{
				permissionRepo: new(mock.PermissionRepository),
				opts: []Option{
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &permissionService{
				baseService: &baseService{
					logger: log.DefaultLogger(),
					tracer: new(mock.Tracer),
				},
				permissionRepo: new(mock.PermissionRepository),
			},
		},
		{
			name: "new permission service with missing tracer",
			args: args{
				permissionRepo: new(mock.PermissionRepository),
				opts: []Option{
					WithLogger(new(mock.Logger)),
				},
			},
			want: &permissionService{
				baseService: &baseService{
					logger: new(mock.Logger),
					tracer: tracing.NoopTracer(),
				},
				permissionRepo: new(mock.PermissionRepository),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPermissionService(tt.args.permissionRepo, tt.args.opts...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_Create(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, perm *model.Permission) *baseService
		permissionRepo func(ctx context.Context, perm *model.Permission) repository.PermissionRepository
	}
	type args struct {
		ctx  context.Context
		perm *model.Permission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create permission",
			fields: fields{
				baseService: func(ctx context.Context, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Create", ctx, perm).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "create permission with error",
			fields: fields{
				baseService: func(ctx context.Context, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Create", ctx, perm).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrPermissionCreate,
		},
		{
			name: "create permission with nil permission",
			fields: fields{
				baseService: func(ctx context.Context, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, perm *model.Permission) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:  context.Background(),
				perm: nil,
			},
			wantErr: model.ErrInvalidPermissionDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.perm),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.perm),
			}
			err := s.Create(tt.args.ctx, tt.args.perm)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func Test_permissionService_CtxUserCreate(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService    func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService
		permissionRepo func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository
	}
	type args struct {
		ctx  context.Context
		perm *model.Permission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create permission having all permissions",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserCreate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindRead,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Create", ctx, perm).Return(nil)

					return repo
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "create permission having a direct permission",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserCreate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindRead,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Create", ctx, perm).Return(nil)

					return repo
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "create permission having all permissions and relation",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserCreate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindRead,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Create", ctx, perm).Return(nil)

					return repo
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "create permission having a system role",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserCreate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindRead,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("Create", ctx, perm).Return(nil)

					return repo
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "create permission having relation but no permission",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserCreate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindRead,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(false, nil)

					return repo
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create permission no relation or permission or role",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserCreate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindRead,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(false, nil)

					return repo
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				perm: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, userID, tt.args.perm),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, userID, tt.args.perm),
			}
			err := s.CtxUserCreate(tt.args.ctx, tt.args.perm)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func Test_permissionService_Get(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, id model.ID, perm *model.Permission) *baseService
		permissionRepo func(ctx context.Context, id model.ID, perm *model.Permission) repository.PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Permission
		wantErr error
	}{
		{
			name: "get permission",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Get", ctx, id).Return(perm, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			want: &model.Permission{
				ID:      model.MustNewID(model.ResourceTypePermission),
				Kind:    model.PermissionKindCreate,
				Subject: model.MustNewID(model.ResourceTypeUser),
				Target:  model.MustNewID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "get permission with error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Get", ctx, id).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrPermissionGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_GetBySubject(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, id model.ID, perms []*model.Permission) *baseService
		permissionRepo func(ctx context.Context, id model.ID, perms []*model.Permission) repository.PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permissions",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, perms []*model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/GetBySubject", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, perms []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubject", ctx, id).Return(perms, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Permission{
				{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "get permission with error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, perms []*model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/GetBySubject", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, perms []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubject", ctx, id).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrPermissionGetBySubject,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := s.GetBySubject(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_GetByTarget(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, id model.ID, perms []*model.Permission) *baseService
		permissionRepo func(ctx context.Context, id model.ID, perms []*model.Permission) repository.PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permissions",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, perms []*model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/GetByTarget", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, perms []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetByTarget", ctx, id).Return(perms, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Permission{
				{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "get permission with error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, perms []*model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/GetByTarget", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, perms []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetByTarget", ctx, id).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrPermissionGetByTarget,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := s.GetByTarget(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_GetBySubjectAndTarget(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, subject, target model.ID, perms []*model.Permission) *baseService
		permissionRepo func(ctx context.Context, subject, target model.ID, perms []*model.Permission) repository.PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		target  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permissions",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, perms []*model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/GetBySubjectAndTarget", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, perms []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubjectAndTarget", ctx, subject, target).Return(perms, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Permission{
				{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindCreate,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "get permission with error",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, perms []*model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/GetBySubjectAndTarget", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, perms []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubjectAndTarget", ctx, subject, target).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrPermissionGetBySubjectAndTarget,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
			}
			got, err := s.GetBySubjectAndTarget(tt.args.ctx, tt.args.subject, tt.args.target)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_HasAnyRelation(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, subject, target model.ID, hasRelation bool) *baseService
		permissionRepo func(ctx context.Context, subject, target model.ID, hasRelation bool) repository.PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		target  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "get relation",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, subject, target).Return(hasRelation, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: true,
		},
		{
			name: "get relation with no relations",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, subject, target).Return(hasRelation, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: false,
		},
		{
			name: "get relation with error",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, subject, target).Return(false, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
			}
			got, err := s.HasAnyRelation(tt.args.ctx, tt.args.subject, tt.args.target)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_CtxUserHasAnyRelation(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService    func(ctx context.Context, userID, target model.ID, hasRelation bool) *baseService
		permissionRepo func(ctx context.Context, userID, target model.ID, hasRelation bool) repository.PermissionRepository
	}
	type args struct {
		ctx    context.Context
		target model.ID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "get relation",
			fields: fields{
				baseService: func(ctx context.Context, userID, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, target model.ID, hasRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, target).Return(hasRelation, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				target: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: true,
		},
		{
			name: "get relation with no relations",
			fields: fields{
				baseService: func(ctx context.Context, userID, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, target model.ID, hasRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, target).Return(hasRelation, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				target: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: false,
		},
		{
			name: "get relation with error",
			fields: fields{
				baseService: func(ctx context.Context, userID, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, target model.ID, hasRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, target).Return(false, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				target: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: false,
		},
		{
			name: "get relation with no ctx user",
			fields: fields{
				baseService: func(ctx context.Context, userID, target model.ID, hasRelation bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, target model.ID, hasRelation bool) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:    context.Background(),
				target: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, userID, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, userID, tt.args.target, tt.want),
			}
			got := s.CtxUserHasAnyRelation(tt.args.ctx, tt.args.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_HasSystemRole(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, subject model.ID, roles []model.SystemRole, hasRole bool) *baseService
		permissionRepo func(ctx context.Context, subject model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		roles   []model.SystemRole
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "get role",
			fields: fields{
				baseService: func(ctx context.Context, subject model.ID, roles []model.SystemRole, hasRole bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, subject, roles).Return(hasRole, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				roles:   []model.SystemRole{model.SystemRoleOwner},
			},
			want: true,
		},
		{
			name: "get role with error",
			fields: fields{
				baseService: func(ctx context.Context, subject model.ID, roles []model.SystemRole, hasRole bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, subject, roles).Return(false, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				roles:   []model.SystemRole{model.SystemRoleOwner},
			},
			wantErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.subject, tt.args.roles, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.roles, tt.want),
			}
			got, err := s.HasSystemRole(tt.args.ctx, tt.args.subject, tt.args.roles...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_CtxUserHasSystemRole(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService    func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) *baseService
		permissionRepo func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository
	}
	type args struct {
		ctx   context.Context
		roles []model.SystemRole
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "get role",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, userID, roles).Return(hasRole, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roles: []model.SystemRole{model.SystemRoleOwner},
			},
			want: true,
		},
		{
			name: "get role with error",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, userID, roles).Return(false, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				roles: []model.SystemRole{model.SystemRoleOwner},
			},
		},
		{
			name: "get role with no ctx user",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, roles []model.SystemRole, hasRole bool) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:   context.Background(),
				roles: []model.SystemRole{model.SystemRoleOwner},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, userID, tt.args.roles, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, userID, tt.args.roles, tt.want),
			}
			got := s.CtxUserHasSystemRole(tt.args.ctx, tt.args.roles...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_HasPermission(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService
		permissionRepo func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		target  model.ID
		kinds   []model.PermissionKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "has permission",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, append(kinds, model.PermissionKindAll)).Return(true, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
				kinds:   []model.PermissionKind{model.PermissionKindCreate},
			},
			want: true,
		},
		{
			name: "has no permission",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, append(kinds, model.PermissionKindAll)).Return(false, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
				kinds:   []model.PermissionKind{model.PermissionKindCreate},
			},
			want: false,
		},
		{
			name: "has permission with error",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, append(kinds, model.PermissionKindAll)).Return(false, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
				kinds:   []model.PermissionKind{model.PermissionKindCreate},
			},
			wantErr: ErrPermissionHasPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.subject, tt.args.target, tt.args.kinds),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.args.kinds),
			}
			got, err := s.HasPermission(tt.args.ctx, tt.args.subject, tt.args.target, tt.args.kinds...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_CtxUserHasPermission(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService    func(ctx context.Context, userID, target model.ID, kinds []model.PermissionKind) *baseService
		permissionRepo func(ctx context.Context, userID, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository
	}
	type args struct {
		ctx    context.Context
		target model.ID
		kinds  []model.PermissionKind
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "has permission",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, append(kinds, model.PermissionKindAll)).Return(true, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				target: model.MustNewID(model.ResourceTypeOrganization),
				kinds:  []model.PermissionKind{model.PermissionKindCreate},
			},
			want: true,
		},
		{
			name: "has no permission",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, append(kinds, model.PermissionKindAll)).Return(false, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				target: model.MustNewID(model.ResourceTypeOrganization),
				kinds:  []model.PermissionKind{model.PermissionKindCreate},
			},
			want: false,
		},
		{
			name: "has permission with error",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, append(kinds, model.PermissionKindAll)).Return(false, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				target: model.MustNewID(model.ResourceTypeOrganization),
				kinds:  []model.PermissionKind{model.PermissionKindCreate},
			},
			want: false,
		},
		{
			name: "has permission with no ctx user",
			fields: fields{
				baseService: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, kinds []model.PermissionKind) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:    context.Background(),
				target: model.MustNewID(model.ResourceTypeOrganization),
				kinds:  []model.PermissionKind{model.PermissionKindCreate},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, userID, tt.args.target, tt.args.kinds),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, userID, tt.args.target, tt.args.kinds),
			}
			got := s.CtxUserHasPermission(tt.args.ctx, tt.args.target, tt.args.kinds...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_Update(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, id model.ID, kind model.PermissionKind) *baseService
		permissionRepo func(ctx context.Context, id model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		kind model.PermissionKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Permission
		wantErr error
	}{
		{
			name: "update permission",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Update", ctx, id, kind).Return(want, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeUser),
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      model.MustNewID(model.ResourceTypePermission),
				Kind:    model.PermissionKindCreate,
				Subject: model.MustNewNilID(model.ResourceTypeUser),
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "update permission with error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Update", ctx, id, kind).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeUser),
				kind: model.PermissionKindCreate,
			},
			wantErr: ErrPermissionUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.kind),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want, tt.args.kind),
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.kind)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_permissionService_CtxUserUpdate(t *testing.T) {
	permID := model.MustNewID(model.ResourceTypePermission)
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService    func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService
		permissionRepo func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		kind model.PermissionKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Permission
		wantErr error
	}{
		{
			name: "update permission with direct permission",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserUpdate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, want.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, want.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Get", ctx, want.ID).Return(want, nil)
					repo.On("Update", ctx, want.ID, kind).Return(want, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   permID,
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      permID,
				Kind:    model.PermissionKindRead,
				Subject: userID,
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "update permission with relation",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserUpdate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, want.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, want.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Get", ctx, want.ID).Return(want, nil)
					repo.On("Update", ctx, want.ID, kind).Return(want, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   permID,
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      permID,
				Kind:    model.PermissionKindRead,
				Subject: userID,
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "update permission with system role",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserUpdate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, want.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, want.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("Get", ctx, want.ID).Return(want, nil)
					repo.On("Update", ctx, want.ID, kind).Return(want, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   permID,
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      permID,
				Kind:    model.PermissionKindRead,
				Subject: userID,
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "update permission with error",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserUpdate", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, want.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, want.Target, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Get", ctx, want.ID).Return(want, nil)
					repo.On("Update", ctx, want.ID, kind).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   permID,
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      permID,
				Kind:    model.PermissionKindRead,
				Subject: userID,
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
			wantErr: ErrPermissionUpdate,
		},
		{
			name: "update permission no permission found",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserUpdate", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Get", ctx, want.ID).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:   permID,
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      permID,
				Kind:    model.PermissionKindRead,
				Subject: userID,
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
			wantErr: ErrPermissionUpdate,
		},
		{
			name: "update permission with no ctx user",
			fields: fields{
				baseService: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserUpdate", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID model.ID, want *model.Permission, kind model.PermissionKind) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   permID,
				kind: model.PermissionKindCreate,
			},
			want: &model.Permission{
				ID:      permID,
				Kind:    model.PermissionKindRead,
				Subject: userID,
				Target:  model.MustNewNilID(model.ResourceTypeOrganization),
			},
			wantErr: ErrPermissionUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, userID, tt.want, tt.args.kind),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, userID, tt.want, tt.args.kind),
			}
			got, err := s.CtxUserUpdate(tt.args.ctx, tt.args.id, tt.args.kind)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_permissionService_Delete(t *testing.T) {
	type fields struct {
		baseService    func(ctx context.Context, id model.ID) *baseService
		permissionRepo func(ctx context.Context, id model.ID) repository.PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete permission",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
		},
		{
			name: "delete permission with error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Delete", ctx, id).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrPermissionDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, tt.args.id),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id),
			}
			err := s.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func Test_permissionService_CtxUserDelete(t *testing.T) {
	permID := model.MustNewID(model.ResourceTypePermission)
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService    func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService
		permissionRepo func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository
		perm           *model.Permission
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete permission with direct permission",
			fields: fields{
				baseService: func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserDelete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Get", ctx, perm.ID).Return(perm, nil)
					repo.On("Delete", ctx, perm.ID).Return(nil)
					return repo
				},
				perm: &model.Permission{
					ID:      permID,
					Kind:    model.PermissionKindRead,
					Subject: userID,
					Target:  model.MustNewNilID(model.ResourceTypeOrganization),
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  permID,
			},
		},
		{
			name: "delete permission with relation",
			fields: fields{
				baseService: func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserDelete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(false, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Get", ctx, perm.ID).Return(perm, nil)
					repo.On("Delete", ctx, perm.ID).Return(nil)
					return repo
				},
				perm: &model.Permission{
					ID:      permID,
					Kind:    model.PermissionKindRead,
					Subject: userID,
					Target:  model.MustNewNilID(model.ResourceTypeOrganization),
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  permID,
			},
		},
		{
			name: "delete permission with system role",
			fields: fields{
				baseService: func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserDelete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(false, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(false, nil)
					repo.On("Get", ctx, perm.ID).Return(perm, nil)
					repo.On("Delete", ctx, perm.ID).Return(nil)
					return repo
				},
				perm: &model.Permission{
					ID:      permID,
					Kind:    model.PermissionKindRead,
					Subject: userID,
					Target:  model.MustNewNilID(model.ResourceTypeOrganization),
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  permID,
			},
		},
		{
			name: "delete permission with error",
			fields: fields{
				baseService: func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserDelete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasAnyRelation", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasSystemRole", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserHasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/HasPermission", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, userID, perm.Target).Return(true, nil)
					repo.On("HasSystemRole", ctx, userID, []model.SystemRole{
						model.SystemRoleOwner,
						model.SystemRoleAdmin,
					}).Return(true, nil)
					repo.On("HasPermission", ctx, userID, perm.Target, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)
					repo.On("Get", ctx, perm.ID).Return(perm, nil)
					repo.On("Delete", ctx, perm.ID).Return(assert.AnError)
					return repo
				},
				perm: &model.Permission{
					ID:      permID,
					Kind:    model.PermissionKindRead,
					Subject: userID,
					Target:  model.MustNewNilID(model.ResourceTypeOrganization),
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  permID,
			},
			wantErr: ErrPermissionDelete,
		},
		{
			name: "delete permission no permission found",
			fields: fields{
				baseService: func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.permissionService/CtxUserDelete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Get", ctx, perm.ID).Return(nil, assert.AnError)
					return repo
				},
				perm: &model.Permission{
					ID:      permID,
					Kind:    model.PermissionKindRead,
					Subject: userID,
					Target:  model.MustNewNilID(model.ResourceTypeOrganization),
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  permID,
			},
			wantErr: ErrPermissionDelete,
		},
		{
			name: "delete permission with no ctx user",
			fields: fields{
				baseService: func(ctx context.Context, userID, id model.ID, perm *model.Permission) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.permissionService/CtxUserDelete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				permissionRepo: func(ctx context.Context, userID, id model.ID, perm *model.Permission) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
				perm: &model.Permission{
					ID:      permID,
					Kind:    model.PermissionKindRead,
					Subject: userID,
					Target:  model.MustNewNilID(model.ResourceTypeOrganization),
				},
			},
			args: args{
				ctx: context.Background(),
				id:  permID,
			},
			wantErr: ErrPermissionDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &permissionService{
				baseService:    tt.fields.baseService(tt.args.ctx, userID, tt.args.id, tt.fields.perm),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, userID, tt.args.id, tt.fields.perm),
			}
			err := s.CtxUserDelete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
