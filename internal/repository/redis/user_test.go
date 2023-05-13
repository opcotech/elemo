package redis

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	testMock "github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedUserRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, user *model.User) *baseRepository
		userRepo  func(ctx context.Context, user *model.User) repository.UserRepository
	}
	type args struct {
		ctx  context.Context
		user *model.User
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
				cacheRepo: func(ctx context.Context, user *model.User) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Create", ctx, user).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
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
			name: "add new user with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, user *model.User) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Create", ctx, user).Return(repository.ErrUserCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
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
			wantErr: repository.ErrUserCreate,
		},
		{
			name: "add new user get all cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, user *model.User) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new user organizations cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, user *model.User) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new user roles cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, user *model.User) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.user),
				userRepo:  tt.fields.userRepo(tt.args.ctx, tt.args.user),
			}
			err := r.Create(tt.args.ctx, tt.args.user)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedUserRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, user *model.User) *baseRepository
		userRepo  func(ctx context.Context, id model.ID, user *model.User) repository.UserRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.User
		wantErr error
	}{
		{
			name: "get uncached user",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Get", ctx, id).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: func(id model.ID) *model.User {
				return &model.User{
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
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(user, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: func(id model.ID) *model.User {
				return &model.User{
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
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Get", ctx, id).Return(user, nil)
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
			var want *model.User
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				userRepo:  tt.fields.userRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedUserRepository_GetByEmail(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, email string, user *model.User) *baseRepository
		userRepo  func(ctx context.Context, email string, user *model.User) repository.UserRepository
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(email string) *model.User
		wantErr error
	}{
		{
			name: "get uncached user",
			fields: fields{
				cacheRepo: func(ctx context.Context, email string, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, email string, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("GetByEmail", ctx, email).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			want: func(email string) *model.User {
				return &model.User{
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
				cacheRepo: func(ctx context.Context, email string, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(user, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, email string, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			want: func(email string) *model.User {
				return &model.User{
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
				cacheRepo: func(ctx context.Context, email string, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, email string, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("GetByEmail", ctx, email).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, email string, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, email string, user *model.User) repository.UserRepository {
					return new(testMock.UserRepository)
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
				cacheRepo: func(ctx context.Context, email string, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, email string, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("GetByEmail", ctx, email).Return(user, nil)
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
			var want *model.User
			if tt.want != nil {
				want = tt.want(tt.args.email)
			}

			r := &CachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.email, want),
				userRepo:  tt.fields.userRepo(tt.args.ctx, tt.args.email, want),
			}
			got, err := r.GetByEmail(tt.args.ctx, tt.args.email)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedUserRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, offset, limit int, users []*model.User) *baseRepository
		userRepo  func(ctx context.Context, offset, limit int, users []*model.User) repository.UserRepository
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
		want    []*model.User
		wantErr error
	}{
		{
			name: "get uncached users",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, users []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: users,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, offset, limit int, users []*model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("GetAll", ctx, offset, limit).Return(users, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*model.User{
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
				cacheRepo: func(ctx context.Context, offset, limit int, users []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(users, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, offset, limit int, users []*model.User) repository.UserRepository {
					return new(testMock.UserRepository)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*model.User{
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
				cacheRepo: func(ctx context.Context, offset, limit int, users []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, offset, limit int, users []*model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("GetAll", ctx, offset, limit).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, offset, limit int, users []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, offset, limit int, users []*model.User) repository.UserRepository {
					return new(testMock.UserRepository)
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
				cacheRepo: func(ctx context.Context, offset, limit int, users []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: users,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, offset, limit int, users []*model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("GetAll", ctx, offset, limit).Return(users, nil)
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
			r := &CachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
				userRepo:  tt.fields.userRepo(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedUserRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, user *model.User) *baseRepository
		userRepo  func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) repository.UserRepository
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
		want    *model.User
		wantErr error
	}{
		{
			name: "update user",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email)
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					byEmailKeyCmd := new(redis.StringCmd)
					byEmailKeyCmd.SetVal(id.String())

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Delete", ctx, byEmailKey).Return(byEmailKeyCmd, nil)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Update", ctx, id, patch).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				patch: map[string]any{
					"username": "updated-user",
					"email":    "updated@example.com",
				},
			},
			want: &model.User{
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
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(testMock.CacheRepository),
						tracer: new(testMock.Tracer),
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				patch: map[string]any{
					"username": "updated-user",
					"email":    "updated@example.com",
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update user set cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Update", ctx, id, patch).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				patch: map[string]any{
					"username": "updated-user",
					"email":    "updated@example.com",
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update user delete by email cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email)

					byEmailKeyCmd := new(redis.StringCmd)
					byEmailKeyCmd.SetVal(byEmailKey)

					dbClient := new(testMock.RedisClient)
					dbClient.On("Delete", byEmailKey).Return(byEmailKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(assert.AnError)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Update", ctx, id, patch).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				patch: map[string]any{
					"username": "updated-user",
					"email":    "updated@example.com",
				},
			},
			want: &model.User{
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update user delete get all cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, user *model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", user.Email)
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					byEmailKeyCmd := new(redis.StringCmd)
					byEmailKeyCmd.SetVal(byEmailKey)

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Delete", byEmailKey).Return(byEmailKeyCmd, nil)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(assert.AnError)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: user,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Update", ctx, id, patch).Return(user, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
				patch: map[string]any{
					"username": "updated-user",
					"email":    "updated@example.com",
				},
			},
			want: &model.User{
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &CachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				userRepo:  tt.fields.userRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedUserRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID) *baseRepository
		userRepo  func(ctx context.Context, id model.ID) repository.UserRepository
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrUserDelete)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					repo := new(testMock.UserRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byEmailKey).Return(byEmailKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					return new(testMock.UserRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
					byEmailKey := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", "*")
					getAllKey := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", "*")

					byEmailKeyCmd := new(redis.StringSliceCmd)
					byEmailKeyCmd.SetVal([]string{byEmailKey})

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					return new(testMock.UserRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					return new(testMock.UserRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byEmailKey).Return(byEmailKeyCmd)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyCmd)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, byEmailKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				userRepo: func(ctx context.Context, id model.ID) repository.UserRepository {
					return new(testMock.UserRepository)
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
			r := &CachedUserRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				userRepo:  tt.fields.userRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
