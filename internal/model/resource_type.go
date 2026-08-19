package model

import (
	"strconv"
	"strings"
)

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
	ResourceTypeFolder                                // Folder
	ResourceTypeInstallation                          // Installation
	ResourceTypeTeam                                  // Team
)

// ResourceType is the type of resource that is being managed in the system.
// The resource type is used to help permission checks and to determine
// which resource types are available (eg. User, Issue, Label, etc.).
//
//go:generate go tool enumer -type=ResourceType -text -transform=noop -linecomment -output=resource_type_gen.go
type ResourceType uint8

const resourceTypeTextPrefix = "ResourceType("

// MarshalBinary encodes the enum as a single byte. msgpack prefers
// encoding.BinaryMarshaler over TextMarshaler; the generated text form of the
// zero value ("ResourceType(0)") cannot be unmarshaled and poisons Redis.
func (i ResourceType) MarshalBinary() ([]byte, error) {
	return []byte{byte(i)}, nil
}

// UnmarshalBinary decodes a single-byte enum, an empty payload as zero, the
// generated text name, or the enumer fallback "ResourceType(n)".
func (i *ResourceType) UnmarshalBinary(data []byte) error {
	switch len(data) {
	case 0:
		*i = 0
		return nil
	case 1:
		*i = ResourceType(data[0])
		return nil
	}

	if rt, ok := parseResourceTypeEnumerFallback(data); ok {
		*i = rt
		return nil
	}

	return i.UnmarshalText(data)
}

func parseResourceTypeEnumerFallback(data []byte) (ResourceType, bool) {
	s := string(data)
	if !strings.HasPrefix(s, resourceTypeTextPrefix) || !strings.HasSuffix(s, ")") {
		return 0, false
	}

	n, err := strconv.Atoi(s[len(resourceTypeTextPrefix) : len(s)-1])
	if err != nil || n < 0 || n > 255 {
		return 0, false
	}

	return ResourceType(n), true
}
