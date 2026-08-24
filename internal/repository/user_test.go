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
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/password"
)

func TestCachedUserRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateUserOpts) []repository.RedisRepositoryOption
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateUserOpts) repository.UserRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateUserOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateUserOpts) []repository.RedisRepositoryOption {
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

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

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
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateUserOpts) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.User{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateUserOpts{
					Username:  "test-user",
					Email:     "user@example.com",
					Password:  password.UnusablePassword,
					Status:    model.UserStatusActive,
					FirstName: "Test",
					LastName:  "User",
					Picture:   "https://example.com/picture.jpg",
					Title:     "Software Engineer",
					Bio:       "I'm a software engineer",
					Phone:     "+1234567890",
					Address:   "Remote",
					Links:     make([]string, 0),
					Languages: make([]model.Language, 0),
				},
			},
		},
		{
			name: "add new user with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateUserOpts) []repository.RedisRepositoryOption {
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

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

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
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateUserOpts) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrUserCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateUserOpts{
					Username:  "test-user",
					Email:     "user@example.com",
					Password:  password.UnusablePassword,
					Status:    model.UserStatusActive,
					FirstName: "Test",
					LastName:  "User",
					Picture:   "https://example.com/picture.jpg",
					Title:     "Software Engineer",
					Bio:       "I'm a software engineer",
					Phone:     "+1234567890",
					Address:   "Remote",
					Links:     make([]string, 0),
					Languages: make([]model.Language, 0),
				},
			},
			wantErr: repository.ErrUserCreate,
		},
		{
			name: "add new user get all cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateUserOpts) []repository.RedisRepositoryOption {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateUserOpts) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateUserOpts{
					Username:  "test-user",
					Email:     "user@example.com",
					Password:  password.UnusablePassword,
					Status:    model.UserStatusActive,
					FirstName: "Test",
					LastName:  "User",
					Picture:   "https://example.com/picture.jpg",
					Title:     "Software Engineer",
					Bio:       "I'm a software engineer",
					Phone:     "+1234567890",
					Address:   "Remote",
					Links:     make([]string, 0),
					Languages: make([]model.Language, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new user organizations cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateUserOpts) []repository.RedisRepositoryOption {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateUserOpts) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateUserOpts{
					Username:  "test-user",
					Email:     "user@example.com",
					Password:  password.UnusablePassword,
					Status:    model.UserStatusActive,
					FirstName: "Test",
					LastName:  "User",
					Picture:   "https://example.com/picture.jpg",
					Title:     "Software Engineer",
					Bio:       "I'm a software engineer",
					Phone:     "+1234567890",
					Address:   "Remote",
					Links:     make([]string, 0),
					Languages: make([]model.Language, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new user roles cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateUserOpts) []repository.RedisRepositoryOption {
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

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateUserOpts) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateUserOpts{
					Username:  "test-user",
					Email:     "user@example.com",
					Password:  password.UnusablePassword,
					Status:    model.UserStatusActive,
					FirstName: "Test",
					LastName:  "User",
					Picture:   "https://example.com/picture.jpg",
					Title:     "Software Engineer",
					Bio:       "I'm a software engineer",
					Phone:     "+1234567890",
					Address:   "Remote",
					Links:     make([]string, 0),
					Languages: make([]model.Language, 0),
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
			defer ctrl.Finish()
			r := func() *repository.RedisCachedUserRepository {
				r, err := repository.NewCachedUserRepository(
					tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.opts),
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

func TestCachedUserRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) repository.UserRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.User
		wantErr error
	}{
		{
			name: "get uncached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.UserDetailProjection()).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: func(id model.ID) *repository.User {
				return &repository.User{
					ID:          id,
					Username:    "test-user",
					Email:       "user@example.com",
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))

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
						if ptr, ok := dst.(**repository.User); ok {
							*ptr = user
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.User) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: func(id model.ID) *repository.User {
				return &repository.User{
					ID:          id,
					Username:    "test-user",
					Email:       "user@example.com",
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.User) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached user cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.UserDetailProjection()).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *repository.User
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedUserRepository {
				r, err := repository.NewCachedUserRepository(
					tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.UserDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedUserRepository_GetByEmail(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) []repository.RedisRepositoryOption
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) repository.UserRepository
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(email string) *repository.User
		wantErr error
	}{
		{
			name: "get uncached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email, projectionCacheValue(repository.UserDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			want: func(email string) *repository.User {
				return &repository.User{
					ID:          model.MustNewID(model.ResourceTypeUser),
					Username:    "test-user",
					Email:       email,
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email, projectionCacheValue(repository.UserDetailProjection()))

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
						if ptr, ok := dst.(**repository.User); ok {
							*ptr = user
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ *repository.User) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			want: func(email string) *repository.User {
				return &repository.User{
					ID:          model.MustNewID(model.ResourceTypeUser),
					Username:    "test-user",
					Email:       email,
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email, projectionCacheValue(repository.UserDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email, projectionCacheValue(repository.UserDetailProjection()))

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ *repository.User) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached user cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email, projectionCacheValue(repository.UserDetailProjection()))

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *repository.User
			if tt.want != nil {
				want = tt.want(tt.args.email)
			}

			r := func() *repository.RedisCachedUserRepository {
				r, err := repository.NewCachedUserRepository(
					tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.email, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.email, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.GetByEmail(tt.args.ctx, tt.args.email, repository.UserDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedUserRepository_List(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) []repository.RedisRepositoryOption
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) repository.UserRepository
	}
	type args struct {
		ctx    context.Context
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.User
		wantErr error
	}{
		{
			name: "get uncached users",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "List", projectionCacheValue(repository.UserListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.User]{Items: users},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().List(ctx, repository.CursorPage{Size: limit}, repository.UserListProjection()).Return(repository.Page[*repository.User]{Items: users}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*repository.User{
				{
					ID:          model.MustNewID(model.ResourceTypeUser),
					Username:    "test-user",
					Email:       "user@example.com",
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeUser),
					Username:    "test-user",
					Email:       "user@example.com",
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached users",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "List", projectionCacheValue(repository.UserListProjection()), "", limit)

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
						if ptr, ok := dst.(*repository.Page[*repository.User]); ok {
							*ptr = repository.Page[*repository.User]{Items: users}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ int, _ []*repository.User) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*repository.User{
				{
					ID:          model.MustNewID(model.ResourceTypeUser),
					Username:    "test-user",
					Email:       "user@example.com",
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeUser),
					Username:    "test-user",
					Email:       "user@example.com",
					Password:    password.UnusablePassword,
					Status:      model.UserStatusActive,
					FirstName:   "Test",
					LastName:    "User",
					Picture:     "https://example.com/picture.jpg",
					Title:       "Software Engineer",
					Bio:         "I'm a software engineer",
					Phone:       "+1234567890",
					Address:     "Remote",
					Links:       make([]string, 0),
					Languages:   make([]model.Language, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached users error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, _ []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "List", projectionCacheValue(repository.UserListProjection()), "", limit)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, _ []*repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().List(ctx, repository.CursorPage{Size: limit}, repository.UserListProjection()).Return(repository.Page[*repository.User]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get users cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, _ []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "List", projectionCacheValue(repository.UserListProjection()), "", limit)

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ int, _ []*repository.User) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached users cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "List", projectionCacheValue(repository.UserListProjection()), "", limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.User]{Items: users},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, _, limit int, users []*repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().List(ctx, repository.CursorPage{Size: limit}, repository.UserListProjection()).Return(repository.Page[*repository.User]{Items: users}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedUserRepository {
				r, err := repository.NewCachedUserRepository(
					tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.List(tt.args.ctx, repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.UserListProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedUserRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateUserOpts, user *repository.User) repository.UserRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateUserOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.User
		wantErr error
	}{
		{
			name: "update user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email, "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(7)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, issueListUserGenKey(id), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueListUserGenKey(id), Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueListProjectionEpochKey(), Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateUserOpts, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: repository.UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			want: &repository.User{
				ID:          model.MustNewID(model.ResourceTypeUser),
				Username:    "test-user",
				Email:       "user@example.com",
				Password:    password.UnusablePassword,
				Status:      model.UserStatusActive,
				FirstName:   "Test",
				LastName:    "User",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I'm a software engineer",
				Phone:       "+1234567890",
				Address:     "Remote",
				Links:       make([]string, 0),
				Languages:   make([]model.Language, 0),
				Permissions: make([]model.ID, 0),
			},
		},
		{
			name: "update user with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.User) []repository.RedisRepositoryOption {
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
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateUserOpts, _ *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: repository.UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update user set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
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
						Value: user,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateUserOpts, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: repository.UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update user delete by email cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email, "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateUserOpts, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: repository.UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			want: &repository.User{
				ID:          model.MustNewID(model.ResourceTypeUser),
				Username:    "test-user",
				Email:       "user@example.com",
				Password:    password.UnusablePassword,
				Status:      model.UserStatusActive,
				FirstName:   "Test",
				LastName:    "User",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I'm a software engineer",
				Phone:       "+1234567890",
				Address:     "Remote",
				Links:       make([]string, 0),
				Languages:   make([]model.Language, 0),
				Permissions: make([]model.ID, 0),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update user delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(repository.UserDetailProjection()))
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email, "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateUserOpts, user *repository.User) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: repository.UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			want: &repository.User{
				ID:          model.MustNewID(model.ResourceTypeUser),
				Username:    "test-user",
				Email:       "user@example.com",
				Password:    password.UnusablePassword,
				Status:      model.UserStatusActive,
				FirstName:   "Test",
				LastName:    "User",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I'm a software engineer",
				Phone:       "+1234567890",
				Address:     "Remote",
				Links:       make([]string, 0),
				Languages:   make([]model.Language, 0),
				Permissions: make([]model.ID, 0),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := func() *repository.RedisCachedUserRepository {
				r, err := repository.NewCachedUserRepository(
					tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedUserRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.UserRepository
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
			name: "delete user success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationsKeyCmd := new(redis.StringSliceCmd)
					organizationsKeyCmd.SetVal([]string{organizationsKey})

					rolesKeyCmd := new(redis.StringSliceCmd)
					rolesKeyCmd.SetVal([]string{rolesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(9)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, issueListUserGenKey(id), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueListUserGenKey(id), Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueListProjectionEpochKey(), Value: int64(1)}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "delete user with user deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationsKeyCmd := new(redis.StringSliceCmd)
					organizationsKeyCmd.SetVal([]string{organizationsKey})

					rolesKeyCmd := new(redis.StringSliceCmd)
					rolesKeyCmd.SetVal([]string{rolesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(9)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, issueListUserGenKey(id), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueListUserGenKey(id), Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueListProjectionEpochKey(), Value: int64(1)}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrUserDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrUserDelete,
		},
		{
			name: "delete user with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.UserRepository {
					repo := mockrepo.NewMockUserRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete user cache by email key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete user cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete user cache by organization key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationsKeyCmd := new(redis.StringSliceCmd)
					organizationsKeyCmd.SetVal([]string{organizationsKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete user cache by roles key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), "*")
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "List", "*", "*", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationsKeyCmd := new(redis.StringSliceCmd)
					organizationsKeyCmd.SetVal([]string{organizationsKey})

					rolesKeyCmd := new(redis.StringSliceCmd)
					rolesKeyCmd.SetVal([]string{rolesKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.UserRepository {
					return mockrepo.NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedUserRepository {
				r, err := repository.NewCachedUserRepository(
					tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.id),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
