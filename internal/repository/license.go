package repository

import (
	"context"
)

// LicenseRepository is the repository for retrieving license information.
type LicenseRepository interface {
	ActiveUserCount(ctx context.Context) (int, error)
	ActiveOrganizationCount(ctx context.Context) (int, error)
	DocumentCount(ctx context.Context) (int, error)
	NamespaceCount(ctx context.Context) (int, error)
	ProjectCount(ctx context.Context) (int, error)
	RoleCount(ctx context.Context) (int, error)
}
