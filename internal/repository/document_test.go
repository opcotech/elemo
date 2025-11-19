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

func TestCachedDocumentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) DocumentRepository
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		document  *model.Document
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")
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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, document).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
		},
		{
			name: "create document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")
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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Create(ctx, belongsTo, document).Return(ErrDocumentCreate)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: ErrDocumentCreate,
		},
		{
			name: "create document with belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _ *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with by creator cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with namespace cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")
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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with project cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")
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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create document with user cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, document *model.Document) *redisBaseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")
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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
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
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.document),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.document),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.document)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedDocumentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) DocumentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Document
		wantErr error
	}{
		{
			name: "get uncached document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *model.Document {
				return &model.Document{
					ID:          id,
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
						if ptr, ok := dst.(**model.Document); ok {
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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *model.Document {
				return &model.Document{
					ID:          id,
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached document error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
						if ptr, ok := dst.(**model.Document); ok {
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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(document, nil)
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
			var want *model.Document
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedDocumentRepository_GetByCreator(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) DocumentRepository
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
		want    []*model.Document
		wantErr error
	}{
		{
			name: "get uncached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
						Value: documents,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().GetByCreator(ctx, createdBy, offset, limit).Return(documents, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Document{
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
						if docsPtr, ok := dst.(*[]*model.Document); ok {
							*docsPtr = documents
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Document{
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached documents error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, _ []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, _ []*model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().GetByCreator(ctx, createdBy, offset, limit).Return(nil, ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, _ []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
						Value: documents,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().GetByCreator(ctx, createdBy, offset, limit).Return(documents, nil)
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
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.createdBy, tt.args.offset, tt.args.limit, tt.want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.createdBy, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetByCreator(tt.args.ctx, tt.args.createdBy, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedDocumentRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) DocumentRepository
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
		want    []*model.Document
		wantErr error
	}{
		{
			name: "get uncached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: documents,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(documents, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Document{
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						if docsPtr, ok := dst.(*[]*model.Document); ok {
							*docsPtr = documents
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Document{
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file subject",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached documents error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
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
			name: "get get documents cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Document) DocumentRepository {
					return mock.NewDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: documents,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(documents, nil)
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
				cacheRepo:    tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedDocumentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, document *model.Document) DocumentRepository
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		patch map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Document
		wantErr error
	}{
		{
			name: "update document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
				patch: map[string]any{
					"name":    "new content",
					"excerpt": "new excerpt",
				},
			},
			want: &model.Document{
				ID:          model.MustNewID(model.ResourceTypeDocument),
				Name:        "new document",
				Excerpt:     "new excerpt",
				FileID:      "test file subject",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
			},
		},
		{
			name: "update document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Document) *redisBaseRepository {
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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
				patch: map[string]any{
					"name":    "new content",
					"excerpt": "new excerpt",
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update document set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
				patch: map[string]any{
					"name":    "new content",
					"excerpt": "new excerpt",
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update document delete belongs to cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
				patch: map[string]any{
					"name":    "new content",
					"excerpt": "new excerpt",
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "update document with delete by creator cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *model.Document) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, document *model.Document) DocumentRepository {
					repo := mock.NewDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
				patch: map[string]any{
					"name":    "new content",
					"excerpt": "new excerpt",
				},
			},
			want: &model.Document{
				ID:          model.MustNewID(model.ResourceTypeDocument),
				Name:        "new document",
				Excerpt:     "new excerpt",
				FileID:      "test file subject",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
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
				documentRepo: tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")
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
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

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
					repo := mock.NewDocumentRepository(ctrl)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")
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
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

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
					repo := mock.NewDocumentRepository(ctrl)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return mock.NewDocumentRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) DocumentRepository {
					return mock.NewDocumentRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")

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
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

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
					return mock.NewDocumentRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")
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
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					return mock.NewDocumentRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")
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
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

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
					return mock.NewDocumentRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")
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
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

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
					return mock.NewDocumentRepository(nil)
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
