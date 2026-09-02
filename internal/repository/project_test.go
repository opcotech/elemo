package repository_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

func testProject(id model.ID, proj repository.ProjectProjection) *repository.Project {
	project := &repository.Project{
		ID:          id,
		Key:         "PROJ",
		Name:        "test project",
		Description: "test description",
		Logo:        "https://example.com/logo.png",
		Status:      model.ProjectStatusActive,
	}
	if proj.Teams {
		project.Teams = []model.ID{}
	}
	if proj.DocumentCount {
		count := int64(0)
		project.DocumentCount = &count
	}
	if proj.IssueCount {
		count := int64(2)
		project.IssueCount = &count
	}
	return project
}

func mustProjectGetKey(t *testing.T, id model.ID, proj repository.ProjectProjection) string {
	t.Helper()
	return mustPlanCacheKey(t, repository.ProjectGetQuery{ID: id, Projection: proj}, model.ResourceTypeProject.String(), "Get", id.String())
}

func mustProjectGetByKeyKey(t *testing.T, namespaceID model.ID, key string, proj repository.ProjectProjection) string {
	t.Helper()
	return mustPlanCacheKey(t, repository.ProjectGetByKeyQuery{NamespaceID: namespaceID, Key: key, Projection: proj}, model.ResourceTypeProject.String(), "GetByKey", namespaceID.String(), key)
}

func mustProjectListKey(t *testing.T, namespaceID model.ID, page repository.CursorPage, proj repository.ProjectProjection) string {
	t.Helper()
	return mustPlanCacheKey(t, repository.ProjectListQuery{
		NamespaceID: namespaceID,
		ActorID:     model.MustNewNilID(model.ResourceTypeUser),
		Page:        page,
		Order:       repository.SortDirectionDesc,
		Projection:  proj,
	}, model.ResourceTypeProject.String(), "ListForNamespace", namespaceID.String())
}

func TestCachedProjectRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context) []repository.RedisRepositoryOption
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateProjectOpts) repository.ProjectRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateProjectOpts
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "clears list and namespace caches before create",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context) []repository.RedisRepositoryOption {
					projectListPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "ListForNamespace", "*")
					namespacePattern := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					projectListCmd := new(redis.StringSliceCmd)
					projectListCmd.SetVal([]string{projectListPattern})
					namespaceCmd := new(redis.StringSliceCmd)
					namespaceCmd.SetVal([]string{namespacePattern})

					client := mockrepo.NewMockUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, projectListPattern).Return(projectListCmd)
					client.EXPECT().Keys(ctx, namespacePattern).Return(namespaceCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(3)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, projectListPattern).Return(nil)
					backend.EXPECT().Delete(ctx, namespacePattern).Return(nil)
					backend.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).Times(3)
					backend.EXPECT().Set(gomock.AssignableToTypeOf(&cache.Item{})).DoAndReturn(func(item *cache.Item) error {
						require.Equal(t, int64(1), item.Value)
						return nil
					}).Times(3)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateProjectOpts) repository.ProjectRepository {
					repo := mockrepo.NewMockProjectRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Project{ID: model.MustNewID(model.ResourceTypeProject)}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateProjectOpts{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					Key:         "PROJ",
					Name:        "test project",
					Description: "test description",
					Logo:        "https://example.com/logo.png",
					Status:      model.ProjectStatusActive,
				},
			},
		},
		{
			name: "returns cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context) []repository.RedisRepositoryOption {
					projectListPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "ListForNamespace", "*")
					projectListCmd := new(redis.StringSliceCmd)
					projectListCmd.SetVal([]string{projectListPattern})

					client := mockrepo.NewMockUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, projectListPattern).Return(projectListCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, projectListPattern).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateProjectOpts) repository.ProjectRepository {
					return mockrepo.NewMockProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateProjectOpts{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			repo := func() *repository.RedisCachedProjectRepository {
				r, err := repository.NewCachedProjectRepository(
					tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.opts),
					tt.fields.cacheRepo(ctrl, tt.args.ctx)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()

			_, err := repo.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedProjectRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, key string, project *repository.Project) []repository.RedisRepositoryOption
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, proj repository.ProjectProjection, project *repository.Project) repository.ProjectRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		proj repository.ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Project
		wantErr error
	}{
		{
			name: "reads uncached project with projection",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, project *repository.Project) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: project}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, proj repository.ProjectProjection, project *repository.Project) repository.ProjectRepository {
					repo := mockrepo.NewMockProjectRepository(ctrl)
					repo.EXPECT().Get(ctx, id, proj).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeProject),
				proj: repository.ProjectDetailProjection(),
			},
		},
		{
			name: "returns cached project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, project *repository.Project) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						*(dst.(**repository.Project)) = project
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ repository.ProjectProjection, _ *repository.Project) repository.ProjectRepository {
					return mockrepo.NewMockProjectRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeProject),
				proj: repository.ProjectListProjection(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			want := testProject(tt.args.id, tt.args.proj)
			key := mustProjectGetKey(t, tt.args.id, tt.args.proj)

			repo := func() *repository.RedisCachedProjectRepository {
				r, err := repository.NewCachedProjectRepository(
					tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.proj, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, key, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()

			got, err := repo.Get(tt.args.ctx, tt.args.id, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_GetByKey(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, cacheKey string, project *repository.Project) []repository.RedisRepositoryOption
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string, proj repository.ProjectProjection, project *repository.Project) repository.ProjectRepository
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		key         string
		proj        repository.ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Project
		wantErr error
	}{
		{
			name: "reads uncached project by key",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, cacheKey string, project *repository.Project) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, cacheKey, gomock.Any()).Return(cache.ErrCacheMiss)
					backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: cacheKey, Value: project}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string, proj repository.ProjectProjection, project *repository.Project) repository.ProjectRepository {
					repo := mockrepo.NewMockProjectRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, namespaceID, key, proj).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: model.MustNewID(model.ResourceTypeNamespace),
				key:         "PROJ",
				proj:        repository.ProjectDetailProjection(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			want := testProject(model.MustNewID(model.ResourceTypeProject), tt.args.proj)
			want.Key = tt.args.key
			cacheKey := mustProjectGetByKeyKey(t, tt.args.namespaceID, tt.args.key, tt.args.proj)

			repo := func() *repository.RedisCachedProjectRepository {
				r, err := repository.NewCachedProjectRepository(
					tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.key, tt.args.proj, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, cacheKey, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()

			got, err := repo.GetByKey(tt.args.ctx, tt.args.namespaceID, tt.args.key, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_List(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, key string, page repository.Page[*repository.Project]) []repository.RedisRepositoryOption
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, page repository.CursorPage, proj repository.ProjectProjection, want repository.Page[*repository.Project]) repository.ProjectRepository
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		page        repository.CursorPage
		proj        repository.ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    repository.Page[*repository.Project]
		wantErr error
	}{
		{
			name: "reads uncached list page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, want repository.Page[*repository.Project]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: want}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, page repository.CursorPage, proj repository.ProjectProjection, want repository.Page[*repository.Project]) repository.ProjectRepository {
					repo := mockrepo.NewMockProjectRepository(ctrl)
					repo.EXPECT().ListForNamespace(ctx, repository.ProjectListQuery{
						NamespaceID: namespaceID,
						ActorID:     model.MustNewNilID(model.ResourceTypeUser),
						Page:        page,
						Order:       repository.SortDirectionDesc,
						Projection:  proj,
					}).Return(want, nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: model.MustNewID(model.ResourceTypeNamespace),
				page:        repository.CursorPage{Size: 10},
				proj:        repository.ProjectListProjection(),
			},
			want: repository.Page[*repository.Project]{
				Items: []*repository.Project{
					testProject(model.MustNewID(model.ResourceTypeProject), repository.ProjectListProjection()),
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			key := mustProjectListKey(t, tt.args.namespaceID, tt.args.page, tt.args.proj)

			repo := func() *repository.RedisCachedProjectRepository {
				r, err := repository.NewCachedProjectRepository(
					tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.page, tt.args.proj, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, key, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()

			got, err := repo.ListForNamespace(tt.args.ctx, repository.ProjectListQuery{
				NamespaceID: tt.args.namespaceID,
				ActorID:     model.MustNewNilID(model.ResourceTypeUser),
				Page:        tt.args.page,
				Order:       repository.SortDirectionDesc,
				Projection:  tt.args.proj,
			})
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedProjectRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, detailKey string, project *repository.Project) []repository.RedisRepositoryOption
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateProjectOpts, proj repository.ProjectProjection, project *repository.Project) repository.ProjectRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateProjectOpts
		proj repository.ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Project
		wantErr error
	}{
		{
			name: "updates cache and clears related pages",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, detailKey string, _ *repository.Project) []repository.RedisRepositoryOption {
					byKeyPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "GetByKey", "*")
					listPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "ListForNamespace", "*")
					namespacePattern := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKeyPattern})
					listCmd := new(redis.StringSliceCmd)
					listCmd.SetVal([]string{listPattern})
					namespaceCmd := new(redis.StringSliceCmd)
					namespaceCmd.SetVal([]string{namespacePattern})
					client := mockrepo.NewMockUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, byKeyPattern).Return(byKeyCmd)
					client.EXPECT().Keys(ctx, listPattern).Return(listCmd)
					client.EXPECT().Keys(ctx, namespacePattern).Return(namespaceCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, byKeyPattern).Return(nil)
					backend.EXPECT().Delete(ctx, listPattern).Return(nil)
					backend.EXPECT().Delete(ctx, namespacePattern).Return(nil)
					backend.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).Times(2)
					backend.EXPECT().Set(gomock.AssignableToTypeOf(&cache.Item{})).DoAndReturn(func(item *cache.Item) error {
						if item.Key == detailKey {
							return nil
						}
						require.Equal(t, int64(1), item.Value)
						return nil
					}).Times(3)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateProjectOpts, proj repository.ProjectProjection, project *repository.Project) repository.ProjectRepository {
					repo := mockrepo.NewMockProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, proj).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				opts: repository.UpdateProjectOpts{
					Name: optional.Some("updated"),
				},
				proj: repository.ProjectDetailProjection(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			want := testProject(tt.args.id, repository.ProjectDetailProjection())
			detailKey := mustProjectGetKey(t, tt.args.id, repository.ProjectDetailProjection())

			repo := func() *repository.RedisCachedProjectRepository {
				r, err := repository.NewCachedProjectRepository(
					tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.args.proj, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, detailKey, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()

			got, err := repo.Update(tt.args.ctx, tt.args.id, tt.args.opts, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
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
			name: "clears get bykey list cross caches",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					getPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "Get", id.String(), "*")
					byKeyPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "GetByKey", "*")
					listPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "ListForNamespace", "*")
					namespacePattern := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getCmd := new(redis.StringSliceCmd)
					getCmd.SetVal([]string{getPattern})
					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKeyPattern})
					listCmd := new(redis.StringSliceCmd)
					listCmd.SetVal([]string{listPattern})
					namespaceCmd := new(redis.StringSliceCmd)
					namespaceCmd.SetVal([]string{namespacePattern})
					client := mockrepo.NewMockUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, getPattern).Return(getCmd)
					client.EXPECT().Keys(ctx, byKeyPattern).Return(byKeyCmd)
					client.EXPECT().Keys(ctx, listPattern).Return(listCmd)
					client.EXPECT().Keys(ctx, namespacePattern).Return(namespaceCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					backend := mockrepo.NewMockCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, getPattern).Return(nil)
					backend.EXPECT().Delete(ctx, byKeyPattern).Return(nil)
					backend.EXPECT().Delete(ctx, listPattern).Return(nil)
					backend.EXPECT().Delete(ctx, namespacePattern).Return(nil)
					backend.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).Times(2)
					backend.EXPECT().Set(gomock.AssignableToTypeOf(&cache.Item{})).DoAndReturn(func(item *cache.Item) error {
						require.Equal(t, int64(1), item.Value)
						return nil
					}).Times(2)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(backend),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.ProjectRepository {
					repo := mockrepo.NewMockProjectRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			repo := func() *repository.RedisCachedProjectRepository {
				r, err := repository.NewCachedProjectRepository(
					tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()

			err := repo.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedProjectRepository_ProjectionAffectsCacheKey(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeProject)
	detailKey := mustProjectGetKey(t, id, repository.ProjectDetailProjection())
	listKey := mustProjectGetKey(t, id, repository.ProjectListProjection())

	assert.NotEqual(t, detailKey, listKey)
}
