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

func TestCachedFolderRepository_Create(t *testing.T) {
	t.Parallel()

	opts := CreateFolderOpts{
		Library:   model.MustNewID(model.ResourceTypeNamespace),
		Name:      "Guides",
		CreatedBy: model.MustNewID(model.ResourceTypeUser),
	}

	t.Run("create folder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		listKey := composeCacheKey(model.ResourceTypeFolder.String(), "List", "*")
		listKeyResult := new(redis.StringSliceCmd)
		listKeyResult.SetVal([]string{listKey})

		dbClient := mock.NewUniversalClient(ctrl)
		dbClient.EXPECT().Keys(ctx, listKey).Return(listKeyResult)

		db, err := NewRedisDatabase(WithRedisClient(dbClient))
		require.NoError(t, err)

		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

		cacheRepo := mock.NewCacheBackend(ctrl)
		cacheRepo.EXPECT().Delete(ctx, listKey).Return(nil)

		repo := NewMockFolderRepository(ctrl)
		repo.EXPECT().Create(ctx, opts).Return(&Folder{Name: opts.Name}, nil)

		r := &RedisCachedFolderRepository{
			cacheRepo: &redisBaseRepository{
				db:     db,
				cache:  cacheRepo,
				tracer: tracer,
				logger: mock.NewMockLogger(ctrl),
			},
			folderRepo: repo,
		}
		got, err := r.Create(ctx, opts)
		require.NoError(t, err)
		assert.Equal(t, opts.Name, got.Name)
	})
}

func TestCachedFolderRepository_Get(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeFolder)
	folder := &Folder{ID: id, Name: "Guides"}

	t.Run("uncached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		key := composeCacheKey(model.ResourceTypeFolder.String(), "Get", id.String())

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
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
			Value: folder,
		}).Return(nil)

		repo := NewMockFolderRepository(ctrl)
		repo.EXPECT().Get(ctx, id).Return(folder, nil)

		r := &RedisCachedFolderRepository{
			cacheRepo: &redisBaseRepository{
				db:     db,
				cache:  cacheRepo,
				tracer: tracer,
				logger: mock.NewMockLogger(ctrl),
			},
			folderRepo: repo,
		}
		got, err := r.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, folder, got)
	})
}

func TestCachedFolderRepository_List(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	folders := []*Folder{{ID: model.MustNewID(model.ResourceTypeFolder), Name: "Guides"}}
	limit := 10

	t.Run("uncached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		key := composeCacheKey(model.ResourceTypeFolder.String(), "List", libraryID.String(), "root", model.MustNewNilID(model.ResourceTypeUser).String(), "", limit)

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
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
			Value: Page[*Folder]{Items: folders},
		}).Return(nil)

		repo := NewMockFolderRepository(ctrl)
		repo.EXPECT().List(ctx, libraryID, (*model.ID)(nil), model.MustNewNilID(model.ResourceTypeUser), nil, CursorPage{Size: limit}).Return(Page[*Folder]{Items: folders}, nil)

		r := &RedisCachedFolderRepository{
			cacheRepo: &redisBaseRepository{
				db:     db,
				cache:  cacheRepo,
				tracer: tracer,
				logger: mock.NewMockLogger(ctrl),
			},
			folderRepo: repo,
		}
		got, err := r.List(ctx, libraryID, nil, model.MustNewNilID(model.ResourceTypeUser), nil, CursorPage{Size: limit})
		require.NoError(t, err)
		assert.Equal(t, folders, got.Items)
	})
}
