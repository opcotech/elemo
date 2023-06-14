package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	SystemRoleOwner   SystemRole = iota + 1 // the owner them of the instance
	SystemRoleAdmin                         // the administrator team of the instance
	SystemRoleSupport                       // the support team of the instance
)

var (
	systemRoleValues = map[SystemRole]string{
		SystemRoleOwner:   "Owner",
		SystemRoleAdmin:   "Admin",
		SystemRoleSupport: "Support",
	}
	systemRoleKeys = map[string]SystemRole{
		"Owner":   SystemRoleOwner,
		"Admin":   SystemRoleAdmin,
		"Support": SystemRoleSupport,
	}
)

// SystemRole is a special role that is created by the system.
type SystemRole uint8

// String returns the string representation of the SystemRole.
func (r SystemRole) String() string {
	return systemRoleValues[r]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (r SystemRole) MarshalText() (text []byte, err error) {
	if r < 1 || r > 3 {
		return nil, ErrInvalidSystemRole
	}
	return []byte(r.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (r *SystemRole) UnmarshalText(text []byte) error {
	if v, ok := systemRoleKeys[string(text)]; ok {
		*r = v
		return nil
	}
	return ErrInvalidSystemRole
}

// Role is a group of users. However, permissions are attached to roles
// separately to avoid infinitely nested permissions.
type Role struct {
	ID          ID         `json:"id" validate:"required,dive"`
	Name        string     `json:"name" validate:"required,min=3,max=120"`
	Description string     `json:"description" validate:"omitempty,min=5,max=500"`
	Members     []ID       `json:"members" validate:"omitempty,dive"`
	Permissions []ID       `json:"permissions" validate:"omitempty,dive"`
	CreatedAt   *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time `json:"updated_at" validate:"omitempty"`
}

func (r *Role) Validate() error {
	if err := validate.Struct(r); err != nil {
		return errors.Join(ErrInvalidRoleDetails, err)
	}
	if err := r.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidRoleDetails, err)
	}
	for _, member := range r.Members {
		if err := member.Validate(); err != nil {
			return errors.Join(ErrInvalidRoleDetails, err)
		}
	}
	for _, permission := range r.Permissions {
		if err := permission.Validate(); err != nil {
			return errors.Join(ErrInvalidRoleDetails, err)
		}
	}
	return nil
}

// NewRole creates a new Role.
func NewRole(name string) (*Role, error) {
	role := &Role{
		ID:          MustNewNilID(ResourceTypeRole),
		Name:        name,
		Members:     make([]ID, 0),
		Permissions: make([]ID, 0),
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	return role, nil
}
