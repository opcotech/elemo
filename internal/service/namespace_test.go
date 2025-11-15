package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNamespaceService(t *testing.T) {
	type args struct {
		opts func(ctrl *gomock.Controller) []Option
	}
	tests := []struct {
		name    string
		args    args
		want    func(ctrl *gomock.Controller) NamespaceService
		wantErr error
	}{
		{
			name: "new namespace service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithNamespaceRepository(mock.NewNamespaceRepository(nil)),
						WithPermissionService(mock.NewPermissionService(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			want: func(ctrl *gomock.Controller) NamespaceService {
				return &namespaceService{
					baseService: &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            mock.NewMockTracer(ctrl),
						namespaceRepo:     mock.NewNamespaceRepository(nil),
						permissionService: mock.NewPermissionService(nil),
						licenseService:    mock.NewMockLicenseService(nil),
					},
				}
			},
		},
		{
			name: "new namespace service with invalid options",
			args: args{
				opts: func(_ *gomock.Controller) []Option {
					return []Option{
						WithLogger(nil),
						WithNamespaceRepository(mock.NewNamespaceRepository(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new namespace service with no namespace repository",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithPermissionService(mock.NewPermissionService(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: ErrNoNamespaceRepository,
		},
		{
			name: "new namespace service with no permission service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithNamespaceRepository(mock.NewNamespaceRepository(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new namespace service with no license service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithNamespaceRepository(mock.NewNamespaceRepository(nil)),
						WithPermissionService(mock.NewPermissionService(nil)),
					}
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := NewNamespaceService(tt.args.opts(ctrl)...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.want != nil {
				assert.Equal(t, tt.want(ctrl), got)
			}
		})
	}
}

func TestNamespaceService_Create(t *testing.T) {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	namespace := testModel.NewNamespace()

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, namespace *model.Namespace) *baseService
	}
	type args struct {
		ctx       context.Context
		orgID     model.ID
		namespace *model.Namespace
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, namespace *model.Namespace) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Create(ctx, userID, orgID, namespace).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:     orgID,
				namespace: namespace,
			},
		},
		{
			name: "create namespace with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:     orgID,
				namespace: namespace,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, orgID model.ID, _ *model.Namespace) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindWrite}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:     orgID,
				namespace: namespace,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create namespace with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:     model.ID{},
				namespace: namespace,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "create namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, namespace *model.Namespace) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Create(ctx, userID, orgID, namespace).Return(repository.ErrNamespaceCreate)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID:     orgID,
				namespace: namespace,
			},
			wantErr: repository.ErrNamespaceCreate,
		},
		{
			name: "create namespace with no user ID in context",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, orgID model.ID, _ *model.Namespace) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				orgID:     orgID,
				namespace: namespace,
			},
			wantErr: model.ErrInvalidID,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userID, _ := tt.args.ctx.Value(pkg.CtxKeyUserID).(model.ID)
			s := &namespaceService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, userID, tt.args.orgID, tt.args.namespace),
			}

			err := s.Create(tt.args.ctx, tt.args.orgID, tt.args.namespace)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNamespaceService_Get(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	namespace := testModel.NewNamespace()
	namespace.ID = namespaceID

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Namespace
		wantErr error
	}{
		{
			name: "get namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Get(ctx, id).Return(namespace, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			want: namespace,
		},
		{
			name: "get namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindRead}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "get namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNamespaceRead)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: repository.ErrNamespaceRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &namespaceService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id),
			}

			got, err := s.Get(tt.args.ctx, tt.args.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNamespaceService_GetAll(t *testing.T) {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaces := []*model.Namespace{
		testModel.NewNamespace(),
		testModel.NewNamespace(),
	}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService
	}
	type args struct {
		ctx    context.Context
		orgID  model.ID
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Namespace
		wantErr error
	}{
		{
			name: "get all namespaces",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/GetAll", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().GetAll(ctx, orgID, 0, 10).Return(namespaces, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  orgID,
				offset: 0,
				limit:  10,
			},
			want: namespaces,
		},
		{
			name: "get all namespaces with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/GetAll", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindRead}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  orgID,
				offset: 0,
				limit:  10,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get all namespaces with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.ID{},
				offset: 0,
				limit:  10,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "get all namespaces with invalid pagination",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  orgID,
				offset: -1,
				limit:  10,
			},
			wantErr: ErrInvalidPaginationParams,
		},
		{
			name: "get all namespaces with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/GetAll", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().GetAll(ctx, orgID, 0, 10).Return(nil, repository.ErrNamespaceRead)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, orgID, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  orgID,
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrNamespaceRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &namespaceService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID),
			}

			got, err := s.GetAll(tt.args.ctx, tt.args.orgID, tt.args.offset, tt.args.limit)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNamespaceService_Update(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	namespace := testModel.NewNamespace()
	namespace.ID = namespaceID
	patch := map[string]any{"name": "Updated Name"}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any) *baseService
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
		want    *model.Namespace
		wantErr error
	}{
		{
			name: "update namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    namespaceID,
				patch: patch,
			},
			want: namespace,
		},
		{
			name: "update namespace with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    namespaceID,
				patch: patch,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ map[string]any) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindWrite}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    namespaceID,
				patch: patch,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.ID{},
				patch: patch,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "update namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNamespaceUpdate)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    namespaceID,
				patch: patch,
			},
			wantErr: repository.ErrNamespaceUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &namespaceService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.patch),
			}

			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNamespaceService_Delete(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService
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
			name: "delete namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindDelete}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
		},
		{
			name: "delete namespace with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindDelete}).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "delete namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mock.NewNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Delete(ctx, id).Return(repository.ErrNamespaceDelete)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindDelete}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: repository.ErrNamespaceDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &namespaceService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id),
			}

			err := s.Delete(tt.args.ctx, tt.args.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
