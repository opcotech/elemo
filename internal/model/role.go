package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	RoleIDType = "Role"
)

var (
	ErrInvalidRoleDetails = errors.New("invalid role details") // the role details are invalid
)

// Role is a group of users. However, permissions are attached to roles
// separately to avoid infinitely nested permissions.
type Role struct {
	ID          ID         `validate:"required,dive"`
	Name        string     `validate:"required,min=3,max=120"`
	Description string     `validate:"omitempty,min=5,max=500"`
	Members     []ID       `validate:"omitempty,dive"`
	Permissions []ID       `validate:"omitempty,dive"`
	CreatedAt   *time.Time `validate:"omitempty"`
	UpdatedAt   *time.Time `validate:"omitempty"`
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
		ID:          MustNewNilID(RoleIDType),
		Name:        name,
		Members:     make([]ID, 0),
		Permissions: make([]ID, 0),
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	return role, nil
}
