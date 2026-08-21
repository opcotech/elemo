package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func testProject(id model.ID, proj ProjectProjection) *Project {
	project := &Project{
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

func mustProjectGetKey(t *testing.T, id model.ID, proj ProjectProjection) string {
	t.Helper()
	key, err := projectGetCacheKey(id, proj)
	require.NoError(t, err)
	return key
}

func mustProjectGetByKeyKey(t *testing.T, key string, proj ProjectProjection) string {
	t.Helper()
	cacheKey, err := projectGetByKeyCacheKey(key, proj)
	require.NoError(t, err)
	return cacheKey
}

func mustProjectListKey(t *testing.T, namespaceID model.ID, page CursorPage, proj ProjectProjection) string {
	t.Helper()
	key, err := projectListCacheKey(namespaceID, model.MustNewNilID(model.ResourceTypeUser), nil, page, proj)
	require.NoError(t, err)
	return key
}

func TestCachedProjectRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context) *redisBaseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateProjectOpts) ProjectRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateProjectOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context) *redisBaseRepository {
					projectListPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "List", "*")
					namespacePattern := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					projectListCmd := new(redis.StringSliceCmd)
					projectListCmd.SetVal([]string{projectListPattern})
					namespaceCmd := new(redis.StringSliceCmd)
					namespaceCmd.SetVal([]string{namespacePattern})

					client := mock.NewUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, projectListPattern).Return(projectListCmd)
					client.EXPECT().Keys(ctx, namespacePattern).Return(namespaceCmd)

					db, err := NewRedisDatabase(WithRedisClient(client))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(3)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, projectListPattern).Return(nil)
					backend.EXPECT().Delete(ctx, namespacePattern).Return(nil)
					backend.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).Times(3)
					backend.EXPECT().Set(gomock.AssignableToTypeOf(&cache.Item{})).DoAndReturn(func(item *cache.Item) error {
						require.Equal(t, int64(1), item.Value)
						return nil
					}).Times(3)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateProjectOpts) ProjectRepository {
					repo := NewMockProjectRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&Project{ID: model.MustNewID(model.ResourceTypeProject)}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateProjectOpts{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context) *redisBaseRepository {
					projectListPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "List", "*")
					projectListCmd := new(redis.StringSliceCmd)
					projectListCmd.SetVal([]string{projectListPattern})

					client := mock.NewUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, projectListPattern).Return(projectListCmd)

					db, err := NewRedisDatabase(WithRedisClient(client))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, projectListPattern).Return(ErrCacheDelete)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateProjectOpts) ProjectRepository {
					return NewMockProjectRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateProjectOpts{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCacheDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := &RedisCachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.opts),
			}

			_, err := repo.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedProjectRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, key string, project *Project) *redisBaseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, proj ProjectProjection, project *Project) ProjectRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		proj ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Project
		wantErr error
	}{
		{
			name: "reads uncached project with projection",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, project *Project) *redisBaseRepository {
					db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: project}).Return(nil)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, proj ProjectProjection, project *Project) ProjectRepository {
					repo := NewMockProjectRepository(ctrl)
					repo.EXPECT().Get(ctx, id, proj).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeProject),
				proj: ProjectDetailProjection(),
			},
		},
		{
			name: "returns cached project",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, project *Project) *redisBaseRepository {
					db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						*(dst.(**Project)) = project
					}).Return(nil)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ ProjectProjection, _ *Project) ProjectRepository {
					return NewMockProjectRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeProject),
				proj: ProjectListProjection(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			want := testProject(tt.args.id, tt.args.proj)
			key := mustProjectGetKey(t, tt.args.id, tt.args.proj)

			repo := &RedisCachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, key, want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.proj, want),
			}

			got, err := repo.Get(tt.args.ctx, tt.args.id, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_GetByKey(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, cacheKey string, project *Project) *redisBaseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, key string, proj ProjectProjection, project *Project) ProjectRepository
	}
	type args struct {
		ctx  context.Context
		key  string
		proj ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Project
		wantErr error
	}{
		{
			name: "reads uncached project by key",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, cacheKey string, project *Project) *redisBaseRepository {
					db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, cacheKey, gomock.Any()).Return(cache.ErrCacheMiss)
					backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: cacheKey, Value: project}).Return(nil)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, proj ProjectProjection, project *Project) ProjectRepository {
					repo := NewMockProjectRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, key, proj).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				key:  "PROJ",
				proj: ProjectDetailProjection(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			want := testProject(model.MustNewID(model.ResourceTypeProject), tt.args.proj)
			want.Key = tt.args.key
			cacheKey := mustProjectGetByKeyKey(t, tt.args.key, tt.args.proj)

			repo := &RedisCachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, cacheKey, want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.key, tt.args.proj, want),
			}

			got, err := repo.GetByKey(tt.args.ctx, tt.args.key, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_List(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*Project]) *redisBaseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, page CursorPage, proj ProjectProjection, want Page[*Project]) ProjectRepository
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		page        CursorPage
		proj        ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page[*Project]
		wantErr error
	}{
		{
			name: "reads uncached list page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, want Page[*Project]) *redisBaseRepository {
					db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: want}).Return(nil)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, page CursorPage, proj ProjectProjection, want Page[*Project]) ProjectRepository {
					repo := NewMockProjectRepository(ctrl)
					repo.EXPECT().List(ctx, namespaceID, model.MustNewNilID(model.ResourceTypeUser), nil, page, proj).Return(want, nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: model.MustNewID(model.ResourceTypeNamespace),
				page:        CursorPage{Size: 10},
				proj:        ProjectListProjection(),
			},
			want: Page[*Project]{
				Items: []*Project{
					testProject(model.MustNewID(model.ResourceTypeProject), ProjectListProjection()),
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			key := mustProjectListKey(t, tt.args.namespaceID, tt.args.page, tt.args.proj)

			repo := &RedisCachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, key, tt.want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.page, tt.args.proj, tt.want),
			}

			got, err := repo.List(tt.args.ctx, tt.args.namespaceID, model.MustNewNilID(model.ResourceTypeUser), nil, tt.args.page, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedProjectRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, detailKey string, project *Project) *redisBaseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateProjectOpts, proj ProjectProjection, project *Project) ProjectRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateProjectOpts
		proj ProjectProjection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Project
		wantErr error
	}{
		{
			name: "updates cache and clears related pages",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, detailKey string, _ *Project) *redisBaseRepository {
					byKeyPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "GetByKey", "*")
					listPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "List", "*")
					namespacePattern := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKeyPattern})
					listCmd := new(redis.StringSliceCmd)
					listCmd.SetVal([]string{listPattern})
					namespaceCmd := new(redis.StringSliceCmd)
					namespaceCmd.SetVal([]string{namespacePattern})
					client := mock.NewUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, byKeyPattern).Return(byKeyCmd)
					client.EXPECT().Keys(ctx, listPattern).Return(listCmd)
					client.EXPECT().Keys(ctx, namespacePattern).Return(namespaceCmd)

					db, err := NewRedisDatabase(WithRedisClient(client))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					backend := mock.NewCacheBackend(ctrl)
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

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateProjectOpts, proj ProjectProjection, project *Project) ProjectRepository {
					repo := NewMockProjectRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, proj).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeProject),
				opts: UpdateProjectOpts{
					Name: optional.Some("updated"),
				},
				proj: ProjectDetailProjection(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			want := testProject(tt.args.id, ProjectDetailProjection())
			detailKey := mustProjectGetKey(t, tt.args.id, ProjectDetailProjection())

			repo := &RedisCachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, detailKey, want),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.args.proj, want),
			}

			got, err := repo.Update(tt.args.ctx, tt.args.id, tt.args.opts, tt.args.proj)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedProjectRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		projectRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) ProjectRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					getPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "Get", id.String(), "*")
					byKeyPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "GetByKey", "*")
					listPattern := composeCacheKey(model.ResourceTypeProject.String(), "*", "List", "*")
					namespacePattern := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					getCmd := new(redis.StringSliceCmd)
					getCmd.SetVal([]string{getPattern})
					byKeyCmd := new(redis.StringSliceCmd)
					byKeyCmd.SetVal([]string{byKeyPattern})
					listCmd := new(redis.StringSliceCmd)
					listCmd.SetVal([]string{listPattern})
					namespaceCmd := new(redis.StringSliceCmd)
					namespaceCmd.SetVal([]string{namespacePattern})
					client := mock.NewUniversalClient(ctrl)
					client.EXPECT().Keys(ctx, getPattern).Return(getCmd)
					client.EXPECT().Keys(ctx, byKeyPattern).Return(byKeyCmd)
					client.EXPECT().Keys(ctx, listPattern).Return(listCmd)
					client.EXPECT().Keys(ctx, namespacePattern).Return(namespaceCmd)

					db, err := NewRedisDatabase(WithRedisClient(client))
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					backend := mock.NewCacheBackend(ctrl)
					backend.EXPECT().Delete(ctx, getPattern).Return(nil)
					backend.EXPECT().Delete(ctx, byKeyPattern).Return(nil)
					backend.EXPECT().Delete(ctx, listPattern).Return(nil)
					backend.EXPECT().Delete(ctx, namespacePattern).Return(nil)
					backend.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).Times(2)
					backend.EXPECT().Set(gomock.AssignableToTypeOf(&cache.Item{})).DoAndReturn(func(item *cache.Item) error {
						require.Equal(t, int64(1), item.Value)
						return nil
					}).Times(2)

					return &redisBaseRepository{db: db, cache: backend, tracer: tracer, logger: mock.NewMockLogger(ctrl)}
				},
				projectRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) ProjectRepository {
					repo := NewMockProjectRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
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
			defer ctrl.Finish()

			repo := &RedisCachedProjectRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				projectRepo: tt.fields.projectRepo(ctrl, tt.args.ctx, tt.args.id),
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedProjectRepository_ProjectionAffectsCacheKey(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeProject)
	detailKey := mustProjectGetKey(t, id, ProjectDetailProjection())
	listKey := mustProjectGetKey(t, id, ProjectListProjection())

	assert.NotEqual(t, detailKey, listKey)
}
