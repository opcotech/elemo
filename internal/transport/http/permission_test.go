package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newTestPermissionController(t *testing.T, ps service.PermissionService) PermissionController {
	t.Helper()
	c, err := NewPermissionController(WithPermissionService(ps))
	require.NoError(t, err)
	return c
}

func newServiceGrant() *service.Grant {
	roleID := model.MustNewID(model.ResourceTypeRole)
	return &service.Grant{
		ID:        model.MustNewID(model.ResourceTypePermission),
		Principal: model.MustNewID(model.ResourceTypeUser),
		Scope:     model.MustNewID(model.ResourceTypeOrganization),
		RoleID:    &roleID,
		Actions:   []model.Action{model.ActionOrganizationRead},
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

func createGrantRequestBody(principal, scope model.ID, actions []api.Action, roleID *string) *api.V1PermissionsCreateJSONRequestBody {
	body := &api.V1PermissionsCreateJSONRequestBody{
		Actions: &actions,
		RoleId:  roleID,
	}
	body.Principal.Id = principal.String()
	body.Principal.ResourceType = api.GrantPrincipalType(principal.Type.String())
	body.Scope.Id = scope.String()
	body.Scope.ResourceType = api.ResourceType(scope.Type.String())
	return body
}

func TestNewPermissionController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		c, err := NewPermissionController(WithPermissionService(service.NewMockPermissionService(ctrl)))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing permission service", func(t *testing.T) {
		t.Parallel()
		_, err := NewPermissionController()
		assert.ErrorIs(t, err, ErrNoPermissionService)
	})

	t.Run("nil permission service option", func(t *testing.T) {
		t.Parallel()
		_, err := NewPermissionController(WithPermissionService(nil))
		assert.ErrorIs(t, err, ErrNoPermissionService)
	})
}

func TestPermissionController_V1PermissionsCreate(t *testing.T) {
	t.Parallel()

	principal := model.MustNewID(model.ResourceTypeUser)
	scope := model.MustNewID(model.ResourceTypeOrganization)
	grant := newServiceGrant()
	body := createGrantRequestBody(principal, scope, []api.Action{api.Action(model.ActionOrganizationRead.String())}, nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserCreate(gomock.Any(), service.CreateGrantOpts{
			Principal: principal,
			Scope:     scope,
			Actions:   []model.Action{model.ActionOrganizationRead},
		}).Return(grant, nil)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionsCreate(context.Background(), api.V1PermissionsCreateRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1PermissionsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, grant.ID.String(), got.Id)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		c := newTestPermissionController(t, service.NewMockPermissionService(ctrl))
		resp, err := c.V1PermissionsCreate(context.Background(), api.V1PermissionsCreateRequestObject{})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserCreate(gomock.Any(), gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionsCreate(context.Background(), api.V1PermissionsCreateRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("privilege escalation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserCreate(gomock.Any(), gomock.Any()).Return(nil, model.ErrPrivilegeEscalation)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionsCreate(context.Background(), api.V1PermissionsCreateRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserCreate(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionsCreate(context.Background(), api.V1PermissionsCreateRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionsCreate500JSONResponse)
		assert.True(t, ok)
	})
}

func TestPermissionController_V1PermissionGet(t *testing.T) {
	t.Parallel()

	grant := newServiceGrant()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().Get(gomock.Any(), grant.ID).Return(grant, nil)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionGet(context.Background(), api.V1PermissionGetRequestObject{Id: grant.ID.String()})
		require.NoError(t, err)
		got, ok := resp.(api.V1PermissionGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, grant.ID.String(), got.Id)
		require.NotNil(t, got.RoleId)
		assert.Equal(t, grant.RoleID.String(), *got.RoleId)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		c := newTestPermissionController(t, service.NewMockPermissionService(ctrl))
		resp, err := c.V1PermissionGet(context.Background(), api.V1PermissionGetRequestObject{Id: "not-a-xid"})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().Get(gomock.Any(), grant.ID).Return(nil, errors.Join(service.ErrPermissionGet, repository.ErrNotFound))

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionGet(context.Background(), api.V1PermissionGetRequestObject{Id: grant.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().Get(gomock.Any(), grant.ID).Return(nil, errors.New("boom"))

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionGet(context.Background(), api.V1PermissionGetRequestObject{Id: grant.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionGet500JSONResponse)
		assert.True(t, ok)
	})
}

func TestPermissionController_V1PermissionDelete(t *testing.T) {
	t.Parallel()

	grantID := model.MustNewID(model.ResourceTypePermission)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserDelete(gomock.Any(), grantID).Return(nil)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionDelete(context.Background(), api.V1PermissionDeleteRequestObject{Id: grantID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionDelete204Response)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserDelete(gomock.Any(), grantID).Return(service.ErrNoPermission)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionDelete(context.Background(), api.V1PermissionDeleteRequestObject{Id: grantID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionDelete403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserDelete(gomock.Any(), grantID).Return(errors.Join(service.ErrPermissionDelete, repository.ErrNotFound))

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionDelete(context.Background(), api.V1PermissionDeleteRequestObject{Id: grantID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionDelete404JSONResponse)
		assert.True(t, ok)
	})
}

func TestPermissionController_V1PermissionResourceGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	resourceID := orgID.Composite()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserEffectiveActions(gomock.Any(), orgID).Return([]model.Action{model.ActionOrganizationRead}, nil)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionResourceGet(context.Background(), api.V1PermissionResourceGetRequestObject{ResourceId: resourceID})
		require.NoError(t, err)
		got, ok := resp.(api.V1PermissionResourceGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, []api.Action{api.Action(model.ActionOrganizationRead.String())}, got.Actions)
	})

	t.Run("malformed resource id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		c := newTestPermissionController(t, service.NewMockPermissionService(ctrl))
		resp, err := c.V1PermissionResourceGet(context.Background(), api.V1PermissionResourceGetRequestObject{ResourceId: "not-a-resource"})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionResourceGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		ps := service.NewMockPermissionService(ctrl)
		ps.EXPECT().CtxUserEffectiveActions(gomock.Any(), orgID).Return(nil, service.ErrNoPermission)

		c := newTestPermissionController(t, ps)
		resp, err := c.V1PermissionResourceGet(context.Background(), api.V1PermissionResourceGetRequestObject{ResourceId: resourceID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PermissionResourceGet403JSONResponse)
		assert.True(t, ok)
	})
}

func TestCreateGrantJSONRequestBodyToCreateGrantOpts(t *testing.T) {
	t.Parallel()

	principal := model.MustNewID(model.ResourceTypeUser)
	scope := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	roleIDStr := roleID.String()

	t.Run("actions and role", func(t *testing.T) {
		t.Parallel()
		opts, err := createGrantJSONRequestBodyToCreateGrantOpts(createGrantRequestBody(
			principal,
			scope,
			[]api.Action{api.Action(model.ActionOrganizationRead.String())},
			&roleIDStr,
		))
		require.NoError(t, err)
		assert.Equal(t, principal, opts.Principal)
		assert.Equal(t, scope, opts.Scope)
		assert.Equal(t, []model.Action{model.ActionOrganizationRead}, opts.Actions)
		require.NotNil(t, opts.RoleID)
		assert.Equal(t, roleID, *opts.RoleID)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		_, err := createGrantJSONRequestBodyToCreateGrantOpts(nil)
		assert.ErrorIs(t, err, model.ErrInvalidGrant)
	})
}

func TestGrantToDTO(t *testing.T) {
	t.Parallel()

	grant := newServiceGrant()
	dto := grantToDTO(grant)
	assert.Equal(t, grant.ID.String(), dto.Id)
	assert.Equal(t, grant.Principal.String(), dto.Principal)
	assert.Equal(t, api.ResourceType(grant.Principal.Type.String()), dto.PrincipalType)
	require.NotNil(t, dto.RoleId)
	assert.Equal(t, grant.RoleID.String(), *dto.RoleId)
	assert.Equal(t, []api.Action{api.Action(model.ActionOrganizationRead.String())}, dto.Actions)
}

func TestActionStringsOrEmpty(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []api.Action{}, actionStringsOrEmpty(nil))
	assert.Equal(t, []api.Action{api.Action(model.ActionOrganizationRead.String())}, actionStringsOrEmpty([]model.Action{model.ActionOrganizationRead}))
}
