package repository_test

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/pkg/password"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
)

func cachedUserCreateCacheOpts(ctx context.Context, ctrl *gomock.Controller) []repository.RedisRepositoryOption {
	getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")
	organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
	rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

	getAllKeyResult := new(redis.StringSliceCmd)
	getAllKeyResult.SetVal([]string{getAllKey})
	organizationsKeyResult := new(redis.StringSliceCmd)
	organizationsKeyResult.SetVal([]string{organizationsKey})
	rolesKeyResult := new(redis.StringSliceCmd)
	rolesKeyResult.SetVal([]string{rolesKey})

	dbClient := mockrepo.NewMockUniversalClient(ctrl)
	dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
	dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
	dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
	if err != nil {
		panic(err)
	}

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).Times(3)
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

	cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
	cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
	cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
	cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)

	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(cacheRepo),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(tracer),
	}
}

func TestCachedUserRepository_CreateWithSeatLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	opts := repository.CreateUserOpts{
		Username:   "test-user",
		Email:      "user@example.com",
		Password:   password.UnusablePassword,
		Status:     model.UserStatusActive,
		FirstName:  "Test",
		LastName:   "User",
		Picture:    "https://example.com/picture.jpg",
		Title:      "Software Engineer",
		Bio:        "I'm a software engineer",
		Phone:      "+1234567890",
		Address:    "Remote",
		Links:      make([]string, 0),
		Languages:  make([]model.Language, 0),
		Activation: ptrActivation(repository.FiniteActivation(3)),
	}

	t.Run("create active user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		limited := opts
		repo := mockrepo.NewMockUserRepository(ctrl)
		repo.EXPECT().Create(ctx, limited).Return(&repository.User{Status: model.UserStatusActive}, nil)
		cached, err := repository.NewCachedUserRepository(repo, cachedUserCreateCacheOpts(ctx, ctrl)...)
		require.NoError(t, err)
		_, err = cached.Create(ctx, limited)
		require.NoError(t, err)
	})

	t.Run("create active user when the seat limit is reached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		limited := opts
		limited.Activation = ptrActivation(repository.FiniteActivation(1))
		repo := mockrepo.NewMockUserRepository(ctrl)
		repo.EXPECT().Create(ctx, limited).Return(nil, repository.ErrSeatLimitReached)
		cached, err := repository.NewCachedUserRepository(repo, cachedUserCreateCacheOpts(ctx, ctrl)...)
		require.NoError(t, err)
		_, err = cached.Create(ctx, limited)
		require.ErrorIs(t, err, repository.ErrSeatLimitReached)
	})
}

func TestCachedUserRepository_ActiveHumanCount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockrepo.NewMockUserRepository(ctrl)
	repo.EXPECT().ActiveHumanCount(gomock.Any()).Return(4, nil)

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
	require.NoError(t, err)
	cached, err := repository.NewCachedUserRepository(repo,
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(mocktrace.NewMockTracer(ctrl)),
	)
	require.NoError(t, err)
	count, err := cached.ActiveHumanCount(context.Background())
	require.NoError(t, err)
	require.Equal(t, 4, count)
}

func TestCachedUserRepository_ActivateCacheFailureIsNonFatal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := model.MustNewID(model.ResourceTypeUser)
	user := &repository.User{ID: userID, Email: "user@example.com", Status: model.UserStatusActive}
	ctrl := gomock.NewController(t)

	userRepo := mockrepo.NewMockUserRepository(ctrl)
	userRepo.EXPECT().Activate(ctx, userID, gomock.Any(), repository.UnrestrictedActivation()).Return(user, nil)

	dbClient := mockrepo.NewMockUniversalClient(ctrl)
	empty := new(redis.StringSliceCmd)
	empty.SetVal(nil)
	dbClient.EXPECT().Keys(gomock.Any(), gomock.Any()).Return(empty).AnyTimes()

	cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
	cacheRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).AnyTimes()
	cacheRepo.EXPECT().Set(gomock.Any()).Return(assert.AnError).AnyTimes()
	cacheRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(assert.AnError).AnyTimes()

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
	require.NoError(t, err)
	cached, err := repository.NewCachedUserRepository(
		userRepo,
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(cacheRepo),
	)
	require.NoError(t, err)

	got, err := cached.Activate(ctx, userID, repository.UpdateUserOpts{}, repository.UnrestrictedActivation())
	require.NoError(t, err)
	require.Same(t, user, got)
}

func TestCachedUserRepository_AcceptInvitationCacheFailureIsNonFatal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	user := &repository.User{ID: userID}
	ctrl := gomock.NewController(t)

	userRepo := mockrepo.NewMockUserRepository(ctrl)
	userRepo.EXPECT().AcceptInvitation(ctx, userID, orgID, "", nil, nil).Return(user, nil)

	dbClient := mockrepo.NewMockUniversalClient(ctrl)
	empty := new(redis.StringSliceCmd)
	empty.SetVal(nil)
	dbClient.EXPECT().Keys(gomock.Any(), gomock.Any()).Return(empty).AnyTimes()

	cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
	cacheRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).AnyTimes()
	cacheRepo.EXPECT().Set(gomock.Any()).Return(assert.AnError).AnyTimes()
	cacheRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(assert.AnError).AnyTimes()

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
	require.NoError(t, err)
	cached, err := repository.NewCachedUserRepository(
		userRepo,
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(cacheRepo),
	)
	require.NoError(t, err)

	got, err := cached.AcceptInvitation(ctx, userID, orgID, "", nil, nil)
	require.NoError(t, err)
	require.Same(t, user, got)
}
