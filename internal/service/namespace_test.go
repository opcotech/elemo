package service_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateNamespaceOpts() service.CreateNamespaceOpts {
	return service.CreateNamespaceOpts{
		Name:        "test namespace",
		Slug:        "test-namespace",
		Description: "test description",
	}
}

func TestNewNamespaceService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.NamespaceService, error)
		wantErr error
	}{
		{
			name: "new namespace service",
			build: func(ctrl *gomock.Controller) (service.NamespaceService, error) {
				return service.NewNamespaceService(
					mockrepo.NewMockNamespaceRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
		},
		{
			name: "new namespace service with invalid options",
			build: func(_ *gomock.Controller) (service.NamespaceService, error) {
				return service.NewNamespaceService(
					mockrepo.NewMockNamespaceRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(nil),
				)
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new namespace service with no namespace repository",
			build: func(ctrl *gomock.Controller) (service.NamespaceService, error) {
				return service.NewNamespaceService(
					nil,
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoNamespaceRepository,
		},
		{
			name: "new namespace service with no permission service",
			build: func(ctrl *gomock.Controller) (service.NamespaceService, error) {
				return service.NewNamespaceService(
					mockrepo.NewMockNamespaceRepository(nil),
					nil,
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new namespace service with no license service",
			build: func(ctrl *gomock.Controller) (service.NamespaceService, error) {
				return service.NewNamespaceService(
					mockrepo.NewMockNamespaceRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					nil,
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new namespace service with no search service",
			build: func(ctrl *gomock.Controller) (service.NamespaceService, error) {
				return service.NewNamespaceService(
					mockrepo.NewMockNamespaceRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoSearchService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := tt.build(ctrl)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.NotNil(t, got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestNamespaceService_Create(t *testing.T) {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	opts := newCreateNamespaceOpts()

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, opts service.CreateNamespaceOpts) service.NamespaceService
	}
	type args struct {
		ctx   context.Context
		orgID model.ID
		opts  service.CreateNamespaceOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, opts service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Create(ctx, repository.CreateNamespaceOpts{
						Name:        opts.Name,
						Slug:        opts.Slug,
						Description: opts.Description,
						CreatorID:   userID,
						OrgID:       orgID,
					}).Return(testModel.NewRepositoryNamespace(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							licenseSvc,
							mockSearchIndex(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, orgID model.ID, _ service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create namespace with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				opts:  service.CreateNamespaceOpts{Name: "ab"},
			},
			wantErr: model.ErrInvalidNamespaceDetails,
		},
		{
			name: "create namespace with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, orgID model.ID, opts service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Create(ctx, repository.CreateNamespaceOpts{
						Name:        opts.Name,
						Slug:        opts.Slug,
						Description: opts.Description,
						CreatorID:   userID,
						OrgID:       orgID,
					}).Return(nil, repository.ErrNamespaceCreate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, orgID model.ID, _ service.CreateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				opts:  opts,
			},
			wantErr: service.ErrNoUser,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userID, _ := tt.args.ctx.Value(pkg.CtxKeyUserID).(model.ID)
			s := tt.fields.baseService(ctrl, tt.args.ctx, userID, tt.args.orgID, tt.args.opts)

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
	want := service.NamespaceFromRepository(repoNamespace)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Namespace
		wantErr error
	}{
		{
			name: "get namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(repoNamespace, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Get", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(nil, repository.ErrNamespaceRead)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id)

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

func TestNamespaceService_GetByRef(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoNS := testModel.NewRepositoryNamespace()
	accessible := &repository.AccessibleNamespace{
		Namespace: *repoNS,
		Organization: repository.PartialOrganization{
			ID:   orgID,
			Slug: "acme",
			Name: "ACME",
		},
	}
	want := service.AccessibleNamespaceFromRepository(accessible)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("get reachable namespace by slug", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.namespaceService/GetByRef", gomock.Len(0)).Return(ctx, span)

		namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
		namespaceRepo.EXPECT().GetByRef(ctx, orgID, model.ID{}, repoNS.Slug, userID, repository.NamespaceDetailProjection()).Return(accessible, nil)

		svc, err := service.NewNamespaceService(
			namespaceRepo,
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockLicenseService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)

		got, err := svc.GetByRef(ctx, orgID, model.ID{}, repoNS.Slug)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("namespace xid under wrong organization", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.namespaceService/GetByRef", gomock.Len(0)).Return(ctx, span)

		namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
		namespaceRepo.EXPECT().GetByRef(ctx, orgID, repoNS.ID, "", userID, repository.NamespaceDetailProjection()).Return(nil, repository.ErrNotFound)

		svc, err := service.NewNamespaceService(
			namespaceRepo,
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockLicenseService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)

		got, err := svc.GetByRef(ctx, orgID, repoNS.ID, "")
		require.ErrorIs(t, err, repository.ErrNotFound)
		require.Nil(t, got)
	})
}

func TestNamespaceService_List(t *testing.T) {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoNamespaces := []*repository.Namespace{
		testModel.NewRepositoryNamespace(),
		testModel.NewRepositoryNamespace(),
	}
	want := service.NamespacesFromRepository(repoNamespaces)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) service.NamespaceService
	}
	type args struct {
		ctx   context.Context
		orgID model.ID
		page  service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.Namespace]
		wantErr error
	}{
		{
			name: "get all namespaces",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().ListForOrganization(ctx, repository.NamespaceListQuery{
						OrgID:      orgID,
						ActorID:    userID,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionDesc,
						Projection: repository.NamespaceListProjection(),
					}).Return(repository.Page[*repository.Namespace]{Items: repoNamespaces}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				page:  service.CursorPage{Size: 10},
			},
			want: service.Page[*service.Namespace]{Items: want},
		},
		{
			name: "get all namespaces with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				page:  service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "get all namespaces with invalid orgID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: model.ID{},
				page:  service.CursorPage{Size: 10},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "list namespaces with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: orgID,
				page:  service.CursorPage{Size: -1},
			},
			wantErr: repository.ErrInvalidPageSize,
		},
		{
			name: "get all namespaces with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/List", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().ListForOrganization(ctx, repository.NamespaceListQuery{
						OrgID:      orgID,
						ActorID:    userID,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionDesc,
						Projection: repository.NamespaceListProjection(),
					}).Return(repository.Page[*repository.Namespace]{}, repository.ErrNamespaceRead)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				orgID: orgID,
				page:  service.CursorPage{Size: 10},
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.orgID)

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
	want := service.Page[*service.AccessibleNamespace]{
		Items: []*service.AccessibleNamespace{service.AccessibleNamespaceFromRepository(repoAccessible[0])},
	}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context) service.NamespaceService
	}
	type args struct {
		ctx  context.Context
		page service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.AccessibleNamespace]
		wantErr error
	}{
		{
			name: "list reachable namespaces",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/ListAccessible", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().ListAccessible(ctx, repository.NamespaceListAccessibleQuery{
						ActorID:    userID,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionDesc,
						Projection: repository.NamespaceListProjection(),
					}).Return(repository.Page[*repository.AccessibleNamespace]{Items: repoAccessible}, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page: service.CursorPage{Size: 10},
			},
			want: want,
		},
		{
			name: "list reachable namespaces with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/ListAccessible", gomock.Len(0)).Return(ctx, span)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:  context.Background(),
				page: service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "list reachable namespaces with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/ListAccessible", gomock.Len(0)).Return(ctx, span)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:  context.Background(),
				page: service.CursorPage{Size: -1},
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

			s := tt.fields.baseService(ctrl, tt.args.ctx)

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
	want := service.NamespaceFromRepository(repoNamespace)
	opts := service.UpdateNamespaceOpts{Name: optional.Some("Updated Name")}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateNamespaceOpts) service.NamespaceService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts service.UpdateNamespaceOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Namespace
		wantErr error
	}{
		{
			name: "update namespace",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Update(ctx, id, repository.UpdateNamespaceOpts{
						Name:        opts.Name,
						Description: opts.Description,
					}).Return(repoNamespace, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							licenseSvc,
							mockSearchIndex(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   namespaceID,
				opts: opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateNamespaceOpts) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Update", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Update(ctx, id, repository.UpdateNamespaceOpts{
						Name:        opts.Name,
						Description: opts.Description,
					}).Return(nil, repository.ErrNamespaceUpdate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts)

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
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							licenseSvc,
							mockSearchDeleteByScope(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			args: args{
				ctx: context.Background(),
				id:  namespaceID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete namespace with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							mockrepo.NewMockNamespaceRepository(ctrl),
							mocksvc.NewMockPermissionService(ctrl),
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.NamespaceService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

					namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
					namespaceRepo.EXPECT().Delete(ctx, id).Return(repository.ErrNamespaceDelete)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.NamespaceService {
						svc, err := service.NewNamespaceService(
							namespaceRepo,
							permSvc,
							licenseSvc,
							mocksvc.NewMockSearchService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id)

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
		assert.Nil(t, service.NamespaceFromRepository(nil))
	})

	t.Run("nil partial project", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, service.PartialProjectFromRepository(nil))
	})
}
