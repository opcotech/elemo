package model

const (
	ResourceTypeResourceType  ResourceType = iota + 1 // resource type
	ResourceTypeAssignment                            // assignment resource type
	ResourceTypeAttachment                            // attachment resource type
	ResourceTypeComment                               // comment resource type
	ResourceTypeDocument                              // document resource type
	ResourceTypeIssue                                 // issue resource type
	ResourceTypeIssueRelation                         // issue relation resource type
	ResourceTypeLabel                                 // Type resource type
	ResourceTypeNamespace                             // namespace resource type
	ResourceTypeNotification                          // notification resource type
	ResourceTypeOrganization                          // organization resource type
	ResourceTypePermission                            // permission resource type
	ResourceTypeProject                               // project resource type
	ResourceTypeRole                                  // role resource type
	ResourceTypeTodo                                  // todo resource type
	ResourceTypeUser                                  // user resource type
	ResourceTypeUserToken                             // user token resource type
)

var (
	resourceTypeKeys = map[string]ResourceType{
		"ResourceType":  ResourceTypeResourceType,
		"Assignment":    ResourceTypeAssignment,
		"Attachment":    ResourceTypeAttachment,
		"Comment":       ResourceTypeComment,
		"Document":      ResourceTypeDocument,
		"Issue":         ResourceTypeIssue,
		"IssueRelation": ResourceTypeIssueRelation,
		"Label":         ResourceTypeLabel,
		"Namespace":     ResourceTypeNamespace,
		"Notification":  ResourceTypeNotification,
		"Organization":  ResourceTypeOrganization,
		"Permission":    ResourceTypePermission,
		"Project":       ResourceTypeProject,
		"Role":          ResourceTypeRole,
		"Todo":          ResourceTypeTodo,
		"User":          ResourceTypeUser,
		"UserToken":     ResourceTypeUserToken,
	}
	resourceTypeValues = map[ResourceType]string{
		ResourceTypeResourceType:  "ResourceType",
		ResourceTypeAssignment:    "Assignment",
		ResourceTypeAttachment:    "Attachment",
		ResourceTypeComment:       "Comment",
		ResourceTypeDocument:      "Document",
		ResourceTypeIssue:         "Issue",
		ResourceTypeIssueRelation: "IssueRelation",
		ResourceTypeLabel:         "Label",
		ResourceTypeNamespace:     "Namespace",
		ResourceTypeNotification:  "Notification",
		ResourceTypeOrganization:  "Organization",
		ResourceTypePermission:    "Permission",
		ResourceTypeProject:       "Project",
		ResourceTypeRole:          "Role",
		ResourceTypeTodo:          "Todo",
		ResourceTypeUser:          "User",
		ResourceTypeUserToken:     "UserToken",
	}
)

// ResourceType is the type of resource that is being managed in the system.
// The resource type is used to help permission checks and to determine
// which resource types are available (eg. User, Issue, Label, etc.).
type ResourceType uint8

// String returns the string representation of the resource type.
func (t ResourceType) String() string {
	return resourceTypeValues[t]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t ResourceType) MarshalText() (text []byte, err error) {
	if t < 1 || t > 17 {
		return nil, ErrInvalidResourceType
	}
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *ResourceType) UnmarshalText(text []byte) error {
	if v, ok := resourceTypeKeys[string(text)]; ok {
		*t = v
		return nil
	}
	return ErrInvalidResourceType
}
