package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceType_String(t *testing.T) {
	tests := []struct {
		name string
		rt   ResourceType
		want string
	}{
		{"ResourceType", ResourceTypeKind, "ResourceType"},
		{"Assignment", ResourceTypeAssignment, "Assignment"},
		{"Attachment", ResourceTypeAttachment, "Attachment"},
		{"Comment", ResourceTypeComment, "Comment"},
		{"Document", ResourceTypeDocument, "Document"},
		{"Issue", ResourceTypeIssue, "Issue"},
		{"Label", ResourceTypeLabel, "Label"},
		{"Namespace", ResourceTypeNamespace, "Namespace"},
		{"Notification", ResourceTypeNotification, "Notification"},
		{"Organization", ResourceTypeOrganization, "Organization"},
		{"Project", ResourceTypeProject, "Project"},
		{"Role", ResourceTypeRole, "Role"},
		{"Todo", ResourceTypeTodo, "Todo"},
		{"User", ResourceTypeUser, "User"},
		{"UserToken", ResourceTypeUserToken, "UserToken"},
		{"Folder", ResourceTypeFolder, "Folder"},
		{"Installation", ResourceTypeInstallation, "Installation"},
		{"Team", ResourceTypeTeam, "Team"},
		{"CustomFieldDefinition", ResourceTypeCustomFieldDefinition, "CustomFieldDefinition"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.rt.String())
		})
	}
}

func TestResourceType_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		rt      ResourceType
		want    []byte
		wantErr error
	}{
		{"ResourceType", ResourceTypeKind, []byte("ResourceType"), nil},
		{"Assignment", ResourceTypeAssignment, []byte("Assignment"), nil},
		{"Attachment", ResourceTypeAttachment, []byte("Attachment"), nil},
		{"Comment", ResourceTypeComment, []byte("Comment"), nil},
		{"Document", ResourceTypeDocument, []byte("Document"), nil},
		{"Issue", ResourceTypeIssue, []byte("Issue"), nil},
		{"IssueRelation", ResourceTypeIssueRelation, []byte("IssueRelation"), nil},
		{"Label", ResourceTypeLabel, []byte("Label"), nil},
		{"Namespace", ResourceTypeNamespace, []byte("Namespace"), nil},
		{"Notification", ResourceTypeNotification, []byte("Notification"), nil},
		{"Organization", ResourceTypeOrganization, []byte("Organization"), nil},
		{"Project", ResourceTypeProject, []byte("Project"), nil},
		{"Role", ResourceTypeRole, []byte("Role"), nil},
		{"Todo", ResourceTypeTodo, []byte("Todo"), nil},
		{"User", ResourceTypeUser, []byte("User"), nil},
		{"UserToken", ResourceTypeUserToken, []byte("UserToken"), nil},
		{"Folder", ResourceTypeFolder, []byte("Folder"), nil},
		{"Installation", ResourceTypeInstallation, []byte("Installation"), nil},
		{"Team", ResourceTypeTeam, []byte("Team"), nil},
		{"CustomFieldDefinition", ResourceTypeCustomFieldDefinition, []byte("CustomFieldDefinition"), nil},
		{"type high", ResourceType(100), []byte("ResourceType(100)"), nil},
		{"type low", ResourceType(0), []byte("ResourceType(0)"), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.rt.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResourceType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    ResourceType
		wantErr bool
	}{
		{"ResourceType", []byte("ResourceType"), ResourceTypeKind, false},
		{"Assignment", []byte("Assignment"), ResourceTypeAssignment, false},
		{"Attachment", []byte("Attachment"), ResourceTypeAttachment, false},
		{"Comment", []byte("Comment"), ResourceTypeComment, false},
		{"Document", []byte("Document"), ResourceTypeDocument, false},
		{"Issue", []byte("Issue"), ResourceTypeIssue, false},
		{"IssueRelation", []byte("IssueRelation"), ResourceTypeIssueRelation, false},
		{"Label", []byte("Label"), ResourceTypeLabel, false},
		{"Namespace", []byte("Namespace"), ResourceTypeNamespace, false},
		{"Notification", []byte("Notification"), ResourceTypeNotification, false},
		{"Organization", []byte("Organization"), ResourceTypeOrganization, false},
		{"Project", []byte("Project"), ResourceTypeProject, false},
		{"Role", []byte("Role"), ResourceTypeRole, false},
		{"Todo", []byte("Todo"), ResourceTypeTodo, false},
		{"User", []byte("User"), ResourceTypeUser, false},
		{"UserToken", []byte("UserToken"), ResourceTypeUserToken, false},
		{"Folder", []byte("Folder"), ResourceTypeFolder, false},
		{"Installation", []byte("Installation"), ResourceTypeInstallation, false},
		{"Team", []byte("Team"), ResourceTypeTeam, false},
		{"CustomFieldDefinition", []byte("CustomFieldDefinition"), ResourceTypeCustomFieldDefinition, false},
		{"invalid", []byte("invalid"), 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var rt ResourceType
			err := rt.UnmarshalText(tt.text)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, rt)
			}
		})
	}
}

func TestResourceType_MarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		rt   ResourceType
		want []byte
	}{
		{"zero", ResourceType(0), []byte{0}},
		{"issue", ResourceTypeIssue, []byte{byte(ResourceTypeIssue)}},
		{"user", ResourceTypeUser, []byte{byte(ResourceTypeUser)}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.rt.MarshalBinary()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResourceType_UnmarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    ResourceType
		wantErr bool
	}{
		{"empty", []byte{}, ResourceType(0), false},
		{"single byte zero", []byte{0}, ResourceType(0), false},
		{"single byte issue", []byte{byte(ResourceTypeIssue)}, ResourceTypeIssue, false},
		{"text name", []byte("User"), ResourceTypeUser, false},
		{"enumer fallback zero", []byte("ResourceType(0)"), ResourceType(0), false},
		{"enumer fallback high", []byte("ResourceType(100)"), ResourceType(100), false},
		{"invalid text", []byte("not-a-type"), 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var rt ResourceType
			err := rt.UnmarshalBinary(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, rt)
		})
	}
}
