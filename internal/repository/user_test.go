package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedUserRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateUserOpts) *redisBaseRepository
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, opts CreateUserOpts) UserRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateUserOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateUserOpts) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&User{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateUserOpts{
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
					Languages: make([]model.Language, 0)},
			},
		},
		{
			name: "add new user with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateUserOpts) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, ErrUserCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateUserOpts{
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
					Languages: make([]model.Language, 0)},
			},
			wantErr: ErrUserCreate,
		},
		{
			name: "add new user get all cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateUserOpts) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateUserOpts{
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
					Languages: make([]model.Language, 0)},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create new user organizations cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateUserOpts) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateUserOpts{
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
					Languages: make([]model.Language, 0)},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "create new user roles cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateUserOpts) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateUserOpts{
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
					Languages: make([]model.Language, 0)},
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
			r := &RedisCachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts),
				userRepo:  tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedUserRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) UserRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *User
		wantErr error
	}{
		{
			name: "get uncached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: func(id model.ID) *User {
				return &User{
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
					Documents:   make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

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
						if ptr, ok := dst.(**User); ok {
							*ptr = user
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *User) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: func(id model.ID) *User {
				return &User{
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
					Documents:   make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *User) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached user cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
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
			var want *User
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				userRepo:  tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedUserRepository_GetByEmail(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) *redisBaseRepository
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) UserRepository
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(email string) *User
		wantErr error
	}{
		{
			name: "get uncached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().GetByEmail(ctx, email).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			want: func(email string) *User {
				return &User{
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
					Documents:   make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

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
						if ptr, ok := dst.(**User); ok {
							*ptr = user
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ *User) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			want: func(email string) *User {
				return &User{
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
					Documents:   make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().GetByEmail(ctx, email).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached user error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ string, _ *User) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached user cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, email string, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().GetByEmail(ctx, email).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
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
			var want *User
			if tt.want != nil {
				want = tt.want(tt.args.email)
			}

			r := &RedisCachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.email, want),
				userRepo:  tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.email, want),
			}
			got, err := r.GetByEmail(tt.args.ctx, tt.args.email)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedUserRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) *redisBaseRepository
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) UserRepository
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
		want    []*User
		wantErr error
	}{
		{
			name: "get uncached users",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: users,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().GetAll(ctx, offset, limit).Return(users, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*User{
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
					Documents:   make([]model.ID, 0),
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
					Documents:   make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached users",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

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
						if listPtr, ok := dst.(*[]*User); ok {
							*listPtr = users
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ int, _ []*User) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*User{
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
					Documents:   make([]model.ID, 0),
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
					Documents:   make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached users error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().GetAll(ctx, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get users cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ int, _ []*User) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached users cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: users,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().GetAll(ctx, offset, limit).Return(users, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
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
			r := &RedisCachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
				userRepo:  tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedUserRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, user *User) UserRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateUserOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr error
	}{
		{
			name: "update user",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email)
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			want: &User{
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
				Documents:   make([]model.ID, 0),
				Permissions: make([]model.ID, 0),
			},
		},
		{
			name: "update user with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *User) *redisBaseRepository {
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
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, _ *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update user set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update user delete by email cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			want: &User{
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
				Documents:   make([]model.ID, 0),
				Permissions: make([]model.ID, 0),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "update user delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email)
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, user *User) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				opts: UpdateUserOpts{
					Username: optional.Some("updated-user"),
					Email:    optional.Some("updated@example.com"),
				},
			},
			want: &User{
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
				Documents:   make([]model.ID, 0),
				Permissions: make([]model.ID, 0),
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

			r := &RedisCachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				userRepo:  tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
			}
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		userRepo  func(ctrl *gomock.Controller, ctx context.Context, id model.ID) UserRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) UserRepository {
					repo := NewMockUserRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) UserRepository {
					repo := NewMockUserRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrUserDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrUserDelete,
		},
		{
			name: "delete user with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

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
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) UserRepository {
					repo := NewMockUserRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete user cache by email key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete user cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete user cache by organization key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationsKeyCmd := new(redis.StringSliceCmd)
					organizationsKeyCmd.SetVal([]string{organizationsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete user cache by roles key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, byEmailKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				userRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) UserRepository {
					return NewMockUserRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				userRepo:  tt.fields.userRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
