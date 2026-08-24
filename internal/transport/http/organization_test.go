package http

import (
	"context"
	"errors"
	"testing"
	"time"

	oapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestOrganizationToDTO(t *testing.T) {
	t.Parallel()

	createdAt := convert.ToPointer(time.Now().UTC())
	updatedAt := convert.ToPointer(time.Now().UTC())
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("with projected counts", func(t *testing.T) {
		t.Parallel()

		org := &service.Organization{
			ID:             orgID,
			Name:           "ACME Inc.",
			Email:          "info@example.com",
			Logo:           "https://example.com/logo.png",
			Website:        "https://example.com",
			Status:         model.OrganizationStatusActive,
			MemberCount:    convert.ToPointer(int64(12)),
			TeamCount:      convert.ToPointer(int64(3)),
			NamespaceCount: convert.ToPointer(int64(2)),
			DocumentCount:  convert.ToPointer(int64(4)),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}

		dto := organizationToDTO(org)
		assert.Equal(t, orgID.String(), dto.Id)
		assert.Equal(t, "ACME Inc.", dto.Name)
		require.NotNil(t, dto.MemberCount)
		assert.Equal(t, int64(12), *dto.MemberCount)
		require.NotNil(t, dto.DocumentCount)
		assert.Equal(t, int64(4), *dto.DocumentCount)
		assert.Equal(t, api.OrganizationStatus(org.Status.String()), dto.Status)
	})

	t.Run("without projected counts", func(t *testing.T) {
		t.Parallel()

		org := &service.Organization{
			ID:        orgID,
			Name:      "ACME Inc.",
			Email:     "info@example.com",
			Status:    model.OrganizationStatusActive,
			CreatedAt: createdAt,
		}

		dto := organizationToDTO(org)
		assert.Equal(t, orgID.String(), dto.Id)
		assert.Nil(t, dto.MemberCount)
		assert.Nil(t, dto.DocumentCount)
		assert.Nil(t, dto.TeamCount)
		assert.Nil(t, dto.NamespaceCount)
	})
}

func newTestOrganizationController(t *testing.T, os service.OrganizationService, rs service.RoleService, ts service.TeamService) OrganizationController {
	t.Helper()
	c, err := NewOrganizationController(os, rs, ts, mocksvc.NewMockUserService(gomock.NewController(t)))
	require.NoError(t, err)
	return c
}

func newServiceOrganization() *service.Organization {
	return &service.Organization{
		ID:        model.MustNewID(model.ResourceTypeOrganization),
		Name:      "ACME Inc.",
		Email:     "info@example.com",
		Status:    model.OrganizationStatusActive,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

func TestNewOrganizationController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		c, err := NewOrganizationController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockRoleService(ctrl),
			mocksvc.NewMockTeamService(ctrl),
			mocksvc.NewMockUserService(ctrl),
		)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing organization service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewOrganizationController(
			nil,
			mocksvc.NewMockRoleService(ctrl),
			mocksvc.NewMockTeamService(ctrl),
			mocksvc.NewMockUserService(ctrl),
		)
		assert.ErrorIs(t, err, ErrNoOrganizationService)
	})

	t.Run("missing role service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewOrganizationController(
			mocksvc.NewMockOrganizationService(ctrl),
			nil,
			mocksvc.NewMockTeamService(ctrl),
			mocksvc.NewMockUserService(ctrl),
		)
		assert.ErrorIs(t, err, ErrNoRoleService)
	})

	t.Run("missing team service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewOrganizationController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockRoleService(ctrl),
			nil,
			mocksvc.NewMockUserService(ctrl),
		)
		assert.ErrorIs(t, err, ErrNoTeamService)
	})

	t.Run("missing user service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewOrganizationController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockRoleService(ctrl),
			mocksvc.NewMockTeamService(ctrl),
			nil,
		)
		assert.ErrorIs(t, err, ErrNoUserService)
	})
}

func TestOrganizationController_V1OrganizationsCreate(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	org := newServiceOrganization()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Create(gomock.Any(), userID, service.CreateOrganizationOpts{
			Name:  "ACME Inc.",
			Email: "info@example.com",
		}).Return(org, nil)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationsCreate(ctx, api.V1OrganizationsCreateRequestObject{
			Body: &api.V1OrganizationsCreateJSONRequestBody{
				Name:  "ACME Inc.",
				Email: oapiTypes.Email("info@example.com"),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, org.ID.String(), got.Id)
	})

	t.Run("missing user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationsCreate(context.Background(), api.V1OrganizationsCreateRequestObject{
			Body: &api.V1OrganizationsCreateJSONRequestBody{Name: "ACME Inc.", Email: oapiTypes.Email("info@example.com")},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Create(gomock.Any(), userID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationsCreate(ctx, api.V1OrganizationsCreateRequestObject{
			Body: &api.V1OrganizationsCreateJSONRequestBody{Name: "ACME Inc.", Email: oapiTypes.Email("info@example.com")},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationGet(t *testing.T) {
	t.Parallel()

	org := newServiceOrganization()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Get(gomock.Any(), org.ID).Return(org, nil)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationGet(context.Background(), api.V1OrganizationGetRequestObject{Id: org.ID.String()})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, org.ID.String(), got.Id)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationGet(context.Background(), api.V1OrganizationGetRequestObject{Id: "not-a-xid"})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Get(gomock.Any(), org.ID).Return(nil, service.ErrNoPermission)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationGet(context.Background(), api.V1OrganizationGetRequestObject{Id: org.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Get(gomock.Any(), org.ID).Return(nil, errors.Join(service.ErrOrganizationGet, repository.ErrNotFound))

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationGet(context.Background(), api.V1OrganizationGetRequestObject{Id: org.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationUpdate(t *testing.T) {
	t.Parallel()

	org := newServiceOrganization()
	name := "Updated"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Update(gomock.Any(), org.ID, gomock.Any()).Return(org, nil)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationUpdate(context.Background(), api.V1OrganizationUpdateRequestObject{
			Id:   org.ID.String(),
			Body: &api.V1OrganizationUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationUpdate200JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Update(gomock.Any(), org.ID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationUpdate(context.Background(), api.V1OrganizationUpdateRequestObject{
			Id:   org.ID.String(),
			Body: &api.V1OrganizationUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationUpdate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationDelete(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Delete(gomock.Any(), orgID, false).Return(nil)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationDelete(context.Background(), api.V1OrganizationDeleteRequestObject{Id: orgID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationDelete204Response)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().Delete(gomock.Any(), orgID, false).Return(errors.Join(service.ErrOrganizationDelete, repository.ErrNotFound))

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationDelete(context.Background(), api.V1OrganizationDeleteRequestObject{Id: orgID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationDelete404JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationMembersAdd(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	// The handler parses request.Id as both the organization and the user, and ignores Body.UserId.
	userIDFromOrg, err := model.NewIDFromString(orgID.String(), model.ResourceTypeUser.String())
	require.NoError(t, err)

	t.Run("success uses path id as user id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().AddMember(gomock.Any(), orgID, userIDFromOrg).Return(nil)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationMembersAdd(context.Background(), api.V1OrganizationMembersAddRequestObject{
			Id: orgID.String(),
			Body: &api.V1OrganizationMembersAddJSONRequestBody{
				UserId: model.MustNewID(model.ResourceTypeUser).String(),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationMembersAdd201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, userIDFromOrg.String(), got.Id)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationMembersAdd(context.Background(), api.V1OrganizationMembersAddRequestObject{Id: "not-a-xid"})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationMembersAdd400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		os := mocksvc.NewMockOrganizationService(ctrl)
		os.EXPECT().AddMember(gomock.Any(), orgID, userIDFromOrg).Return(service.ErrNoPermission)

		c := newTestOrganizationController(t, os, mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationMembersAdd(context.Background(), api.V1OrganizationMembersAddRequestObject{Id: orgID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationMembersAdd403JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationRolesCreate(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	role := newServiceRole()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		rs := mocksvc.NewMockRoleService(ctrl)
		rs.EXPECT().Create(gomock.Any(), userID, orgID, service.CreateRoleOpts{Name: "Custom role"}).Return(role, nil)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), rs, mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationRolesCreate(ctx, api.V1OrganizationRolesCreateRequestObject{
			Id:   orgID.String(),
			Body: &api.V1OrganizationRolesCreateJSONRequestBody{Name: "Custom role"},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationRolesCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, role.ID.String(), got.Id)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		rs := mocksvc.NewMockRoleService(ctrl)
		rs.EXPECT().Create(gomock.Any(), userID, orgID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), rs, mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationRolesCreate(ctx, api.V1OrganizationRolesCreateRequestObject{
			Id:   orgID.String(),
			Body: &api.V1OrganizationRolesCreateJSONRequestBody{Name: "Custom role"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationRolesCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationRoleGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	role := newServiceRole()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		rs := mocksvc.NewMockRoleService(ctrl)
		rs.EXPECT().Get(gomock.Any(), role.ID, orgID).Return(role, nil)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), rs, mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationRoleGet(context.Background(), api.V1OrganizationRoleGetRequestObject{
			Id:     orgID.String(),
			RoleId: role.ID.String(),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationRoleGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, role.ID.String(), got.Id)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		rs := mocksvc.NewMockRoleService(ctrl)
		rs.EXPECT().Get(gomock.Any(), role.ID, orgID).Return(nil, errors.Join(service.ErrRoleGet, repository.ErrNotFound))

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), rs, mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationRoleGet(context.Background(), api.V1OrganizationRoleGetRequestObject{
			Id:     orgID.String(),
			RoleId: role.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationRoleGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationTeamsCreate(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	team := newServiceTeam()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().Create(gomock.Any(), orgID, service.CreateTeamOpts{Name: "Platform"}).Return(team, nil)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamsCreate(context.Background(), api.V1OrganizationTeamsCreateRequestObject{
			Id:   orgID.String(),
			Body: &api.V1OrganizationTeamsCreateJSONRequestBody{Name: "Platform"},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationTeamsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, team.ID.String(), got.Id)
	})

	t.Run("bad org id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationTeamsCreate(context.Background(), api.V1OrganizationTeamsCreateRequestObject{
			Id:   "not-a-xid",
			Body: &api.V1OrganizationTeamsCreateJSONRequestBody{Name: "Platform"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().Create(gomock.Any(), orgID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamsCreate(context.Background(), api.V1OrganizationTeamsCreateRequestObject{
			Id:   orgID.String(),
			Body: &api.V1OrganizationTeamsCreateJSONRequestBody{Name: "Platform"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().Create(gomock.Any(), orgID, gomock.Any()).Return(nil, errors.Join(service.ErrTeamCreate, repository.ErrNotFound))

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamsCreate(context.Background(), api.V1OrganizationTeamsCreateRequestObject{
			Id:   orgID.String(),
			Body: &api.V1OrganizationTeamsCreateJSONRequestBody{Name: "Platform"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamsCreate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationTeamGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	team := newServiceTeam()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().Get(gomock.Any(), team.ID, orgID).Return(team, nil)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamGet(context.Background(), api.V1OrganizationTeamGetRequestObject{
			Id:     orgID.String(),
			TeamId: team.ID.String(),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationTeamGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, team.ID.String(), got.Id)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().Get(gomock.Any(), team.ID, orgID).Return(nil, service.ErrNoPermission)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamGet(context.Background(), api.V1OrganizationTeamGetRequestObject{
			Id:     orgID.String(),
			TeamId: team.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().Get(gomock.Any(), team.ID, orgID).Return(nil, errors.Join(service.ErrTeamGet, repository.ErrNotFound))

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamGet(context.Background(), api.V1OrganizationTeamGetRequestObject{
			Id:     orgID.String(),
			TeamId: team.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestOrganizationController_V1OrganizationTeamMembersAdd(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	userID := model.MustNewID(model.ResourceTypeUser)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().AddMember(gomock.Any(), teamID, userID, orgID).Return(nil)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamMembersAdd(context.Background(), api.V1OrganizationTeamMembersAddRequestObject{
			Id:     orgID.String(),
			TeamId: teamID.String(),
			Body:   &api.V1OrganizationTeamMembersAddJSONRequestBody{UserId: userID.String()},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationTeamMembersAdd201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, userID.String(), got.Id)
	})

	t.Run("missing user id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), mocksvc.NewMockTeamService(ctrl))
		resp, err := c.V1OrganizationTeamMembersAdd(context.Background(), api.V1OrganizationTeamMembersAddRequestObject{
			Id:     orgID.String(),
			TeamId: teamID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamMembersAdd400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ts := mocksvc.NewMockTeamService(ctrl)
		ts.EXPECT().AddMember(gomock.Any(), teamID, userID, orgID).Return(service.ErrNoPermission)

		c := newTestOrganizationController(t, mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockRoleService(ctrl), ts)
		resp, err := c.V1OrganizationTeamMembersAdd(context.Background(), api.V1OrganizationTeamMembersAddRequestObject{
			Id:     orgID.String(),
			TeamId: teamID.String(),
			Body:   &api.V1OrganizationTeamMembersAddJSONRequestBody{UserId: userID.String()},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationTeamMembersAdd403JSONResponse)
		assert.True(t, ok)
	})
}
