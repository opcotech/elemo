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
		{"ASSIGNED_TO", EdgeKindAssignedTo, "ASSIGNED_TO"},
		{"BELONGS_TO", EdgeKindBelongsTo, "BELONGS_TO"},
		{"COMMENTED", EdgeKindCommented, "COMMENTED"},
		{"CREATED", EdgeKindCreated, "CREATED"},
		{"HAS_ATTACHMENT", EdgeKindHasAttachment, "HAS_ATTACHMENT"},
		{"HAS_COMMENT", EdgeKindHasComment, "HAS_COMMENT"},
		{"HAS_LABEL", EdgeKindHasLabel, "HAS_LABEL"},
		{"HAS_NAMESPACE", EdgeKindHasNamespace, "HAS_NAMESPACE"},
		{"HAS_PERMISSION", EdgeKindHasPermission, "HAS_PERMISSION"},
		{"HAS_PROJECT", EdgeKindHasProject, "HAS_PROJECT"},
		{"HAS_TEAM", EdgeKindHasTeam, "HAS_TEAM"},
		{"INVITED", EdgeKindInvited, "INVITED"},
		{"INVITED_TO", EdgeKindInvitedTo, "INVITED_TO"},
		{"KIND_OF", EdgeKindKindOf, "KIND_OF"},
		{"MEMBER_OF", EdgeKindMemberOf, "MEMBER_OF"},
		{"RELATED_TO", EdgeKindRelatedTo, "RELATED_TO"},
		{"SPEAKS", EdgeKindSpeaks, "SPEAKS"},
		{"WATCHES", EdgeKindWatches, "WATCHES"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}
