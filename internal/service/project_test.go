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
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateProjectOpts() service.CreateProjectOpts {
	return service.CreateProjectOpts{
		Key:         "ENG",
		Name:        "test project",
		Description: "test description for project",
		Logo:        "https://example.com/logo.png",
		Status:      model.ProjectStatusActive,
	}
}

func TestNewProjectService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.ProjectService, error)
		wantErr error
	}{
		{
			name: "new project service",
			build: func(ctrl *gomock.Controller) (service.ProjectService, error) {
				return service.NewProjectService(mockrepo.NewMockProjectRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new project service with no project repository",
			build: func(ctrl *gomock.Controller) (service.ProjectService, error) {
				return service.NewProjectService(nil, mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoProjectRepository,
		},
		{
			name: "new project service with no permission service",
			build: func(ctrl *gomock.Controller) (service.ProjectService, error) {
				return service.NewProjectService(mockrepo.NewMockProjectRepository(nil), nil, mocksvc.NewMockLicenseService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new project service with no license service",
			build: func(ctrl *gomock.Controller) (service.ProjectService, error) {
				return service.NewProjectService(mockrepo.NewMockProjectRepository(nil), mocksvc.NewMockPermissionService(nil), nil, mocksvc.NewMockSearchService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new project service with no search service",
			build: func(ctrl *gomock.Controller) (service.ProjectService, error) {
				return service.NewProjectService(mockrepo.NewMockProjectRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), nil, service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoSearchService,
		},
		{
			name: "new project service with invalid options",
			build: func(_ *gomock.Controller) (service.ProjectService, error) {
				return service.NewProjectService(mockrepo.NewMockProjectRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), mocksvc.NewMockSearchService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
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
				require.NotNil(t, got)
			}
		})
	}
}

func TestProjectService_Create(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	opts := newCreateProjectOpts()

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts service.CreateProjectOpts) service.ProjectService
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		opts        service.CreateProjectOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(testModel.NewRepositoryProject(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, namespaceID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				opts:        opts,
			},
		},
		{
			name: "create project normalizes key to uppercase",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         "ENG",
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(testModel.NewRepositoryProject(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, namespaceID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				opts: service.CreateProjectOpts{
					Key:         "eng",
					Name:        "test project",
					Description: "test description for project",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
				},
			},
		},
		{
			name: "create project with default status",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      model.ProjectStatusActive,
					}).Return(testModel.NewRepositoryProject(), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, namespaceID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				opts: service.CreateProjectOpts{
					Key:         "ENG",
					Name:        "test project",
					Description: "test description for project",
					Logo:        "https://example.com/logo.png",
				},
			},
		},
		{
			name: "create project with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, _ service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, namespaceID, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create project with invalid namespaceID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: model.ID{},
				opts:        opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "create project with invalid details",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        service.CreateProjectOpts{Key: "AB", Name: "ab"},
			},
			wantErr: model.ErrInvalidProjectDetails,
		},
		{
			name: "create project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(nil, repository.ErrProjectCreate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, namespaceID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				opts:        opts,
			},
			wantErr: repository.ErrProjectCreate,
		},
		{
			name: "create project with no user ID in context",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, _ service.CreateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, namespaceID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        opts,
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

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.opts)

			_, err := s.Create(tt.args.ctx, tt.args.namespaceID, tt.args.opts)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectService_Get(t *testing.T) {
	projectID := model.MustNewID(model.ResourceTypeProject)
	repoProject := testModel.NewRepositoryProject()
	repoProject.ID = projectID
	want := service.ProjectFromRepository(repoProject)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Project
		wantErr error
	}{
		{
			name: "get project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Get(ctx, id, repository.ProjectDetailProjection()).Return(repoProject, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				id:  projectID,
			},
			want: want,
		},
		{
			name: "get project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				id:  projectID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get project with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
			name: "get project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Get(ctx, id, repository.ProjectDetailProjection()).Return(nil, repository.ErrProjectRead)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				id:  projectID,
			},
			wantErr: repository.ErrProjectRead,
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

func TestProjectService_GetByKey(t *testing.T) {
	repoProject := testModel.NewRepositoryProject()
	projectKey := repoProject.Key
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	want := service.ProjectFromRepository(repoProject)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) service.ProjectService
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		key         string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Project
		wantErr error
	}{
		{
			name: "get project by key",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetByKey(ctx, namespaceID, key, repository.ProjectDetailProjection()).Return(repoProject, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, repoProject.ID, gomock.Any()).Return(true, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         projectKey,
			},
			want: want,
		},
		{
			name: "get project by key with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetByKey(ctx, namespaceID, key, repository.ProjectDetailProjection()).Return(repoProject, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, repoProject.ID, gomock.Any()).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         projectKey,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get project by key with empty key",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         "",
			},
			wantErr: validate.ErrInvalidProjectKey,
		},
		{
			name: "get project by key with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetByKey(ctx, namespaceID, key, repository.ProjectDetailProjection()).Return(nil, repository.ErrProjectRead)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         projectKey,
			},
			wantErr: repository.ErrProjectRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.key)

			got, err := s.GetByKey(tt.args.ctx, tt.args.namespaceID, tt.args.key)
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

func TestProjectService_List(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoProjects := []*repository.Project{
		testModel.NewRepositoryProject(),
		testModel.NewRepositoryProject(),
	}
	want := service.ProjectsFromRepository(repoProjects)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) service.ProjectService
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		page        service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.Project]
		wantErr error
	}{
		{
			name: "get all projects",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/List", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().ListForNamespace(ctx, repository.ProjectListQuery{
						NamespaceID: namespaceID,
						ActorID:     userID,
						Page:        repository.CursorPage{Size: 10},
						Order:       repository.SortDirectionDesc,
						Projection:  repository.ProjectListProjection(),
					}).Return(repository.Page[*repository.Project]{Items: repoProjects}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionProjectRead).Return([]model.ID{namespaceID}, nil)
					permSvc.EXPECT().ListScopeAncestry(ctx, namespaceID).Return([]model.ID{namespaceID}, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				page:        service.CursorPage{Size: 10},
			},
			want: service.Page[*service.Project]{Items: want},
		},
		{
			name: "get all projects with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				page:        service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "get all projects with invalid namespaceID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: model.ID{},
				page:        service.CursorPage{Size: 10},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "list projects with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				page:        service.CursorPage{Size: -1},
			},
			wantErr: repository.ErrInvalidPageSize,
		},
		{
			name: "get all projects with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/List", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().ListForNamespace(ctx, repository.ProjectListQuery{
						NamespaceID: namespaceID,
						ActorID:     userID,
						Page:        repository.CursorPage{Size: 10},
						Order:       repository.SortDirectionDesc,
						Projection:  repository.ProjectListProjection(),
					}).Return(repository.Page[*repository.Project]{}, repository.ErrProjectRead)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionProjectRead).Return([]model.ID{namespaceID}, nil)
					permSvc.EXPECT().ListScopeAncestry(ctx, namespaceID).Return([]model.ID{namespaceID}, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				page:        service.CursorPage{Size: 10},
			},
			wantErr: repository.ErrProjectRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID)

			got, err := s.List(tt.args.ctx, tt.args.namespaceID, tt.args.page)
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

func TestProjectService_Update(t *testing.T) {
	projectID := model.MustNewID(model.ResourceTypeProject)
	repoProject := testModel.NewRepositoryProject()
	repoProject.ID = projectID
	want := service.ProjectFromRepository(repoProject)
	opts := service.UpdateProjectOpts{Name: optional.Some("Updated Name")}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateProjectOpts) service.ProjectService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts service.UpdateProjectOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Project
		wantErr error
	}{
		{
			name: "update project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Update(ctx, id, repository.UpdateProjectOpts{
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}, repository.ProjectDetailProjection()).Return(repoProject, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				id:   projectID,
				opts: opts,
			},
			want: want,
		},
		{
			name: "update project with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				id:   projectID,
				opts: opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				id:   projectID,
				opts: opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update project with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
			name: "update project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateProjectOpts) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Update(ctx, id, repository.UpdateProjectOpts{
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}, repository.ProjectDetailProjection()).Return(nil, repository.ErrProjectUpdate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				id:   projectID,
				opts: opts,
			},
			wantErr: repository.ErrProjectUpdate,
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

func TestProjectService_Delete(t *testing.T) {
	projectID := model.MustNewID(model.ResourceTypeProject)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService
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
			name: "delete project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				id:  projectID,
			},
		},
		{
			name: "delete project with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				id:  projectID,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
				id:  projectID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete project with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							mockrepo.NewMockProjectRepository(ctrl),
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
			name: "delete project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.ProjectService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					projectRepo := mockrepo.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Delete(ctx, id).Return(repository.ErrProjectDelete)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.ProjectService {
						svc, err := service.NewProjectService(
							projectRepo,
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
				id:  projectID,
			},
			wantErr: repository.ErrProjectDelete,
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

func TestProjectFromRepository(t *testing.T) {
	t.Parallel()

	t.Run("nil project", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, service.ProjectFromRepository(nil))
	})

	t.Run("nil partial project", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, service.PartialProjectFromRepository(nil))
	})

	t.Run("nil partial issue", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, service.PartialIssueFromRepository(nil))
	})
}
