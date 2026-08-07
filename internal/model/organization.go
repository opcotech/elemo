package model

const (
	OrganizationStatusActive  OrganizationStatus = iota + 1 // active
	OrganizationStatusDeleted                               // deleted
)

// OrganizationStatus represents the status of the organization.
//
//go:generate go tool enumer -type=OrganizationStatus -text -transform=noop -linecomment -output=organization_status_gen.go
type OrganizationStatus int
