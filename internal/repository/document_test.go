package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedDocumentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) DocumentRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateDocumentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&Document{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "create document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, ErrDocumentCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrDocumentCreate,
		},
		{
			name: "create document with belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ CreateDocumentOpts) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with by creator cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ CreateDocumentOpts) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with namespace cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ CreateDocumentOpts) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with project cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ CreateDocumentOpts) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with user cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateDocumentOpts) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", opts.BelongsTo.String(), "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ CreateDocumentOpts) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateDocumentOpts{
					BelongsTo: model.MustNewID(model.ResourceTypeUser),
					Name:      "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedDocumentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) DocumentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *Document
		wantErr error
	}{
		{
			name: "get uncached document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, DocumentDetailProjection()).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *Document {
				return &Document{
					ID:              id,
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get cached document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))

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
						if ptr, ok := dst.(**Document); ok {
							*ptr = document
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *Document) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *Document {
				return &Document{
					ID:              id,
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get uncached document error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))

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
						if ptr, ok := dst.(**Document); ok {
							*ptr = document
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, DocumentDetailProjection()).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached document error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *Document) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached document cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, DocumentDetailProjection()).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *Document
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id, DocumentDetailProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedDocumentRepository_ListByCreator(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) DocumentRepository
	}
	type args struct {
		ctx       context.Context
		createdBy model.ID
		offset    int
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Document
		wantErr error
	}{
		{
			name: "get uncached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: Page[*Document]{Items: documents},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListByCreator(ctx, createdBy, CursorPage{Size: limit}, DocumentListProjection()).Return(Page[*Document]{Items: documents}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get cached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
						if ptr, ok := dst.(*Page[*Document]); ok {
							*ptr = Page[*Document]{Items: documents}
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Document) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get uncached documents error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, _ []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), projectionCacheValue(DocumentListProjection()), "", limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, _ []*Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListByCreator(ctx, createdBy, CursorPage{Size: limit}, DocumentListProjection()).Return(Page[*Document]{}, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get documents cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, _ []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Document) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached documents cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: Page[*Document]{Items: documents},
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListByCreator(ctx, createdBy, CursorPage{Size: limit}, DocumentListProjection()).Return(Page[*Document]{Items: documents}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.createdBy, tt.args.offset, testPageSize(tt.args.limit), tt.want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.createdBy, tt.args.offset, testPageSize(tt.args.limit), tt.want),
			}
			got, err := r.ListByCreator(tt.args.ctx, tt.args.createdBy, CursorPage{Size: testPageSize(tt.args.limit)}, DocumentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedDocumentRepository_ListBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) DocumentRepository
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
		want    []*Document
		wantErr error
	}{
		{
			name: "get uncached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: Page[*Document]{Items: documents},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, DocumentListProjection()).Return(Page[*Document]{Items: documents}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get cached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
						if ptr, ok := dst.(*Page[*Document]); ok {
							*ptr = Page[*Document]{Items: documents}
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Document) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Name:            "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get uncached documents error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(DocumentListProjection()), "", limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, DocumentListProjection()).Return(Page[*Document]{}, ErrNotFound)
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
			name: "get get documents cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Document) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached documents cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(DocumentListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: Page[*Document]{Items: documents},
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, documents []*Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, DocumentListProjection()).Return(Page[*Document]{Items: documents}, nil)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want),
			}
			got, err := r.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, CursorPage{Size: testPageSize(tt.args.limit)}, DocumentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedDocumentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateDocumentOpts, document *Document) DocumentRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateDocumentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Document
		wantErr error
	}{
		{
			name: "update document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", document.CreatedBy.ID.String(), "*", "*", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					byCreatorKeyCmd := new(redis.StringSliceCmd)
					byCreatorKeyCmd.SetVal([]string{byCreatorKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateDocumentOpts, document *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: UpdateDocumentOpts{},
			},
			want: &Document{
				ID:              model.MustNewID(model.ResourceTypeDocument),
				Name:            "new document",
				Excerpt:         "new excerpt",
				FileID:          "test file subject",
				CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
			},
		},
		{
			name: "update document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *Document) *redisBaseRepository {
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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateDocumentOpts, _ *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: UpdateDocumentOpts{},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update document set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))

					dbClient := mock.NewUniversalClient(ctrl)
					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateDocumentOpts, document *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: UpdateDocumentOpts{},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update document delete belongs to cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)
					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateDocumentOpts, document *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: UpdateDocumentOpts{},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "update document with delete by creator cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", document.CreatedBy.ID.String(), "*", "*", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					byCreatorKeyCmd := new(redis.StringSliceCmd)
					byCreatorKeyCmd.SetVal([]string{byCreatorKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyCmd)
					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateDocumentOpts, document *Document) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: UpdateDocumentOpts{},
			},
			want: &Document{
				ID:              model.MustNewID(model.ResourceTypeDocument),
				Name:            "new document",
				Excerpt:         "new excerpt",
				FileID:          "test file subject",
				CreatedBy:       PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedDocumentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) DocumentRepository
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
			name: "delete document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(6)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(6)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) DocumentRepository {
					repo := NewMockDocumentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "delete document with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete document with belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
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
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete document with by creator cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete document with namespaces cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete document with projects cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete document with users cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(6)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNeo4jDocumentRepository_applyDocumentLoadersUnknown(t *testing.T) {
	t.Parallel()

	r := new(Neo4jDocumentRepository)
	err := r.applyDocumentLoaders(context.Background(), nil, QueryPlan{
		Loaders: []CompiledQuery{{Name: "document.load_unknown", Params: map[string]any{}}},
	}, []*Document{{ID: model.MustNewID(model.ResourceTypeDocument)}})
	assert.ErrorIs(t, err, ErrQueryCompile)
}
