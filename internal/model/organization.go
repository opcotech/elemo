package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	OrganizationIDType = "Organization"
)

const (
	OrganizationStatusActive  OrganizationStatus = iota + 1 // active organization
	OrganizationStatusDeleted                               // deleted organization
)

var (
	ErrInvalidOrganizationDetails = errors.New("invalid organization details") // the organization details are invalid
	ErrInvalidOrganizationStatus  = errors.New("invalid organization status")  // the organization status is invalid

	organizationStatusKeys = map[string]OrganizationStatus{
		"active":  OrganizationStatusActive,
		"deleted": OrganizationStatusDeleted,
	}
	organizationStatusValues = map[OrganizationStatus]string{
		OrganizationStatusActive:  "active",
		OrganizationStatusDeleted: "deleted",
	}
)

// OrganizationStatus represents the status of the organization.
type OrganizationStatus int

// String returns the string representation of the organization status.
func (s OrganizationStatus) String() string {
	return organizationStatusValues[s]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s OrganizationStatus) MarshalText() (text []byte, err error) {
	if s < 1 || s > 2 {
		return nil, ErrInvalidOrganizationStatus
	}
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *OrganizationStatus) UnmarshalText(text []byte) error {
	if v, ok := organizationStatusKeys[string(text)]; ok {
		*s = v
		return nil
	}
	return ErrInvalidOrganizationStatus
}

// Organization represents an organization.
type Organization struct {
	ID         ID                 `json:"id" validate:"required,dive"`
	Name       string             `json:"name" validate:"required,min=1,max=120"`
	Email      string             `json:"email" validate:"required,email"`
	Logo       string             `json:"logo" validate:"omitempty,url"`
	Website    string             `json:"website" validate:"omitempty,url"`
	Status     OrganizationStatus `json:"status" validate:"required,min=1,max=2"`
	Namespaces []ID               `json:"namespaces" validate:"omitempty,dive"`
	Teams      []ID               `json:"teams" validate:"omitempty,dive"`
	Members    []ID               `json:"members" validate:"omitempty,dive"`
	CreatedAt  *time.Time         `json:"created_at" validate:"omitempty"`
	UpdatedAt  *time.Time         `json:"updated_at" validate:"omitempty"`
}

func (o *Organization) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(ErrInvalidOrganizationDetails, err)
	}
	if err := o.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidOrganizationDetails, err)
	}
	for _, id := range o.Namespaces {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidOrganizationDetails, err)
		}
	}
	for _, id := range o.Members {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidOrganizationDetails, err)
		}
	}
	for _, id := range o.Teams {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidOrganizationDetails, err)
		}
	}
	return nil
}

// NewOrganization creates a new Organization.
func NewOrganization(name, email string) (*Organization, error) {
	org := &Organization{
		ID:         MustNewNilID(OrganizationIDType),
		Name:       name,
		Email:      email,
		Status:     OrganizationStatusActive,
		Namespaces: make([]ID, 0),
		Teams:      make([]ID, 0),
		Members:    make([]ID, 0),
	}

	if err := org.Validate(); err != nil {
		return nil, err
	}

	return org, nil
}
