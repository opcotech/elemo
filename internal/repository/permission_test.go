package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCreateGrantOpts_Validate(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	projectID := model.MustNewID(model.ResourceTypeProject)
	roleID := model.MustNewID(model.ResourceTypeRole)

	tests := []struct {
		name    string
		opts    CreateGrantOpts
		wantErr error
	}{
		{
			name: "valid user principal",
			opts: CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
		},
		{
			name: "valid team principal",
			opts: CreateGrantOpts{
				Principal: teamID,
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
		},
		{
			name: "valid organization principal",
			opts: CreateGrantOpts{
				Principal: orgID,
				Scope:     projectID,
				Actions:   []model.Action{model.ActionProjectRead},
			},
		},
		{
			name: "role id without actions",
			opts: CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				RoleID:    &roleID,
			},
		},
		{
			name: "not a principal",
			opts: CreateGrantOpts{
				Principal: projectID,
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
			wantErr: model.ErrNotAPrincipal,
		},
		{
			name: "empty actions without role",
			opts: CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
			},
			wantErr: model.ErrInvalidAction,
		},
		{
			name: "invalid action",
			opts: CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				Actions:   []model.Action{"not-an-action"},
			},
			wantErr: model.ErrInvalidAction,
		},
		{
			name: "role id wrong type",
			opts: CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				RoleID:    &projectID,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "invalid principal",
			opts: CreateGrantOpts{
				Principal: model.ID{},
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "invalid scope",
			opts: CreateGrantOpts{
				Principal: userID,
				Scope:     model.ID{},
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
			wantErr: model.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			err := tt.opts.Validate()
			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				is.ErrorIs(err, model.ErrInvalidGrant)
				return
			}
			must.NoError(err)
		})
	}
}

func TestAuthzVisibleExistsClause(t *testing.T) {
	t.Parallel()

	clause := AuthzVisibleExistsClause("n", "$user_id", "$action")
	assert.Contains(t, clause, "ALL(authz_node IN nodes(path)")
	assert.NotContains(t, clause, "ALL(n IN")
	assert.NotContains(t, clause, "ALL(x IN")
}

func TestCachedPermissionRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	opts := CreateGrantOpts{
		Principal: model.MustNewID(model.ResourceTypeUser),
		Scope:     model.MustNewID(model.ResourceTypeOrganization),
		Actions:   []model.Action{model.ActionOrganizationRead},
	}
	grant := &Grant{ID: model.MustNewID(model.ResourceTypePermission), Principal: opts.Principal, Scope: opts.Scope}

	t.Run("creates and clears authz caches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Create(ctx, opts).Return(grant, nil)
		r := &RedisCachedPermissionRepository{
			cacheRepo:      redisCacheExpectingBumpThenPatterns(ctrl, ctx, opts.Principal, permissionCrossCachePatterns()),
			permissionRepo: inner,
		}
		got, err := r.Create(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, grant, got)
	})

	t.Run("create error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Create(ctx, opts).Return(nil, ErrPermissionCreate)
		r := &RedisCachedPermissionRepository{
			cacheRepo:      &redisBaseRepository{},
			permissionRepo: inner,
		}
		_, err := r.Create(ctx, opts)
		require.ErrorIs(t, err, ErrPermissionCreate)
	})
}

func TestCachedPermissionRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := model.MustNewID(model.ResourceTypePermission)
	grant := &Grant{
		ID:        id,
		Principal: model.MustNewID(model.ResourceTypeUser),
		Scope:     model.MustNewID(model.ResourceTypeOrganization),
	}

	t.Run("deletes and clears authz caches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(grant, nil)
		inner.EXPECT().Delete(ctx, id).Return(nil)
		r := &RedisCachedPermissionRepository{
			cacheRepo:      redisCacheExpectingBumpThenPatterns(ctrl, ctx, grant.Principal, permissionCrossCachePatterns()),
			permissionRepo: inner,
		}
		require.NoError(t, r.Delete(ctx, id))
	})

	t.Run("delete error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(grant, nil)
		inner.EXPECT().Delete(ctx, id).Return(ErrPermissionDelete)
		r := &RedisCachedPermissionRepository{
			cacheRepo:      &redisBaseRepository{},
			permissionRepo: inner,
		}
		require.ErrorIs(t, r.Delete(ctx, id), ErrPermissionDelete)
	})
}

func TestCachedPermissionRepository_LinkInScopeOf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	child := model.MustNewID(model.ResourceTypeProject)
	parent := model.MustNewID(model.ResourceTypeNamespace)

	t.Run("links and clears authz caches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().LinkInScopeOf(ctx, child, parent).Return(nil)
		r := &RedisCachedPermissionRepository{
			cacheRepo:      redisCacheExpectingPatterns(ctrl, ctx, permissionCrossCachePatterns(), -1, nil),
			permissionRepo: inner,
		}
		require.NoError(t, r.LinkInScopeOf(ctx, child, parent))
	})

	t.Run("link error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().LinkInScopeOf(ctx, child, parent).Return(ErrInScopeOfLink)
		r := &RedisCachedPermissionRepository{
			cacheRepo:      &redisBaseRepository{},
			permissionRepo: inner,
		}
		require.ErrorIs(t, r.LinkInScopeOf(ctx, child, parent), ErrInScopeOfLink)
	})
}

func TestCachedPermissionRepository_BumpGeneration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	principal := model.MustNewID(model.ResourceTypeUser)
	key := authzGenKey(principal)

	t.Run("increments generation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
		cacheRepo := mock.NewCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: int64(1)}).Return(nil)

		r := &RedisCachedPermissionRepository{
			cacheRepo: &redisBaseRepository{db: db, cache: cacheRepo, tracer: tracer, logger: mock.NewMockLogger(ctrl)},
		}
		require.NoError(t, r.BumpGeneration(ctx, principal))
	})
}

func TestCachedPermissionRepository_Passthrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := model.MustNewID(model.ResourceTypePermission)
	actor := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	grant := &Grant{ID: id, Principal: actor, Scope: resource}

	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(grant, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.Get(ctx, id)
		require.NoError(t, err)
		require.Equal(t, grant, got)
	})

	t.Run("ListByPrincipal", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListByPrincipal(ctx, actor).Return([]*Grant{grant}, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.ListByPrincipal(ctx, actor)
		require.NoError(t, err)
		require.Equal(t, []*Grant{grant}, got)
	})

	t.Run("ListByScope", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListByScope(ctx, resource).Return([]*Grant{grant}, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.ListByScope(ctx, resource)
		require.NoError(t, err)
		require.Equal(t, []*Grant{grant}, got)
	})

	t.Run("Has", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Has(ctx, actor, resource, model.ActionOrganizationRead).Return(true, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.Has(ctx, actor, resource, model.ActionOrganizationRead)
		require.NoError(t, err)
		require.True(t, got)
	})

	t.Run("EffectiveActions", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().EffectiveActions(ctx, actor, resource).Return([]model.Action{model.ActionOrganizationRead}, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.EffectiveActions(ctx, actor, resource)
		require.NoError(t, err)
		require.Equal(t, []model.Action{model.ActionOrganizationRead}, got)
	})

	t.Run("Explain", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		decision := &Decision{Allowed: true, Actor: actor, Resource: resource}
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().Explain(ctx, actor, resource, model.ActionOrganizationRead).Return(decision, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.Explain(ctx, actor, resource, model.ActionOrganizationRead)
		require.NoError(t, err)
		require.Equal(t, decision, got)
	})

	t.Run("ListVisible", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		parent := model.InstallationID()
		inner := NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListVisible(ctx, actor, model.ActionOrganizationRead, parent, model.ResourceTypeOrganization).Return([]model.ID{resource}, nil)
		r := &RedisCachedPermissionRepository{permissionRepo: inner}
		got, err := r.ListVisible(ctx, actor, model.ActionOrganizationRead, parent, model.ResourceTypeOrganization)
		require.NoError(t, err)
		require.Equal(t, []model.ID{resource}, got)
	})
}
