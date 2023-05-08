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

func TestCachedDocumentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository
		documentRepo func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

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
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Create", ctx, belongsTo, document).Return(nil)
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
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

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
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Create", ctx, belongsTo, document).Return(repository.ErrDocumentCreate)
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
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrDocumentCreate,
		},
		{
			name: "create document with belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")

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
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with by creator cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)

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
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with namespace cross cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)

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
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with project cross cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

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
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with user cross cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

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
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
				document: &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.document),
				documentRepo: tt.fields.documentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.document),
			}
			err := r.Create(tt.args.ctx, tt.args.belongsTo, tt.args.document)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedDocumentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctx context.Context, id model.ID, document *model.Document) *baseRepository
		documentRepo func(ctx context.Context, id model.ID, document *model.Document) repository.DocumentRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
						Value: document,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Get", ctx, id).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *model.Document {
				return &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(document, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *model.Document {
				return &model.Document{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(ctx context.Context, id model.ID, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached document error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
				documentRepo: func(ctx context.Context, id model.ID, document *model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached document cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

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
						Value: document,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Get", ctx, id).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var want *model.Document
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				documentRepo: tt.fields.documentRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedDocumentRepository_GetByCreator(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *baseRepository
		documentRepo func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository
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
				cacheRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
						Value: documents,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("GetByCreator", ctx, createdBy, offset, limit).Return(documents, nil)
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
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(documents, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
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
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
				documentRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("GetByCreator", ctx, createdBy, offset, limit).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get documents cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
				documentRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached documents cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)

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
						Value: documents,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, createdBy model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("GetByCreator", ctx, createdBy, offset, limit).Return(documents, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(tt.args.ctx, tt.args.createdBy, tt.args.offset, tt.args.limit, tt.want),
				documentRepo: tt.fields.documentRepo(tt.args.ctx, tt.args.createdBy, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetByCreator(tt.args.ctx, tt.args.createdBy, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedDocumentRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *baseRepository
		documentRepo func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: documents,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(documents, nil)
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
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(documents, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
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
					FileID:      "test file id",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeDocument),
					Name:        "test document",
					Excerpt:     "test excerpt",
					FileID:      "test file id",
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
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				documentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
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
			name: "get get documents cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				documentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached documents cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: documents,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, documents []*model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(documents, nil)
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
			r := &CachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				documentRepo: tt.fields.documentRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedDocumentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctx context.Context, id model.ID, document *model.Document) *baseRepository
		documentRepo func(ctx context.Context, id model.ID, patch map[string]any, document *model.Document) repository.DocumentRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					byCreatorKeyCmd := new(redis.StringSliceCmd)
					byCreatorKeyCmd.SetVal([]string{byCreatorKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyCmd, nil)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
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

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, patch map[string]any, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Update", ctx, id, patch).Return(document, nil)
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
				FileID:      "test file id",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
			},
		},
		{
			name: "update document with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(testMock.CacheRepo),
						tracer: new(testMock.Tracer),
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, patch map[string]any, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
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
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update document set cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, patch map[string]any, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Update", ctx, id, patch).Return(document, nil)
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update document delete belongs to cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
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

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(errors.New("error"))
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, patch map[string]any, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Update", ctx, id, patch).Return(document, nil)
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update document with delete by creator cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, document *model.Document) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", document.CreatedBy.String(), "*")

					belongsToKeyCmd := new(redis.StringSliceCmd)
					belongsToKeyCmd.SetVal([]string{belongsToKey})

					byCreatorKeyCmd := new(redis.StringSliceCmd)
					byCreatorKeyCmd.SetVal([]string{byCreatorKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyCmd, nil)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
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

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(repository.ErrCacheDelete)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID, patch map[string]any, document *model.Document) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Update", ctx, id, patch).Return(document, nil)
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
				FileID:      "test file id",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &CachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				documentRepo: tt.fields.documentRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
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
		cacheRepo    func(ctx context.Context, id model.ID) *baseRepository
		documentRepo func(ctx context.Context, id model.ID) repository.DocumentRepository
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

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
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

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
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					repo := new(testMock.DocumentRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "delete document with cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())

					dbClient := new(testMock.RedisClient)

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
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete document with belongs to cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")

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
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete document with by creator cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
					belongsToKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)

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
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete document with namespaces cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)

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
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete document with projects cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

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
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete document with users cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.On("Keys", ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

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
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, byCreatorKey).Return(nil)
					cacheRepo.On("Delete", ctx, namespacesKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				documentRepo: func(ctx context.Context, id model.ID) repository.DocumentRepository {
					return new(testMock.DocumentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedDocumentRepository{
				cacheRepo:    tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				documentRepo: tt.fields.documentRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
