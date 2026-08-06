package model

const (
	OrganizationStatusActive  OrganizationStatus = iota + 1 // active organization
	OrganizationStatusDeleted                               // deleted organization
)

// OrganizationStatus represents the status of the organization.
//
//go:generate go tool enumer -type=OrganizationStatus -trimprefix=OrganizationStatus -text -transform=snake -output=organization_status_gen.go
type OrganizationStatus int
