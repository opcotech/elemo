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

func TestCachedAttachmentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, attachment *model.Attachment) *baseRepository
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, attachment).Return(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, attachment).Return(repository.ErrAttachmentCreate)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Attachment) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Attachment) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Attachment) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.attachment),
				attachmentRepo: tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.attachment),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.attachment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedAttachmentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			want: func(id model.ID) *model.Attachment {
				return &model.Attachment{
					ID:        id,
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get cached attachment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

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
						if ptr, ok := dst.(**model.Attachment); ok {
							*ptr = attachment
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Attachment) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			want: func(id model.ID) *model.Attachment {
				return &model.Attachment{
					ID:        id,
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				}
			},
		},
		{
			name: "get uncached attachment error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Attachment) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(attachment, nil)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *model.Attachment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				attachmentRepo: tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedAttachmentRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachments,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(attachments, nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						if listPtr, ok := dst.(*[]*model.Attachment); ok {
							*listPtr = attachments
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Attachment) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Attachment) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachments,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, attachments []*model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(attachments, nil)
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
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				attachmentRepo: tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedAttachmentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, attachment.Name).Return(attachment, nil)
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Attachment) *baseRepository {
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
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, "name").Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

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
						Value: attachment,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, "name").Return(attachment, nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *model.Attachment) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, "name").Return(attachment, nil)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				attachmentRepo: tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.name)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCachedAttachmentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AttachmentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := mock.NewAttachmentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrAttachmentDelete)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					byBelongsToCmd := new(redis.StringSliceCmd)
					byBelongsToCmd.SetVal([]string{byBelongsTo})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byBelongsTo).Return(byBelongsToCmd)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byBelongsTo).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mock.NewAttachmentRepository(nil)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedAttachmentRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				attachmentRepo: tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
