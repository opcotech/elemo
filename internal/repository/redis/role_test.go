package redis

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedRoleRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) *baseRepository
		roleRepo  func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) repository.RoleRepository
	}
	type args struct {
		ctx       context.Context
		createdBy model.ID
		belongsTo model.ID
		role      *model.Role
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new role",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Create", ctx, createdBy, belongsTo, role).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role: &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "add new role with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Create", ctx, createdBy, belongsTo, role).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role: &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "add new role with belongs to cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Create", ctx, createdBy, belongsTo, role).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role: &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new role with organization cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Create", ctx, createdBy, belongsTo, role).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role: &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new role with project cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) *baseRepository {
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.On("Keys", ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, belongsToKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Create", ctx, createdBy, belongsTo, role).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				createdBy: model.MustNewID(model.ResourceTypeUser),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				role: &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
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
			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.createdBy, tt.args.belongsTo, tt.args.role),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.createdBy, tt.args.belongsTo, tt.args.role),
			}
			err := r.Create(tt.args.ctx, tt.args.createdBy, tt.args.belongsTo, tt.args.role)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, role *model.Role) *baseRepository
		roleRepo  func(ctx context.Context, id model.ID, role *model.Role) repository.RoleRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Role
		wantErr error
	}{
		{
			name: "get uncached role",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Get", ctx, id).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			want: func(id model.ID) *model.Role {
				return &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached role",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(role, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, role *model.Role) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			want: func(id model.ID) *model.Role {
				return &model.Role{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached role error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached role error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, role *model.Role) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached role cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Get", ctx, id).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var want *model.Role
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedRoleRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseRepository
		roleRepo  func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) repository.RoleRepository
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
		want    []*model.Role
		wantErr error
	}{
		{
			name: "get uncached roles",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: roles,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(roles, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Role{
				{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached roles",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(roles, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Role{
				{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					Members:     make([]model.ID, 0),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached roles error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get roles cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached roles cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: roles,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, belongsTo model.ID, offset, limit int, roles []*model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("GetAllBelongsTo", ctx, belongsTo, offset, limit).Return(roles, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedRoleRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, role *model.Role) *baseRepository
		roleRepo  func(ctx context.Context, id model.ID, patch map[string]any, role *model.Role) repository.RoleRepository
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
		want    *model.Role
		wantErr error
	}{
		{
			name: "update role",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, patch map[string]any, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Update", ctx, id, patch).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
				patch: map[string]any{
					"name":        "updated role",
					"description": "updated description",
				},
			},
			want: &model.Role{
				ID:          model.MustNewID(model.ResourceTypeRole),
				Name:        "test role",
				Description: "test description",
			},
		},
		{
			name: "update role with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, patch map[string]any, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
				patch: map[string]any{
					"name":        "updated role",
					"description": "updated description",
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update role set cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					dbClient := new(mock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, patch map[string]any, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Update", ctx, id, patch).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
				patch: map[string]any{
					"name":        "updated role",
					"description": "updated description",
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update role delete get all cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, role *model.Role) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(assert.AnError)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID, patch map[string]any, role *model.Role) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Update", ctx, id, patch).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
				patch: map[string]any{
					"name":        "updated role",
					"description": "updated description",
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedRoleRepository_AddMember(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id, memberID model.ID) *baseRepository
		roleRepo  func(ctx context.Context, id, memberID model.ID) repository.RoleRepository
	}
	type args struct {
		ctx      context.Context
		id       model.ID
		memberID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete role success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("AddMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete role with role deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("AddMember", ctx, id, memberID).Return(repository.ErrRoleDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrRoleDelete,
		},
		{
			name: "delete role with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					dbClient := new(mock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("AddMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete role cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
			}
			err := r.AddMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_RemoveMember(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id, memberID model.ID) *baseRepository
		roleRepo  func(ctx context.Context, id, memberID model.ID) repository.RoleRepository
	}
	type args struct {
		ctx      context.Context
		id       model.ID
		memberID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete role success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("RemoveMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete role with role deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("RemoveMember", ctx, id, memberID).Return(repository.ErrRoleDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrRoleDelete,
		},
		{
			name: "delete role with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					dbClient := new(mock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("RemoveMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete role cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id, memberID model.ID) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeRole),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
			}
			err := r.RemoveMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID) *baseRepository
		roleRepo  func(ctx context.Context, id model.ID) repository.RoleRepository
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
			name: "delete role success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					projectKeyCmd := new(redis.StringSliceCmd)
					projectKeyCmd.SetVal([]string{projectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationKey).Return(organizationKeyCmd)
					dbClient.On("Keys", ctx, projectKey).Return(projectKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
		},
		{
			name: "delete role with role deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					projectKeyCmd := new(redis.StringSliceCmd)
					projectKeyCmd.SetVal([]string{projectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationKey).Return(organizationKeyCmd)
					dbClient.On("Keys", ctx, projectKey).Return(projectKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrRoleDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrRoleDelete,
		},
		{
			name: "delete role with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

					dbClient := new(mock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID) repository.RoleRepository {
					repo := new(mock.RoleRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete role with get all cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete role with organization cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete role with project cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					projectKeyCmd := new(redis.StringSliceCmd)
					projectKeyCmd.SetVal([]string{projectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, organizationKey).Return(organizationKeyCmd)
					dbClient.On("Keys", ctx, projectKey).Return(projectKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, organizationKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				roleRepo: func(ctx context.Context, id model.ID) repository.RoleRepository {
					return new(mock.RoleRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeRole),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				roleRepo:  tt.fields.roleRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
