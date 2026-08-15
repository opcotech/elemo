package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// Role represents a role returned by the service.
type Role struct {
	ID          model.ID
	Name        string
	Description string
	MemberCount *int64
	Permissions []model.ID
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// CreateRoleOpts holds the data required to create a role.
type CreateRoleOpts struct {
	Name        string `json:"name" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"omitempty,min=5,max=500"`
}

// Validate validates the create options.
func (o *CreateRoleOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidRoleDetails, err)
	}
	return nil
}

// UpdateRoleOpts holds the fields that can be updated on a role.
// Undefined fields (Defined == false) are left unchanged.
type UpdateRoleOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
}

// RoleService is the interface that provides methods for managing roles.
type RoleService interface {
	// Create creates a new role in the system and assigns it to a resource it
	// belongs to. The user who created the role is also assigned as a member
	// of the role. If the role already exists, an error is returned.
	Create(ctx context.Context, owner, belongsTo model.ID, opts CreateRoleOpts) (*Role, error)
	// Get returns a role by its ID. If the role does not exist, an error is
	// returned.
	Get(ctx context.Context, id, belongsTo model.ID) (*Role, error)
	// ListBelongsTo returns a cursor-paginated page of roles for a resource.
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Role], error)
	// Update updates a role in the system. If the role does not exist, an
	// error is returned.
	Update(ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts) (*Role, error)
	// ListMembers returns a cursor-paginated page of members of a role that
	// belongs to a resource. If the resource does not exist, an error is
	// returned.
	ListMembers(ctx context.Context, id, belongsTo model.ID, page CursorPage) (Page[*User], error)
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
	GetPermissions(ctx context.Context, roleID, belongsToID model.ID) ([]*Permission, error)
}

// roleService implements RoleService interface.
type roleService struct {
	*baseService
}

func roleFromRepository(r *repository.Role) *Role {
	if r == nil {
		return nil
	}
	return &Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		MemberCount: r.MemberCount,
		Permissions: r.Permissions,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func rolesFromRepository(roles []*repository.Role) []*Role {
	out := make([]*Role, len(roles))
	for i, r := range roles {
		out[i] = roleFromRepository(r)
	}
	return out
}

func (s *roleService) Create(ctx context.Context, owner, belongsTo model.ID, opts CreateRoleOpts) (*Role, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrRoleCreate, license.ErrLicenseExpired)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrRoleCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindWrite) {
		return nil, errors.Join(ErrRoleCreate, ErrNoPermission)
	}

	if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaRoles); !ok || err != nil {
		return nil, errors.Join(ErrRoleCreate, ErrQuotaExceeded)
	}

	role, err := s.roleRepo.Create(ctx, repository.CreateRoleOpts{
		Name:        opts.Name,
		Description: opts.Description,
		CreatedBy:   owner,
		BelongsTo:   belongsTo,
	})
	if err != nil {
		return nil, errors.Join(ErrRoleCreate, err)
	}

	return roleFromRepository(role), nil
}

func (s *roleService) Get(ctx context.Context, id, belongsTo model.ID) (*Role, error) {
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

	role, err := s.roleRepo.Get(ctx, id, belongsTo, repository.RoleDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrRoleGet, err)
	}

	return roleFromRepository(role), nil
}

func (s *roleService) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Role], error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/ListBelongsTo")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleGetBelongsTo, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleGetBelongsTo, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead) {
		return Page[*Role]{}, errors.Join(ErrRoleGetBelongsTo, ErrNoPermission)
	}

	roles, err := s.roleRepo.ListBelongsTo(
		ctx,
		belongsTo,
		normalized,
		repository.RoleListProjection(),
	)
	if err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleGetBelongsTo, err)
	}

	return mapPage(roles, roleFromRepository), nil
}

func (s *roleService) Update(ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts) (*Role, error) {
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

	role, err := s.roleRepo.Update(ctx, id, belongsTo, repository.UpdateRoleOpts{
		Name:        opts.Name,
		Description: opts.Description,
	})
	if err != nil {
		return nil, errors.Join(ErrRoleUpdate, err)
	}

	return roleFromRepository(role), nil
}

func (s *roleService) ListMembers(ctx context.Context, id, belongsTo model.ID, page CursorPage) (Page[*User], error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/ListMembers")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return Page[*User]{}, errors.Join(ErrRoleGetBelongsTo, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*User]{}, errors.Join(ErrRoleGetBelongsTo, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return Page[*User]{}, errors.Join(ErrRoleGetBelongsTo, ErrNoPermission)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead) {
		return Page[*User]{}, errors.Join(ErrRoleGetBelongsTo, ErrNoPermission)
	}

	members, err := s.roleRepo.ListMembers(ctx, id, belongsTo, normalized)
	if err != nil {
		return Page[*User]{}, errors.Join(ErrOrganizationMembersGet, err)
	}

	return mapPage(members, userFromRepository), nil
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

	if s.notificationService != nil && s.organizationRepo != nil {
		role, err := s.roleRepo.Get(ctx, roleID, belongsToID, repository.RoleDetailProjection())
		if err != nil {
			s.logger.Warn(ctx, "failed to get role for notification when adding member",
				log.WithError(err),
				slog.String("role_id", roleID.String()))
		} else {
			organization, err := s.organizationRepo.Get(ctx, belongsToID, repository.OrganizationDetailProjection())
			if err != nil {
				s.logger.Warn(ctx, "failed to get organization for notification when adding member to role",
					log.WithError(err),
					slog.String("organization_id", belongsToID.String()))
			} else {
				notificationTitle := fmt.Sprintf("You've been added to the %s role", role.Name)
				notificationDescription := fmt.Sprintf("You have been added to the %s role in the %s organization.", role.Name, organization.Name)

				if _, err := s.notificationService.Create(ctx, CreateNotificationOpts{
					Title:       notificationTitle,
					Description: notificationDescription,
					Recipient:   memberID,
				}); err != nil {
					s.logger.Warn(ctx, "failed to send notification for role member addition",
						log.WithError(err),
						log.WithUserID(memberID.String()))
				}
			}
		}
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

	if s.notificationService != nil && s.organizationRepo != nil {
		role, err := s.roleRepo.Get(ctx, roleID, belongsToID, repository.RoleDetailProjection())
		if err != nil {
			s.logger.Warn(ctx, "failed to get role for notification when removing member",
				log.WithError(err),
				slog.String("role_id", roleID.String()))
		} else {
			organization, err := s.organizationRepo.Get(ctx, belongsToID, repository.OrganizationDetailProjection())
			if err != nil {
				s.logger.Warn(ctx, "failed to get organization for notification when removing member from role",
					log.WithError(err),
					slog.String("organization_id", belongsToID.String()))
			} else {
				notificationTitle := fmt.Sprintf("You've been removed from the %s role", role.Name)
				notificationDescription := fmt.Sprintf("You have been removed from the %s role in the %s organization.", role.Name, organization.Name)

				if _, err := s.notificationService.Create(ctx, CreateNotificationOpts{
					Title:       notificationTitle,
					Description: notificationDescription,
					Recipient:   memberID,
				}); err != nil {
					s.logger.Warn(ctx, "failed to send notification for role member removal",
						log.WithError(err),
						log.WithUserID(memberID.String()))
				}
			}
		}
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

	if _, err := s.roleRepo.Get(ctx, roleID, belongsToID, repository.RoleDetailProjection()); err != nil {
		return errors.Join(ErrRoleAddPermission, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleAddPermission, ErrNoPermission)
	}

	if _, err := s.permissionService.Create(ctx, CreatePermissionOpts{
		Subject: roleID,
		Target:  targetID,
		Kind:    kind,
	}); err != nil {
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

	if _, err := s.roleRepo.Get(ctx, roleID, belongsToID, repository.RoleDetailProjection()); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindWrite) {
		return errors.Join(ErrRoleRemovePermission, ErrNoPermission)
	}

	perm, err := s.permissionService.Get(ctx, permissionID)
	if err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	if perm.Subject.String() != roleID.String() {
		return errors.Join(ErrRoleRemovePermission, ErrNoPermission)
	}

	if err := s.permissionService.Delete(ctx, permissionID); err != nil {
		return errors.Join(ErrRoleRemovePermission, err)
	}

	return nil
}

func (s *roleService) GetPermissions(ctx context.Context, roleID, belongsToID model.ID) ([]*Permission, error) {
	ctx, span := s.tracer.Start(ctx, "service.roleService/GetPermissions")
	defer span.End()

	if err := roleID.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	if err := belongsToID.Validate(); err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	if _, err := s.roleRepo.Get(ctx, roleID, belongsToID, repository.RoleDetailProjection()); err != nil {
		return nil, errors.Join(ErrRoleGetPermissions, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsToID, model.PermissionKindRead) {
		return nil, errors.Join(ErrRoleGetPermissions, ErrNoPermission)
	}

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
