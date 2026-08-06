package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedRoleRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) RoleRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateRoleOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) *redisBaseRepository {
					belongsTo := opts.BelongsTo
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&Role{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateRoleOpts{
					Name:        "test role",
					Description: "test description",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "add new role with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) *redisBaseRepository {
					belongsTo := opts.BelongsTo
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateRoleOpts{
					Name:        "test role",
					Description: "test description",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "add new role with belongs to cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) *redisBaseRepository {
					belongsTo := opts.BelongsTo
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")

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
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateRoleOpts) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateRoleOpts{
					Name:        "test role",
					Description: "test description",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new role with organization cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) *redisBaseRepository {
					belongsTo := opts.BelongsTo
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
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
					cacheRepo.EXPECT().Delete(ctx, belongsToKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateRoleOpts) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateRoleOpts{
					Name:        "test role",
					Description: "test description",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new role with project cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateRoleOpts) *redisBaseRepository {
					belongsTo := opts.BelongsTo
					belongsToKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
					organizationsKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					belongsToKeyResult := new(redis.StringSliceCmd)
					belongsToKeyResult.SetVal([]string{belongsToKey})

					organizationsKeyResult := new(redis.StringSliceCmd)
					organizationsKeyResult.SetVal([]string{organizationsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, belongsToKey).Return(belongsToKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationsKey).Return(organizationsKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, organizationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateRoleOpts) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateRoleOpts{
					Name:        "test role",
					Description: "test description",
					CreatedBy:   model.MustNewID(model.ResourceTypeUser),
					BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *Role) RoleRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *Role
		wantErr error
	}{
		{
			name: "get uncached role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
						Value: role,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *Role {
				return &Role{
					ID:          id,
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
						if rolePtr, ok := dst.(**Role); ok {
							*rolePtr = role
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *Role) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(_ model.ID) *Role {
				return &Role{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, _ *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached role error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *Role) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached role cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
						Value: role,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, role *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
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
			var want *Role
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedRoleRepository_GetAllBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) RoleRepository
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
		want    []*Role
		wantErr error
	}{
		{
			name: "get uncached roles",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: roles,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(roles, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*Role{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						if rolesPtr, ok := dst.(*[]*Role); ok {
							*rolesPtr = roles
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Role) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*Role{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get roles cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, _ []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Role) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached roles cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)

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
						Value: roles,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, offset, limit int, roles []*Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().GetAllBelongsTo(ctx, belongsTo, offset, limit).Return(roles, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllBelongsTo(tt.args.ctx, tt.args.belongsTo, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedRoleRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts, role *Role) RoleRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
		opts      UpdateRoleOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Role
		wantErr error
	}{
		{
			name: "update role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts, role *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: UpdateRoleOpts{
					Name:        optional.Some("updated role"),
					Description: optional.Some("updated description"),
				},
			},
			want: &Role{
				ID:          model.MustNewID(model.ResourceTypeRole),
				Name:        "test role",
				Description: "test description",
			},
		},
		{
			name: "update role with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *Role) *redisBaseRepository {
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
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts, _ *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: UpdateRoleOpts{
					Name:        optional.Some("updated role"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update role set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
						Value: role,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts, role *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: UpdateRoleOpts{
					Name:        optional.Some("updated role"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update role delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: role,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts, role *Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(role, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				opts: UpdateRoleOpts{
					Name:        optional.Some("updated role"),
					Description: optional.Some("updated description"),
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

			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.args.opts, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.belongsTo, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedRoleRepository_AddMember(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id, belongsToID model.ID) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, id, memberID, belongsToID model.ID) RoleRepository
	}
	type args struct {
		ctx         context.Context
		id          model.ID
		memberID    model.ID
		belongsToID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add member success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsToID model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					orgKey := composeCacheKey(model.ResourceTypeOrganization.String(), belongsToID.String())

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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, orgKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID, belongsToID model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().AddMember(ctx, id, memberID, belongsToID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "add member with role deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsToID model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					orgKey := composeCacheKey(model.ResourceTypeOrganization.String(), belongsToID.String())

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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, orgKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID, belongsToID model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().AddMember(ctx, id, memberID, belongsToID).Return(ErrRoleDelete)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrRoleDelete,
		},
		{
			name: "delete role with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _, _ model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete role cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _, _ model.ID) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsToID),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID, tt.args.belongsToID),
			}
			err := r.AddMember(tt.args.ctx, tt.args.id, tt.args.memberID, tt.args.belongsToID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_RemoveMember(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id, belongsToID model.ID) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, id, memberID, belongsToID model.ID) RoleRepository
	}
	type args struct {
		ctx         context.Context
		id          model.ID
		memberID    model.ID
		belongsToID model.ID
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsToID model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					orgKey := composeCacheKey(model.ResourceTypeOrganization.String(), belongsToID.String())

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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, orgKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID, belongsToID model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().RemoveMember(ctx, id, memberID, belongsToID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "delete role with role deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsToID model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					orgKey := composeCacheKey(model.ResourceTypeOrganization.String(), belongsToID.String())

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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, orgKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID, belongsToID model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().RemoveMember(ctx, id, memberID, belongsToID).Return(ErrRoleDelete)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrRoleDelete,
		},
		{
			name: "delete role with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _, _ model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete role cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _, _ model.ID) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:         context.Background(),
				id:          model.MustNewID(model.ResourceTypeRole),
				memberID:    model.MustNewID(model.ResourceTypeDocument),
				belongsToID: model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsToID),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID, tt.args.belongsToID),
			}
			err := r.RemoveMember(tt.args.ctx, tt.args.id, tt.args.memberID, tt.args.belongsToID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID) RoleRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectKey).Return(projectKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Delete(ctx, id, belongsTo).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "delete role with role deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectKey).Return(projectKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().Delete(ctx, id, belongsTo).Return(ErrRoleDelete)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrRoleDelete,
		},
		{
			name: "delete role with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())

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
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete role with get all cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete role with organization cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete role with project cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
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

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectKey).Return(projectKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) RoleRepository {
					return NewMockRoleRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeRole),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedRoleRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo),
			}
			err := r.Delete(tt.args.ctx, tt.args.id, tt.args.belongsTo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
