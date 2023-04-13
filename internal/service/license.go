package service

import (
	"context"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/repository"
)

// LicenseRepository defines the interface for retrieving quota information.
type LicenseRepository interface {
	// ActiveUserCount returns the number of active users.
	ActiveUserCount(ctx context.Context) (uint, error)
	// ActiveOrganizationCount returns the number of active organizations.
	ActiveOrganizationCount(ctx context.Context) (uint, error)
	// DocumentCount returns the number of documents.
	DocumentCount(ctx context.Context) (uint, error)
	// NamespaceCount returns the number of namespaces.
	NamespaceCount(ctx context.Context) (uint, error)
	// ProjectCount returns the number of projects.
	ProjectCount(ctx context.Context) (uint, error)
	// RoleCount returns the number of roles.
	RoleCount(ctx context.Context) (uint, error)
}

// LicenseService serves the business logic of retrieving license information.
type LicenseService interface {
	// Expired returns true if the license has expired.
	Expired(ctx context.Context) bool
	// HasFeature returns true if the license has the specified feature.
	HasFeature(ctx context.Context, feature license.Feature) bool
	// WithinQuota returns true if the resource usage is within the quota.
	WithinQuota(ctx context.Context, name license.Quota, current int) bool
}

// licenseService is the concrete implementation of LicenseService.
type licenseService struct {
	*baseService
	licenseRepo LicenseRepository
	license     *license.License
}

func (s *licenseService) Expired(ctx context.Context) bool {
	_, span := s.tracer.Start(ctx, "service.licenseService/Expired")
	defer span.End()

	return s.license.Expired()
}

func (s *licenseService) HasFeature(ctx context.Context, feature license.Feature) bool {
	_, span := s.tracer.Start(ctx, "service.licenseService/HasFeature")
	defer span.End()

	return s.license.HasFeature(feature)
}

func (s *licenseService) WithinQuota(ctx context.Context, quota license.Quota, current int) bool {
	_, span := s.tracer.Start(ctx, "service.licenseService/WithinQuota")
	defer span.End()

	return s.license.WithinThreshold(quota, current)
}

/*
TODO: The license service should have a license repository that returns the
	current count of resources used as desired from a given resource type.
	For example, the license service should be able to return the number of
	projects currently in the system.
*/

// NewLicenseService returns a new LicenseService.
func NewLicenseService(l *license.License, repo LicenseRepository, opts ...Option) (LicenseService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &licenseService{
		baseService: s,
		license:     l,
	}

	if svc.license == nil {
		return nil, license.ErrNoLicense
	}

	if repo == nil {
		return nil, repository.ErrNoLicenseRepository
	}

	return svc, nil
}
