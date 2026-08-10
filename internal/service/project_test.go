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

func newCreateProjectOpts() CreateProjectOpts {
	return CreateProjectOpts{
		Key:         "ENG",
		Name:        "test project",
		Description: "test description for project",
		Logo:        "https://example.com/logo.png",
		Status:      model.ProjectStatusActive,
	}
}

func TestNewProjectService(t *testing.T) {
	type args struct {
		opts func(ctrl *gomock.Controller) []Option
	}
	tests := []struct {
		name    string
		args    args
		want    func(ctrl *gomock.Controller) ProjectService
		wantErr error
	}{
		{
			name: "new project service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithProjectRepository(repository.NewMockProjectRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			want: func(ctrl *gomock.Controller) ProjectService {
				return &projectService{
					baseService: &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            mock.NewMockTracer(ctrl),
						projectRepo:       repository.NewMockProjectRepository(nil),
						permissionService: NewMockPermissionService(nil),
						licenseService:    mock.NewMockLicenseService(nil),
					},
				}
			},
		},
		{
			name: "new project service with invalid options",
			args: args{
				opts: func(_ *gomock.Controller) []Option {
					return []Option{
						WithLogger(nil),
						WithProjectRepository(repository.NewMockProjectRepository(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new project service with no project repository",
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
			wantErr: ErrNoProjectRepository,
		},
		{
			name: "new project service with no permission service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithProjectRepository(repository.NewMockProjectRepository(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new project service with no license service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithProjectRepository(repository.NewMockProjectRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
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
			got, err := NewProjectService(tt.args.opts(ctrl)...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.want != nil {
				assert.Equal(t, tt.want(ctrl), got)
			}
		})
	}
}

func TestProjectService_Create(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	opts := newCreateProjectOpts()

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) *baseService
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		opts        CreateProjectOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(testModel.NewRepositoryProject(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         "ENG",
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(testModel.NewRepositoryProject(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				opts: CreateProjectOpts{
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      model.ProjectStatusActive,
					}).Return(testModel.NewRepositoryProject(), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				opts: CreateProjectOpts{
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, _ CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindWrite}).Return(false)

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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        opts,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create project with invalid namespaceID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

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
				ctx:         context.Background(),
				namespaceID: model.ID{},
				opts:        opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "create project with invalid details",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        CreateProjectOpts{Key: "AB", Name: "ab"},
			},
			wantErr: model.ErrInvalidProjectDetails,
		},
		{
			name: "create project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Create(ctx, repository.CreateProjectOpts{
						NamespaceID: namespaceID,
						CreatorID:   creatorID,
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(nil, repository.ErrProjectCreate)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, _ CreateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

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
				ctx:         context.Background(),
				namespaceID: namespaceID,
				opts:        opts,
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

			s := &projectService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.opts),
			}

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
	repoProject.Documents = []*repository.PartialDocument{
		{
			ID:        model.MustNewID(model.ResourceTypeDocument),
			Name:      "Plan",
			Excerpt:   "Overview",
			CreatedBy: model.MustNewID(model.ResourceTypeUser),
		},
	}
	want := projectFromRepository(repoProject)

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
		want    *Project
		wantErr error
	}{
		{
			name: "get project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Get(ctx, id).Return(repoProject, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
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
				id:  projectID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get project with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

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
			name: "get project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Get", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Get(ctx, id).Return(nil, repository.ErrProjectRead)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
					}
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

			s := &projectService{
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

func TestProjectService_GetByKey(t *testing.T) {
	repoProject := testModel.NewRepositoryProject()
	projectKey := repoProject.Key
	want := projectFromRepository(repoProject)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, key string) *baseService
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Project
		wantErr error
	}{
		{
			name: "get project by key",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, key string) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetByKey(ctx, key).Return(repoProject, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, repoProject.ID, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				key: projectKey,
			},
			want: want,
		},
		{
			name: "get project by key with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, key string) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetByKey(ctx, key).Return(repoProject, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, repoProject.ID, []model.PermissionKind{model.PermissionKindRead}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				key: projectKey,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get project by key with empty key",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ string) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				key: "",
			},
			wantErr: model.ErrInvalidProjectDetails,
		},
		{
			name: "get project by key with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, key string) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetByKey", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetByKey(ctx, key).Return(nil, repository.ErrProjectRead)

					return &baseService{
						logger:      mock.NewMockLogger(ctrl),
						tracer:      tracer,
						projectRepo: projectRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				key: projectKey,
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

			s := &projectService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.key),
			}

			got, err := s.GetByKey(tt.args.ctx, tt.args.key)
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

func TestProjectService_GetAll(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	repoProjects := []*repository.Project{
		testModel.NewRepositoryProject(),
		testModel.NewRepositoryProject(),
	}
	repoProjects[0].Documents = []*repository.PartialDocument{
		{
			ID:        model.MustNewID(model.ResourceTypeDocument),
			Name:      "Spec",
			Excerpt:   "Details",
			CreatedBy: model.MustNewID(model.ResourceTypeUser),
		},
	}
	want := projectsFromRepository(repoProjects)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) *baseService
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		offset      int
		limit       int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Project
		wantErr error
	}{
		{
			name: "get all projects",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetAll", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetAll(ctx, namespaceID, 0, 10).Return(repoProjects, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				offset:      0,
				limit:       10,
			},
			want: want,
		},
		{
			name: "get all projects with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetAll", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindRead}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				offset:      0,
				limit:       10,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get all projects with invalid namespaceID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: model.ID{},
				offset:      0,
				limit:       10,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "get all projects with invalid pagination",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				offset:      -1,
				limit:       10,
			},
			wantErr: ErrInvalidPaginationParams,
		},
		{
			name: "get all projects with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/GetAll", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().GetAll(ctx, namespaceID, 0, 10).Return(nil, repository.ErrProjectRead)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, namespaceID, []model.PermissionKind{model.PermissionKindRead}).Return(true)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				offset:      0,
				limit:       10,
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

			s := &projectService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID),
			}

			got, err := s.GetAll(tt.args.ctx, tt.args.namespaceID, tt.args.offset, tt.args.limit)
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
	want := projectFromRepository(repoProject)
	opts := UpdateProjectOpts{Name: optional.Some("Updated Name")}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateProjectOpts) *baseService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateProjectOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Project
		wantErr error
	}{
		{
			name: "update project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Update(ctx, id, repository.UpdateProjectOpts{
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(repoProject, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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
			name: "update project normalizes key to uppercase",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Update(ctx, id, repository.UpdateProjectOpts{
						Key: optional.Some("ENG"),
					}).Return(repoProject, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   projectID,
				opts: UpdateProjectOpts{Key: optional.Some("eng")},
			},
			want: want,
		},
		{
			name: "update project with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

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
				ctx:  context.Background(),
				id:   projectID,
				opts: opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
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
				ctx:  context.Background(),
				id:   projectID,
				opts: opts,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update project with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

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
				ctx:  context.Background(),
				id:   model.ID{},
				opts: opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "update project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateProjectOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Update", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Update(ctx, id, repository.UpdateProjectOpts{
						Key:         opts.Key,
						Name:        opts.Name,
						Description: opts.Description,
						Logo:        opts.Logo,
						Status:      opts.Status,
					}).Return(nil, repository.ErrProjectUpdate)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindWrite}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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

			s := &projectService{
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

func TestProjectService_Delete(t *testing.T) {
	projectID := model.MustNewID(model.ResourceTypeProject)

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
			name: "delete project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindDelete}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

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
				id:  projectID,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete project with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
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
				id:  projectID,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete project with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

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
			name: "delete project with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.projectService/Delete", gomock.Len(0)).Return(ctx, span)

					projectRepo := repository.NewMockProjectRepository(ctrl)
					projectRepo.EXPECT().Delete(ctx, id).Return(repository.ErrProjectDelete)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{model.PermissionKindDelete}).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						projectRepo:       projectRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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

			s := &projectService{
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

func TestProjectFromRepository(t *testing.T) {
	t.Parallel()

	t.Run("nil project", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, projectFromRepository(nil))
	})

	t.Run("nil partial project", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, partialProjectFromRepository(nil))
	})

	t.Run("nil partial document", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, partialDocumentFromRepository(nil))
	})
}
