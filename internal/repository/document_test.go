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
	"github.com/opcotech/elemo/internal/pkg/convert"
)

func TestCachedDocumentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) repository.DocumentRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateDocumentOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					foldersKey := composeCacheKey(model.ResourceTypeFolder.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					foldersKeyResult := new(redis.StringSliceCmd)
					foldersKeyResult.SetVal([]string{foldersKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, foldersKey).Return(foldersKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(8)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, foldersKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Document{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "create document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					foldersKey := composeCacheKey(model.ResourceTypeFolder.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					foldersKeyResult := new(redis.StringSliceCmd)
					foldersKeyResult.SetVal([]string{foldersKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, foldersKey).Return(foldersKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(8)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, foldersKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrDocumentCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrDocumentCreate,
		},
		{
			name: "create document with library cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateDocumentOpts) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with by creator cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateDocumentOpts) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with namespace cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateDocumentOpts) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with project cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateDocumentOpts) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
					CreatedBy: model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create document with user cross cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateDocumentOpts) []repository.RedisRepositoryOption {
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", opts.Library.String(), "*", "*", "*", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", opts.CreatedBy.String(), "*", "*", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ repository.CreateDocumentOpts) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateDocumentOpts{
					Library:   model.MustNewID(model.ResourceTypeNamespace),
					Title:     "test document",
					Excerpt:   "test excerpt",
					FileID:    "test file subject",
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
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedDocumentRepository {
				r, err := repository.NewCachedDocumentRepository(
					tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.opts),
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

func TestCachedDocumentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) repository.DocumentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Document
		wantErr error
	}{
		{
			name: "get uncached document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))

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
						Value: document,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *repository.Document {
				return &repository.Document{
					ID:              id,
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get cached document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))

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
						if ptr, ok := dst.(**repository.Document); ok {
							*ptr = document
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Document) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			want: func(id model.ID) *repository.Document {
				return &repository.Document{
					ID:              id,
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get uncached document error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))

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
						if ptr, ok := dst.(**repository.Document); ok {
							*ptr = document
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Document) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))

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
						Value: document,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(document, nil)
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
			ctrl := gomock.NewController(t)
			var want *repository.Document
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedDocumentRepository {
				r, err := repository.NewCachedDocumentRepository(
					tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.DocumentDetailProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedDocumentRepository_ListByCreator(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) repository.DocumentRepository
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
		want    []*repository.Document
		wantErr error
	}{
		{
			name: "get uncached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Document]{Items: documents},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListByCreator(ctx, createdBy, model.MustNewNilID(model.ResourceTypeUser), repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{Items: documents}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get cached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.Document]); ok {
							*ptr = repository.Page[*repository.Document]{Items: documents}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Document) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get uncached documents error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, _ []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, _ []*repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListByCreator(ctx, createdBy, model.MustNewNilID(model.ResourceTypeUser), repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{}, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, _ []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Document) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Document]{Items: documents},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, createdBy model.ID, _, limit int, documents []*repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListByCreator(ctx, createdBy, model.MustNewNilID(model.ResourceTypeUser), repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{Items: documents}, nil)
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
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedDocumentRepository {
				r, err := repository.NewCachedDocumentRepository(
					tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.createdBy, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.createdBy, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListByCreator(tt.args.ctx, tt.args.createdBy, model.MustNewNilID(model.ResourceTypeUser), repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.DocumentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedDocumentRepository_ListLibrary(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) repository.DocumentRepository
	}
	type args struct {
		ctx       context.Context
		libraryID model.ID
		offset    int
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.Document
		wantErr error
	}{
		{
			name: "get uncached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", libraryID.String(), model.MustNewNilID(model.ResourceTypeUser).String(), "root", projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Document]{Items: documents},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListLibrary(ctx, libraryID, model.MustNewNilID(model.ResourceTypeUser), nil, repository.LibraryListFilter{}, repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{Items: documents}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				libraryID: model.MustNewID(model.ResourceTypeNamespace),
			},
			want: []*repository.Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get cached documents",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", libraryID.String(), model.MustNewNilID(model.ResourceTypeUser).String(), "root", projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.Document]); ok {
							*ptr = repository.Page[*repository.Document]{Items: documents}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Document) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				libraryID: model.MustNewID(model.ResourceTypeNamespace),
			},
			want: []*repository.Document{
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeDocument),
					Title:           "test document",
					Excerpt:         "test excerpt",
					FileID:          "test file subject",
					CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Labels:          make([]repository.PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get uncached documents error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, _ []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", libraryID.String(), model.MustNewNilID(model.ResourceTypeUser).String(), "root", projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, _ []*repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListLibrary(ctx, libraryID, model.MustNewNilID(model.ResourceTypeUser), nil, repository.LibraryListFilter{}, repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				libraryID: model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get documents cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, _ []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", libraryID.String(), model.MustNewNilID(model.ResourceTypeUser).String(), "root", projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Document) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx:       context.Background(),
				libraryID: model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached documents cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", libraryID.String(), model.MustNewNilID(model.ResourceTypeUser).String(), "root", projectionCacheValue(repository.DocumentListProjection()), "", limit)

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
						Value: repository.Page[*repository.Document]{Items: documents},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, libraryID model.ID, _, limit int, documents []*repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().ListLibrary(ctx, libraryID, model.MustNewNilID(model.ResourceTypeUser), nil, repository.LibraryListFilter{}, repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{Items: documents}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				libraryID: model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedDocumentRepository {
				r, err := repository.NewCachedDocumentRepository(
					tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.libraryID, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.libraryID, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListLibrary(tt.args.ctx, tt.args.libraryID, model.MustNewNilID(model.ResourceTypeUser), nil, repository.LibraryListFilter{}, repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.DocumentListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedDocumentRepository_ListRelated(t *testing.T) {
	t.Parallel()

	relatedTo := model.MustNewID(model.ResourceTypeProject)
	documents := []*repository.Document{{
		ID:        model.MustNewID(model.ResourceTypeDocument),
		Title:     "test document",
		CreatedBy: repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
	}}
	limit := testPageSize(0)

	t.Run("get uncached documents", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ctx := context.Background()
		key := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", relatedTo.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(repository.DocumentListProjection()), "", limit)

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
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
			Value: repository.Page[*repository.Document]{Items: documents},
		}).Return(nil)

		repo := mockrepo.NewMockDocumentRepository(ctrl)
		repo.EXPECT().ListRelated(ctx, relatedTo, model.MustNewNilID(model.ResourceTypeUser), repository.CursorPage{Size: limit}, repository.DocumentListProjection()).Return(repository.Page[*repository.Document]{Items: documents}, nil)

		r := func() *repository.RedisCachedDocumentRepository {
			r, err := repository.NewCachedDocumentRepository(
				repo,
				[]repository.RedisRepositoryOption{
					repository.WithRedisDatabase(db),
					repository.WithCacheBackend(cacheRepo),
					repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
					repository.WithRedisRepositoryTracer(tracer),
				}...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListRelated(ctx, relatedTo, model.MustNewNilID(model.ResourceTypeUser), repository.CursorPage{Size: limit}, repository.DocumentListProjection())
		require.NoError(t, err)
		assert.Equal(t, documents, got.Items)
	})
}

func TestCachedDocumentRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateDocumentOpts, document *repository.Document) repository.DocumentRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateDocumentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Document
		wantErr error
	}{
		{
			name: "update document",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", document.CreatedBy.ID.String(), "*", "*", "*")

					libraryKeyCmd := new(redis.StringSliceCmd)
					libraryKeyCmd.SetVal([]string{libraryKey})

					relatedKeyCmd := new(redis.StringSliceCmd)
					relatedKeyCmd.SetVal([]string{relatedKey})

					byCreatorKeyCmd := new(redis.StringSliceCmd)
					byCreatorKeyCmd.SetVal([]string{byCreatorKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyCmd)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateDocumentOpts, document *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: repository.UpdateDocumentOpts{},
			},
			want: &repository.Document{
				ID:              model.MustNewID(model.ResourceTypeDocument),
				Title:           "new document",
				Excerpt:         "new excerpt",
				FileID:          "test file subject",
				CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Labels:          make([]repository.PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
			},
		},
		{
			name: "update document with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Document) []repository.RedisRepositoryOption {
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
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateDocumentOpts, _ *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: repository.UpdateDocumentOpts{},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update document set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateDocumentOpts, document *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: repository.UpdateDocumentOpts{},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update document delete belongs to cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")

					libraryKeyCmd := new(redis.StringSliceCmd)
					libraryKeyCmd.SetVal([]string{libraryKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyCmd)
					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateDocumentOpts, document *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: repository.UpdateDocumentOpts{},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update document with delete by creator cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, document *repository.Document) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(repository.DocumentDetailProjection()))
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", document.CreatedBy.ID.String(), "*", "*", "*")

					libraryKeyCmd := new(redis.StringSliceCmd)
					libraryKeyCmd.SetVal([]string{libraryKey})

					relatedKeyCmd := new(redis.StringSliceCmd)
					relatedKeyCmd.SetVal([]string{relatedKey})

					byCreatorKeyCmd := new(redis.StringSliceCmd)
					byCreatorKeyCmd.SetVal([]string{byCreatorKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyCmd)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyCmd)
					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: document,
					}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateDocumentOpts, document *repository.Document) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(document, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeDocument),
				opts: repository.UpdateDocumentOpts{},
			},
			want: &repository.Document{
				ID:              model.MustNewID(model.ResourceTypeDocument),
				Title:           "new document",
				Excerpt:         "new excerpt",
				FileID:          "test file subject",
				CreatedBy:       repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Labels:          make([]repository.PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			r := func() *repository.RedisCachedDocumentRepository {
				r, err := repository.NewCachedDocumentRepository(
					tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
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
		cacheRepo    func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		documentRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.DocumentRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					foldersKey := composeCacheKey(model.ResourceTypeFolder.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					relatedKeyResult := new(redis.StringSliceCmd)
					relatedKeyResult.SetVal([]string{relatedKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					foldersKeyResult := new(redis.StringSliceCmd)
					foldersKeyResult.SetVal([]string{foldersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, foldersKey).Return(foldersKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(10)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(10)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, foldersKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil).Times(1)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")
					foldersKey := composeCacheKey(model.ResourceTypeFolder.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					relatedKeyResult := new(redis.StringSliceCmd)
					relatedKeyResult.SetVal([]string{relatedKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					foldersKeyResult := new(redis.StringSliceCmd)
					foldersKeyResult.SetVal([]string{foldersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, issuesKey).Return(issuesKeyResult)
					dbClient.EXPECT().Keys(ctx, foldersKey).Return(foldersKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(10)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(10)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, issuesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, foldersKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.DocumentRepository {
					repo := mockrepo.NewMockDocumentRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")

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
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete document with library cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					relatedKeyResult := new(redis.StringSliceCmd)
					relatedKeyResult.SetVal([]string{relatedKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					relatedKeyResult := new(redis.StringSliceCmd)
					relatedKeyResult.SetVal([]string{relatedKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					relatedKeyResult := new(redis.StringSliceCmd)
					relatedKeyResult.SetVal([]string{relatedKey})

					byCreatorKeyResult := new(redis.StringSliceCmd)
					byCreatorKeyResult.SetVal([]string{byCreatorKey})

					namespacesKeyResult := new(redis.StringSliceCmd)
					namespacesKeyResult.SetVal([]string{namespacesKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(6)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), "*")
					libraryKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*")
					relatedKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*")
					byCreatorKey := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*")
					namespacesKey := composeCacheKey(model.ResourceTypeNamespace.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					libraryKeyResult := new(redis.StringSliceCmd)
					libraryKeyResult.SetVal([]string{libraryKey})

					relatedKeyResult := new(redis.StringSliceCmd)
					relatedKeyResult.SetVal([]string{relatedKey})

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

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, libraryKey).Return(libraryKeyResult)
					dbClient.EXPECT().Keys(ctx, relatedKey).Return(relatedKeyResult)
					dbClient.EXPECT().Keys(ctx, byCreatorKey).Return(byCreatorKeyResult)
					dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(7)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(7)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, libraryKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relatedKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byCreatorKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				documentRepo: func(_ *gomock.Controller, _ context.Context, _ model.ID) repository.DocumentRepository {
					return mockrepo.NewMockDocumentRepository(nil)
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
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedDocumentRepository {
				r, err := repository.NewCachedDocumentRepository(
					tt.fields.documentRepo(ctrl, tt.args.ctx, tt.args.id),
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
