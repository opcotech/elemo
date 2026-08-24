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

func TestCachedFolderRepository_Create(t *testing.T) {
	t.Parallel()

	opts := repository.CreateFolderOpts{
		Library:   model.MustNewID(model.ResourceTypeNamespace),
		Name:      "Guides",
		CreatedBy: model.MustNewID(model.ResourceTypeUser),
	}

	t.Run("create folder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		listKey := composeCacheKey(model.ResourceTypeFolder.String(), "*", "ListForLibrary", "*")
		listKeyResult := new(redis.StringSliceCmd)
		listKeyResult.SetVal([]string{listKey})

		dbClient := mockrepo.NewMockUniversalClient(ctrl)
		dbClient.EXPECT().Keys(ctx, listKey).Return(listKeyResult)

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
		require.NoError(t, err)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

		cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
		cacheRepo.EXPECT().Delete(ctx, listKey).Return(nil)

		repo := mockrepo.NewMockFolderRepository(ctrl)
		repo.EXPECT().Create(ctx, opts).Return(&repository.Folder{Name: opts.Name}, nil)

		r := func() *repository.RedisCachedFolderRepository {
			r, err := repository.NewCachedFolderRepository(
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
		got, err := r.Create(ctx, opts)
		require.NoError(t, err)
		assert.Equal(t, opts.Name, got.Name)
	})
}

func TestCachedFolderRepository_Get(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeFolder)
	folder := &repository.Folder{ID: id, Name: "Guides"}

	t.Run("uncached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		key := composeCacheKey(model.ResourceTypeFolder.String(), "Get", id.String())

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
			Value: folder,
		}).Return(nil)

		repo := mockrepo.NewMockFolderRepository(ctrl)
		repo.EXPECT().Get(ctx, id).Return(folder, nil)

		r := func() *repository.RedisCachedFolderRepository {
			r, err := repository.NewCachedFolderRepository(
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
		got, err := r.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, folder, got)
	})
}

func TestCachedFolderRepository_List(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	folders := []*repository.Folder{{ID: model.MustNewID(model.ResourceTypeFolder), Name: "Guides"}}
	limit := 10

	t.Run("uncached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		key := mustPlanCacheKey(t, repository.FolderListQuery{
			LibraryID: libraryID,
			ActorID:   model.MustNewNilID(model.ResourceTypeUser),
			Page:      repository.CursorPage{Size: limit},
			Order:     repository.SortDirectionDesc,
		}, model.ResourceTypeFolder.String(), "ListForLibrary", libraryID.String())

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
			Value: repository.Page[*repository.Folder]{Items: folders},
		}).Return(nil)

		repo := mockrepo.NewMockFolderRepository(ctrl)
		repo.EXPECT().ListForLibrary(ctx, repository.FolderListQuery{
			LibraryID: libraryID,
			ActorID:   model.MustNewNilID(model.ResourceTypeUser),
			Page:      repository.CursorPage{Size: limit},
			Order:     repository.SortDirectionDesc,
		}).Return(repository.Page[*repository.Folder]{Items: folders}, nil)

		r := func() *repository.RedisCachedFolderRepository {
			r, err := repository.NewCachedFolderRepository(
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
		got, err := r.ListForLibrary(ctx, repository.FolderListQuery{
			LibraryID: libraryID,
			ActorID:   model.MustNewNilID(model.ResourceTypeUser),
			Page:      repository.CursorPage{Size: limit},
			Order:     repository.SortDirectionDesc,
		})
		require.NoError(t, err)
		assert.Equal(t, folders, got.Items)
	})
}
