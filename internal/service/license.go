package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
)

// LicenseRepository defines the interface for retrieving quota information.
type LicenseRepository interface {
	// ActiveUserCount returns the number of active users.
	ActiveUserCount(ctx context.Context) (int, error)
	// ActiveOrganizationCount returns the number of active organizations.
	ActiveOrganizationCount(ctx context.Context) (int, error)
	// DocumentCount returns the number of documents.
	DocumentCount(ctx context.Context) (int, error)
	// NamespaceCount returns the number of namespaces.
	NamespaceCount(ctx context.Context) (int, error)
	// ProjectCount returns the number of projects.
	ProjectCount(ctx context.Context) (int, error)
	// RoleCount returns the number of roles.
	RoleCount(ctx context.Context) (int, error)
}

// LicenseService serves the business logic of retrieving license information.
type LicenseService interface {
	// Expired returns true if the license has expired.
	Expired(ctx context.Context) (bool, error)
	// HasFeature returns true if the license has the specified feature.
	HasFeature(ctx context.Context, feature license.Feature) (bool, error)
	// WithinThreshold returns true if the resource usage is within the quota.
	WithinThreshold(ctx context.Context, name license.Quota) (bool, error)
	// GetLicense returns the license.
	GetLicense(ctx context.Context) (license.License, error)
	// Ping implements the Pingable interface to check the license validity.
	Ping(ctx context.Context) error
}

// licenseService is the concrete implementation of LicenseService.
type licenseService struct {
	*baseService
	licenseRepo LicenseRepository
	license     *license.License
}

func (s *licenseService) Expired(ctx context.Context) (bool, error) {
	_, span := s.tracer.Start(ctx, "service.licenseService/Expired")
	defer span.End()

	return s.license.Expired(), nil
}

func (s *licenseService) HasFeature(ctx context.Context, feature license.Feature) (bool, error) {
	_, span := s.tracer.Start(ctx, "service.licenseService/HasFeature")
	defer span.End()

	return s.license.HasFeature(feature), nil
}

func (s *licenseService) WithinThreshold(ctx context.Context, quota license.Quota) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "service.licenseService/WithinThreshold")
	defer span.End()

	var count int
	var err error

	switch quota {
	case license.QuotaDocuments:
		count, err = s.licenseRepo.DocumentCount(ctx)
	case license.QuotaNamespaces:
		count, err = s.licenseRepo.NamespaceCount(ctx)
	case license.QuotaOrganizations:
		count, err = s.licenseRepo.ActiveOrganizationCount(ctx)
	case license.QuotaProjects:
		count, err = s.licenseRepo.ProjectCount(ctx)
	case license.QuotaRoles:
		count, err = s.licenseRepo.RoleCount(ctx)
	case license.QuotaUsers:
		count, err = s.licenseRepo.ActiveUserCount(ctx)
	default:
		err = ErrQuotaInvalid
	}

	if err != nil {
		return false, errors.Join(ErrQuotaUsageGet, err)
	}

	return s.license.WithinThreshold(quota, count), nil
}

func (s *licenseService) GetLicense(ctx context.Context) (license.License, error) {
	ctx, span := s.tracer.Start(ctx, "service.licenseService/GetLicense")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return license.License{}, ErrNoUser
	}

	hasRole, err := s.permissionRepo.HasSystemRole(ctx, userID, model.SystemRoleOwner, model.SystemRoleAdmin, model.SystemRoleSupport)
	if !hasRole || err != nil {
		return license.License{}, ErrNoPermission
	}

	return *s.license, nil
}

func (s *licenseService) Ping(ctx context.Context) error {
	_, span := s.tracer.Start(ctx, "service.licenseService/Ping")
	defer span.End()

	if expired, err := s.Expired(ctx); expired || err != nil {
		return license.ErrLicenseInvalid
	}

	return nil
}

// NewLicenseService returns a new LicenseService.
func NewLicenseService(l *license.License, repo LicenseRepository, opts ...Option) (LicenseService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &licenseService{
		baseService: s,
		license:     l,
		licenseRepo: repo,
	}

	if svc.license == nil {
		return nil, license.ErrNoLicense
	}

	if svc.licenseRepo == nil {
		return nil, repository.ErrNoLicenseRepository
	}

	if svc.permissionRepo == nil {
		return nil, ErrNoPermissionRepository
	}

	return svc, nil
}
