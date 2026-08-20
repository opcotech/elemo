package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

//nolint:revive // test helpers take gomock.Controller first
func newPermissionTestBase(ctrl *gomock.Controller, ctx context.Context) (*baseService, *repository.MockPermissionRepository) {
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()

	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	repo := repository.NewMockPermissionRepository(ctrl)
	return &baseService{
		logger: mock.NewMockLogger(ctrl),
		tracer: tracer,
	}, repo
}

func TestNewPermissionService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    repository.PermissionRepository
		opts    []Option
		wantErr error
	}{
		{
			name: "new permission service",
			repo: repository.NewMockPermissionRepository(nil),
			opts: []Option{
				WithLogger(mock.NewMockLogger(nil)),
				WithTracer(mock.NewMockTracer(nil)),
			},
		},
		{
			name: "nil permission repository",
			repo: nil,
			opts: []Option{
				WithLogger(mock.NewMockLogger(nil)),
				WithTracer(mock.NewMockTracer(nil)),
			},
			wantErr: ErrNoPermissionRepository,
		},
		{
			name: "nil logger",
			repo: repository.NewMockPermissionRepository(nil),
			opts: []Option{
				WithLogger(nil),
				WithTracer(mock.NewMockTracer(nil)),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "nil tracer",
			repo: repository.NewMockPermissionRepository(nil),
			opts: []Option{
				WithLogger(mock.NewMockLogger(nil)),
				WithTracer(nil),
			},
			wantErr: tracing.ErrNoTracer,
		},
		{
			name: "missing logger uses default",
			repo: repository.NewMockPermissionRepository(nil),
			opts: []Option{
				WithTracer(mock.NewMockTracer(nil)),
			},
		},
		{
			name: "missing tracer uses noop",
			repo: repository.NewMockPermissionRepository(nil),
			opts: []Option{
				WithLogger(mock.NewMockLogger(nil)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			got, err := NewPermissionService(tt.repo, tt.opts...)
			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				is.Nil(got)
				return
			}
			must.NoError(err)
			is.NotNil(got)
		})
	}
}

func Test_permissionService_CtxUserHas(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	tests := []struct {
		name  string
		ctx   context.Context
		setup func(repo *repository.MockPermissionRepository)
		want  bool
	}{
		{
			name: "true when repo allows",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionOrganizationRead).Return(true, nil)
			},
			want: true,
		},
		{
			name: "false when repo denies",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionOrganizationRead).Return(false, nil)
			},
			want: false,
		},
		{
			name:  "false when context has no user",
			ctx:   context.Background(),
			setup: func(_ *repository.MockPermissionRepository) {},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			ctrl := gomock.NewController(t)

			base, repo := newPermissionTestBase(ctrl, tt.ctx)
			tt.setup(repo)

			s := &permissionService{
				baseService:    base,
				permissionRepo: repo,
			}
			is.Equal(tt.want, s.CtxUserHas(tt.ctx, orgID, model.ActionOrganizationRead))
		})
	}
}

func Test_permissionService_CtxUserCreate(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	principal := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	grant := testModel.NewRepositoryGrant(principal, orgID, model.ActionOrganizationRead)

	opts := CreateGrantOpts{
		Principal: principal,
		Scope:     orgID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	}

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *repository.MockPermissionRepository)
		wantErr error
	}{
		{
			name: "creates when caller has permission.manage and held actions",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(true, nil)
				repo.EXPECT().EffectiveActions(gomock.Any(), userID, orgID).Return([]model.Action{
					model.ActionPermissionManage,
					model.ActionOrganizationRead,
				}, nil)
				repo.EXPECT().Create(gomock.Any(), repository.CreateGrantOpts{
					Principal: principal,
					Scope:     orgID,
					Actions:   []model.Action{model.ActionOrganizationRead},
				}).Return(grant, nil)
				repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(nil)
			},
		},
		{
			name: "denied without permission.manage",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(false, nil)
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "denied when granting unheld action",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(true, nil)
				repo.EXPECT().EffectiveActions(gomock.Any(), userID, orgID).Return([]model.Action{
					model.ActionPermissionManage,
				}, nil)
			},
			wantErr: model.ErrPrivilegeEscalation,
		},
		{
			name:    "missing user",
			ctx:     context.Background(),
			setup:   func(_ *repository.MockPermissionRepository) {},
			wantErr: ErrNoUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)
			ctrl := gomock.NewController(t)

			base, repo := newPermissionTestBase(ctrl, tt.ctx)
			tt.setup(repo)

			s := &permissionService{
				baseService:    base,
				permissionRepo: repo,
			}
			got, err := s.CtxUserCreate(tt.ctx, opts)
			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				is.Nil(got)
				return
			}
			must.NoError(err)
			is.Equal(grant.ID, got.ID)
			is.Equal(grant.Principal, got.Principal)
			is.Equal(grant.Scope, got.Scope)
		})
	}
}

func Test_permissionService_CtxUserDelete(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	grantID := model.MustNewID(model.ResourceTypePermission)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	grant := testModel.NewRepositoryGrant(userID, orgID, model.ActionOrganizationRead)
	grant.ID = grantID

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *repository.MockPermissionRepository)
		wantErr error
	}{
		{
			name: "deletes when caller has permission.manage on scope",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Get(gomock.Any(), grantID).Return(grant, nil)
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(true, nil)
				repo.EXPECT().Delete(gomock.Any(), grantID).Return(nil)
			},
		},
		{
			name: "denied without permission.manage",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Get(gomock.Any(), grantID).Return(grant, nil)
				repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(false, nil)
			},
			wantErr: ErrNoPermission,
		},
		{
			name:    "missing user",
			ctx:     context.Background(),
			setup:   func(_ *repository.MockPermissionRepository) {},
			wantErr: ErrNoUser,
		},
		{
			name: "get not found wraps ErrPermissionDelete and ErrPermissionGet",
			ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().Get(gomock.Any(), grantID).Return(nil, repository.ErrNotFound)
			},
			wantErr: ErrPermissionDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)
			ctrl := gomock.NewController(t)

			base, repo := newPermissionTestBase(ctrl, tt.ctx)
			tt.setup(repo)

			s := &permissionService{
				baseService:    base,
				permissionRepo: repo,
			}
			err := s.CtxUserDelete(tt.ctx, grantID)
			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				if tt.name == "get not found wraps ErrPermissionDelete and ErrPermissionGet" {
					must.ErrorIs(err, ErrPermissionGet)
				}
				return
			}
			must.NoError(err)
		})
	}
}

func Test_permissionService_BootstrapCreator(t *testing.T) {
	t.Parallel()

	creator := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	actions := []model.Action{model.ActionOrganizationRead}
	grant := testModel.NewRepositoryGrant(creator, resource, actions...)

	t.Run("creates grant for creator", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)
		ctrl := gomock.NewController(t)
		ctx := context.Background()

		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Create(gomock.Any(), repository.CreateGrantOpts{
			Principal: creator,
			Scope:     resource,
			Actions:   actions,
		}).Return(grant, nil)
		repo.EXPECT().BumpGeneration(gomock.Any(), creator).Return(nil)

		s := &permissionService{baseService: base, permissionRepo: repo}
		must.NoError(s.BootstrapCreator(ctx, creator, resource, actions))
	})
}

func Test_permissionService_GrantRole(t *testing.T) {
	t.Parallel()

	principal := model.MustNewID(model.ResourceTypeUser)
	scope := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	grant := testModel.NewRepositoryGrant(principal, scope)
	grant.RoleID = &roleID

	t.Run("creates role grant", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)
		ctrl := gomock.NewController(t)
		ctx := context.Background()

		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Create(gomock.Any(), repository.CreateGrantOpts{
			Principal: principal,
			Scope:     scope,
			RoleID:    &roleID,
		}).Return(grant, nil)
		repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(nil)

		s := &permissionService{baseService: base, permissionRepo: repo}
		must.NoError(s.GrantRole(ctx, principal, scope, roleID))
	})
}

func Test_permissionService_BumpGeneration(t *testing.T) {
	t.Parallel()

	principal := model.MustNewID(model.ResourceTypeUser)

	tests := []struct {
		name    string
		setup   func(repo *repository.MockPermissionRepository)
		wantErr error
	}{
		{
			name: "bumps generation",
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(nil)
			},
		},
		{
			name: "wraps repository error",
			setup: func(repo *repository.MockPermissionRepository) {
				repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(assert.AnError)
			},
			wantErr: ErrPermissionUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			base, repo := newPermissionTestBase(ctrl, ctx)
			tt.setup(repo)

			s := &permissionService{baseService: base, permissionRepo: repo}
			err := s.BumpGeneration(ctx, principal)
			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				return
			}
			must.NoError(err)
		})
	}
}

func Test_permissionService_Has(t *testing.T) {
	t.Parallel()
	actor := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Has(gomock.Any(), actor, resource, model.ActionOrganizationRead).Return(true, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.Has(ctx, actor, resource, model.ActionOrganizationRead)
		require.NoError(t, err)
		require.True(t, got)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Has(gomock.Any(), actor, resource, model.ActionOrganizationRead).Return(false, repository.ErrPermissionRead)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.Has(ctx, actor, resource, model.ActionOrganizationRead)
		require.ErrorIs(t, err, ErrPermissionHasPermission)
		require.ErrorIs(t, err, repository.ErrPermissionRead)
	})
}

func Test_permissionService_EffectiveActions(t *testing.T) {
	t.Parallel()
	actor := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.Background()
	actions := []model.Action{model.ActionOrganizationRead}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().EffectiveActions(gomock.Any(), actor, resource).Return(actions, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.EffectiveActions(ctx, actor, resource)
		require.NoError(t, err)
		require.Equal(t, actions, got)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().EffectiveActions(gomock.Any(), actor, resource).Return(nil, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.EffectiveActions(ctx, actor, resource)
		require.ErrorIs(t, err, ErrPermissionGet)
	})
}

func Test_permissionService_CtxUserEffectiveActions(t *testing.T) {
	t.Parallel()
	userID := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	actions := []model.Action{model.ActionOrganizationRead}

	t.Run("delegates for context user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().EffectiveActions(gomock.Any(), userID, resource).Return(actions, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.CtxUserEffectiveActions(ctx, resource)
		require.NoError(t, err)
		require.Equal(t, actions, got)
	})

	t.Run("missing user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ctx := context.Background()
		base, repo := newPermissionTestBase(ctrl, ctx)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.CtxUserEffectiveActions(ctx, resource)
		require.ErrorIs(t, err, ErrNoUser)
	})
}

func Test_permissionService_Explain(t *testing.T) {
	t.Parallel()
	actor := model.MustNewID(model.ResourceTypeUser)
	resource := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.Background()
	decision := &repository.Decision{Allowed: true, Actor: actor, Resource: resource}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Explain(gomock.Any(), actor, resource, model.ActionOrganizationRead).Return(decision, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.Explain(ctx, actor, resource, model.ActionOrganizationRead)
		require.NoError(t, err)
		require.Equal(t, decision, got)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Explain(gomock.Any(), actor, resource, model.ActionOrganizationRead).Return(nil, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.Explain(ctx, actor, resource, model.ActionOrganizationRead)
		require.ErrorIs(t, err, ErrPermissionHasPermission)
	})
}

func Test_permissionService_Create(t *testing.T) {
	t.Parallel()
	principal := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	grant := testModel.NewRepositoryGrant(principal, orgID, model.ActionOrganizationRead)
	opts := CreateGrantOpts{Principal: principal, Scope: orgID, Actions: []model.Action{model.ActionOrganizationRead}}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Create(gomock.Any(), repository.CreateGrantOpts{
			Principal: principal, Scope: orgID, Actions: opts.Actions,
		}).Return(grant, nil)
		repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.Create(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, grant.ID, got.ID)
	})

	t.Run("validate fail", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.Create(ctx, CreateGrantOpts{})
		require.ErrorIs(t, err, model.ErrInvalidGrant)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.Create(ctx, opts)
		require.ErrorIs(t, err, ErrPermissionCreate)
	})
}

func Test_permissionService_Get(t *testing.T) {
	t.Parallel()
	id := model.MustNewID(model.ResourceTypePermission)
	grant := testModel.NewRepositoryGrant(model.MustNewID(model.ResourceTypeUser), model.MustNewID(model.ResourceTypeOrganization), model.ActionOrganizationRead)
	grant.ID = id
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Get(gomock.Any(), id).Return(grant, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.Get(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, got.ID)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Get(gomock.Any(), id).Return(nil, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.Get(ctx, id)
		require.ErrorIs(t, err, ErrPermissionGet)
	})
}

func Test_permissionService_ListByPrincipal(t *testing.T) {
	t.Parallel()
	principal := model.MustNewID(model.ResourceTypeUser)
	grant := testModel.NewRepositoryGrant(principal, model.MustNewID(model.ResourceTypeOrganization), model.ActionOrganizationRead)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().ListByPrincipal(gomock.Any(), principal).Return([]*repository.Grant{grant}, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.ListByPrincipal(ctx, principal)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().ListByPrincipal(gomock.Any(), principal).Return(nil, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.ListByPrincipal(ctx, principal)
		require.ErrorIs(t, err, ErrPermissionGetBySubject)
	})
}

func Test_permissionService_ListByScope(t *testing.T) {
	t.Parallel()
	scope := model.MustNewID(model.ResourceTypeOrganization)
	grant := testModel.NewRepositoryGrant(model.MustNewID(model.ResourceTypeUser), scope, model.ActionOrganizationRead)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().ListByScope(gomock.Any(), scope).Return([]*repository.Grant{grant}, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		got, err := s.ListByScope(ctx, scope)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().ListByScope(gomock.Any(), scope).Return(nil, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.ListByScope(ctx, scope)
		require.ErrorIs(t, err, ErrPermissionGetByTarget)
	})
}

func Test_permissionService_Delete(t *testing.T) {
	t.Parallel()
	id := model.MustNewID(model.ResourceTypePermission)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Delete(gomock.Any(), id).Return(nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		require.NoError(t, s.Delete(ctx, id))
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Delete(gomock.Any(), id).Return(assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		require.ErrorIs(t, s.Delete(ctx, id), ErrPermissionDelete)
	})
}

func Test_permissionService_LinkInScopeOf(t *testing.T) {
	t.Parallel()
	child := model.MustNewID(model.ResourceTypeProject)
	parent := model.MustNewID(model.ResourceTypeNamespace)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().LinkInScopeOf(gomock.Any(), child, parent).Return(nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		require.NoError(t, s.LinkInScopeOf(ctx, child, parent))
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().LinkInScopeOf(gomock.Any(), child, parent).Return(assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		require.ErrorIs(t, s.LinkInScopeOf(ctx, child, parent), ErrPermissionCreate)
	})
}

func Test_permissionService_CtxUserCreate_RoleID(t *testing.T) {
	t.Parallel()
	userID := model.MustNewID(model.ResourceTypeUser)
	principal := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	grant := testModel.NewRepositoryGrant(principal, orgID)
	grant.RoleID = &roleID
	opts := CreateGrantOpts{Principal: principal, Scope: orgID, RoleID: &roleID}
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("resolves role actions", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		roleRepo := repository.NewMockRoleRepository(ctrl)
		repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(true, nil)
		repo.EXPECT().EffectiveActions(gomock.Any(), userID, orgID).Return([]model.Action{
			model.ActionPermissionManage, model.ActionOrganizationRead,
		}, nil)
		roleRepo.EXPECT().GetByID(gomock.Any(), roleID).Return(&repository.Role{
			ID: roleID, Actions: []string{model.ActionOrganizationRead.String()},
		}, nil)
		repo.EXPECT().Create(gomock.Any(), repository.CreateGrantOpts{
			Principal: principal, Scope: orgID, RoleID: &roleID,
		}).Return(grant, nil)
		repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		s.roleRepo = roleRepo
		got, err := s.CtxUserCreate(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, grant.ID, got.ID)
	})

	t.Run("nil role repository", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionPermissionManage).Return(true, nil)
		repo.EXPECT().EffectiveActions(gomock.Any(), userID, orgID).Return([]model.Action{model.ActionPermissionManage}, nil)
		s := &permissionService{baseService: base, permissionRepo: repo}
		_, err := s.CtxUserCreate(ctx, opts)
		require.ErrorIs(t, err, ErrNoRoleRepository)
	})
}

func Test_permissionService_CtxUserHas_RepoErrors(t *testing.T) {
	t.Parallel()
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("permission read error returns allowed value", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionOrganizationRead).Return(false, repository.ErrPermissionRead)
		s := &permissionService{baseService: base, permissionRepo: repo}
		require.False(t, s.CtxUserHas(ctx, orgID, model.ActionOrganizationRead))
	})

	t.Run("other errors deny", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, repo := newPermissionTestBase(ctrl, ctx)
		repo.EXPECT().Has(gomock.Any(), userID, orgID, model.ActionOrganizationRead).Return(false, assert.AnError)
		s := &permissionService{baseService: base, permissionRepo: repo}
		require.False(t, s.CtxUserHas(ctx, orgID, model.ActionOrganizationRead))
	})
}
