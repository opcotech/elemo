package redis

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedTodoRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) repository.TodoRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseRepository {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseRepository {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Create(ctx, todo).Return(repository.ErrTodoCreate)
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
			wantErr: repository.ErrTodoCreate,
		},
		{
			name: "add new todo get by owner cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseRepository {
					getByOwner := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerResult := new(redis.StringSliceCmd)
					getByOwnerResult.SetVal([]string{getByOwner})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwner).Return(getByOwnerResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwner).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ *model.Todo) repository.TodoRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedTodoRepository{
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) repository.TodoRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**model.Todo); ok {
							*ptr = todo
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Todo) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Todo) repository.TodoRepository {
					return mock.NewTodoRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
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
			var want *model.Todo
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedTodoRepository{
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *baseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) repository.TodoRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todos,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if listPtr, ok := dst.(*[]*model.Todo); ok {
							*listPtr = todos
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ *bool, _ []*model.Todo) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, _ []*model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().GetByOwner(ctx, owner, offset, limit, completed).Return(todos, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, _ []*model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ *bool, _ []*model.Todo) repository.TodoRepository {
					return mock.NewTodoRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", owner.String(), offset, limit, completed)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todos,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, offset, limit int, completed *bool, todos []*model.Todo) repository.TodoRepository {
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
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedTodoRepository{
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository
		todoRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) repository.TodoRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Todo) *baseRepository {
					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Todo) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)
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
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update todo set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) repository.TodoRepository {
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update todo delete get by owner cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: todo,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) repository.TodoRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &CachedTodoRepository{
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.TodoRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.TodoRepository {
					repo := mock.NewTodoRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
					getByOwnerKey := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")

					getByOwnerKeyCmd := new(redis.StringSliceCmd)
					getByOwnerKeyCmd.SetVal([]string{getByOwnerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getByOwnerKey).Return(getByOwnerKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByOwnerKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				todoRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.TodoRepository {
					return mock.NewTodoRepository(ctrl)
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
			r := &CachedTodoRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				todoRepo:  tt.fields.todoRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
