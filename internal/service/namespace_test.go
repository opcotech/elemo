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

func newCreateNamespaceOpts() CreateNamespaceOpts {
	return CreateNamespaceOpts{
		Name:        "test namespace",
		Description: "test description",
	}
}

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
						WithNamespaceRepository(repository.NewMockNamespaceRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
						WithSearchService(NewMockSearchService(nil)),
					}
				},
			},
			want: func(ctrl *gomock.Controller) NamespaceService {
				return &namespaceService{
					baseService: &baseService{
						searchService:     NewMockSearchService(nil),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            mock.NewMockTracer(ctrl),
						namespaceRepo:     repository.NewMockNamespaceRepository(nil),
						permissionService: NewMockPermissionService(nil),
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
						WithNamespaceRepository(repository.NewMockNamespaceRepository(nil)),
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
						WithPermissionService(NewMockPermissionService(nil)),
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
						WithNamespaceRepository(repository.NewMockNamespaceRepository(nil)),
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
						WithNamespaceRepository(repository.NewMockNamespaceRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
					}
				},
			},
			wantErr: ErrNoLicenseService,
		},
		{
			name: "new namespace service with no search service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithNamespaceRepository(repository.NewMockNamespaceRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: ErrNoSearchService,
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
	opts := newCreateNamespaceOpts()

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, opts CreateNamespaceOpts) *baseService
	}
	type args struct {
		ctx   context.Context
		orgID model.ID
		opts  CreateNamespaceOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, opts CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Create(ctx, repository.CreateNamespaceOpts{
						Name:        opts.Name,
						Description: opts.Description,
						CreatorID:   userID,
						OrgID:       orgID,
					}).Return(testModel.NewRepositoryNamespace(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     mockSearchIndex(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  opts,
			},
		},
		{
			name: "create namespace with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						searchService:  NewMockSearchService(ctrl),
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, orgID model.ID, _ CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  opts,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create namespace with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:  NewMockSearchService(ctrl),
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: model.ID{},
				opts:  opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "create namespace with invalid details",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:  NewMockSearchService(ctrl),
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  CreateNamespaceOpts{Name: "ab"},
			},
			wantErr: model.ErrInvalidNamespaceDetails,
		},
		{
			name: "create namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, opts CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Create(ctx, repository.CreateNamespaceOpts{
						Name:        opts.Name,
						Description: opts.Description,
						CreatorID:   userID,
						OrgID:       orgID,
					}).Return(nil, repository.ErrNamespaceCreate)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  opts,
			},
			wantErr: repository.ErrNamespaceCreate,
		},
		{
			name: "create namespace with no user ID in context",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, orgID model.ID, _ CreateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				opts:  opts,
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
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, userID, tt.args.orgID, tt.args.opts),
			}

			_, err := s.Create(tt.args.ctx, tt.args.orgID, tt.args.opts)
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
	repoNamespace := testModel.NewRepositoryNamespace()
	repoNamespace.ID = namespaceID
	want := namespaceFromRepository(repoNamespace)

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
		want    *Namespace
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

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(repoNamespace, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
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
			want: want,
		},
		{
			name: "get namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
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
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
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

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(nil, repository.ErrNamespaceRead)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
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

func TestNamespaceService_List(t *testing.T) {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoNamespaces := []*repository.Namespace{
		testModel.NewRepositoryNamespace(),
		testModel.NewRepositoryNamespace(),
	}
	want := namespacesFromRepository(repoNamespaces)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService
	}
	type args struct {
		ctx   context.Context
		orgID model.ID
		page  CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page[*Namespace]
		wantErr error
	}{
		{
			name: "get all namespaces",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().List(
						ctx,
						orgID,
						userID,
						repository.CursorPage{Size: 10},
						repository.NamespaceListProjection(),
					).Return(repository.Page[*repository.Namespace]{Items: repoNamespaces}, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				page:  CursorPage{Size: 10},
			},
			want: Page[*Namespace]{Items: want},
		},
		{
			name: "get all namespaces with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				page:  CursorPage{Size: 10},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "get all namespaces with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: model.ID{},
				page:  CursorPage{Size: 10},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "list namespaces with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				page:  CursorPage{Size: -1},
			},
			wantErr: repository.ErrInvalidPageSize,
		},
		{
			name: "get all namespaces with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().List(
						ctx,
						orgID,
						userID,
						repository.CursorPage{Size: 10},
						repository.NamespaceListProjection(),
					).Return(repository.Page[*repository.Namespace]{}, repository.ErrNamespaceRead)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				page:  CursorPage{Size: 10},
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

			got, err := s.List(tt.args.ctx, tt.args.orgID, tt.args.page)
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

func TestNamespaceService_ListAccessible(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	repoNS := testModel.NewRepositoryNamespace()
	repoAccessible := []*repository.AccessibleNamespace{
		{
			Namespace: *repoNS,
			Organization: repository.PartialOrganization{
				ID:   orgID,
				Name: "ACME",
			},
		},
	}
	want := Page[*AccessibleNamespace]{
		Items: []*AccessibleNamespace{accessibleNamespaceFromRepository(repoAccessible[0])},
	}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) *baseService
	}
	type args struct {
		ctx  context.Context
		page CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page[*AccessibleNamespace]
		wantErr error
	}{
		{
			name: "list reachable namespaces",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/ListAccessible", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().ListAccessible(
						ctx,
						userID,
						repository.CursorPage{Size: 10},
						repository.NamespaceListProjection(),
					).Return(repository.Page[*repository.AccessibleNamespace]{Items: repoAccessible}, nil)

					return &baseService{
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
						namespaceRepo: namespaceRepo,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page: CursorPage{Size: 10},
			},
			want: want,
		},
		{
			name: "list reachable namespaces with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/ListAccessible", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				page: CursorPage{Size: 10},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "list reachable namespaces with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/ListAccessible", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						searchService: NewMockSearchService(ctrl),
						logger:        mock.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				page: CursorPage{Size: -1},
			},
			wantErr: repository.ErrInvalidPageSize,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &namespaceService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
			}

			got, err := s.ListAccessible(tt.args.ctx, tt.args.page)
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
	repoNamespace := testModel.NewRepositoryNamespace()
	repoNamespace.ID = namespaceID
	want := namespaceFromRepository(repoNamespace)
	opts := UpdateNamespaceOpts{Name: optional.Some("Updated Name")}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateNamespaceOpts) *baseService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateNamespaceOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Namespace
		wantErr error
	}{
		{
			name: "update namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Update(ctx, id, repository.UpdateNamespaceOpts{
						Name:        opts.Name,
						Description: opts.Description,
					}).Return(repoNamespace, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     mockSearchIndex(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   namespaceID,
				opts: opts,
			},
			want: want,
		},
		{
			name: "update namespace with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						searchService:  NewMockSearchService(ctrl),
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   namespaceID,
				opts: opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update namespace with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   namespaceID,
				opts: opts,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:  NewMockSearchService(ctrl),
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.ID{},
				opts: opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "update namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateNamespaceOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Update(ctx, id, repository.UpdateNamespaceOpts{
						Name:        opts.Name,
						Description: opts.Description,
					}).Return(nil, repository.ErrNamespaceUpdate)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						namespaceRepo:     namespaceRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   namespaceID,
				opts: opts,
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
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts),
			}

			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.opts)
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

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     mockSearchDeleteByScope(ctrl),
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
						searchService:  NewMockSearchService(ctrl),
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

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
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
						searchService:  NewMockSearchService(ctrl),
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

					namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Delete(ctx, id).Return(repository.ErrNamespaceDelete)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						searchService:     NewMockSearchService(ctrl),
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

func TestNamespaceFromRepository(t *testing.T) {
	t.Parallel()

	t.Run("nil namespace", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, namespaceFromRepository(nil))
	})

	t.Run("nil partial project", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, partialProjectFromRepository(nil))
	})
}
