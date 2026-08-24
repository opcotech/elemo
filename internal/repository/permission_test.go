package repository_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/go-redis/cache/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
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
		opts    repository.CreateGrantOpts
		wantErr error
	}{
		{
			name: "valid user principal",
			opts: repository.CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
		},
		{
			name: "valid team principal",
			opts: repository.CreateGrantOpts{
				Principal: teamID,
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
		},
		{
			name: "valid organization principal",
			opts: repository.CreateGrantOpts{
				Principal: orgID,
				Scope:     projectID,
				Actions:   []model.Action{model.ActionProjectRead},
			},
		},
		{
			name: "role id without actions",
			opts: repository.CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				RoleID:    &roleID,
			},
		},
		{
			name: "not a principal",
			opts: repository.CreateGrantOpts{
				Principal: projectID,
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
			wantErr: model.ErrNotAPrincipal,
		},
		{
			name: "empty actions without role",
			opts: repository.CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
			},
			wantErr: model.ErrInvalidAction,
		},
		{
			name: "invalid action",
			opts: repository.CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				Actions:   []model.Action{"not-an-action"},
			},
			wantErr: model.ErrInvalidAction,
		},
		{
			name: "role id wrong type",
			opts: repository.CreateGrantOpts{
				Principal: userID,
				Scope:     orgID,
				RoleID:    &projectID,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "invalid principal",
			opts: repository.CreateGrantOpts{
				Principal: model.ID{},
				Scope:     orgID,
				Actions:   []model.Action{model.ActionOrganizationRead},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "invalid scope",
			opts: repository.CreateGrantOpts{
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

	clause := repository.AuthzVisibleExistsClause("n", "$user_id", "$action")
	assert.Contains(t, clause, "ALL(authz_node IN nodes(path)")
	assert.NotContains(t, clause, "ALL(n IN")
	assert.NotContains(t, clause, "ALL(x IN")
}

func TestCachedPermissionRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	opts := repository.CreateGrantOpts{
		Principal: model.MustNewID(model.ResourceTypeUser),
		Scope:     model.MustNewID(model.ResourceTypeOrganization),
		Actions:   []model.Action{model.ActionOrganizationRead},
	}
	grant := &repository.Grant{ID: model.MustNewID(model.ResourceTypePermission), Principal: opts.Principal, Scope: opts.Scope}

	t.Run("creates and clears authz caches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Create(ctx, opts).Return(grant, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisCacheExpectingBumpThenPatternsAndIssueAuthzEpoch(ctrl, ctx, opts.Principal, permissionCrossCachePatterns())...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.Create(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, grant, got)
	})

	t.Run("create error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Create(ctx, opts).Return(nil, repository.ErrPermissionCreate)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		_, err := r.Create(ctx, opts)
		require.ErrorIs(t, err, repository.ErrPermissionCreate)
	})
}

func TestCachedPermissionRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := model.MustNewID(model.ResourceTypePermission)
	grant := &repository.Grant{
		ID:        id,
		Principal: model.MustNewID(model.ResourceTypeUser),
		Scope:     model.MustNewID(model.ResourceTypeOrganization),
	}

	t.Run("deletes and clears authz caches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(grant, nil)
		inner.EXPECT().Delete(ctx, id).Return(nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisCacheExpectingBumpThenPatternsAndIssueAuthzEpoch(ctrl, ctx, grant.Principal, permissionCrossCachePatterns())...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.NoError(t, r.Delete(ctx, id))
	})

	t.Run("delete error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(grant, nil)
		inner.EXPECT().Delete(ctx, id).Return(repository.ErrPermissionDelete)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.ErrorIs(t, r.Delete(ctx, id), repository.ErrPermissionDelete)
	})

	t.Run("get failure still deletes and skips generation bump", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
		inner.EXPECT().Delete(ctx, id).Return(nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl, ctx, permissionCrossCachePatterns(), -1, nil, 1)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.NoError(t, r.Delete(ctx, id))
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
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().LinkInScopeOf(ctx, child, parent).Return(nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl, ctx, permissionCrossCachePatterns(), -1, nil, 1)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.NoError(t, r.LinkInScopeOf(ctx, child, parent))
	})

	t.Run("link error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().LinkInScopeOf(ctx, child, parent).Return(repository.ErrInScopeOfLink)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.ErrorIs(t, r.LinkInScopeOf(ctx, child, parent), repository.ErrInScopeOfLink)
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

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
		require.NoError(t, err)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
		cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: int64(1)}).Return(nil)

		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				nil,
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
		require.NoError(t, r.BumpGeneration(ctx, principal))
	})
}

func TestCachedPermissionRepository_Passthrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := model.MustNewID(model.ResourceTypePermission)
	actor := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	grant := &repository.Grant{ID: id, Principal: actor, Scope: resource}

	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Get(ctx, id).Return(grant, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.Get(ctx, id)
		require.NoError(t, err)
		require.Equal(t, grant, got)
	})

	t.Run("ListByPrincipal", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListByPrincipal(ctx, actor).Return([]*repository.Grant{grant}, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListByPrincipal(ctx, actor)
		require.NoError(t, err)
		require.Equal(t, []*repository.Grant{grant}, got)
	})

	t.Run("ListByScope", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListByScope(ctx, resource).Return([]*repository.Grant{grant}, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListByScope(ctx, resource)
		require.NoError(t, err)
		require.Equal(t, []*repository.Grant{grant}, got)
	})

	t.Run("Has", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Has(ctx, actor, resource, model.ActionOrganizationRead).Return(true, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.Has(ctx, actor, resource, model.ActionOrganizationRead)
		require.NoError(t, err)
		require.True(t, got)
	})

	t.Run("EffectiveActions", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().EffectiveActions(ctx, actor, resource).Return([]model.Action{model.ActionOrganizationRead}, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.EffectiveActions(ctx, actor, resource)
		require.NoError(t, err)
		require.Equal(t, []model.Action{model.ActionOrganizationRead}, got)
	})

	t.Run("Explain", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		decision := &repository.Decision{Allowed: true, Actor: actor, Resource: resource}
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().Explain(ctx, actor, resource, model.ActionOrganizationRead).Return(decision, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.Explain(ctx, actor, resource, model.ActionOrganizationRead)
		require.NoError(t, err)
		require.Equal(t, decision, got)
	})

	t.Run("ListGrantScopes", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListGrantScopes(ctx, actor, model.ActionIssueRead).Return([]model.ID{resource}, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListGrantScopes(ctx, actor, model.ActionIssueRead)
		require.NoError(t, err)
		require.Equal(t, []model.ID{resource}, got)
	})

	t.Run("ListScopeAncestry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		inner := mockrepo.NewMockPermissionRepository(ctrl)
		inner.EXPECT().ListScopeAncestry(ctx, resource).Return([]model.ID{resource}, nil)
		r := func() *repository.RedisCachedPermissionRepository {
			r, err := repository.NewCachedPermissionRepository(
				inner,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListScopeAncestry(ctx, resource)
		require.NoError(t, err)
		require.Equal(t, []model.ID{resource}, got)
	})
}
