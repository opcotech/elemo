package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newServiceRole() *service.Role {
	return &service.Role{
		ID:          model.MustNewID(model.ResourceTypeRole),
		Key:         model.RoleKeyOrgMember,
		Name:        "Organization member",
		Description: "Read the organization they belong to.",
		Actions:     []model.Action{model.ActionOrganizationRead},
		MemberCount: convert.ToPointer(int64(2)),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}

func TestCreateRoleJSONRequestBodyToCreateRoleOpts(t *testing.T) {
	t.Parallel()

	key := "custom-role"
	description := "Custom role description"
	actions := []api.Action{api.Action(model.ActionOrganizationRead.String())}

	t.Run("with optional fields", func(t *testing.T) {
		t.Parallel()
		opts, err := createRoleJSONRequestBodyToCreateRoleOpts(&api.V1OrganizationRolesCreateJSONRequestBody{
			Name:        "Custom role",
			Key:         &key,
			Description: &description,
			Actions:     &actions,
		})
		require.NoError(t, err)
		assert.Equal(t, "Custom role", opts.Name)
		assert.Equal(t, key, opts.Key)
		assert.Equal(t, description, opts.Description)
		assert.Equal(t, []model.Action{model.ActionOrganizationRead}, opts.Actions)
	})

	t.Run("name only", func(t *testing.T) {
		t.Parallel()
		opts, err := createRoleJSONRequestBodyToCreateRoleOpts(&api.V1OrganizationRolesCreateJSONRequestBody{
			Name: "Custom role",
		})
		require.NoError(t, err)
		assert.Equal(t, "Custom role", opts.Name)
		assert.Empty(t, opts.Key)
		assert.Empty(t, opts.Actions)
	})
}

func TestUpdateRoleJSONRequestBodyToUpdateRoleOpts(t *testing.T) {
	t.Parallel()

	t.Run("actions null becomes empty slice", func(t *testing.T) {
		t.Parallel()
		opts, err := updateRoleJSONRequestBodyToUpdateRoleOpts(&api.V1OrganizationRoleUpdateJSONRequestBody{
			Actions: optional.Null[[]string](),
		})
		require.NoError(t, err)
		require.True(t, opts.Actions.Defined)
		require.NotNil(t, opts.Actions.Value)
		assert.Empty(t, *opts.Actions.Value)
	})

	t.Run("actions values", func(t *testing.T) {
		t.Parallel()
		values := []string{model.ActionOrganizationRead.String()}
		opts, err := updateRoleJSONRequestBodyToUpdateRoleOpts(&api.V1OrganizationRoleUpdateJSONRequestBody{
			Actions: optional.Some(values),
		})
		require.NoError(t, err)
		require.True(t, opts.Actions.Defined)
		require.NotNil(t, opts.Actions.Value)
		assert.Equal(t, []model.Action{model.ActionOrganizationRead}, *opts.Actions.Value)
	})

	t.Run("undefined actions left unset", func(t *testing.T) {
		t.Parallel()
		name := "updated"
		opts, err := updateRoleJSONRequestBodyToUpdateRoleOpts(&api.V1OrganizationRoleUpdateJSONRequestBody{
			Name: &name,
		})
		require.NoError(t, err)
		assert.Equal(t, optional.Some("updated"), opts.Name)
		assert.False(t, opts.Actions.Defined)
	})
}

func TestRoleToDTO(t *testing.T) {
	t.Parallel()

	role := newServiceRole()
	dto := roleToDTO(role)
	assert.Equal(t, role.ID.String(), dto.Id)
	assert.Equal(t, role.Key, dto.Key)
	assert.Equal(t, role.Name, dto.Name)
	require.NotNil(t, dto.Description)
	assert.Equal(t, role.Description, *dto.Description)
	assert.Equal(t, []api.Action{api.Action(model.ActionOrganizationRead.String())}, dto.Actions)
	require.NotNil(t, dto.MemberCount)
	assert.Equal(t, int64(2), *dto.MemberCount)
}
