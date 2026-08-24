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

func TestCachedCommentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) []repository.RedisRepositoryOption
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) repository.CommentRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateCommentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new comment to an issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Comment{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateCommentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new comment to a document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Comment{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateCommentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeDocument),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new comment with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrCommentCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateCommentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCommentCreate,
		},
		{
			name: "add new comment belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateCommentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateCommentOpts) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateCommentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedCommentRepository {
				r, err := repository.NewCachedCommentRepository(
					tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.opts),
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

func TestCachedCommentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Comment
		wantErr error
	}{
		{
			name: "get uncached comment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())

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
						Value: comment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			want: func(id model.ID) *repository.Comment {
				return &repository.Comment{
					ID:        id,
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get cached comment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())

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
						if ptr, ok := dst.(**repository.Comment); ok {
							*ptr = comment
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Comment) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			want: func(id model.ID) *repository.Comment {
				return &repository.Comment{
					ID:        id,
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get uncached comment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())

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
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached comment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())

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
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Comment) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached comment cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())

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
						Value: comment,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
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
			var want *repository.Comment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedCommentRepository {
				r, err := repository.NewCachedCommentRepository(
					tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedCommentRepository_ListBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) []repository.RedisRepositoryOption
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) repository.CommentRepository
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		offset    int
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.Comment
		wantErr error
	}{
		{
			name: "get uncached comments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", belongsTo.String(), "", limit)

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
						Value: repository.Page[*repository.Comment]{Items: comments},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}).Return(repository.Page[*repository.Comment]{Items: comments}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Comment{
				{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
				{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "get cached comments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", belongsTo.String(), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.Comment]); ok {
							*ptr = repository.Page[*repository.Comment]{Items: comments}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Comment) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Comment{
				{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
				{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "get uncached comments error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", belongsTo.String(), "", limit)

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
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}).Return(repository.Page[*repository.Comment]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get comments cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", belongsTo.String(), "", limit)

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
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Comment) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached comments cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", belongsTo.String(), "", limit)

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
						Value: repository.Page[*repository.Comment]{Items: comments},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, comments []*repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}).Return(repository.Page[*repository.Comment]{Items: comments}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedCommentRepository {
				r, err := repository.NewCachedCommentRepository(
					tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, repository.CursorPage{Size: testPageSize(tt.args.limit)})
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedCommentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateCommentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Comment
		wantErr error
	}{
		{
			name: "update comment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateCommentOpts{Content: comment.Content}).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeComment),
				opts: repository.UpdateCommentOpts{Content: "new content"},
			},
			want: &repository.Comment{
				ID:        model.MustNewID(model.ResourceTypeComment),
				Content:   "new content",
				CreatedBy: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "update comment with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Comment) []repository.RedisRepositoryOption {
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
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateCommentOpts{Content: "new content"}).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeComment),
				opts: repository.UpdateCommentOpts{Content: "new content"},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update comment set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
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
						Value: comment,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateCommentOpts{Content: "new content"}).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeComment),
				opts: repository.UpdateCommentOpts{Content: "new content"},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update comment delete cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *repository.Comment) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateCommentOpts{Content: "new content"}).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeComment),
				opts: repository.UpdateCommentOpts{Content: "new content"},
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

			r := func() *repository.RedisCachedCommentRepository {
				r, err := repository.NewCachedCommentRepository(
					tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCachedCommentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.CommentRepository
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
			name: "delete comment success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String()+"*")
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
		},
		{
			name: "delete comment with comment deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String()+"*")
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.CommentRepository {
					repo := mockrepo.NewMockCommentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrCommentDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: repository.ErrCommentDelete,
		},
		{
			name: "delete comment with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String()+"*")

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
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete comment cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String()+"*")
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: repository.ErrCacheDelete,
		},

		{
			name: "delete comment cache by document key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String()+"*")
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete comment cache by issues key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String()+"*")
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "ListBelongsTo", "*", "*", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.CommentRepository {
					return mockrepo.NewMockCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
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
			r := func() *repository.RedisCachedCommentRepository {
				r, err := repository.NewCachedCommentRepository(
					tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.id),
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
