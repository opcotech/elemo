package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/rs/xid"
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
	// InviteMember sends an invitation email to a user to join an organization.
	// If the user doesn't exist, a pending user is created. If the organization
	// does not exist, an error is returned. Optionally, a roleID can be provided
	// to assign the user to a specific role when they accept the invitation.
	InviteMember(ctx context.Context, orgID model.ID, email string, roleID ...model.ID) error
	// RevokeInvitation revokes an invitation for a user to join an organization.
	// If the organization or user does not exist, an error is returned.
	RevokeInvitation(ctx context.Context, orgID, userID model.ID) error
	// AcceptInvitation accepts an invitation to join an organization using an invitation token.
	// If the user is pending, they will be activated. If a password is provided, it will be set.
	AcceptInvitation(ctx context.Context, orgID model.ID, token string, password string) error
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

	perm, err := model.NewPermission(memberID, orgID, model.PermissionKindRead)
	if err != nil {
		return errors.Join(ErrOrganizationMemberAdd, err)
	}

	if err := s.permissionService.Create(ctx, perm); err != nil {
		// Log error but don't fail - member is already added
		s.logger.Warn(ctx, "failed to assign read permission to new member",
			log.WithError(err),
			log.WithUserID(memberID.String()),
			slog.String("organization_id", orgID.String()))
	}

	return nil
}

func (s *organizationService) GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/GetMembers")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationMembersGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindRead) {
		return nil, errors.Join(ErrOrganizationMembersGet, ErrNoPermission)
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

	// Remove all permissions for the member on the organization before removing the member
	permissions, err := s.permissionService.GetBySubjectAndTarget(ctx, memberID, orgID)
	if err != nil && !errors.Is(err, repository.ErrPermissionRead) {
		s.logger.Warn(ctx, "failed to get permissions when removing member",
			log.WithError(err),
			log.WithUserID(memberID.String()),
			slog.String("organization_id", orgID.String()))
	} else {
		for _, perm := range permissions {
			if err := s.permissionService.Delete(ctx, perm.ID); err != nil {
				s.logger.Warn(ctx, "failed to delete permission when removing member",
					log.WithError(err),
					slog.String("permission_id", perm.ID.String()),
					log.WithUserID(memberID.String()),
					slog.String("organization_id", orgID.String()))
			}
		}
	}

	if err := s.organizationRepo.RemoveMember(ctx, orgID, memberID); err != nil {
		return errors.Join(ErrOrganizationMemberRemove, err)
	}

	// Send notification to the removed member
	if s.notificationService != nil {
		organization, err := s.organizationRepo.Get(ctx, orgID)
		if err != nil {
			s.logger.Warn(ctx, "failed to get organization for notification when removing member",
				log.WithError(err),
				slog.String("organization_id", orgID.String()))
		} else {
			notificationTitle := fmt.Sprintf("You've been removed from %s", organization.Name)
			notificationDescription := fmt.Sprintf("You have been removed from the organization %s.", organization.Name)

			notification, err := model.NewNotification(notificationTitle, memberID)
			if err != nil {
				s.logger.Warn(ctx, "failed to create notification for member removal",
					log.WithError(err),
					log.WithUserID(memberID.String()))
			} else {
				notification.Description = notificationDescription
				if err := s.notificationService.Create(ctx, notification); err != nil {
					s.logger.Warn(ctx, "failed to send notification for member removal",
						log.WithError(err),
						log.WithUserID(memberID.String()))
				}
			}
		}
	}

	return nil
}

func (s *organizationService) InviteMember(ctx context.Context, orgID model.ID, email string, roleID ...model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/InviteMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationMemberInvite, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if email == "" {
		return errors.Join(ErrOrganizationMemberInvite, ErrInvalidEmail)
	}

	var targetRoleID model.ID
	if len(roleID) > 0 && !roleID[0].IsNil() {
		targetRoleID = roleID[0]
		if err := targetRoleID.Validate(); err != nil {
			return errors.Join(ErrOrganizationMemberInvite, err)
		}
		if targetRoleID.Type != model.ResourceTypeRole {
			return errors.Join(ErrOrganizationMemberInvite, model.ErrInvalidID)
		}
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite) {
		return errors.Join(ErrOrganizationMemberInvite, ErrNoPermission)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	userExists := true
	if errors.Is(err, repository.ErrNotFound) {
		userExists = false

		firstName, lastName := convert.EmailToNameParts(email)

		user, err = model.NewUser(xid.New().String(), firstName, lastName, email, password.UnusablePassword)
		if err != nil {
			return errors.Join(ErrOrganizationMemberInvite, err)
		}

		user.Status = model.UserStatusPending

		if err := s.userRepo.Create(ctx, user); err != nil {
			return errors.Join(ErrOrganizationMemberInvite, err)
		}
	}

	if userExists {
		if user.Status != model.UserStatusActive && user.Status != model.UserStatusPending {
			return errors.Join(ErrOrganizationMemberInvite, ErrOrganizationMemberInvalidStatus)
		}
	}

	hasPermission, err := s.permissionService.HasPermission(ctx, user.ID, orgID, model.PermissionKindRead)
	if err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}
	if hasPermission {
		return errors.Join(ErrOrganizationMemberInvite, ErrOrganizationMemberAlreadyExists)
	}

	organization, err := s.organizationRepo.Get(ctx, orgID)
	if err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if err := s.organizationRepo.AddInvitation(ctx, orgID, user.ID); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	existingToken, err := s.userTokenRepo.Get(ctx, user.ID, model.UserTokenContextInvite)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if existingToken != nil {
		if err := s.userTokenRepo.Delete(ctx, existingToken.UserID, existingToken.Context); err != nil {
			return errors.Join(ErrOrganizationMemberInvite, err)
		}
	}

	tokenData := pkg.MergeMaps(map[string]any{
		"organization_id": orgID.String(),
	}, map[string]any{"user_id": user.ID.String()})

	if !targetRoleID.IsNil() {
		tokenData["role_id"] = targetRoleID.String()
	}

	public, secret, err := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
	if err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	newToken, err := model.NewUserToken(user.ID, email, secret, model.UserTokenContextInvite)
	if err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if err := s.userTokenRepo.Create(ctx, newToken); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	token := public

	if err := s.emailService.SendOrganizationInvitationEmail(ctx, organization, user, token); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	// Send notification to the invited user
	if s.notificationService != nil {
		notificationTitle := fmt.Sprintf("You've been invited to join %s", organization.Name)
		notificationDescription := fmt.Sprintf("You have been invited to join the organization %s. Click the link in your email to accept the invitation.", organization.Name)

		notification, err := model.NewNotification(notificationTitle, user.ID)
		if err != nil {
			s.logger.Warn(ctx, "failed to create notification for invitation",
				log.WithError(err),
				log.WithUserID(user.ID.String()))
		} else {
			notification.Description = notificationDescription
			if err := s.notificationService.Create(ctx, notification); err != nil {
				s.logger.Warn(ctx, "failed to send notification for invitation",
					log.WithError(err),
					log.WithUserID(user.ID.String()))
			}
		}
	}

	return nil
}

func (s *organizationService) RevokeInvitation(ctx context.Context, orgID, userID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/RevokeInvitation")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationInviteRevoke, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationInviteRevoke, err)
	}

	if err := userID.Validate(); err != nil {
		return errors.Join(ErrOrganizationInviteRevoke, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite) {
		return errors.Join(ErrOrganizationInviteRevoke, ErrNoPermission)
	}

	// Get user to verify they exist and check status
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return errors.Join(ErrOrganizationInviteRevoke, err)
	}

	if err := s.organizationRepo.RemoveInvitation(ctx, orgID, userID); err != nil {
		s.logger.Warn(ctx, "failed to remove invitation edge during revocation",
			log.WithError(err),
			log.WithUserID(userID.String()),
			slog.String("organization_id", orgID.String()))
	}

	if err := s.userTokenRepo.Delete(ctx, userID, model.UserTokenContextInvite); err != nil {
		s.logger.Warn(ctx, "failed to delete invitation token during revocation",
			log.WithError(err),
			log.WithUserID(userID.String()))
	}

	if err := s.organizationRepo.RemoveMember(ctx, orgID, userID); err != nil {
		s.logger.Warn(ctx, "failed to remove member during invitation revocation",
			log.WithError(err),
			log.WithUserID(userID.String()),
			slog.String("organization_id", orgID.String()))
	}

	if user.Status == model.UserStatusPending {
		organizations, err := s.organizationRepo.GetAll(ctx, userID, 0, 1)
		if err != nil {
			s.logger.Warn(ctx, "failed to check user organization membership during invitation revocation",
				log.WithError(err),
				log.WithUserID(userID.String()))
			return nil
		}

		if len(organizations) == 0 {
			if err := s.userRepo.Delete(ctx, userID); err != nil {
				s.logger.Error(ctx, "failed to delete pending user account during invitation revocation",
					log.WithError(err),
					log.WithUserID(userID.String()))
				return nil
			}
			s.logger.Info(ctx, "deleted pending user account after invitation revocation",
				log.WithUserID(userID.String()))
		}
	}

	return nil
}

func (s *organizationService) AcceptInvitation(ctx context.Context, orgID model.ID, token string, userPassword string) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/AcceptInvitation")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if token == "" {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	kind, _, tokenData := auth.SplitToken(token)

	userIDStr, ok := tokenData["user_id"].(string)
	if !ok {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	userID, err := model.NewIDFromString(userIDStr, model.ResourceTypeUser.String())
	if err != nil {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	var tokenContext model.UserTokenContext
	if err := tokenContext.UnmarshalText([]byte(kind)); err != nil {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	if tokenContext != model.UserTokenContextInvite {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	confirmation, err := s.userTokenRepo.Get(ctx, userID, tokenContext)
	if err != nil {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	if !auth.IsTokenMatching(confirmation.Token, token) {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	if time.Now().After(confirmation.CreatedAt.Add(UserInvitationDeadline)) {
		return errors.Join(ErrOrganizationInviteAccept, ErrExpiredToken)
	}

	orgIDStr, ok := tokenData["organization_id"].(string)
	if !ok {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	expectedOrgID, err := model.NewIDFromString(orgIDStr, model.ResourceTypeOrganization.String())
	if err != nil {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	if expectedOrgID != orgID {
		return errors.Join(ErrOrganizationInviteAccept, ErrInvalidToken)
	}

	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if user.Status != model.UserStatusPending && user.Status != model.UserStatusActive {
		return errors.Join(ErrOrganizationInviteAccept, errors.New("user account is not in a valid state to accept invitations"))
	}

	if user.Status == model.UserStatusPending {
		if userPassword == "" {
			return errors.Join(ErrOrganizationInviteAccept, errors.New("password is required for pending users"))
		}

		hashedPassword := password.HashPassword(userPassword)

		patch := map[string]any{
			"status":   model.UserStatusActive.String(),
			"password": hashedPassword,
		}

		if _, err := s.userRepo.Update(ctx, userID, patch); err != nil {
			return errors.Join(ErrOrganizationInviteAccept, err)
		}
	}

	if err := s.organizationRepo.RemoveInvitation(ctx, orgID, userID); err != nil {
		s.logger.Warn(ctx, "failed to remove invitation edge during acceptance",
			log.WithError(err),
			log.WithUserID(userID.String()),
			slog.String("organization_id", orgID.String()))
	}

	organization, err := s.organizationRepo.Get(ctx, orgID)
	if err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if !slices.Contains(organization.Members, userID) {
		if err := s.organizationRepo.AddMember(ctx, orgID, userID); err != nil {
			return errors.Join(ErrOrganizationInviteAccept, err)
		}

		perm, err := model.NewPermission(userID, orgID, model.PermissionKindRead)
		if err != nil {
			return errors.Join(ErrOrganizationInviteAccept, err)
		}

		if err := s.permissionService.Create(ctx, perm); err != nil {
			s.logger.Warn(ctx, "failed to assign read permission to new member during invitation acceptance",
				log.WithError(err),
				log.WithUserID(userID.String()),
				slog.String("organization_id", orgID.String()))
		}
	}

	if roleIDStr, ok := tokenData["role_id"].(string); ok && roleIDStr != "" {
		roleID, err := model.NewIDFromString(roleIDStr, model.ResourceTypeRole.String())
		if err == nil && !roleID.IsNil() {
			if s.roleRepo != nil {
				_, err := s.roleRepo.Get(ctx, roleID, orgID)
				if err == nil {
					_ = s.roleRepo.AddMember(ctx, roleID, userID, orgID)
				}
			}
		}
	}

	if err := s.userTokenRepo.Delete(ctx, userID, model.UserTokenContextInvite); err != nil {
		s.logger.Warn(ctx, "failed to delete invitation token after acceptance",
			log.WithError(err),
			log.WithUserID(userID.String()))
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

	if svc.userTokenRepo == nil {
		return nil, ErrNoUserTokenRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	if svc.emailService == nil {
		return nil, ErrNoEmailService
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
