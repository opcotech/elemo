package neo4j

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEdgeKind_String(t *testing.T) {
	tests := []struct {
		name string
		s    EdgeKind
		want string
	}{
		{"HAS_PERMISSION", EdgeKindHasPermission, "HAS_PERMISSION"},
		{"HAS_TEAM", EdgeKindHasTeam, "HAS_TEAM"},
		{"HAS_NAMESPACE", EdgeKindHasNamespace, "HAS_NAMESPACE"},
		{"BELONGS_TO", EdgeKindBelongsTo, "BELONGS_TO"},
		{"KIND_OF", EdgeKindKindOf, "KIND_OF"},
		{"HAS_COMMENT", EdgeKindHasComment, "HAS_COMMENT"},
		{"HAS_LABEL", EdgeKindHasLabel, "HAS_LABEL"},
		{"ASSIGNED_TO", EdgeKindAssignedTo, "ASSIGNED_TO"},
		{"MEMBER_OF", EdgeKindMemberOf, "MEMBER_OF"},
		{"CREATED", EdgeKindCreated, "CREATED"},
		{"INVITED", EdgeKindInvited, "INVITED"},
		{"SPEAKS", EdgeKindSpeaks, "SPEAKS"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestAssignedToKind_String(t *testing.T) {
	tests := []struct {
		name string
		s    AssignedToKind
		want string
	}{
		{"assignee", AssignedToKindAssignee, "assignee"},
		{"reviewer", AssignedToKindReviewer, "reviewer"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}
