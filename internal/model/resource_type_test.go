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
		{"Assignment", ResourceTypeAssignment, "Assignment"},
		{"Attachment", ResourceTypeAttachment, "Attachment"},
		{"Comment", ResourceTypeComment, "Comment"},
		{"Document", ResourceTypeDocument, "Document"},
		{"Issue", ResourceTypeIssue, "Issue"},
		{"Label", ResourceTypeLabel, "Label"},
		{"Namespace", ResourceTypeNamespace, "Namespace"},
		{"Organization", ResourceTypeOrganization, "Organization"},
		{"Project", ResourceTypeProject, "Project"},
		{"Role", ResourceTypeRole, "Role"},
		{"Todo", ResourceTypeTodo, "Todo"},
		{"User", ResourceTypeUser, "User"},
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
		{"Assignment", ResourceTypeAssignment, []byte("Assignment"), nil},
		{"Attachment", ResourceTypeAttachment, []byte("Attachment"), nil},
		{"Comment", ResourceTypeComment, []byte("Comment"), nil},
		{"Document", ResourceTypeDocument, []byte("Document"), nil},
		{"Issue", ResourceTypeIssue, []byte("Issue"), nil},
		{"IssueRelation", ResourceTypeIssueRelation, []byte("IssueRelation"), nil},
		{"Label", ResourceTypeLabel, []byte("Label"), nil},
		{"Namespace", ResourceTypeNamespace, []byte("Namespace"), nil},
		{"Organization", ResourceTypeOrganization, []byte("Organization"), nil},
		{"Project", ResourceTypeProject, []byte("Project"), nil},
		{"Role", ResourceTypeRole, []byte("Role"), nil},
		{"Todo", ResourceTypeTodo, []byte("Todo"), nil},
		{"User", ResourceTypeUser, []byte("User"), nil},
		{"type high", ResourceType(100), nil, ErrInvalidResourceType},
		{"type low", ResourceType(0), nil, ErrInvalidResourceType},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.rt.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestResourceType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    ResourceType
		wantErr error
	}{
		{"Assignment", []byte("Assignment"), ResourceTypeAssignment, nil},
		{"Attachment", []byte("Attachment"), ResourceTypeAttachment, nil},
		{"Comment", []byte("Comment"), ResourceTypeComment, nil},
		{"Document", []byte("Document"), ResourceTypeDocument, nil},
		{"Issue", []byte("Issue"), ResourceTypeIssue, nil},
		{"IssueRelation", []byte("IssueRelation"), ResourceTypeIssueRelation, nil},
		{"Label", []byte("Label"), ResourceTypeLabel, nil},
		{"Namespace", []byte("Namespace"), ResourceTypeNamespace, nil},
		{"Organization", []byte("Organization"), ResourceTypeOrganization, nil},
		{"Project", []byte("Project"), ResourceTypeProject, nil},
		{"Role", []byte("Role"), ResourceTypeRole, nil},
		{"Todo", []byte("Todo"), ResourceTypeTodo, nil},
		{"User", []byte("User"), ResourceTypeUser, nil},
		{"invalid", []byte("invalid"), 0, ErrInvalidResourceType},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var rt ResourceType
			err := rt.UnmarshalText(tt.text)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, rt)
			}
		})
	}
}
