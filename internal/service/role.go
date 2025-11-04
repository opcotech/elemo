package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
)

// RoleService is the interface that provides methods for managing roles.
type RoleService interface {
	// Create creates a new role in the system and assigns it to a resource it
	// belongs to. The user who created the role is also assigned as a member
	// of the role. If the role already exists, an error is returned.
	Create(ctx context.Context, owner, belongsTo model.ID, role *model.Role) error
	// Get returns a role by its ID. If the role does not exist, an error is
	// returned.
	Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error)
	// GetAllBelongsTo returns all roles that belong to a resource. The offset
	// and limit parameters are used to paginate the results. If the offset is
	// greater than the number of roles in the system, an empty slice is
	// returned.
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error)
	// Update updates a role in the system. If the role does not exist, an
	// error is returned.
	Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error)
	// GetMembers returns all members of a role that belongs to a resource. If
	// the resource does not exist, an error is returned.
	GetMembers(ctx context.Context, id, belongsTo model.ID) ([]*model.User, error)
	// AddMember adds a member to a role. If the member is already a member of
	// the role, an error is returned.
	AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	// RemoveMember removes a member from a role. If the member is not a member
	// of the role, an error is returned.
	RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	// Delete deletes a role from the system. This method does not actually
	// delete the role from the database to preserve the role's history and
	// relations unless the force parameter is set to true.
	Delete(ctx context.Context, id, belongsTo model.ID) error
	// AddPermission adds a permission to a role. The target must be an
	// organization-scoped resource. The caller must have write permission on the
	// organization.
	AddPermission(ctx context.Context, roleID, belongsToID, targetID model.ID, kind model.PermissionKind) error
	// RemovePermission removes a permission from a role. The permission must
	// belong to the role. The caller must have write permission on the
	// organization.
	RemovePermission(ctx context.Context, roleID, belongsToID, permissionID model.ID) error
	// GetPermissions returns all permissions assigned to a role.
	GetPermissions(ctx context.Context, roleID, belongsToID model.ID) ([]*model.Permission, error)
}

// roleService implements RoleService interface.
type roleService struct {
	*baseService
}

func (s *roleService) Create(ctx context.Context, owner, belongsTo model.ID, role *model.Role) error {
	ctx, span := s.tracer.Start(ctx, "service.roleService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrRoleCreate, license.ErrLicenseExpired)
	}

	if err := role.Validate(); err != nil {
		return errors.Join(ErrRoleCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite) {
		return errors.Join(ErrRoleCreate, ErrNoPermission)
	}

	if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaRoles); !ok || err != nil {
		return errors.Join(ErrRoleCreate, ErrQuotaExceeded)
	}

	if err := s.roleRepo.Create(ctx, owner, belongsTo, role); err != nil {
		return errors.Join(ErrRoleCreate, err)
	}

	return nil
}

func (s *roleService) Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGet, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGet, ErrNoPermission)
	}

	role, err := s.roleRepo.Get(ctx, id, belongsTo)
	if err != nil {
		return nil, errors.Join(ErrRoleGet, err)
	}

	return role, nil
}

func (s *roleService) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/GetAllBelongsTo")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGetBelongsTo, err)
	}

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrRoleGetBelongsTo, ErrInvalidPaginationParams)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGetBelongsTo, ErrNoPermission)
	}

	roles, err := s.roleRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrRoleGetBelongsTo, err)
	}

	return roles, nil
}

func (s *roleService) Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrRoleUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrRoleUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrRoleUpdate, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite) {
		return nil, errors.Join(ErrRoleUpdate, ErrNoPermission)
	}

	role, err := s.roleRepo.Update(ctx, id, belongsTo, patch)
	if err != nil {
		return nil, errors.Join(ErrRoleUpdate, err)
	}

	return role, nil
}

func (s *roleService) GetMembers(ctx context.Context, id, belongsTo model.ID) ([]*model.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/AddMember")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGetBelongsTo, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGetBelongsTo, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGetBelongsTo, ErrNoPermission)
	}

	role, err := s.roleRepo.Get(ctx, id, belongsTo)
	if err != nil {
		return nil, errors.Join(ErrOrganizationMembersGet, err)
	}

	members := make([]*model.User, 0, len(role.Members))
	for _, member := range role.Members {
		user, err := s.userRepo.Get(ctx, member)
		if err != nil {
			return nil, errors.Join(ErrOrganizationMembersGet, err)
		}
		members = append(members, user)
	}

	return members, nil
}

func (s *roleService) AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.roleService/AddMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrRoleAddMember, license.ErrLicenseExpired)
	}

	if err := roleID.Validate(); err != nil {
		return errors.Join(ErrRoleAddMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrRoleAddMember, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, roleID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleAddMember, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleAddMember, ErrNoPermission)
	}

	err := s.roleRepo.AddMember(ctx, roleID, memberID, belongsToID)
	if err != nil {
		return errors.Join(ErrRoleAddMember, err)
	}

	return nil
}

func (s *roleService) RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.roleService/RemoveMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrRoleRemoveMember, license.ErrLicenseExpired)
	}

	if err := roleID.Validate(); err != nil {
		return errors.Join(ErrRoleRemoveMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrRoleRemoveMember, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, roleID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleRemoveMember, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleAddMember, ErrNoPermission)
	}

	err := s.roleRepo.RemoveMember(ctx, roleID, memberID, belongsToID)
	if err != nil {
		return errors.Join(ErrRoleRemoveMember, err)
	}

	return nil
}

func (s *roleService) Delete(ctx context.Context, id, belongsTo model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.roleService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrRoleDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrRoleDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrRoleDelete, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite) {
		return errors.Join(ErrRoleDelete, ErrNoPermission)
	}

	err := s.roleRepo.Delete(ctx, id, belongsTo)
	if err != nil {
		return errors.Join(ErrRoleDelete, err)
	}

	return nil
}

func (s *roleService) AddPermission(ctx context.Context, roleID, belongsToID, targetID model.ID, kind model.PermissionKind) error {
	ctx, span := s.tracer.Start(ctx, "service.roleService/AddPermission")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrRoleAddPermission, license.ErrLicenseExpired)
	}

	if err := roleID.Validate(); err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	if err := belongsToID.Validate(); err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	if err := targetID.Validate(); err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	// Verify the role belongs to the organization
	if _, err := s.roleRepo.Get(ctx, roleID, belongsToID); err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	// Check write permission on organization
	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleAddPermission, ErrNoPermission)
	}

	// Create permission with role as subject and target as target
	perm, err := model.NewPermission(roleID, targetID, kind)
	if err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	if err := s.permissionService.Create(ctx, perm); err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	return nil
}

func (s *roleService) RemovePermission(ctx context.Context, roleID, belongsToID, permissionID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.roleService/RemovePermission")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrRoleRemovePermission, license.ErrLicenseExpired)
	}

	if err := roleID.Validate(); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	if err := belongsToID.Validate(); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	if err := permissionID.Validate(); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	// Verify the role belongs to the organization
	if _, err := s.roleRepo.Get(ctx, roleID, belongsToID); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	// Check write permission on organization
	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleRemovePermission, ErrNoPermission)
	}

	// Get the permission to verify it belongs to the role
	perm, err := s.permissionService.Get(ctx, permissionID)
	if err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	// Verify the permission's subject is the role
	if perm.Subject.String() != roleID.String() {
		return errors.Join(ErrRoleRemovePermission, ErrNoPermission)
	}

	// Delete the permission
	if err := s.permissionService.Delete(ctx, permissionID); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	return nil
}

func (s *roleService) GetPermissions(ctx context.Context, roleID, belongsToID model.ID) ([]*model.Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/GetPermissions")
	defer span.End()

	if err := roleID.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	if err := belongsToID.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	// Verify the role belongs to the organization
	if _, err := s.roleRepo.Get(ctx, roleID, belongsToID); err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	// Check read permission on organization
	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGetPermissions, ErrNoPermission)
	}

	// Get all permissions where role is subject
	permissions, err := s.permissionService.GetBySubject(ctx, roleID)
	if err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	return permissions, nil
}

// NewRoleService creates a new RoleService that provides methods
// for managing roles.
func NewRoleService(opts ...Option) (RoleService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &roleService{
		baseService: s,
	}

	if svc.roleRepo == nil {
		return nil, ErrNoRoleRepository
	}

	if svc.userRepo == nil {
		return nil, ErrNoUserRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}
