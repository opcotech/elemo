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

func TestCachedTodoRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) []repository.RedisRepositoryOption
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) repository.TodoRepository
	}
	type args struct {
		ctx  context.Context
		todo repository.CreateTodoOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new todo",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) []repository.RedisRepositoryOption {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Create(ctx, todo).Return(&repository.Todo{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				todo: repository.CreateTodoOpts{
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new todo with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) []repository.RedisRepositoryOption {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Create(ctx, todo).Return(nil, repository.ErrTodoCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				todo: repository.CreateTodoOpts{
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrTodoCreate,
		},
		{
			name: "add new todo get by owner cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo repository.CreateTodoOpts) []repository.RedisRepositoryOption {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateTodoOpts) repository.TodoRepository {
					return mockrepo.NewMockTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				todo: repository.CreateTodoOpts{
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedTodoRepository {
				r, err := repository.NewCachedTodoRepository(
					tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.todo),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.todo)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			_, err := r.Create(tt.args.ctx, tt.args.todo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedTodoRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) repository.TodoRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Todo
		wantErr error
	}{
		{
			name: "get uncached todo",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			want: func(id model.ID) *repository.Todo {
				return &repository.Todo{
					ID:          id,
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get cached todo",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**repository.Todo); ok {
							*ptr = todo
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Todo) repository.TodoRepository {
					return mockrepo.NewMockTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			want: func(id model.ID) *repository.Todo {
				return &repository.Todo{
					ID:          id,
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get uncached todo error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached todo error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Todo) repository.TodoRepository {
					return mockrepo.NewMockTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached todo cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
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
			var want *repository.Todo
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedTodoRepository {
				r, err := repository.NewCachedTodoRepository(
					tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedTodoRepository_GetByOwner(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) []repository.RedisRepositoryOption
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) repository.TodoRepository
	}
	type args struct {
		ctx       context.Context
		owner     model.ID
		offset    int
		limit     int
		completed *bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.Todo
		wantErr error
	}{
		{
			name: "get uncached todos",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", owner.String(), "", limit, completed)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Todo]{Items: todos},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().ListByOwner(ctx, owner, repository.CursorPage{Size: limit}, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*repository.Todo{
				{
					ID:          model.MustNewID(model.ResourceTypeTodo),
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeTodo),
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "get cached todos",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", owner.String(), "", limit, completed)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*repository.Page[*repository.Todo]); ok {
							*ptr = repository.Page[*repository.Todo]{Items: todos}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ *bool, _ []*repository.Todo) repository.TodoRepository {
					return mockrepo.NewMockTodoRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*repository.Todo{
				{
					ID:          model.MustNewID(model.ResourceTypeTodo),
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeTodo),
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "get uncached todos error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, _ []*repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", owner.String(), "", limit, completed)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, _ []*repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().ListByOwner(ctx, owner, repository.CursorPage{Size: limit}, completed).Return(repository.Page[*repository.Todo]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get todos cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, _ []*repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", owner.String(), "", limit, completed)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ *bool, _ []*repository.Todo) repository.TodoRepository {
					return mockrepo.NewMockTodoRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached todos cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", owner.String(), "", limit, completed)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Todo]{Items: todos},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, _, limit int, completed *bool, todos []*repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().ListByOwner(ctx, owner, repository.CursorPage{Size: limit}, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
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
			r := func() *repository.RedisCachedTodoRepository {
				r, err := repository.NewCachedTodoRepository(
					tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.owner, tt.args.offset, testPageSize(tt.args.limit), tt.args.completed, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.owner, tt.args.offset, testPageSize(tt.args.limit), tt.args.completed, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListByOwner(tt.args.ctx, tt.args.owner, repository.CursorPage{Size: testPageSize(tt.args.limit)}, tt.args.completed)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedTodoRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch repository.UpdateTodoOpts, todo *repository.Todo) repository.TodoRepository
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		patch repository.UpdateTodoOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Todo
		wantErr error
	}{
		{
			name: "update todo",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", todo.OwnedBy.String(), "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch repository.UpdateTodoOpts, todo *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: repository.UpdateTodoOpts{
					Title:       optional.Some("updated todo"),
					Description: optional.Some("updated description"),
				},
			},
			want: &repository.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "update todo with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Todo) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(mocktrace.NewMockTracer(ctrl)),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch repository.UpdateTodoOpts, _ *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: repository.UpdateTodoOpts{
					Title:       optional.Some("updated todo"),
					Description: optional.Some("updated description"),
				},
			},
			want: &repository.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update todo set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())

					dbClient := mockrepo.NewMockUniversalClient(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch repository.UpdateTodoOpts, todo *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: repository.UpdateTodoOpts{
					Title:       optional.Some("updated todo"),
					Description: optional.Some("updated description"),
				},
			},
			want: &repository.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update todo delete get by owner cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", todo.OwnedBy.String(), "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch repository.UpdateTodoOpts, todo *repository.Todo) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: repository.UpdateTodoOpts{
					Title:       optional.Some("updated todo"),
					Description: optional.Some("updated description"),
				},
			},
			want: &repository.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
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

			r := func() *repository.RedisCachedTodoRepository {
				r, err := repository.NewCachedTodoRepository(
					tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedTodoRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.TodoRepository
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
			name: "delete todo success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String()+"*")
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
		},
		{
			name: "delete todo with todo deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String()+"*")
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrTodoDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: repository.ErrTodoDelete,
		},
		{
			name: "delete todo with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String()+"*")

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.TodoRepository {
					repo := mockrepo.NewMockTodoRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete todo with get by owner cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String()+"*")
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.TodoRepository {
					return mockrepo.NewMockTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
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
			r := func() *repository.RedisCachedTodoRepository {
				r, err := repository.NewCachedTodoRepository(
					tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTodoGetQueryCompileUsesDetailProjection(t *testing.T) {
	t.Parallel()

	plan, err := repository.CompileQuery(repository.TodoGetQuery{
		ID:         model.MustNewID(model.ResourceTypeTodo),
		Projection: repository.TodoDetailProjection(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, plan.Root.Cypher)
	require.NotEmpty(t, plan.Fingerprint())
}

func TestTodoListByOwnerQueryCompileUsesListProjection(t *testing.T) {
	t.Parallel()

	plan, err := repository.CompileQuery(repository.TodoListByOwnerQuery{
		OwnerID:    model.MustNewID(model.ResourceTypeUser),
		Page:       repository.CursorPage{Size: 10},
		Order:      repository.SortDirectionDesc,
		Projection: repository.TodoListProjection(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, plan.Root.Cypher)
	require.NotEmpty(t, plan.Fingerprint())
}
