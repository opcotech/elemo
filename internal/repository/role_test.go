package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedRoleRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "add new role", failIndex: -1},
		{name: "add new role with error", failIndex: -1, repoErr: ErrNotFound, wantErr: ErrNotFound},
		{name: "add new role with belongs to cache error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add new role with get by key cache error", failIndex: 1, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add new role with organization cache error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add new role with project cache error", failIndex: 3, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			opts := CreateRoleOpts{
				Name:        "test role",
				Description: "test description",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
			}
			repo := NewMockRoleRepository(ctrl)
			if tt.failIndex < 0 {
				if tt.repoErr != nil {
					repo.EXPECT().Create(ctx, opts).Return(nil, tt.repoErr)
				} else {
					repo.EXPECT().Create(ctx, opts).Return(&Role{}, nil)
				}
			}

			r := &RedisCachedRoleRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, roleCreateCachePatterns(opts.BelongsTo), tt.failIndex, tt.failErr),
				roleRepo:  repo,
			}
			_, err := r.Create(ctx, opts)
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
					key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))

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
					repo.EXPECT().Get(ctx, id, belongsTo, RoleDetailProjection()).Return(role, nil)
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
					MemberCount: convert.ToPointer(int64(0)),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, role *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))

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
					MemberCount: convert.ToPointer(int64(0)),
					Permissions: make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached role error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))

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
					repo.EXPECT().Get(ctx, id, belongsTo, RoleDetailProjection()).Return(nil, ErrNotFound)
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
					key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))

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
					key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))

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
					repo.EXPECT().Get(ctx, id, belongsTo, RoleDetailProjection()).Return(role, nil)
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
			got, err := r.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo, RoleDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedRoleRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		hit     bool
		getErr  error
		setErr  error
		repoErr error
		wantErr error
	}{
		{name: "get uncached role"},
		{name: "get cached role", hit: true},
		{name: "get uncached role error", repoErr: ErrNotFound, wantErr: ErrNotFound},
		{name: "get cached role error", getErr: assert.AnError, wantErr: ErrCacheRead},
		{name: "get uncached role cache set error", setErr: assert.AnError, wantErr: ErrCacheWrite},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeRole)
			role := &Role{ID: id, Name: "test role", Key: "org-admin"}
			key := composeCacheKey(model.ResourceTypeRole.String(), "GetByID", id.String())

			db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
			require.NoError(t, err)
			span := mock.NewMockSpan(ctrl)
			tracer := mock.NewMockTracer(ctrl)
			cacheRepo := mock.NewCacheBackend(ctrl)
			repo := NewMockRoleRepository(ctrl)

			switch {
			case tt.hit:
				span.EXPECT().End(gomock.Len(0)).Times(1)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
				cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
					*(dst.(**Role)) = role
				}).Return(nil)
			case tt.getErr != nil:
				span.EXPECT().End(gomock.Len(0)).Times(1)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
				cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(tt.getErr)
			default:
				spanCount := 1
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
				cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
				if tt.repoErr != nil {
					repo.EXPECT().GetByID(ctx, id).Return(nil, tt.repoErr)
				} else {
					repo.EXPECT().GetByID(ctx, id).Return(role, nil)
					spanCount++
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					set := cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: role})
					if tt.setErr != nil {
						set.Return(tt.setErr)
					} else {
						set.Return(nil)
					}
				}
				span.EXPECT().End(gomock.Len(0)).Times(spanCount)
			}

			r := &RedisCachedRoleRepository{
				cacheRepo: &redisBaseRepository{db: db, cache: cacheRepo, tracer: tracer, logger: mock.NewMockLogger(ctrl)},
				roleRepo:  repo,
			}
			got, err := r.GetByID(ctx, id)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			require.Equal(t, role, got)
		})
	}
}

func TestCachedRoleRepository_GetByKey(t *testing.T) {
	tests := []struct {
		name    string
		hit     bool
		getErr  error
		setErr  error
		repoErr error
		wantErr error
	}{
		{name: "get uncached role"},
		{name: "get cached role", hit: true},
		{name: "get uncached role error", repoErr: ErrNotFound, wantErr: ErrNotFound},
		{name: "get cached role error", getErr: assert.AnError, wantErr: ErrCacheRead},
		{name: "get uncached role cache set error", setErr: assert.AnError, wantErr: ErrCacheWrite},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			belongsTo := model.MustNewID(model.ResourceTypeOrganization)
			id := model.MustNewID(model.ResourceTypeRole)
			roleKey := "org-admin"
			role := &Role{ID: id, Name: "test role", Key: roleKey}
			cacheKey := composeCacheKey(model.ResourceTypeRole.String(), "GetByKey", belongsTo.String(), roleKey)

			db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
			require.NoError(t, err)
			span := mock.NewMockSpan(ctrl)
			tracer := mock.NewMockTracer(ctrl)
			cacheRepo := mock.NewCacheBackend(ctrl)
			repo := NewMockRoleRepository(ctrl)

			switch {
			case tt.hit:
				span.EXPECT().End(gomock.Len(0)).Times(1)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
				cacheRepo.EXPECT().Get(ctx, cacheKey, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
					*(dst.(**Role)) = role
				}).Return(nil)
			case tt.getErr != nil:
				span.EXPECT().End(gomock.Len(0)).Times(1)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
				cacheRepo.EXPECT().Get(ctx, cacheKey, gomock.Any()).Return(tt.getErr)
			default:
				spanCount := 1
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
				cacheRepo.EXPECT().Get(ctx, cacheKey, gomock.Any()).Return(cache.ErrCacheMiss)
				if tt.repoErr != nil {
					repo.EXPECT().GetByKey(ctx, belongsTo, roleKey).Return(nil, tt.repoErr)
				} else {
					repo.EXPECT().GetByKey(ctx, belongsTo, roleKey).Return(role, nil)
					spanCount++
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					set := cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: cacheKey, Value: role})
					if tt.setErr != nil {
						set.Return(tt.setErr)
					} else {
						set.Return(nil)
					}
				}
				span.EXPECT().End(gomock.Len(0)).Times(spanCount)
			}

			r := &RedisCachedRoleRepository{
				cacheRepo: &redisBaseRepository{db: db, cache: cacheRepo, tracer: tracer, logger: mock.NewMockLogger(ctrl)},
				roleRepo:  repo,
			}
			got, err := r.GetByKey(ctx, belongsTo, roleKey)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			require.Equal(t, role, got)
		})
	}
}

func TestCachedRoleRepository_ListBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) *redisBaseRepository
		roleRepo  func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) RoleRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(RoleListProjection()), "", limit)

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
						Value: Page[*Role]{Items: roles},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, RoleListProjection()).Return(Page[*Role]{Items: roles}, nil)
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
					MemberCount: convert.ToPointer(int64(0)),
					Permissions: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					MemberCount: convert.ToPointer(int64(0)),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached roles",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(RoleListProjection()), "", limit)

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
						if ptr, ok := dst.(*Page[*Role]); ok {
							*ptr = Page[*Role]{Items: roles}
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
					MemberCount: convert.ToPointer(int64(0)),
					Permissions: make([]model.ID, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeRole),
					Name:        "test role",
					Description: "test description",
					MemberCount: convert.ToPointer(int64(0)),
					Permissions: make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached roles error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(RoleListProjection()), "", limit)

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
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, RoleListProjection()).Return(Page[*Role]{}, ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, _ []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(RoleListProjection()), "", limit)

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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(RoleListProjection()), "", limit)

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
						Value: Page[*Role]{Items: roles},
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				roleRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, _, limit int, roles []*Role) RoleRepository {
					repo := NewMockRoleRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, RoleListProjection()).Return(Page[*Role]{Items: roles}, nil)
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
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want),
				roleRepo:  tt.fields.roleRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.offset, testPageSize(tt.args.limit), tt.want),
			}
			got, err := r.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, CursorPage{Size: testPageSize(tt.args.limit)}, RoleListProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedRoleRepository_Update(t *testing.T) {
	newRole := func(id model.ID) *Role {
		return &Role{ID: id, Name: "test role", Description: "test description", Key: "org-admin"}
	}
	opts := UpdateRoleOpts{
		Name:        optional.Some("updated role"),
		Description: optional.Some("updated description"),
	}

	t.Run("update role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeRole)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		role := newRole(id)
		setKey := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))
		repo := NewMockRoleRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(role, nil)
		r := &RedisCachedRoleRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, role, roleUpdateInvalidatePatterns(id, belongsTo), false, -1, nil),
			roleRepo:  repo,
		}
		got, err := r.Update(ctx, id, belongsTo, opts)
		require.NoError(t, err)
		require.Equal(t, role, got)
	})

	t.Run("update role with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeRole)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)
		repo := NewMockRoleRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(nil, ErrNotFound)
		r := &RedisCachedRoleRepository{
			cacheRepo: &redisBaseRepository{db: db, cache: mock.NewCacheBackend(ctrl), tracer: mock.NewMockTracer(ctrl), logger: mock.NewMockLogger(ctrl)},
			roleRepo:  repo,
		}
		_, err = r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("update role set cache error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeRole)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		role := newRole(id)
		setKey := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))
		repo := NewMockRoleRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(role, nil)
		r := &RedisCachedRoleRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, role, nil, true, -1, assert.AnError),
			roleRepo:  repo,
		}
		_, err := r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, ErrCacheWrite)
	})

	t.Run("update role delete get all cache error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeRole)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		role := newRole(id)
		setKey := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))
		repo := NewMockRoleRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(role, nil)
		r := &RedisCachedRoleRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, role, roleUpdateInvalidatePatterns(id, belongsTo), false, 2, assert.AnError),
			roleRepo:  repo,
		}
		_, err := r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, ErrCacheDelete)
	})

	t.Run("update role actions clears authz list caches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeRole)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		role := newRole(id)
		actionOpts := UpdateRoleOpts{Actions: optional.Some([]string{model.ActionOrganizationRead.String()})}
		setKey := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))
		patterns := append(roleUpdateInvalidatePatterns(id, belongsTo), permissionCrossCachePatterns()...)
		repo := NewMockRoleRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, actionOpts).Return(role, nil)
		r := &RedisCachedRoleRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, role, patterns, false, -1, nil),
			roleRepo:  repo,
		}
		got, err := r.Update(ctx, id, belongsTo, actionOpts)
		require.NoError(t, err)
		require.Equal(t, role, got)
	})
}

func TestCachedRoleRepository_AddMember(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "add member success", failIndex: -1},
		{name: "add member with role deletion error", failIndex: -1, repoErr: ErrRoleDelete, wantErr: ErrRoleDelete},
		{name: "add member with cache deletion error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add member with related cache deletion error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeRole)
			memberID := model.MustNewID(model.ResourceTypeUser)
			belongsToID := model.MustNewID(model.ResourceTypeOrganization)
			repo := NewMockRoleRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().AddMember(ctx, id, memberID, belongsToID).Return(tt.repoErr)
			}
			r := &RedisCachedRoleRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, roleMemberCachePatterns(id, belongsToID), tt.failIndex, tt.failErr),
				roleRepo:  repo,
			}
			err := r.AddMember(ctx, id, memberID, belongsToID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_RemoveMember(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "remove member success", failIndex: -1},
		{name: "remove member with role deletion error", failIndex: -1, repoErr: ErrRoleDelete, wantErr: ErrRoleDelete},
		{name: "remove member with cache deletion error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "remove member with related cache deletion error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeRole)
			memberID := model.MustNewID(model.ResourceTypeUser)
			belongsToID := model.MustNewID(model.ResourceTypeOrganization)
			repo := NewMockRoleRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().RemoveMember(ctx, id, memberID, belongsToID).Return(tt.repoErr)
			}
			r := &RedisCachedRoleRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, roleMemberCachePatterns(id, belongsToID), tt.failIndex, tt.failErr),
				roleRepo:  repo,
			}
			err := r.RemoveMember(ctx, id, memberID, belongsToID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedRoleRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "delete role success", failIndex: -1},
		{name: "delete role with role deletion error", failIndex: -1, repoErr: ErrRoleDelete, wantErr: ErrRoleDelete},
		{name: "delete role with cache deletion error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "delete role with get all cache deletion error", failIndex: 3, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "delete role with organization cache deletion error", failIndex: 4, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "delete role with project cache deletion error", failIndex: 5, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeRole)
			belongsTo := model.MustNewID(model.ResourceTypeOrganization)
			repo := NewMockRoleRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().Delete(ctx, id, belongsTo).Return(tt.repoErr)
			}
			r := &RedisCachedRoleRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, roleDeleteCachePatterns(id, belongsTo), tt.failIndex, tt.failErr),
				roleRepo:  repo,
			}
			err := r.Delete(ctx, id, belongsTo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
