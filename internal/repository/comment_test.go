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

func TestCachedCommentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, comment *model.Comment) *redisBaseRepository
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, comment *model.Comment) CommentRepository
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		comment   *model.Comment
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Comment) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, comment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				comment: &model.Comment{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new comment to a document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Comment) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, comment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeDocument),
				comment: &model.Comment{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new comment with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Comment) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, comment).Return(ErrCommentCreate)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				comment: &model.Comment{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCommentCreate,
		},
		{
			name: "add new comment belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Comment) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Comment) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				comment: &model.Comment{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.comment),
				commentRepo: tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.comment),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.comment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedCommentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Comment
		wantErr error
	}{
		{
			name: "get uncached comment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

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
						Value: comment,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			want: func(id model.ID) *model.Comment {
				return &model.Comment{
					ID:        id,
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get cached comment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

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
						if ptr, ok := dst.(**model.Comment); ok {
							*ptr = comment
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Comment) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			want: func(id model.ID) *model.Comment {
				return &model.Comment{
					ID:        id,
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get uncached comment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

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
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached comment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

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
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Comment) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached comment cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

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
						Value: comment,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
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
			var want *model.Comment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				commentRepo: tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedCommentRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *redisBaseRepository
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) CommentRepository
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
		want    []*model.Comment
		wantErr error
	}{
		{
			name: "get uncached comments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: comments,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(comments, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Comment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						if listPtr, ok := dst.(*[]*model.Comment); ok {
							*listPtr = comments
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Comment) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Comment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get comments cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Comment) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached comments cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: comments,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(comments, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				commentRepo: tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedCommentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository
	}
	type args struct {
		ctx     context.Context
		id      model.ID
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Comment
		wantErr error
	}{
		{
			name: "update comment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comment,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, comment.Content).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeComment),
				content: "new content",
			},
			want: &model.Comment{
				ID:        model.MustNewID(model.ResourceTypeComment),
				Content:   "new content",
				CreatedBy: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "update comment with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Comment) *redisBaseRepository {
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
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, "new content").Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeComment),
				content: "new content",
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update comment set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
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
						Value: comment,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, "new content").Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeComment),
				content: "new content",
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update comment delete cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comment,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, comment *model.Comment) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, "new content").Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeComment),
				content: "new content",
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

			r := &RedisCachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				commentRepo: tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.content)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCachedCommentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		commentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) CommentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) CommentRepository {
					repo := mock.NewCommentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrCommentDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: ErrCommentDelete,
		},
		{
			name: "delete comment with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

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
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete comment cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: ErrCacheDelete,
		},

		{
			name: "delete comment cache by document key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete comment cache by issues key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				commentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) CommentRepository {
					return mock.NewCommentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
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
			r := &RedisCachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				commentRepo: tt.fields.commentRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
