package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	AssignmentKindAssignee AssignmentKind = iota + 1 // a user is assigned as an assignee
	AssignmentKindReviewer                           // a user is assigned as a reviewer
)

var (
	assignmentKindKeys = map[string]AssignmentKind{
		"assignee": AssignmentKindAssignee,
		"reviewer": AssignmentKindReviewer,
	}
	assignmentKindValues = map[AssignmentKind]string{
		AssignmentKindAssignee: "assignee",
		AssignmentKindReviewer: "reviewer",
	}
)

// AssignmentKind is the kind of assignment between a user and a resource.
type AssignmentKind uint8

// String returns the string representation of the relation kind.
func (k AssignmentKind) String() string {
	return assignmentKindValues[k]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (k AssignmentKind) MarshalText() (text []byte, err error) {
	if k < 1 || k > 2 {
		return nil, ErrInvalidAssignmentKind
	}
	return []byte(k.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (k *AssignmentKind) UnmarshalText(text []byte) error {
	if v, ok := assignmentKindKeys[string(text)]; ok {
		*k = v
		return nil
	}
	return ErrInvalidAssignmentKind
}

// Assignment is the model of an assignment between a user and a resource.
type Assignment struct {
	ID        ID             `json:"id" validate:"required"`
	Kind      AssignmentKind `json:"kind" validate:"required,min=1,max=2"`
	User      ID             `json:"user_id" validate:"required"`
	Resource  ID             `json:"resource_id" validate:"required"`
	CreatedAt *time.Time     `json:"created_at" validate:"omitempty"`
}

// Validate validates the assignment details.
func (a *Assignment) Validate() error {
	if err := validate.Struct(a); err != nil {
		return errors.Join(ErrInvalidAssignmentDetails, err)
	}
	if err := a.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidAssignmentDetails, err)
	}
	if err := a.User.Validate(); err != nil {
		return errors.Join(ErrInvalidAssignmentDetails, err)
	}
	if err := a.Resource.Validate(); err != nil {
		return errors.Join(ErrInvalidAssignmentDetails, err)
	}
	return nil
}

// NewAssignment creates a new assignment.
func NewAssignment(user, resource ID, kind AssignmentKind) (*Assignment, error) {
	assignment := &Assignment{
		ID:       MustNewNilID(ResourceTypeAssignment),
		Kind:     kind,
		User:     user,
		Resource: resource,
	}

	if err := assignment.Validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}
