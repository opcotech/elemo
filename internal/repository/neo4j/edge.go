package neo4j

import "errors"

const (
	EdgeKindHasPermission EdgeKind = iota + 1 // a subject has permission on a resource
	EdgeKindHasTeam                           // an organization or project has a team
	EdgeKindHasNamespace                      // an organization has a namespace
	EdgeKindBelongsTo                         // a resource belongs to another
	EdgeKindKindOf                            // a resource is a kind of another
	EdgeKindHasComment                        // a resource has a comment
	EdgeKindLabels                            // a resource is labeled by a label
	EdgeKindAssignedTo                        // a user is assigned to a resource
	EdgeKindMemberOf                          // a user is a member of a team
	EdgeKindCreated                           // a user created a resource
	EdgeKindInvited                           // a user invited another user
	EdgeKindSpeaks                            // a user speaks a language
)

const (
	AssignedToKindAssignee AssignedToKind = iota + 1 // a user is assigned as an assignee
	AssignedToKindReviewer                           // a user is assigned as a reviewer
)

var (
	ErrInvalidEdgeKind       = errors.New("invalid edge kind")        // the edge kind is invalid
	ErrInvalidAssignedToKind = errors.New("invalid assigned to kind") // the assigned to kind is invalid

	relationKindKeys = map[string]EdgeKind{
		"HAS_PERMISSION": EdgeKindHasPermission,
		"HAS_TEAM":       EdgeKindHasTeam,
		"HAS_NAMESPACE":  EdgeKindHasNamespace,
		"BELONGS_TO":     EdgeKindBelongsTo,
		"KIND_OF":        EdgeKindKindOf,
		"HAS_COMMENT":    EdgeKindHasComment,
		"LABELS":         EdgeKindLabels,
		"ASSIGNED_TO":    EdgeKindAssignedTo,
		"MEMBER_OF":      EdgeKindMemberOf,
		"CREATED":        EdgeKindCreated,
		"INVITED":        EdgeKindInvited,
		"SPEAKS":         EdgeKindSpeaks,
	}
	relationKindValues = map[EdgeKind]string{
		EdgeKindHasPermission: "HAS_PERMISSION",
		EdgeKindHasTeam:       "HAS_TEAM",
		EdgeKindHasNamespace:  "HAS_NAMESPACE",
		EdgeKindBelongsTo:     "BELONGS_TO",
		EdgeKindKindOf:        "KIND_OF",
		EdgeKindHasComment:    "HAS_COMMENT",
		EdgeKindLabels:        "LABELS",
		EdgeKindAssignedTo:    "ASSIGNED_TO",
		EdgeKindMemberOf:      "MEMBER_OF",
		EdgeKindCreated:       "CREATED",
		EdgeKindInvited:       "INVITED",
		EdgeKindSpeaks:        "SPEAKS",
	}

	assignedToKindKeys = map[string]AssignedToKind{
		"assignee": AssignedToKindAssignee,
		"reviewer": AssignedToKindReviewer,
	}
	assignedToKindValues = map[AssignedToKind]string{
		AssignedToKindAssignee: "assignee",
		AssignedToKindReviewer: "reviewer",
	}
)

// EdgeKind is the kind of relation between two entities.
type EdgeKind uint8

// String returns the string representation of the relation kind.
func (k EdgeKind) String() string {
	return relationKindValues[k]
}

// AssignedToKind is the kind of assignment between a user and a resource.
type AssignedToKind uint8

// String returns the string representation of the relation kind.
func (k AssignedToKind) String() string {
	return assignedToKindValues[k]
}
