package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	PermissionKindAll    PermissionKind = iota + 1 // permission to do everything on a resource
	PermissionKindCreate                           // permission to create a resource
	PermissionKindRead                             // permission to read a resource
	PermissionKindWrite                            // permission to write a resource
	PermissionKindDelete                           // permission to delete a resource
)

var (
	permissionKindKeys = map[string]PermissionKind{
		"*":      PermissionKindAll,
		"create": PermissionKindCreate,
		"read":   PermissionKindRead,
		"write":  PermissionKindWrite,
		"delete": PermissionKindDelete,
	}
	permissionKindValues = map[PermissionKind]string{
		PermissionKindAll:    "*",
		PermissionKindCreate: "create",
		PermissionKindRead:   "read",
		PermissionKindWrite:  "write",
		PermissionKindDelete: "delete",
	}
)

// PermissionKind represents a permission attached to a relation.
type PermissionKind uint8

// String returns the string representation of the permission.
func (p PermissionKind) String() string {
	return permissionKindValues[p]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (p PermissionKind) MarshalText() (text []byte, err error) {
	if p < 1 || p > 5 {
		return nil, ErrInvalidPermissionKind
	}
	return []byte(p.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *PermissionKind) UnmarshalText(text []byte) error {
	if v, ok := permissionKindKeys[string(text)]; ok {
		*p = v
		return nil
	}
	return ErrInvalidPermissionKind
}

// Permission represents a permission attached to a relation. The permission
// defines the kind of access a subject has on a target.
type Permission struct {
	ID        ID             `json:"id" validate:"required"`
	Kind      PermissionKind `json:"kind" validate:"required,min=1,max=5"`
	Subject   ID             `json:"subject" validate:"required"`
	Target    ID             `json:"target" validate:"required"`
	CreatedAt *time.Time     `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time     `json:"updated_at" validate:"omitempty"`
}

// Validate validates the permission details.
func (p *Permission) Validate() error {
	if err := validate.Struct(p); err != nil {
		return errors.Join(ErrInvalidPermissionDetails, err)
	}
	// Allow roles to have permissions on themselves, but reject for all other resource types
	if p.Subject.Inner == p.Target.Inner && (p.Subject.Type != ResourceTypeRole || p.Target.Type != ResourceTypeRole) {
		return errors.Join(ErrInvalidPermissionDetails, ErrPermissionSubjectTargetEqual)
	}
	if err := p.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidPermissionDetails, err)
	}
	if err := p.Subject.Validate(); err != nil {
		return errors.Join(ErrInvalidPermissionDetails, err)
	}
	if err := p.Target.Validate(); err != nil {
		return errors.Join(ErrInvalidPermissionDetails, err)
	}
	return nil
}

// NewPermission creates a new permission.
func NewPermission(subject, target ID, kind PermissionKind) (*Permission, error) {
	permission := &Permission{
		ID:      MustNewNilID(ResourceTypePermission),
		Kind:    kind,
		Subject: subject,
		Target:  target,
	}

	if err := permission.Validate(); err != nil {
		return nil, err
	}

	return permission, nil
}
