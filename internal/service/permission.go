package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
)

// PermissionService serves the business logic of interacting with permissions.
type PermissionService interface {
	// Create creates a new permission. If the permission already exists, an
	// additional permission is created.
	Create(ctx context.Context, perm *model.Permission) error
	// CtxUserCreate creates a new permission if the user in the context has the
	// permission to create a new permission. If the permission already exists,
	// an additional permission is created. If the user in the context does not
	// have the permission to create a new permission, an error is returned.
	CtxUserCreate(ctx context.Context, perm *model.Permission) error
	// Get returns the permission with the given ID. If the permission does not
	// exist, an error is returned.
	Get(ctx context.Context, id model.ID) (*model.Permission, error)
	// GetBySubject returns all permissions where the subject is the given ID.
	// If no permissions exist, an error is returned.
	GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error)
	// GetByTarget returns all permissions where the target is the given ID. If
	// no permissions exist, an error is returned.
	GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error)
	// GetBySubjectAndTarget returns all permissions where the subject and the
	// target are both provided. If no permissions exist, an error is returned.
	GetBySubjectAndTarget(ctx context.Context, source, target model.ID) ([]*model.Permission, error)
	// HasAnyRelation checks if the subject has any relation to the target. If
	// the subject does not have any relation to the target, an error is
	// returned.
	HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error)
	// CtxUserHasAnyRelation checks if the user in the context has any relation
	// to the target. If the user does not have any relation to the target, an
	// error is returned.
	CtxUserHasAnyRelation(ctx context.Context, target model.ID) bool
	// HasSystemRole checks if the subject has the system role. If the subject
	// does not have the system role, an error is returned.
	HasSystemRole(ctx context.Context, source model.ID, roles ...model.SystemRole) (bool, error)
	// CtxUserHasSystemRole checks if the user in the context has the system
	// role. If the user does not have the system role, an error is returned.
	CtxUserHasSystemRole(ctx context.Context, roles ...model.SystemRole) bool
	// HasPermission checks if the subject has the permission to perform the
	// action on the target. If the subject does not have the permission, an
	// error is returned.
	HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error)
	// CtxUserHasPermission checks if the user in the context has the permission
	// to perform the action on the target. If the user does not have the
	// permission, an error is returned.
	CtxUserHasPermission(ctx context.Context, target model.ID, permissions ...model.PermissionKind) bool
	// Update updates the permission with the given ID. If the permission does
	// not exist, an error is returned.
	Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error)
	// CtxUserUpdate updates the permission with the given ID if the user in the
	// context has the permission to update the permission. If the permission
	// does not exist, an error is returned.
	CtxUserUpdate(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error)
	// Delete deletes the permission with the given ID. If the permission does
	// not exist, an error is returned.
	Delete(ctx context.Context, id model.ID) error
	// CtxUserDelete deletes the permission with the given ID if the user in the
	// context has the permission to delete the permission. If the permission
	// does not exist, an error is returned.
	CtxUserDelete(ctx context.Context, id model.ID) error
}

// permissionService is the concrete implementation of the PermissionService
// interface.
type permissionService struct {
	*baseService
	permissionRepo repository.PermissionRepository
}

func (s *permissionService) Create(ctx context.Context, perm *model.Permission) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Create")
	defer span.End()

	if err := perm.Validate(); err != nil {
		return err
	}

	if err := s.permissionRepo.Create(ctx, perm); err != nil {
		return errors.Join(ErrPermissionCreate, err)
	}

	return nil
}

func (s *permissionService) CtxUserCreate(ctx context.Context, perm *model.Permission) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserCreate")
	defer span.End()

	hasRelation := s.CtxUserHasAnyRelation(ctx, perm.Target)
	hasSystemRole := s.CtxUserHasSystemRole(ctx, model.SystemRoleOwner, model.SystemRoleAdmin)

	hasCreatePermission := s.CtxUserHasPermission(ctx, perm.Target, model.PermissionKindCreate)
	hasWritePermission := s.CtxUserHasPermission(ctx, perm.Target, model.PermissionKindWrite)
	hasReadPermission := s.CtxUserHasPermission(ctx, perm.Target, model.PermissionKindRead)
	hasDeletePermission := s.CtxUserHasPermission(ctx, perm.Target, model.PermissionKindDelete)
	hasPermission := hasCreatePermission && hasWritePermission && hasReadPermission && hasDeletePermission

	if (hasRelation && hasPermission) || hasSystemRole || hasPermission {
		return s.Create(ctx, perm)
	}

	return errors.Join(ErrPermissionCreate, ErrNoPermission)
}

func (s *permissionService) Get(ctx context.Context, id model.ID) (*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Get")
	defer span.End()

	perm, err := s.permissionRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrPermissionGet, err)
	}

	return perm, nil
}

func (s *permissionService) GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/GetBySubject")
	defer span.End()

	permissions, err := s.permissionRepo.GetBySubject(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrPermissionGetBySubject, err)
	}

	return permissions, nil
}

func (s *permissionService) GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/GetByTarget")
	defer span.End()

	permissions, err := s.permissionRepo.GetByTarget(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrPermissionGetByTarget, err)
	}

	return permissions, nil
}

func (s *permissionService) GetBySubjectAndTarget(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/GetBySubjectAndTarget")
	defer span.End()

	permissions, err := s.permissionRepo.GetBySubjectAndTarget(ctx, source, target)
	if err != nil {
		return nil, errors.Join(ErrPermissionGetBySubjectAndTarget, err)
	}

	return permissions, nil
}

func (s *permissionService) HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/HasAnyRelation")
	defer span.End()

	hasAnyRelation, err := s.permissionRepo.HasAnyRelation(ctx, source, target)
	if err != nil {
		return false, errors.Join(ErrPermissionHasAnyRelation, err)
	}

	return hasAnyRelation, nil
}

func (s *permissionService) CtxUserHasAnyRelation(ctx context.Context, target model.ID) bool {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserHasAnyRelation")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return false
	}

	hasAnyRelation, err := s.HasAnyRelation(ctx, userID, target)
	if err != nil {
		return false
	}

	return hasAnyRelation
}

func (s *permissionService) HasSystemRole(ctx context.Context, source model.ID, roles ...model.SystemRole) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/HasSystemRole")
	defer span.End()

	hasSystemRole, err := s.permissionRepo.HasSystemRole(ctx, source, roles...)
	if err != nil {
		return false, errors.Join(ErrPermissionHasSystemRole, err)
	}

	return hasSystemRole, nil
}

func (s *permissionService) CtxUserHasSystemRole(ctx context.Context, roles ...model.SystemRole) bool {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserHasSystemRole")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return false
	}

	hasPermission, err := s.HasSystemRole(ctx, userID, roles...)
	if err != nil {
		return false
	}

	return hasPermission
}

func (s *permissionService) HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/HasPermission")
	defer span.End()

	hasPermission, err := s.permissionRepo.HasPermission(ctx, subject, target, append(kinds, model.PermissionKindAll)...)
	if err != nil {
		return false, errors.Join(ErrPermissionHasPermission, err)
	}

	return hasPermission, nil
}

func (s *permissionService) CtxUserHasPermission(ctx context.Context, target model.ID, kinds ...model.PermissionKind) bool {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserHasPermission")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return false
	}

	hasPerm, err := s.HasPermission(ctx, userID, target, kinds...)
	if err != nil && !errors.Is(err, repository.ErrPermissionRead) {
		return false
	}

	return hasPerm
}

func (s *permissionService) Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Update")
	defer span.End()

	permission, err := s.permissionRepo.Update(ctx, id, kind)
	if err != nil {
		return nil, errors.Join(ErrPermissionUpdate, err)
	}

	return permission, nil
}

func (s *permissionService) CtxUserUpdate(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserUpdate")
	defer span.End()

	if _, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok {
		return nil, errors.Join(ErrPermissionUpdate, ErrNoUser)
	}

	perm, err := s.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrPermissionUpdate, err)
	}

	hasRelation := s.CtxUserHasAnyRelation(ctx, perm.Target)
	hasSystemRole := s.CtxUserHasSystemRole(ctx, model.SystemRoleOwner, model.SystemRoleAdmin)
	hasPermission := s.CtxUserHasPermission(ctx, perm.Target, model.PermissionKindWrite)

	if (hasRelation && hasPermission) || hasSystemRole || hasPermission {
		return s.Update(ctx, id, kind)
	}

	return nil, errors.Join(ErrPermissionUpdate, ErrNoPermission)
}

func (s *permissionService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Delete")
	defer span.End()

	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrPermissionDelete, err)
	}

	return nil
}

func (s *permissionService) CtxUserDelete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserDelete")
	defer span.End()

	if _, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok {
		return errors.Join(ErrPermissionDelete, ErrNoUser)
	}

	perm, err := s.Get(ctx, id)
	if err != nil {
		return errors.Join(ErrPermissionDelete, err)
	}

	hasRelation := s.CtxUserHasAnyRelation(ctx, perm.Target)
	hasSystemRole := s.CtxUserHasSystemRole(ctx, model.SystemRoleOwner, model.SystemRoleAdmin)
	hasPermission := s.CtxUserHasPermission(ctx, perm.Target, model.PermissionKindDelete)

	if (hasRelation && hasPermission) || hasSystemRole || hasPermission {
		return s.Delete(ctx, id)
	}

	return errors.Join(ErrPermissionDelete, ErrNoPermission)
}

// NewPermissionService creates a new permission service.
func NewPermissionService(permissionRepo repository.PermissionRepository, opts ...Option) (PermissionService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &permissionService{
		baseService:    s,
		permissionRepo: permissionRepo,
	}

	if svc.permissionRepo == nil {
		return nil, ErrNoPermissionRepository
	}

	return svc, nil
}
