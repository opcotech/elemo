package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
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

					cacheRepo := new(testMock.CacheRepo)
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
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.attachment),
				attachmentRepo: tt.fields.attachmentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.attachment),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.attachment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
