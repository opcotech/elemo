package neo4j

import "errors"

const (
	EdgeKindHasPermission EdgeKind = iota + 1 // a subject has permission on a resource
	EdgeKindHasTeam                           // an organization or project has a team
	EdgeKindHasNamespace                      // an organization has a namespace
	EdgeKindHasProject                        // a namespace has a project
	EdgeKindBelongsTo                         // a resource belongs to another
	EdgeKindKindOf                            // a resource is a kind of another
	EdgeKindHasComment                        // a resource has a comment
	EdgeKindHasLabel                          // a resource is labeled by a label
	EdgeKindAssignedTo                        // a user is assigned to a resource
	EdgeKindMemberOf                          // a user is a member of a team
	EdgeKindCreated                           // a user created a resource
	EdgeKindInvited                           // a user invited another user
	EdgeKindSpeaks                            // a user speaks a language
	EdgeKindCommented                         // a user commented a resource
	EdgeKindHasAttachment                     // a resource has an attachment
)

var (
	ErrInvalidEdgeKind = errors.New("invalid edge kind") // the edge kind is invalid

	relationKindKeys = map[string]EdgeKind{
		"HAS_PERMISSION": EdgeKindHasPermission,
		"HAS_TEAM":       EdgeKindHasTeam,
		"HAS_NAMESPACE":  EdgeKindHasNamespace,
		"HAS_PROJECT":    EdgeKindHasProject,
		"BELONGS_TO":     EdgeKindBelongsTo,
		"KIND_OF":        EdgeKindKindOf,
		"HAS_COMMENT":    EdgeKindHasComment,
		"HAS_LABEL":      EdgeKindHasLabel,
		"ASSIGNED_TO":    EdgeKindAssignedTo,
		"MEMBER_OF":      EdgeKindMemberOf,
		"CREATED":        EdgeKindCreated,
		"INVITED":        EdgeKindInvited,
		"SPEAKS":         EdgeKindSpeaks,
		"COMMENTED":      EdgeKindCommented,
		"HAS_ATTACHMENT": EdgeKindHasAttachment,
	}
	relationKindValues = map[EdgeKind]string{
		EdgeKindHasPermission: "HAS_PERMISSION",
		EdgeKindHasTeam:       "HAS_TEAM",
		EdgeKindHasNamespace:  "HAS_NAMESPACE",
		EdgeKindHasProject:    "HAS_PROJECT",
		EdgeKindBelongsTo:     "BELONGS_TO",
		EdgeKindKindOf:        "KIND_OF",
		EdgeKindHasComment:    "HAS_COMMENT",
		EdgeKindHasLabel:      "HAS_LABEL",
		EdgeKindAssignedTo:    "ASSIGNED_TO",
		EdgeKindMemberOf:      "MEMBER_OF",
		EdgeKindCreated:       "CREATED",
		EdgeKindInvited:       "INVITED",
		EdgeKindSpeaks:        "SPEAKS",
		EdgeKindCommented:     "COMMENTED",
		EdgeKindHasAttachment: "HAS_ATTACHMENT",
	}
)

// EdgeKind is the kind of relation between two entities.
type EdgeKind uint8

// String returns the string representation of the relation kind.
func (k EdgeKind) String() string {
	return relationKindValues[k]
}
