package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAction_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action Action
		want   bool
	}{
		{"organization.create", ActionOrganizationCreate, true},
		{"organization.read", ActionOrganizationRead, true},
		{"permission.manage", ActionPermissionManage, true},
		{"empty", Action(""), false},
		{"star", Action("*"), false},
		{"legacy write", Action("write"), false},
		{"unknown", Action("issue.explode"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.Valid())
		})
	}
}

func TestAction_UnmarshalText(t *testing.T) {
	t.Parallel()

	var action Action
	require.NoError(t, action.UnmarshalText([]byte("project.read")))
	assert.Equal(t, ActionProjectRead, action)

	require.ErrorIs(t, action.UnmarshalText([]byte("*")), ErrInvalidAction)
	require.ErrorIs(t, action.UnmarshalText([]byte("write")), ErrInvalidAction)
}

func TestRoleTemplateByKey(t *testing.T) {
	t.Parallel()

	tmpl, err := RoleTemplateByKey(RoleKeyOrgAdmin)
	require.NoError(t, err)
	assert.NotEmpty(t, tmpl.Actions)
	assert.NotContains(t, tmpl.Actions, ActionOrganizationCreate)

	_, err = RoleTemplateByKey("god")
	require.ErrorIs(t, err, ErrUnknownRoleKey)

	_, err = RoleTemplateByKey("todo-owner")
	require.ErrorIs(t, err, ErrUnknownRoleKey)
}

func TestAction_Validate(t *testing.T) {
	t.Parallel()

	require.NoError(t, ActionOrganizationRead.Validate())
	require.ErrorIs(t, Action("").Validate(), ErrInvalidAction)
	require.ErrorIs(t, Action("write").Validate(), ErrInvalidAction)
}

func TestAction_MarshalText(t *testing.T) {
	t.Parallel()

	got, err := ActionProjectRead.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("project.read"), got)
	assert.Equal(t, "project.read", ActionProjectRead.String())
}

func TestActionStrings(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"issue.read", "issue.update"}, ActionStrings([]Action{ActionIssueRead, ActionIssueUpdate}))
	assert.Empty(t, ActionStrings(nil))
}

func TestRoleTemplate_ActionStrings(t *testing.T) {
	t.Parallel()

	tmpl, err := RoleTemplateByKey(RoleKeyProjectViewer)
	require.NoError(t, err)
	assert.Equal(t, []string{"project.read", "issue.read", "document.read"}, tmpl.ActionStrings())
}

func TestParseActions(t *testing.T) {
	t.Parallel()

	got, err := ParseActions([]string{"issue.read", "issue.update", "issue.read"})
	require.NoError(t, err)
	assert.Equal(t, []Action{ActionIssueRead, ActionIssueUpdate}, got)

	got, err = ParseActions(nil)
	require.NoError(t, err)
	assert.Equal(t, []Action{}, got)

	got, err = ParseActions([]string{"  issue.read  "})
	require.NoError(t, err)
	assert.Equal(t, []Action{ActionIssueRead}, got)

	_, err = ParseActions([]string{"issue.read", "*"})
	require.ErrorIs(t, err, ErrInvalidAction)
}

func TestAllActionsAreValid(t *testing.T) {
	t.Parallel()

	for _, action := range Actions {
		assert.True(t, action.Valid(), action)
	}
}

func TestReadActionFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resource  ResourceType
		want      Action
		wantFound bool
	}{
		{"installation", ResourceTypeInstallation, ActionOrganizationCreate, true},
		{"organization", ResourceTypeOrganization, ActionOrganizationRead, true},
		{"namespace", ResourceTypeNamespace, ActionNamespaceRead, true},
		{"project", ResourceTypeProject, ActionProjectRead, true},
		{"issue", ResourceTypeIssue, ActionIssueRead, true},
		{"document", ResourceTypeDocument, ActionDocumentRead, true},
		{"folder", ResourceTypeFolder, ActionDocumentRead, true},
		{"user", ResourceTypeUser, "", false},
		{"team", ResourceTypeTeam, "", false},
		{"role", ResourceTypeRole, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ReadActionFor(tt.resource)
			assert.Equal(t, tt.wantFound, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleTemplateByKey_KnownTemplates(t *testing.T) {
	t.Parallel()

	keys := []string{
		RoleKeyOrgAdmin,
		RoleKeyOrgMember,
		RoleKeyNamespaceAdmin,
		RoleKeyProjectMaintainer,
		RoleKeyProjectViewer,
		RoleKeyIssueMaintainer,
		RoleKeyDocumentMaintainer,
	}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			tmpl, err := RoleTemplateByKey(key)
			require.NoError(t, err)
			assert.Equal(t, key, tmpl.Key)
			assert.NotEmpty(t, tmpl.Actions)
			assert.NotContains(t, tmpl.Actions, ActionOrganizationCreate)
		})
	}
}

func TestIsPrincipalType(t *testing.T) {
	t.Parallel()

	assert.True(t, IsPrincipalType(ResourceTypeUser))
	assert.True(t, IsPrincipalType(ResourceTypeTeam))
	assert.True(t, IsPrincipalType(ResourceTypeOrganization))
	assert.False(t, IsPrincipalType(ResourceTypeProject))
	assert.False(t, IsPrincipalType(ResourceTypeRole))
}

func TestInstallationID(t *testing.T) {
	t.Parallel()

	id := InstallationID()
	assert.Equal(t, ResourceTypeInstallation, id.Type)
	assert.True(t, id.IsNil())
}

func TestOrgAdminExcludesOrganizationCreate(t *testing.T) {
	t.Parallel()

	for _, tmpl := range RoleTemplates {
		assert.NotContains(t, tmpl.Actions, ActionOrganizationCreate, tmpl.Key)
	}
}
