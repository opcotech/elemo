package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	testMock "github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedCommentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, belongsTo model.ID, comment *model.Comment) *baseRepository
		commentRepo func(ctx context.Context, belongsTo model.ID, comment *model.Comment) repository.CommentRepository
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
			name: "add new comment",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Create", ctx, belongsTo, comment).Return(nil)
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
			name: "add new comment with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Create", ctx, belongsTo, comment).Return(repository.ErrCommentCreate)
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
			wantErr: repository.ErrCommentCreate,
		},
		{
			name: "add new comment belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new comment cross cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, comment *model.Comment) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.comment),
				commentRepo: tt.fields.commentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.comment),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.comment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedCommentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, id model.ID, comment *model.Comment) *baseRepository
		commentRepo func(ctx context.Context, id model.ID, comment *model.Comment) repository.CommentRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, comment *model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID, comment *model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Get", ctx, id).Return(comment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			want: func(id model.ID) *model.Comment {
				return &model.Comment{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get cached comment",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, comment *model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(comment, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID, comment *model.Comment) repository.CommentRepository {
					return new(testMock.CommentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeComment),
			},
			want: func(id model.ID) *model.Comment {
				return &model.Comment{
					ID:        model.MustNewID(model.ResourceTypeComment),
					Content:   "test comment content",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get uncached comment error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, comment *model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID, comment *model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, comment *model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID, comment *model.Comment) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, comment *model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comment,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID, comment *model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Get", ctx, id).Return(comment, nil)
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
		t.Run(tt.name, func(t *testing.T) {
			var want *model.Comment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				commentRepo: tt.fields.commentRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedCommentRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *baseRepository
		commentRepo func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) repository.CommentRepository
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comments,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(comments, nil)
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(comments, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: comments,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, comments []*model.Comment) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(comments, nil)
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
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				commentRepo: tt.fields.commentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedCommentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo   func(ctx context.Context, id model.ID) *baseRepository
		commentRepo func(ctx context.Context, id model.ID) repository.CommentRepository
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byBelongsTo).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byBelongsTo).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrCommentDelete)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID) repository.CommentRepository {
					repo := new(testMock.CommentRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byBelongsTo).Return(byBelongsToCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byBelongsTo).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byBelongsTo).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byBelongsTo).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				commentRepo: func(ctx context.Context, id model.ID) repository.CommentRepository {
					return new(testMock.CommentRepository)
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
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedCommentRepository{
				cacheRepo:   tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				commentRepo: tt.fields.commentRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
