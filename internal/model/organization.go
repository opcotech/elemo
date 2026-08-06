package model

const (
	OrganizationStatusActive  OrganizationStatus = iota + 1 // active organization
	OrganizationStatusDeleted                               // deleted organization
)

var (
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
