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
	Key         string
	Name        string
	Description string
	Actions     []model.Action
	MemberCount *int64
	Permissions []model.ID
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// CreateRoleOpts holds the data required to create a role.
type CreateRoleOpts struct {
	Key         string         `json:"key" validate:"omitempty,min=1,max=120"`
	Name        string         `json:"name" validate:"required,min=3,max=120"`
	Description string         `json:"description" validate:"omitempty,min=5,max=500"`
	Actions     []model.Action `json:"actions"`
}

// Validate validates the create options.
func (o *CreateRoleOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidRoleDetails, err)
	}
	for _, action := range o.Actions {
		if err := action.Validate(); err != nil {
			return errors.Join(model.ErrInvalidRoleDetails, err)
		}
	}
	return nil
}

// UpdateRoleOpts holds the fields that can be updated on a role.
// Undefined fields (Defined == false) are left unchanged.
type UpdateRoleOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
	Actions     optional.Optional[[]model.Action]
}

// RoleService is the interface that provides methods for managing roles.
//
//go:generate go tool mockgen -destination=mock/mock_role_gen.go -package=mocksvc . RoleService
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
	// Delete permanently deletes a role that belongs to belongsTo. If the role
	// does not exist, an error is returned.
	Delete(ctx context.Context, id, belongsTo model.ID) error
}

// roleService implements RoleService interface.
type roleService struct {
	runtime
	roleRepo            repository.RoleRepository
	permissionService   PermissionService
	licenseService      LicenseService
	organizationRepo    repository.OrganizationRepository
	notificationService NotificationService
}

func roleFromRepository(r *repository.Role) *Role {
	if r == nil {
		return nil
	}
	actions, _ := model.ParseActions(r.Actions)
	return &Role{
		ID:          r.ID,
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		Actions:     actions,
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

	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionRoleManage); err != nil {
		return nil, errors.Join(ErrRoleCreate, err)
	}

	if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaRoles); !ok || err != nil {
		return nil, errors.Join(ErrRoleCreate, ErrQuotaExceeded)
	}

	role, err := s.roleRepo.Create(ctx, repository.CreateRoleOpts{
		Key:         opts.Key,
		Name:        opts.Name,
		Description: opts.Description,
		Actions:     model.ActionStrings(opts.Actions),
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

	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionOrganizationRead); err != nil {
		return nil, errors.Join(ErrRoleGet, err)
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

	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionOrganizationRead); err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleGetBelongsTo, err)
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

	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionRoleManage); err != nil {
		return nil, errors.Join(ErrRoleUpdate, err)
	}

	repoOpts := repository.UpdateRoleOpts{
		Name:        opts.Name,
		Description: opts.Description,
	}
	if opts.Actions.Defined {
		if opts.Actions.Value == nil {
			repoOpts.Actions = optional.Some([]string{})
		} else {
			for _, action := range *opts.Actions.Value {
				if err := action.Validate(); err != nil {
					return nil, errors.Join(ErrRoleUpdate, err)
				}
			}
			repoOpts.Actions = optional.Some(model.ActionStrings(*opts.Actions.Value))
		}
	}

	role, err := s.roleRepo.Update(ctx, id, belongsTo, repoOpts)
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

	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionTeamManage); err != nil {
		return Page[*User]{}, errors.Join(ErrRoleGetBelongsTo, err)
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

	if err := requireAction(ctx, s.permissionService, belongsToID, model.ActionTeamManage); err != nil {
		return errors.Join(ErrRoleAddMember, err)
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

	if err := requireAction(ctx, s.permissionService, belongsToID, model.ActionTeamManage); err != nil {
		return errors.Join(ErrRoleRemoveMember, err)
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

	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionRoleManage); err != nil {
		return errors.Join(ErrRoleDelete, err)
	}

	err := s.roleRepo.Delete(ctx, id, belongsTo)
	if err != nil {
		return errors.Join(ErrRoleDelete, err)
	}

	return nil
}

// NewRoleService creates a new RoleService that provides methods
// for managing roles.
func NewRoleService(
	roleRepo repository.RoleRepository,
	permissionService PermissionService,
	licenseService LicenseService,
	organizationRepo repository.OrganizationRepository,
	notificationService NotificationService,
	opts ...Option,
) (RoleService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &roleService{
		runtime:             rt,
		roleRepo:            roleRepo,
		permissionService:   permissionService,
		licenseService:      licenseService,
		organizationRepo:    organizationRepo,
		notificationService: notificationService,
	}

	if svc.roleRepo == nil {
		return nil, ErrNoRoleRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	if svc.organizationRepo == nil {
		return nil, ErrNoOrganizationRepository
	}

	if svc.notificationService == nil {
		return nil, ErrNoNotificationService
	}

	return svc, nil
}
