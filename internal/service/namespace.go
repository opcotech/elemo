package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
)

// NamespaceService serves the business logic of interacting with namespaces.
type NamespaceService interface {
	// Create creates a new namespace in an organization. If the organization
	// does not exist, an error is returned.
	Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error
	// Get returns a namespace by its ID. If the namespace does not exist, an
	// error is returned.
	Get(ctx context.Context, id model.ID) (*model.Namespace, error)
	// GetAll returns all namespaces for an organization. The offset and limit
	// parameters are used to paginate the results. If the offset is greater
	// than the number of namespaces in the organization, an empty slice is
	// returned.
	GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error)
	// Update updates a namespace. If the namespace does not exist, an error
	// is returned.
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error)
	// Delete deletes a namespace. If the namespace does not exist, an error
	// is returned.
	Delete(ctx context.Context, id model.ID) error
}

// namespaceService is the concrete implementation of NamespaceService.
type namespaceService struct {
	*baseService
}

func (s *namespaceService) Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrNamespaceCreate, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrNamespaceCreate, err)
	}

	if err := namespace.Validate(); err != nil {
		return errors.Join(ErrNamespaceCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite) {
		return errors.Join(ErrNamespaceCreate, ErrNoPermission)
	}

	if err := s.namespaceRepo.Create(ctx, orgID, namespace); err != nil {
		return errors.Join(ErrNamespaceCreate, err)
	}

	return nil
}

func (s *namespaceService) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, errors.Join(ErrNamespaceGet, ErrNoPermission)
	}

	namespace, err := s.namespaceRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	return namespace, nil
}

func (s *namespaceService) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/GetAll")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceGetAll, err)
	}

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrNamespaceGetAll, ErrInvalidPaginationParams)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindRead) {
		return nil, errors.Join(ErrNamespaceGetAll, ErrNoPermission)
	}

	namespaces, err := s.namespaceRepo.GetAll(ctx, orgID, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrNamespaceGetAll, err)
	}

	return namespaces, nil
}

func (s *namespaceService) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrNamespaceUpdate, ErrNoPermission)
	}

	namespace, err := s.namespaceRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	return namespace, nil
}

func (s *namespaceService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrNamespaceDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrNamespaceDelete, ErrNoPermission)
	}

	if err := s.namespaceRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	return nil
}

// NewNamespaceService returns a new instance of the NamespaceService interface.
func NewNamespaceService(opts ...Option) (NamespaceService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &namespaceService{
		baseService: s,
	}

	if svc.namespaceRepo == nil {
		return nil, ErrNoNamespaceRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}
