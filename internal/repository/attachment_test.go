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

func TestCachedAttachmentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) []repository.RedisRepositoryOption
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) repository.AttachmentRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateAttachmentOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Attachment{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAttachmentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
					Name:      "test",
					FileID:    "test",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "add new attachment with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, documentsKey).Return(documentsKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrAttachmentCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAttachmentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateAttachmentOpts) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAttachmentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateAttachmentOpts) []repository.RedisRepositoryOption {
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*")
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
					cacheRepo.EXPECT().Delete(ctx, documentsKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateAttachmentOpts) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateAttachmentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeIssue),
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
			r := func() *repository.RedisCachedAttachmentRepository {
				r, err := repository.NewCachedAttachmentRepository(
					tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.opts),
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

func TestCachedAttachmentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Attachment
		wantErr error
	}{
		{
			name: "get uncached attachment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.AttachmentDetailProjection()).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			want: func(id model.ID) *repository.Attachment {
				return &repository.Attachment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))

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
						if ptr, ok := dst.(**repository.Attachment); ok {
							*ptr = attachment
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Attachment) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAttachment),
			},
			want: func(id model.ID) *repository.Attachment {
				return &repository.Attachment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.AttachmentDetailProjection()).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Attachment) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: attachment,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.AttachmentDetailProjection()).Return(attachment, nil)
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
			var want *repository.Attachment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedAttachmentRepository {
				r, err := repository.NewCachedAttachmentRepository(
					tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.AttachmentDetailProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedAttachmentRepository_ListBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) []repository.RedisRepositoryOption
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) repository.AttachmentRepository
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
		want    []*repository.Attachment
		wantErr error
	}{
		{
			name: "get uncached attachments",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.AttachmentListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Attachment]{Items: attachments},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}, repository.AttachmentListProjection()).Return(repository.Page[*repository.Attachment]{Items: attachments}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Attachment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.AttachmentListProjection()), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.Attachment]); ok {
							*ptr = repository.Page[*repository.Attachment]{Items: attachments}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Attachment) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Attachment{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.AttachmentListProjection()), "", limit)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}, repository.AttachmentListProjection()).Return(repository.Page[*repository.Attachment]{}, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.AttachmentListProjection()), "", limit)

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Attachment) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.AttachmentListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Attachment]{Items: attachments},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, attachments []*repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}, repository.AttachmentListProjection()).Return(repository.Page[*repository.Attachment]{Items: attachments}, nil)
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

			r := func() *repository.RedisCachedAttachmentRepository {
				r, err := repository.NewCachedAttachmentRepository(
					tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.AttachmentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedAttachmentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption
		attachmentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateAttachmentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Attachment
		wantErr error
	}{
		{
			name: "update attachment",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")

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
						Value: attachment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateAttachmentOpts{Name: attachment.Name}).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				opts: repository.UpdateAttachmentOpts{Name: "name"},
			},
			want: &repository.Attachment{
				ID:   model.MustNewID(model.ResourceTypeAttachment),
				Name: "name",
			},
		},
		{
			name: "update attachment with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Attachment) []repository.RedisRepositoryOption {
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
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateAttachmentOpts{Name: "name"}).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				opts: repository.UpdateAttachmentOpts{Name: "name"},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update attachment set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))

					dbClient := mockrepo.NewMockUniversalClient(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
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
						Value: attachment,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateAttachmentOpts{Name: "name"}).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				opts: repository.UpdateAttachmentOpts{Name: "name"},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update attachment delete cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(repository.AttachmentDetailProjection()))
					belongsToKey := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")

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
						Value: attachment,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, attachment *repository.Attachment) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, repository.UpdateAttachmentOpts{Name: "name"}).Return(attachment, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeAttachment),
				opts: repository.UpdateAttachmentOpts{Name: "name"},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := func() *repository.RedisCachedAttachmentRepository {
				r, err := repository.NewCachedAttachmentRepository(
					tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
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

func TestCachedAttachmentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), "*")
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")
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
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), "*")
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")
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
				attachmentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.AttachmentRepository {
					repo := mockrepo.NewMockAttachmentRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), "*")

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), "*")
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")

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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), "*")
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")
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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), "*")
					byBelongsTo := composeCacheKey(model.ResourceTypeAttachment.String(), "ListBelongsTo", "*", "*", "*")
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
				attachmentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.AttachmentRepository {
					return mockrepo.NewMockAttachmentRepository(nil)
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
			r := func() *repository.RedisCachedAttachmentRepository {
				r, err := repository.NewCachedAttachmentRepository(
					tt.fields.attachmentRepo(ctrl, tt.args.ctx, tt.args.id),
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
