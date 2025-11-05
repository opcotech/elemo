package neo4j

const (
	EdgeKindAssignedTo    EdgeKind = iota + 1 // a user is assigned to a resource
	EdgeKindBelongsTo                         // a resource belongs to another
	EdgeKindCommented                         // a user commented a resource
	EdgeKindCreated                           // a user created a resource
	EdgeKindHasAttachment                     // a resource has an attachment
	EdgeKindHasComment                        // a resource has a comment
	EdgeKindHasLabel                          // a resource is labeled by a label
	EdgeKindHasNamespace                      // an organization has a namespace
	EdgeKindHasPermission                     // a subject has permission on a resource
	EdgeKindHasProject                        // a namespace has a project
	EdgeKindHasTeam                           // an organization or project has a team
	EdgeKindInvited                           // a user invited another user
	EdgeKindInvitedTo                         // a user is invited to an organization
	EdgeKindKindOf                            // a resource is a kind of another
	EdgeKindMemberOf                          // a user is a member of a team
	EdgeKindRelatedTo                         // a resource is related to another
	EdgeKindSpeaks                            // a user speaks a language
	EdgeKindWatches                           // a user watches a resource
)

var (
	relationKindValues = map[EdgeKind]string{
		EdgeKindAssignedTo:    "ASSIGNED_TO",
		EdgeKindBelongsTo:     "BELONGS_TO",
		EdgeKindCommented:     "COMMENTED",
		EdgeKindCreated:       "CREATED",
		EdgeKindHasAttachment: "HAS_ATTACHMENT",
		EdgeKindHasComment:    "HAS_COMMENT",
		EdgeKindHasLabel:      "HAS_LABEL",
		EdgeKindHasNamespace:  "HAS_NAMESPACE",
		EdgeKindHasPermission: "HAS_PERMISSION",
		EdgeKindHasProject:    "HAS_PROJECT",
		EdgeKindHasTeam:       "HAS_TEAM",
		EdgeKindInvited:       "INVITED",
		EdgeKindInvitedTo:     "INVITED_TO",
		EdgeKindKindOf:        "KIND_OF",
		EdgeKindMemberOf:      "MEMBER_OF",
		EdgeKindRelatedTo:     "RELATED_TO",
		EdgeKindSpeaks:        "SPEAKS",
		EdgeKindWatches:       "WATCHES",
	}
)

// EdgeKind is the kind of relation between two entities.
type EdgeKind uint8

// String returns the string representation of the relation kind.
func (k EdgeKind) String() string {
	return relationKindValues[k]
}
