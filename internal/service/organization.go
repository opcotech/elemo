package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// OrganizationService serves the business logic of interacting with
// organizations.
type OrganizationService interface {
	// Create creates a new organization. The owner of the organization is
	// automatically added as a member of the organization. If the owner
	// does not exist, an error is returned.
	Create(ctx context.Context, owner model.ID, organization *model.Organization) error
	// Get returns an organization by its ID. If the organization does not
	// exist, an error is returned.
	Get(ctx context.Context, id model.ID) (*model.Organization, error)
	// GetAll returns all organizations. The offset and limit parameters are
	// used to paginate the results. If the offset is greater than the number
	// of users in the system, an empty slice is returned.
	GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error)
	// Update updates an organization. If the organization does not exist, an
	// error is returned.
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error)
	// AddMember adds a member to an organization. If the organization or
	// member does not exist, an error is returned.
	AddMember(ctx context.Context, orgID, memberID model.ID) error
	// GetMembers returns all members of an organization with their roles. If the organization
	// does not exist, an error is returned.
	GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error)
	// RemoveMember removes a member from an organization. If the organization
	// or member does not exist, an error is returned.
	RemoveMember(ctx context.Context, orgID, memberID model.ID) error
	// Delete deletes an organization. If the organization does not exist, an
	// error is returned.
	Delete(ctx context.Context, id model.ID, force bool) error
}

// organizationService is the concrete implementation of OrganizationService.
type organizationService struct {
	*baseService
}

func (s *organizationService) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationCreate, license.ErrLicenseExpired)
	}

	if err := organization.Validate(); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeOrganization), model.PermissionKindCreate) {
		return errors.Join(ErrOrganizationCreate, ErrNoPermission)
	}

	// If the newly created organization is not active, e.g. a company is
	// migrating ex-employees, do not check the license quota as that only
	// counts against active organizations.
	if organization.Status == model.OrganizationStatusActive {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaOrganizations); !ok || err != nil {
			return errors.Join(ErrOrganizationCreate, ErrQuotaExceeded)
		}
	}

	if err := s.organizationRepo.Create(ctx, owner, organization); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	return nil
}

func (s *organizationService) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationGet, err)
	}

	organization, err := s.organizationRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrOrganizationGet, err)
	}

	return organization, nil
}

func (s *organizationService) GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/GetAll")
	defer span.End()

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrOrganizationGetAll, ErrInvalidPaginationParams)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrOrganizationGetAll, model.ErrInvalidID)
	}

	organizations, err := s.organizationRepo.GetAll(ctx, userID, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrOrganizationGetAll, err)
	}

	return organizations, nil
}

func (s *organizationService) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrOrganizationUpdate, ErrNoPermission)
	}

	// Check if the organization is being activated is within the license
	// quota. It could be a possible loophole to activate a previously deleted
	// organization to bypass the quota check.
	if patchStatus, ok := patch["status"]; ok && patchStatus == model.OrganizationStatusActive.String() {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaOrganizations); !ok || err != nil {
			return nil, errors.Join(ErrOrganizationUpdate, ErrQuotaExceeded)
		}
	}

	organization, err := s.organizationRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	return organization, nil
}

func (s *organizationService) Delete(ctx context.Context, id model.ID, force bool) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrOrganizationDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrOrganizationDelete, ErrNoPermission)
	}

	if force {
		if err := s.organizationRepo.Delete(ctx, id); err != nil {
			return errors.Join(ErrOrganizationDelete, err)
		}
	} else {
		patch := map[string]any{
			"status": model.OrganizationStatusDeleted.String(),
		}

		if _, err := s.organizationRepo.Update(ctx, id, patch); err != nil {
			return errors.Join(ErrOrganizationDelete, err)
		}
	}

	return nil
}

func (s *organizationService) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/AddMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationMemberAdd, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberAdd, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberAdd, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite) {
		return errors.Join(ErrOrganizationMemberAdd, ErrNoPermission)
	}

	if err := s.organizationRepo.AddMember(ctx, orgID, memberID); err != nil {
		return errors.Join(ErrOrganizationMemberAdd, err)
	}

	return nil
}

func (s *organizationService) GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/GetMembers")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationMembersGet, err)
	}

	members, err := s.organizationRepo.GetMembers(ctx, orgID)
	if err != nil {
		return nil, errors.Join(ErrOrganizationMembersGet, err)
	}

	result := make([]*model.OrganizationMember, 0, len(members))
	for _, member := range members {
		permissions, err := s.permissionService.GetBySubjectAndTarget(ctx, member.ID, orgID)
		if err != nil {
			return nil, errors.Join(ErrOrganizationMembersGet, err)
		}

		// Compute virtual roles based on permissions
		virtualRoles := computeVirtualRoles(permissions)

		// Combine virtual roles with actual roles (deduplicate)
		allRoles := combineRoles(virtualRoles, member.Roles)

		// Create new OrganizationMember with combined roles
		updatedMember, err := model.NewOrganizationMember(
			member.ID,
			member.FirstName,
			member.LastName,
			member.Email,
			member.Picture,
			member.Status,
			allRoles,
		)
		if err != nil {
			return nil, errors.Join(ErrOrganizationMembersGet, err)
		}

		result = append(result, updatedMember)
	}

	return result, nil
}

func (s *organizationService) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/RemoveMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationMemberRemove, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberRemove, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberRemove, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite) {
		return errors.Join(ErrOrganizationMemberRemove, ErrNoPermission)
	}

	if err := s.organizationRepo.RemoveMember(ctx, orgID, memberID); err != nil {
		return errors.Join(ErrOrganizationMemberRemove, err)
	}

	return nil
}

// NewOrganizationService returns a new instance of the OrganizationService
// interface.
func NewOrganizationService(opts ...Option) (OrganizationService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &organizationService{
		baseService: s,
	}

	if svc.organizationRepo == nil {
		return nil, ErrNoOrganizationRepository
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

// computeVirtualRoles computes virtual roles based on permissions:
// - owner: if user has `all` permissions OR has `read` AND `write` AND `delete`
// - admin: if user has `write` permission
// - member: if user has ONLY `read` permission
func computeVirtualRoles(permissions []*model.Permission) []string {
	virtualRoles := make([]string, 0)
	hasRead := false
	hasWrite := false
	hasDelete := false
	hasAll := false

	for _, perm := range permissions {
		switch perm.Kind {
		case model.PermissionKindAll:
			hasAll = true
		case model.PermissionKindRead:
			hasRead = true
		case model.PermissionKindWrite:
			hasWrite = true
		case model.PermissionKindDelete:
			hasDelete = true
		}
	}

	switch {
	case hasAll || (hasRead && hasWrite && hasDelete):
		virtualRoles = append(virtualRoles, "Owner")
	case hasWrite:
		virtualRoles = append(virtualRoles, "Admin")
	case hasRead && !hasWrite && !hasDelete:
		virtualRoles = append(virtualRoles, "Member")
	}

	return virtualRoles
}

// combineRoles combines virtual roles with actual roles, deduplicating
func combineRoles(virtualRoles, actualRoles []string) []string {
	roleSet := make(map[string]bool)
	result := make([]string, 0)

	// Add virtual roles first
	for _, role := range virtualRoles {
		if !roleSet[role] {
			roleSet[role] = true
			result = append(result, role)
		}
	}

	// Add actual roles
	for _, role := range actualRoles {
		if !roleSet[role] {
			roleSet[role] = true
			result = append(result, role)
		}
	}

	return result
}
