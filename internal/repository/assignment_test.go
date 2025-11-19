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

func TestCachedAssignmentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) *redisBaseRepository
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) AssignmentRepository
	}
	type args struct {
		ctx        context.Context
		assignment *model.Assignment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new issue assignment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) *redisBaseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String())
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String())
					key3 := composeCacheKey(model.ResourceTypeIssue.String(), assignment.Resource.String())

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")
					resourceKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyResult := new(redis.StringSliceCmd)
					byResourceKeyResult.SetVal([]string{key1})

					byUserKeyResult := new(redis.StringSliceCmd)
					byUserKeyResult.SetVal([]string{key2})

					resourceKeyResult := new(redis.StringSliceCmd)
					resourceKeyResult.SetVal([]string{key3})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyResult)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyResult)
					dbClient.EXPECT().Keys(ctx, resourceKey).Return(resourceKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key1).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, key2).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, key3).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().Create(ctx, assignment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeIssue),
				},
			},
		},
		{
			name: "create new unknown resource assignment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) *redisBaseRepository {
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(new(redis.StringSliceCmd))
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(new(redis.StringSliceCmd))

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ *model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrUnexpectedCachedResource,
		},
		{
			name: "create new assignment with by resource cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) *redisBaseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "1")
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "2")

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")

					keysCmd := new(redis.StringSliceCmd)
					keysCmd.SetVal([]string{key1, key2})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(keysCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key1).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ *model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create new assignment with by user cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, assignment *model.Assignment) *redisBaseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.Resource.String(), "1")
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.Resource.String(), "2")

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")

					keysCmd := new(redis.StringSliceCmd)
					keysCmd.SetVal([]string{key1, key2})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(new(redis.StringSliceCmd))
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(keysCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key1).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ *model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.assignment),
				assignmentRepo: tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.assignment),
			}
			err := r.Create(tt.args.ctx, tt.args.assignment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedAssignmentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) *redisBaseRepository
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) AssignmentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

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
						Value: assignment,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(assignment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			want: func(id model.ID) *model.Assignment {
				return &model.Assignment{
					ID:       id,
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				}
			},
		},
		{
			name: "get cached assignment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

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
						if ptr, ok := dst.(**model.Assignment); ok {
							*ptr = assignment
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			want: func(id model.ID) *model.Assignment {
				return &model.Assignment{
					ID:       id,
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				}
			},
		},
		{
			name: "get uncached assignment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

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
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached assignment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

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
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached assignment cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

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
						Value: assignment,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(assignment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			var want *model.Assignment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				assignmentRepo: tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedAssignmentRepository_GetByUser(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) AssignmentRepository
	}
	type args struct {
		ctx    context.Context
		userID model.ID
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)

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
						Value: assignments,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().GetByUser(ctx, userID, offset, limit).Return(assignments, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Assignment{
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
			},
		},
		{
			name: "get cached assignments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)

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
						if listPtr, ok := dst.(*[]*model.Assignment); ok {
							*listPtr = assignments
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Assignment{
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
			},
		},
		{
			name: "get uncached assignments error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)

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
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().GetByUser(ctx, userID, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get assignments cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)

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
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached assignments cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)

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
						Value: assignments,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().GetByUser(ctx, userID, offset, limit).Return(assignments, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit, tt.want),
				assignmentRepo: tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetByUser(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedAssignmentRepository_GetByResource(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) AssignmentRepository
	}
	type args struct {
		ctx    context.Context
		userID model.ID
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", userID.String(), offset, limit)

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
						Value: assignments,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().GetByResource(ctx, userID, offset, limit).Return(assignments, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Assignment{
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
			},
		},
		{
			name: "get cached assignments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", userID.String(), offset, limit)

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
						if listPtr, ok := dst.(*[]*model.Assignment); ok {
							*listPtr = assignments
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Assignment{
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
				{
					ID:       model.MustNewID(model.ResourceTypeAssignment),
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
			},
		},
		{
			name: "get uncached assignments error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", userID.String(), offset, limit)

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
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().GetByResource(ctx, userID, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get assignments cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", userID.String(), offset, limit)

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
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Assignment) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached assignments cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", userID.String(), offset, limit)

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
						Value: assignments,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, assignments []*model.Assignment) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().GetByResource(ctx, userID, offset, limit).Return(assignments, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit, tt.want),
				assignmentRepo: tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetByResource(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedAssignmentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) AssignmentRepository
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
			name: "delete assignment success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
		},
		{
			name: "delete assignment with assignment deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) AssignmentRepository {
					repo := mock.NewAssignmentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrAssignmentDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrAssignmentDelete,
		},
		{
			name: "delete assignment with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

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
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete assignment cache by resource key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete assignment cache by user key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(ErrCacheDelete)
					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete assignment cache by issues key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) AssignmentRepository {
					return mock.NewAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				assignmentRepo: tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
