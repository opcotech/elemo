package model

const (
	ResourceTypeKind          ResourceType = iota + 1 // ResourceType
	ResourceTypeAssignment                            // Assignment
	ResourceTypeAttachment                            // Attachment
	ResourceTypeComment                               // Comment
	ResourceTypeDocument                              // Document
	ResourceTypeIssue                                 // Issue
	ResourceTypeIssueRelation                         // IssueRelation
	ResourceTypeLabel                                 // Label
	ResourceTypeNamespace                             // Namespace
	ResourceTypeNotification                          // Notification
	ResourceTypeOrganization                          // Organization
	ResourceTypePermission                            // Permission
	ResourceTypeProject                               // Project
	ResourceTypeRole                                  // Role
	ResourceTypeTodo                                  // Todo
	ResourceTypeUser                                  // User
	ResourceTypeUserToken                             // UserToken
)

// ResourceType is the type of resource that is being managed in the system.
// The resource type is used to help permission checks and to determine
// which resource types are available (eg. User, Issue, Label, etc.).
//
//go:generate go tool enumer -type=ResourceType -text -transform=noop -linecomment -output=resource_type_gen.go
type ResourceType uint8
