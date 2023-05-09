package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	testMock "github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedProjectRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, namespace model.ID, project *model.Project) *baseRepository
		projectRepo func(ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository
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
				cacheRepo: func(ctx context.Context, namespace model.ID, project *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Create", ctx, namespace, project).Return(nil)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, project *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Create", ctx, namespace, project).Return(repository.ErrProjectCreate)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, project *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, project *model.Project) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, project *model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.namespace, tt.args.project),
				projectRepo: tt.fields.projectRepo(tt.args.ctx, tt.args.namespace, tt.args.project),
			}
			err := r.Create(tt.args.ctx, tt.args.namespace, tt.args.project)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedProjectRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, id model.ID, project *model.Project) *baseRepository
		projectRepo func(ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Get", ctx, id).Return(project, nil)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(project, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Get", ctx, id).Return(project, nil)
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
			var want *model.Project
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				projectRepo: tt.fields.projectRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_GetByKey(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, key string, project *model.Project) *baseRepository
		projectRepo func(ctx context.Context, key string, project *model.Project) repository.ProjectRepository
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
				cacheRepo: func(ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("GetByKey", ctx, projectKey).Return(project, nil)
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
				cacheRepo: func(ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(project, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("GetByKey", ctx, projectKey).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, projectKey string, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", projectKey)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, projectKey string, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("GetByKey", ctx, projectKey).Return(project, nil)
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
			var want *model.Project
			if tt.want != nil {
				want = tt.want(tt.args.key)
			}

			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.key, want),
				projectRepo: tt.fields.projectRepo(tt.args.ctx, tt.args.key, want),
			}
			got, err := r.GetByKey(tt.args.ctx, tt.args.key)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository
		projectRepo func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository
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
				cacheRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: projects,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("GetAll", ctx, namespace, offset, limit).Return(projects, nil)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(projects, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("GetAll", ctx, namespace, offset, limit).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespace.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: projects,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, namespace model.ID, offset, limit int, projects []*model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("GetAll", ctx, namespace, offset, limit).Return(projects, nil)
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
			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.namespace, tt.args.offset, tt.args.limit, tt.want),
				projectRepo: tt.fields.projectRepo(tt.args.ctx, tt.args.namespace, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.namespace, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedProjectRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, id model.ID, project *model.Project) *baseRepository
		projectRepo func(ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byProjectKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byProjectKeyCmd := new(redis.StringSliceCmd)
					byProjectKeyCmd.SetVal([]string{byProjectKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Keys", ctx, byProjectKey).Return(byProjectKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, byProjectKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Update", ctx, id, patch).Return(project, nil)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(testMock.CacheRepository),
						tracer: new(testMock.Tracer),
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Update", ctx, id, patch).Return(project, nil)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byProjectKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byProjectKeyCmd := new(redis.StringSliceCmd)
					byProjectKeyCmd.SetVal([]string{byProjectKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Keys", ctx, byProjectKey).Return(byProjectKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, byProjectKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(errors.New("error"))
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Update", ctx, id, patch).Return(project, nil)
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
				cacheRepo: func(ctx context.Context, id model.ID, project *model.Project) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					byProjectKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					byProjectKeyCmd := new(redis.StringSliceCmd)
					byProjectKeyCmd.SetVal([]string{byProjectKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byProjectKey).Return(byProjectKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, byProjectKey).Return(errors.New("error"))
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: project,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID, patch map[string]any, project *model.Project) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Update", ctx, id, patch).Return(project, nil)
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

			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				projectRepo: tt.fields.projectRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
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
		cacheRepo   func(ctx context.Context, id model.ID) *baseRepository
		projectRepo func(ctx context.Context, id model.ID) repository.ProjectRepository
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, byKey).Return(byKeyCmd)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, byKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, byKey).Return(byKeyCmd)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, byKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrProjectDelete)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					repo := new(testMock.ProjectRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
			name: "delete project cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, byKey).Return(byKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)
					cacheRepo.On("Delete", ctx, byKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, byKey).Return(byKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)
					cacheRepo.On("Delete", ctx, byKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
					byKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", id.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, byKey).Return(byKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, byKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, byKey).Return(byKeyCmd)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, byKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				projectRepo: func(ctx context.Context, id model.ID) repository.ProjectRepository {
					return new(testMock.ProjectRepository)
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
			r := &CachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				projectRepo: tt.fields.projectRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
