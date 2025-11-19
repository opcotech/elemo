package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedTodoRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *redisBaseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) TodoRepository
	}
	type args struct {
		ctx  context.Context
		todo *model.Todo
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *redisBaseRepository {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Create(ctx, todo).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				todo: &model.Todo{
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
			name: "add new todo with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *redisBaseRepository {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Create(ctx, todo).Return(ErrTodoCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				todo: &model.Todo{
					ID:          model.MustNewID(model.ResourceTypeTodo),
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrTodoCreate,
		},
		{
			name: "add new todo get by owner cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *redisBaseRepository {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ *model.Todo) TodoRepository {
					return mock.NewTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				todo: &model.Todo{
					ID:          model.MustNewID(model.ResourceTypeTodo),
					Title:       "test title",
					Description: "test description",
					Priority:    model.TodoPriorityNormal,
					Completed:   false,
					OwnedBy:     model.MustNewID(model.ResourceTypeUser),
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedTodoRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.todo),
				todoRepo:  tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.todo),
			}
			err := r.Create(tt.args.ctx, tt.args.todo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedTodoRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) TodoRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Todo
		wantErr error
	}{
		{
			name: "get uncached todo",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			want: func(id model.ID) *model.Todo {
				return &model.Todo{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**model.Todo); ok {
							*ptr = todo
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Todo) TodoRepository {
					return mock.NewTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			want: func(id model.ID) *model.Todo {
				return &model.Todo{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached todo error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Todo) TodoRepository {
					return mock.NewTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached todo cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *model.Todo
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedTodoRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				todoRepo:  tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedTodoRepository_GetByOwner(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *redisBaseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) TodoRepository
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
		want    []*model.Todo
		wantErr error
	}{
		{
			name: "get uncached todos",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todos,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().GetByOwner(ctx, owner, offset, limit, completed).Return(todos, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*model.Todo{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if listPtr, ok := dst.(*[]*model.Todo); ok {
							*listPtr = todos
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ *bool, _ []*model.Todo) TodoRepository {
					return mock.NewTodoRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*model.Todo{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, _ []*model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().GetByOwner(ctx, owner, offset, limit, completed).Return(todos, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get todos cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, _ []*model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ *bool, _ []*model.Todo) TodoRepository {
					return mock.NewTodoRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached todos cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todos,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().GetByOwner(ctx, owner, offset, limit, completed).Return(todos, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				owner:  model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedTodoRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.owner, tt.args.offset, tt.args.limit, tt.args.completed, tt.want),
				todoRepo:  tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.owner, tt.args.offset, tt.args.limit, tt.args.completed, tt.want),
			}
			got, err := r.GetByOwner(tt.args.ctx, tt.args.owner, tt.args.offset, tt.args.limit, tt.args.completed)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedTodoRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) TodoRepository
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
		want    *model.Todo
		wantErr error
	}{
		{
			name: "update todo",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: map[string]any{
					"title":       "updated todo",
					"description": "updated description",
				},
			},
			want: &model.Todo{
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Todo) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: map[string]any{
					"title":       "updated todo",
					"description": "updated description",
				},
			},
			want: &model.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update todo set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: map[string]any{
					"title":       "updated todo",
					"description": "updated description",
				},
			},
			want: &model.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update todo delete get by owner cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(todo, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
				patch: map[string]any{
					"title":       "updated todo",
					"description": "updated description",
				},
			},
			want: &model.Todo{
				ID:          model.MustNewID(model.ResourceTypeTodo),
				Title:       "test title",
				Description: "test description",
				Priority:    model.TodoPriorityNormal,
				Completed:   false,
				OwnedBy:     model.MustNewID(model.ResourceTypeUser),
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &RedisCachedTodoRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				todoRepo:  tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID) TodoRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrTodoDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: ErrTodoDelete,
		},
		{
			name: "delete todo with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete todo with get by owner cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) TodoRepository {
					return mock.NewTodoRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeTodo),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedTodoRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				todoRepo:  tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
