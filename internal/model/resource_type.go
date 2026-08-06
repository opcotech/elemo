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

// ResourceType is the type of resource that is being managed in the system.
// The resource type is used to help permission checks and to determine
// which resource types are available (eg. User, Issue, Label, etc.).
//
//go:generate go tool enumer -type=ResourceType -trimprefix=ResourceType -text -transform=noop -output=resource_type_gen.go
type ResourceType uint8
