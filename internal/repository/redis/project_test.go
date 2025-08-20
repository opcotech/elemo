package redis

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedProjectRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, project *model.Project) *baseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository
	}
	type args struct {
		ctx       context.Context
		namespace model.ID
		project   *model.Project
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Create(ctx, namespace, project).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeNamespace),
				project: &model.Project{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
			},
		},
		{
			name: "add new project with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Create(ctx, namespace, project).Return(repository.ErrProjectCreate)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeNamespace),
				project: &model.Project{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrProjectCreate,
		},
		{
			name: "add new project with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeNamespace),
				project: &model.Project{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new project with namespace cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeNamespace),
				project: &model.Project{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.namespace, tt.args.project),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.namespace, tt.args.project),
			}
			err := r.Create(tt.args.ctx, tt.args.namespace, tt.args.project)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedProjectRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Project
		wantErr error
	}{
		{
			name: "get uncached project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			want: func(id model.ID) *model.Project {
				return &model.Project{
					ID:          id,
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**model.Project); ok {
							*ptr = project
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			want: func(id model.ID) *model.Project {
				return &model.Project{
					ID:          id,
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached project error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached project error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached project cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *model.Project
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_GetByKey(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, key string, project *model.Project) *baseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, key string, project *model.Project) repository.ProjectRepository
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(key string) *model.Project
		wantErr error
	}{
		{
			name: "get uncached project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, projectKey).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				key: "PROJ",
			},
			want: func(projectKey string) *model.Project {
				return &model.Project{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         projectKey,
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**model.Project); ok {
							*ptr = project
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ *model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				key: "PROJ",
			},
			want: func(projectKey string) *model.Project {
				return &model.Project{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         projectKey,
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached project error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, _ *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, _ *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, projectKey).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				key: "PROJ",
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached project error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, _ *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ *model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				key: "PROJ",
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached project cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, projectKey).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				key: "PROJ",
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *model.Project
			if tt.want != nil {
				want = tt.want(tt.args.key)
			}

			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.key, want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.key, want),
			}
			got, err := r.GetByKey(tt.args.ctx, tt.args.key)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository
	}
	type args struct {
		ctx       context.Context
		namespace model.ID
		offset    int
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Project
		wantErr error
	}{
		{
			name: "get uncached projects",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: projects,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().GetAll(ctx, namespace, offset, limit).Return(projects, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Project{
				{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached projects",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dst any) error {
						if ptr, ok := dst.(*[]*model.Project); ok {
							*ptr = projects
						}
						return nil
					})

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Project{
				{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeProject),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
					Teams:       make([]model.ID, 0),
					Documents:   make([]model.ID, 0),
					Issues:      make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached projects error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, _ []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, _ []*model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().GetAll(ctx, namespace, offset, limit).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get projects cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, _ []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Project) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached projects cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: projects,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().GetAll(ctx, namespace, offset, limit).Return(projects, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				namespace: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.namespace, tt.args.offset, tt.args.limit, tt.want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.namespace, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.namespace, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedProjectRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository
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
		want    *model.Project
		wantErr error
	}{
		{
			name: "update project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byProjectKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byProjectKeyCmd := new(redis.StringSliceCmd)
					byProjectKeyCmd.SetVal([]string{byProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, byProjectKey).Return(byProjectKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byProjectKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				patch: map[string]any{
					"name":        "updated project",
					"description": "updated description",
				},
			},
			want: &model.Project{
				ID:          model.MustNewID(model.ResourceTypeProject),
				Key:         "PROJ",
				Name:        "test project",
				Description: "test description",
				Logo:        "https://example.com/logo.png",
				Status:      model.ProjectStatusActive,
				Teams:       make([]model.ID, 0),
				Documents:   make([]model.ID, 0),
				Issues:      make([]model.ID, 0),
			},
		},
		{
			name: "update project with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Project) *baseRepository {
					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				patch: map[string]any{
					"name":        "updated project",
					"description": "updated description",
				},
			},
			want: &model.Project{
				ID:          model.MustNewID(model.ResourceTypeProject),
				Key:         "PROJ",
				Name:        "test project",
				Description: "test description",
				Logo:        "https://example.com/logo.png",
				Status:      model.ProjectStatusActive,
				Teams:       make([]model.ID, 0),
				Documents:   make([]model.ID, 0),
				Issues:      make([]model.ID, 0),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update project set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				patch: map[string]any{
					"name":        "updated project",
					"description": "updated description",
				},
			},
			want: &model.Project{
				ID:          model.MustNewID(model.ResourceTypeProject),
				Key:         "PROJ",
				Name:        "test project",
				Description: "test description",
				Logo:        "https://example.com/logo.png",
				Status:      model.ProjectStatusActive,
				Teams:       make([]model.ID, 0),
				Documents:   make([]model.ID, 0),
				Issues:      make([]model.ID, 0),
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update project delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byProjectKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byProjectKeyCmd := new(redis.StringSliceCmd)
					byProjectKeyCmd.SetVal([]string{byProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, byProjectKey).Return(byProjectKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byProjectKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				patch: map[string]any{
					"name":        "updated project",
					"description": "updated description",
				},
			},
			want: &model.Project{
				ID:          model.MustNewID(model.ResourceTypeProject),
				Key:         "PROJ",
				Name:        "test project",
				Description: "test description",
				Logo:        "https://example.com/logo.png",
				Status:      model.ProjectStatusActive,
				Teams:       make([]model.ID, 0),
				Documents:   make([]model.ID, 0),
				Issues:      make([]model.ID, 0),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update project delete by key cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					byProjectKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					byProjectKeyCmd := new(redis.StringSliceCmd)
					byProjectKeyCmd.SetVal([]string{byProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byProjectKey).Return(byProjectKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byProjectKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				patch: map[string]any{
					"name":        "updated project",
					"description": "updated description",
				},
			},
			want: &model.Project{
				ID:          model.MustNewID(model.ResourceTypeProject),
				Key:         "PROJ",
				Name:        "test project",
				Description: "test description",
				Logo:        "https://example.com/logo.png",
				Status:      model.ProjectStatusActive,
				Teams:       make([]model.ID, 0),
				Documents:   make([]model.ID, 0),
				Issues:      make([]model.ID, 0),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedProjectRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.ProjectRepository
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
			name: "delete project success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					namespacesKeyCmd := new(redis.StringSliceCmd)
					namespacesKeyCmd.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, byKey).Return(byKeyCmd)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
		},
		{
			name: "delete project with project deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					namespacesKeyCmd := new(redis.StringSliceCmd)
					namespacesKeyCmd.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, byKey).Return(byKeyCmd)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrProjectDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrProjectDelete,
		},
		{
			name: "delete project with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.ProjectRepository {
					repo := mock.NewProjectRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete project cache get all error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, byKey).Return(byKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)
					cacheRepo.EXPECT().Delete(ctx, byKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete project cache by key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byKey).Return(byKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete project cache by namespaces key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					namespacesKeyCmd := new(redis.StringSliceCmd)
					namespacesKeyCmd.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, byKey).Return(byKeyCmd)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.ProjectRepository {
					return mock.NewProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
