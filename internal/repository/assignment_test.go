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
)

func TestCachedAssignmentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) []repository.RedisRepositoryOption
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) repository.AssignmentRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateAssignmentOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) []repository.RedisRepositoryOption {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String())
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", opts.User.String())
					key3 := composeCacheKey(model.ResourceTypeIssue.String(), opts.Resource.String())

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String(), "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", opts.User.String(), "*", "*")
					resourceKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyResult := new(redis.StringSliceCmd)
					byResourceKeyResult.SetVal([]string{key1})

					byUserKeyResult := new(redis.StringSliceCmd)
					byUserKeyResult.SetVal([]string{key2})

					resourceKeyResult := new(redis.StringSliceCmd)
					resourceKeyResult.SetVal([]string{key3})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyResult)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyResult)
					dbClient.EXPECT().Keys(ctx, resourceKey).Return(resourceKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key1).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, key2).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, key3).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Assignment{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAssignmentOpts{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeIssue),
				},
			},
		},
		{
			name: "create new unknown resource assignment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) []repository.RedisRepositoryOption {
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String(), "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", opts.User.String(), "*", "*")

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(new(redis.StringSliceCmd))
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(new(redis.StringSliceCmd))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateAssignmentOpts) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAssignmentOpts{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: repository.ErrUnexpectedCachedResource,
		},
		{
			name: "create new assignment with by resource cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) []repository.RedisRepositoryOption {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String(), "1")
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String(), "2")

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String(), "*", "*")

					keysCmd := new(redis.StringSliceCmd)
					keysCmd.SetVal([]string{key1, key2})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(keysCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key1).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateAssignmentOpts) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAssignmentOpts{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new assignment with by user cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAssignmentOpts) []repository.RedisRepositoryOption {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", opts.Resource.String(), "1")
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", opts.Resource.String(), "2")

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", opts.Resource.String(), "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", opts.User.String(), "*", "*")

					keysCmd := new(redis.StringSliceCmd)
					keysCmd.SetVal([]string{key1, key2})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(new(redis.StringSliceCmd))
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(keysCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key1).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateAssignmentOpts) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAssignmentOpts{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedAssignmentRepository {
				r, err := repository.NewCachedAssignmentRepository(
					tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.opts),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedAssignmentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) []repository.RedisRepositoryOption
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) repository.AssignmentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), projectionCacheValue(repository.AssignmentDetailProjection()))

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
						Value: assignment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.AssignmentDetailProjection()).Return(assignment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			want: func(id model.ID) *repository.Assignment {
				return &repository.Assignment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), projectionCacheValue(repository.AssignmentDetailProjection()))

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
						if ptr, ok := dst.(**repository.Assignment); ok {
							*ptr = assignment
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Assignment) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			want: func(id model.ID) *repository.Assignment {
				return &repository.Assignment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), projectionCacheValue(repository.AssignmentDetailProjection()))

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
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.AssignmentDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached assignment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), projectionCacheValue(repository.AssignmentDetailProjection()))

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
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Assignment) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached assignment cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), projectionCacheValue(repository.AssignmentDetailProjection()))

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
						Value: assignment,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, assignment *repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.AssignmentDetailProjection()).Return(assignment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *repository.Assignment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedAssignmentRepository {
				r, err := repository.NewCachedAssignmentRepository(
					tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.AssignmentDetailProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedAssignmentRepository_GetByUser(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) repository.AssignmentRepository
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
		want    []*repository.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Assignment]{Items: assignments},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().ListByUser(ctx, userID, repository.CursorPage{Size: limit}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: assignments}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Assignment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.Assignment]); ok {
							*ptr = repository.Page[*repository.Assignment]{Items: assignments}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Assignment) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Assignment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().ListByUser(ctx, userID, repository.CursorPage{Size: limit}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get assignments cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
				assignmentRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Assignment) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached assignments cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Assignment]{Items: assignments},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().ListByUser(ctx, userID, repository.CursorPage{Size: limit}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: assignments}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedAssignmentRepository {
				r, err := repository.NewCachedAssignmentRepository(
					tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListByUser(tt.args.ctx, tt.args.userID, repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.AssignmentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedAssignmentRepository_GetByResource(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) repository.AssignmentRepository
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
		want    []*repository.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Assignment]{Items: assignments},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().ListByResource(ctx, userID, repository.CursorPage{Size: limit}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: assignments}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Assignment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.Assignment]); ok {
							*ptr = repository.Page[*repository.Assignment]{Items: assignments}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Assignment) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Assignment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().ListByResource(ctx, userID, repository.CursorPage{Size: limit}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get assignments cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Assignment) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached assignments cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", userID.String(), projectionCacheValue(repository.AssignmentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Assignment]{Items: assignments},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, assignments []*repository.Assignment) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().ListByResource(ctx, userID, repository.CursorPage{Size: limit}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: assignments}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedAssignmentRepository {
				r, err := repository.NewCachedAssignmentRepository(
					tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListByResource(tt.args.ctx, tt.args.userID, repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.AssignmentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedAssignmentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		assignmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AssignmentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), "*")
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", "*", "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", "*", "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), "*")
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", "*", "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", "*", "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AssignmentRepository {
					repo := mockrepo.NewMockAssignmentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrAssignmentDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrAssignmentDelete,
		},
		{
			name: "delete assignment with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), "*")

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
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete assignment cache by resource key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), "*")
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", "*", "*", "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete assignment cache by user key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), "*")
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", "*", "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", "*", "*", "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(repository.ErrCacheDelete)
					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete assignment cache by issues key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), "*")
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByResource", "*", "*", "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "ListByUser", "*", "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyCmd := new(redis.StringSliceCmd)
					byResourceKeyCmd.SetVal([]string{byResourceKey})

					byUserKeyCmd := new(redis.StringSliceCmd)
					byUserKeyCmd.SetVal([]string{byUserKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byResourceKey).Return(byResourceKeyCmd)
					dbClient.EXPECT().Keys(ctx, byUserKey).Return(byUserKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byResourceKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byUserKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				assignmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AssignmentRepository {
					return mockrepo.NewMockAssignmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedAssignmentRepository {
				r, err := repository.NewCachedAssignmentRepository(
					tt.fields.assignmentRepo(ctrl, tt.args.ctx, tt.args.id),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
