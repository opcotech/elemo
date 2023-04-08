package model

import (
	"errors"

	"github.com/rs/xid"
)

var (
	ErrInvalidAssignmentKind       = errors.New("invalid assigned to kind")       // the assigned to kind is invalid
	ErrInvalidAssignmentDetails    = errors.New("invalid assignment details")     // the assignment details are invalid
	ErrInvalidAttachmentDetails    = errors.New("invalid attachment details")     // the attachment details are invalid
	ErrInvalidCommentDetails       = errors.New("invalid comment details")        // the comment details are invalid
	ErrInvalidDocumentDetails      = errors.New("invalid document details")       // the document details are invalid
	ErrInvalidIssueKind            = errors.New("invalid issue kind")             // the issue kind is invalid
	ErrInvalidIssueRelationDetails = errors.New("invalid issue relation details") // the issue relation details are invalid
	ErrInvalidIssueStatus          = errors.New("invalid issue status")           // the issue status is invalid
	ErrInvalidIssueResolution      = errors.New("invalid issue resolution")       // the issue resolution is invalid
	ErrInvalidIssueRelationKind    = errors.New("invalid issue relation kind")    // the issue relation kind is invalid
	ErrInvalidIssuePriority        = errors.New("invalid issue priority")         // the issue priority is invalid
	ErrInvalidIssueDetails         = errors.New("invalid issue details")          // the issue details are invalid
	ErrInvalidLabelDetails         = errors.New("invalid label details")          // the label details are invalid
	ErrInvalidLanguage             = errors.New("invalid language code")          // Language is not valid
	ErrInvalidID                   = errors.New("invalid id")                     // the id is invalid
	ErrInvalidNamespaceDetails     = errors.New("invalid namespace details")      // the namespace details are invalid
	ErrInvalidOrganizationDetails  = errors.New("invalid organization details")   // the organization details are invalid
	ErrInvalidOrganizationStatus   = errors.New("invalid organization status")    // the organization status is invalid
	ErrInvalidPermissionDetails    = errors.New("invalid permission details")     // the permission details are invalid
	ErrInvalidPermissionKind       = errors.New("invalid permission kind")        // the permission kind is invalid
	ErrInvalidProjectDetails       = errors.New("invalid project details")        // the project details are invalid
	ErrInvalidProjectStatus        = errors.New("invalid project status")         // the project status is invalid
	ErrInvalidRoleDetails          = errors.New("invalid role details")           // the role details are invalid
	ErrInvalidHealthStatus         = errors.New("invalid health status")          // health status is invalid
	ErrInvalidTodoPriority         = errors.New("invalid todo priority")          // the todo priority is invalid
	ErrInvalidTodoDetails          = errors.New("invalid todo details")           // the todo details are invalid
	ErrInvalidUserDetails          = errors.New("invalid user details")           // the user details are invalid
	ErrInvalidUserStatus           = errors.New("invalid user status")            // the user status is invalid
)

// ID represents a unique identifier for a resource, combining a resource label
// and a unique identifier.
type ID struct {
	inner xid.ID
	label string `validate:"required,min=4,max=32"`
}

func (id ID) Validate() error {
	if len(id.label) < 4 || len(id.label) > 32 {
		return ErrInvalidID
	}
	return nil
}

// String returns the string representation of the ID. The type is not part of
// the string representation. This is to allow for the ID to be used as a
// label or flag in a database or log aggregation system.
func (id ID) String() string {
	return id.inner.String()
}

// Label returns the label of the ID.
func (id ID) Label() string {
	return id.label
}

// IsNil returns true if the ID is nil.
func (id ID) IsNil() bool {
	return id.inner == xid.NilID()
}

// NewID creates a new ID.
func NewID(typ string) (ID, error) {
	id := ID{inner: xid.New(), label: typ}

	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

// MustNewID creates a new ID. It panics if the type is invalid.
func MustNewID(typ string) ID {
	id, err := NewID(typ)
	if err != nil {
		panic(err)
	}

	return id
}

// NewNilID creates a new ID with a nil xid.ID.
func NewNilID(typ string) (ID, error) {
	id := ID{inner: xid.NilID(), label: typ}

	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

// MustNewNilID creates a new ID with a nil xid.ID. It panics if the type is
// invalid.
func MustNewNilID(typ string) ID {
	id, err := NewNilID(typ)
	if err != nil {
		panic(err)
	}

	return id
}

// NewIDFromString creates a new ID from a string. The string must be a valid
// xid string.
func NewIDFromString(id, typ string) (ID, error) {
	newID, err := NewNilID(typ)
	if err != nil {
		return ID{}, err
	}

	parsed, err := xid.FromString(id)
	if err != nil {
		return ID{}, errors.Join(ErrInvalidID, err)
	}

	newID.inner = parsed
	return newID, nil
}
