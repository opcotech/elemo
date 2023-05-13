package redis

import (
	"context"
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

func TestCachedAttachmentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) *baseRepository
		attachmentRepo func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository
	}
	type args struct {
		ctx        context.Context
		belongsTo  model.ID
		attachment *model.Attachment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new attachment",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Create", ctx, belongsTo, attachment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				attachment: &model.Attachment{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new attachment with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Create", ctx, belongsTo, attachment).Return(repository.ErrAttachmentCreate)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				attachment: &model.Attachment{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrAttachmentCreate,
		},
		{
			name: "add new attachment belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")

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

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				attachment: &model.Attachment{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new attachment cross cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeIssue),
				attachment: &model.Attachment{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
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
			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.attachment),
				attachmentRepo: tt.fields.attachmentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.attachment),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.attachment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedAttachmentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository
		attachmentRepo func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Attachment
		wantErr error
	}{
		{
			name: "get uncached attachment",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Get", ctx, id).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			want: func(id model.ID) *model.Attachment {
				return &model.Attachment{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get cached attachment",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(attachment, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			want: func(id model.ID) *model.Attachment {
				return &model.Attachment{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get uncached attachment error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached attachment error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached attachment cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Get", ctx, id).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var want *model.Attachment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				attachmentRepo: tt.fields.attachmentRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedAttachmentRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository
		attachmentRepo func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository
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
		want    []*model.Attachment
		wantErr error
	}{
		{
			name: "get uncached attachments",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachments,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(attachments, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Attachment{
				{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
				{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "get cached attachments",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(attachments, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Attachment{
				{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
				{
					ID:        model.MustNewID(model.ResourceTypeAttachment),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "get uncached attachments error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
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
			name: "get get attachments cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached attachments cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachments,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(attachments, nil)
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

			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				attachmentRepo: tt.fields.attachmentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedAttachmentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository
		attachmentRepo func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Attachment
		wantErr error
	}{
		{
			name: "update attachment",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Update", ctx, id, attachment.Name).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				name: "name",
			},
			want: &model.Attachment{
				ID:   model.MustNewID(model.ResourceTypeAttachment),
				Name: "name",
			},
		},
		{
			name: "update attachment with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(testMock.CacheRepository),
						tracer: new(testMock.Tracer),
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Update", ctx, id, "name").Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				name: "name",
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update attachment set cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Update", ctx, id, "name").Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				name: "name",
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update attachment delete cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(assert.AnError)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Update", ctx, id, "name").Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				name: "name",
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				attachmentRepo: tt.fields.attachmentRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.name)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCachedAttachmentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID) *baseRepository
		attachmentRepo func(ctx context.Context, id model.ID) repository.AttachmentRepository
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
			name: "delete attachment success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
		},
		{
			name: "delete attachment with attachment deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrAttachmentDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrAttachmentDelete,
		},
		{
			name: "delete attachment with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := new(testMock.AttachmentRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete attachment cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")

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

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byBelongsTo).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				attachmentRepo: func(ctx context.Context, id model.ID) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete attachment cache by document key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, id model.ID) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete attachment cache by issues key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")
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

					cacheRepo := new(testMock.CacheRepository)
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
				attachmentRepo: func(ctx context.Context, id model.ID) repository.AttachmentRepository {
					return new(testMock.AttachmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				attachmentRepo: tt.fields.attachmentRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
