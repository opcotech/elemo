package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/rs/xid"
)

// Organization represents an organization returned by the service.
type Organization struct {
	ID             model.ID
	Name           string
	Email          string
	Logo           string
	Website        string
	Status         model.OrganizationStatus
	NamespaceCount *int64
	TeamCount      *int64
	MemberCount    *int64
	DocumentCount  *int64
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

// OrganizationMember represents a member of an organization.
type OrganizationMember struct {
	ID        model.ID
	FirstName string
	LastName  string
	Email     string
	Picture   *string
	Status    model.UserStatus
	Roles     []string
	CreatedAt *time.Time
}

// CreateOrganizationOpts holds the data required to create an organization.
type CreateOrganizationOpts struct {
	Name    string                   `json:"name" validate:"required,min=1,max=120"`
	Email   string                   `json:"email" validate:"required,email"`
	Logo    string                   `json:"logo" validate:"omitempty,url"`
	Website string                   `json:"website" validate:"omitempty,url"`
	Status  model.OrganizationStatus `json:"status" validate:"omitempty,min=1,max=2"`
}

// Validate validates the create options.
func (o *CreateOrganizationOpts) Validate() error {
	if o.Status == 0 {
		o.Status = model.OrganizationStatusActive
	}
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidOrganizationDetails, err)
	}
	return nil
}

// UpdateOrganizationOpts holds the fields that can be updated on an organization.
// Undefined fields (Defined == false) are left unchanged.
type UpdateOrganizationOpts struct {
	Name    optional.Optional[string]
	Email   optional.Optional[string]
	Logo    optional.Optional[string]
	Website optional.Optional[string]
	Status  optional.Optional[model.OrganizationStatus]
}

// InviteOrganizationMemberOpts holds the data required to invite a member.
type InviteOrganizationMemberOpts struct {
	Email  string   `json:"email" validate:"required,email"`
	RoleID model.ID `json:"role_id"`
}

// Validate validates the invite options.
func (o *InviteOrganizationMemberOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidOrganizationMemberDetails, err)
	}
	if !o.RoleID.IsNil() {
		if err := o.RoleID.Validate(); err != nil {
			return errors.Join(model.ErrInvalidOrganizationMemberDetails, err)
		}
		if o.RoleID.Type != model.ResourceTypeRole {
			return errors.Join(model.ErrInvalidOrganizationMemberDetails, model.ErrInvalidID)
		}
	}
	return nil
}

// AcceptOrganizationInvitationOpts holds the data required to accept an invitation.
type AcceptOrganizationInvitationOpts struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"omitempty"`
}

// Validate validates the accept invitation options.
func (o *AcceptOrganizationInvitationOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(ErrInvalidToken, err)
	}
	return nil
}

// OrganizationService serves the business logic of interacting with
// organizations.
//
//go:generate go tool mockgen -destination=organization_mock_gen.go -package=service -mock_names OrganizationService=MockOrganizationService . OrganizationService
type OrganizationService interface {
	// Create creates a new organization. The owner of the organization is
	// automatically added as a member of the organization. If the owner
	// does not exist, an error is returned.
	Create(ctx context.Context, owner model.ID, opts CreateOrganizationOpts) (*Organization, error)
	// Get returns an organization by its ID. If the organization does not
	// exist, an error is returned.
	Get(ctx context.Context, id model.ID) (*Organization, error)
	// List returns a cursor-paginated page of organizations for the current user.
	List(ctx context.Context, page CursorPage) (Page[*Organization], error)
	// Update updates an organization. If the organization does not exist, an
	// error is returned.
	Update(ctx context.Context, id model.ID, opts UpdateOrganizationOpts) (*Organization, error)
	// AddMember adds a member to an organization. If the organization or
	// member does not exist, an error is returned.
	AddMember(ctx context.Context, orgID, memberID model.ID) error
	// ListMembers returns a cursor-paginated page of organization members with
	// their roles. If the organization does not exist, an error is returned.
	ListMembers(ctx context.Context, orgID model.ID, page CursorPage) (Page[*OrganizationMember], error)
	// RemoveMember removes a member from an organization. If the organization
	// or member does not exist, an error is returned.
	RemoveMember(ctx context.Context, orgID, memberID model.ID) error
	// InviteMember sends an invitation email to a user to join an organization.
	// If the user doesn't exist, a pending user is created. If the organization
	// does not exist, an error is returned. Optionally, a RoleID can be provided
	// to assign the user to a specific role when they accept the invitation.
	InviteMember(ctx context.Context, orgID model.ID, opts InviteOrganizationMemberOpts) error
	// RevokeInvitation revokes an invitation for a user to join an organization.
	// If the organization or user does not exist, an error is returned.
	RevokeInvitation(ctx context.Context, orgID, userID model.ID) error
	// AcceptInvitation accepts an invitation to join an organization using an invitation token.
	// If the user is pending, they will be activated. If a password is provided, it will be set.
	AcceptInvitation(ctx context.Context, orgID model.ID, opts AcceptOrganizationInvitationOpts) error
	// Delete deletes an organization. If the organization does not exist, an
	// error is returned.
	Delete(ctx context.Context, id model.ID, force bool) error
}

// organizationService is the concrete implementation of OrganizationService.
type organizationService struct {
	*baseService
}

func organizationFromRepository(o *repository.Organization) *Organization {
	if o == nil {
		return nil
	}
	return &Organization{
		ID:             o.ID,
		Name:           o.Name,
		Email:          o.Email,
		Logo:           o.Logo,
		Website:        o.Website,
		Status:         o.Status,
		NamespaceCount: o.NamespaceCount,
		TeamCount:      o.TeamCount,
		MemberCount:    o.MemberCount,
		DocumentCount:  o.DocumentCount,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

func (s *organizationService) Create(ctx context.Context, owner model.ID, opts CreateOrganizationOpts) (*Organization, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrOrganizationCreate, license.ErrLicenseExpired)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationCreate, err)
	}

	if !s.permissionService.CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate) {
		return nil, errors.Join(ErrOrganizationCreate, ErrNoPermission)
	}

	// If the newly created organization is not active, e.g. a company is
	// migrating ex-employees, do not check the license quota as that only
	// counts against active organizations.
	if opts.Status == model.OrganizationStatusActive {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaOrganizations); !ok || err != nil {
			return nil, errors.Join(ErrOrganizationCreate, ErrQuotaExceeded)
		}
	}

	organization, err := s.organizationRepo.Create(ctx, repository.CreateOrganizationOpts{
		Owner:   owner,
		Name:    opts.Name,
		Email:   opts.Email,
		Logo:    opts.Logo,
		Website: opts.Website,
		Status:  opts.Status,
	})
	if err != nil {
		return nil, errors.Join(ErrOrganizationCreate, err)
	}

	if err := s.seedOrganizationAuth(ctx, owner, organization.ID); err != nil {
		return nil, errors.Join(ErrOrganizationCreate, err)
	}

	return organizationFromRepository(organization), nil
}

func (s *organizationService) seedOrganizationAuth(ctx context.Context, owner, orgID model.ID) error {
	for _, tmpl := range model.RoleTemplates {
		if _, err := s.roleRepo.Create(ctx, repository.CreateRoleOpts{
			Key:         tmpl.Key,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Actions:     tmpl.ActionStrings(),
			CreatedBy:   owner,
			BelongsTo:   orgID,
		}); err != nil {
			return err
		}
	}

	adminRole, err := s.roleRepo.GetByKey(ctx, orgID, model.RoleKeyOrgAdmin)
	if err != nil {
		return err
	}
	if err := s.permissionService.GrantRole(ctx, owner, orgID, adminRole.ID); err != nil {
		return err
	}

	memberRole, err := s.roleRepo.GetByKey(ctx, orgID, model.RoleKeyOrgMember)
	if err != nil {
		return err
	}
	return s.permissionService.GrantRole(ctx, orgID, orgID, memberRole.ID)
}

func (s *organizationService) Get(ctx context.Context, id model.ID) (*Organization, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationGet, err)
	}

	if !s.permissionService.CtxUserHas(ctx, id, model.ActionOrganizationRead) {
		return nil, errors.Join(ErrOrganizationGet, ErrNoPermission)
	}

	organization, err := s.organizationRepo.Get(ctx, id, repository.OrganizationDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrOrganizationGet, err)
	}

	return organizationFromRepository(organization), nil
}

func (s *organizationService) List(ctx context.Context, page CursorPage) (Page[*Organization], error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Organization]{}, errors.Join(ErrOrganizationGetAll, err)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return Page[*Organization]{}, errors.Join(ErrOrganizationGetAll, model.ErrInvalidID)
	}

	organizations, err := s.organizationRepo.List(
		ctx,
		userID,
		normalized,
		repository.OrganizationListProjection(),
	)
	if err != nil {
		return Page[*Organization]{}, errors.Join(ErrOrganizationGetAll, err)
	}

	return mapPage(organizations, organizationFromRepository), nil
}

func (s *organizationService) Update(ctx context.Context, id model.ID, opts UpdateOrganizationOpts) (*Organization, error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	if !s.permissionService.CtxUserHas(ctx, id, model.ActionOrganizationUpdate) {
		return nil, errors.Join(ErrOrganizationUpdate, ErrNoPermission)
	}

	// Check if the organization is being activated is within the license
	// quota. It could be a possible loophole to activate a previously deleted
	// organization to bypass the quota check.
	if opts.Status.Defined && opts.Status.Value != nil && *opts.Status.Value == model.OrganizationStatusActive {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaOrganizations); !ok || err != nil {
			return nil, errors.Join(ErrOrganizationUpdate, ErrQuotaExceeded)
		}
	}

	organization, err := s.organizationRepo.Update(ctx, id, repository.UpdateOrganizationOpts{
		Name:    opts.Name,
		Email:   opts.Email,
		Logo:    opts.Logo,
		Website: opts.Website,
		Status:  opts.Status,
	})
	if err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	return organizationFromRepository(organization), nil
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

	if !s.permissionService.CtxUserHas(ctx, id, model.ActionOrganizationDelete) {
		return errors.Join(ErrOrganizationDelete, ErrNoPermission)
	}

	if force {
		if err := s.organizationRepo.Delete(ctx, id); err != nil {
			return errors.Join(ErrOrganizationDelete, err)
		}
	} else {
		if _, err := s.organizationRepo.Update(ctx, id, repository.UpdateOrganizationOpts{
			Status: optional.Some(model.OrganizationStatusDeleted),
		}); err != nil {
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

	if !s.permissionService.CtxUserHas(ctx, orgID, model.ActionOrganizationMembersManage) {
		return errors.Join(ErrOrganizationMemberAdd, ErrNoPermission)
	}

	if err := s.organizationRepo.AddMember(ctx, orgID, memberID); err != nil {
		return errors.Join(ErrOrganizationMemberAdd, err)
	}

	return nil
}

func (s *organizationService) ListMembers(ctx context.Context, orgID model.ID, page CursorPage) (Page[*OrganizationMember], error) {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/ListMembers")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationMembersGet, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationMembersGet, err)
	}

	if !s.permissionService.CtxUserHas(ctx, orgID, model.ActionOrganizationRead) {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationMembersGet, ErrNoPermission)
	}

	members, err := s.organizationRepo.ListMembers(ctx, orgID, normalized)
	if err != nil {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationMembersGet, err)
	}

	return mapPage(members, func(member *repository.OrganizationMember) *OrganizationMember {
		return &OrganizationMember{
			ID:        member.ID,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Email:     member.Email,
			Picture:   member.Picture,
			Status:    member.Status,
			Roles:     member.Roles,
			CreatedAt: member.CreatedAt,
		}
	}), nil
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

	if !s.permissionService.CtxUserHas(ctx, orgID, model.ActionOrganizationMembersManage) {
		return errors.Join(ErrOrganizationMemberRemove, ErrNoPermission)
	}

	grants, err := s.permissionService.ListByPrincipal(ctx, memberID)
	if err != nil && !errors.Is(err, repository.ErrPermissionRead) {
		s.logger.Warn(ctx, "failed to get grants when removing member",
			log.WithError(err),
			log.WithUserID(memberID.String()),
			slog.String("organization_id", orgID.String()))
	} else {
		for _, grant := range grants {
			if grant.Scope != orgID {
				continue
			}
			if err := s.permissionService.Delete(ctx, grant.ID); err != nil {
				s.logger.Warn(ctx, "failed to delete grant when removing member",
					log.WithError(err),
					slog.String("grant_id", grant.ID.String()),
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
		organization, err := s.organizationRepo.Get(ctx, orgID, repository.OrganizationDetailProjection())
		if err != nil {
			s.logger.Warn(ctx, "failed to get organization for notification when removing member",
				log.WithError(err),
				slog.String("organization_id", orgID.String()))
		} else {
			notificationTitle := fmt.Sprintf("You've been removed from %s", organization.Name)
			notificationDescription := fmt.Sprintf("You have been removed from the organization %s.", organization.Name)

			if _, err := s.notificationService.Create(ctx, CreateNotificationOpts{
				Title:       notificationTitle,
				Description: notificationDescription,
				Recipient:   memberID,
			}); err != nil {
				s.logger.Warn(ctx, "failed to send notification for member removal",
					log.WithError(err),
					log.WithUserID(memberID.String()))
			}
		}
	}

	return nil
}

func (s *organizationService) InviteMember(ctx context.Context, orgID model.ID, opts InviteOrganizationMemberOpts) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/InviteMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrOrganizationMemberInvite, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if err := opts.Validate(); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if !s.permissionService.CtxUserHas(ctx, orgID, model.ActionOrganizationMembersManage) {
		return errors.Join(ErrOrganizationMemberInvite, ErrNoPermission)
	}

	user, err := s.userRepo.GetByEmail(ctx, opts.Email, repository.UserDetailProjection())
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	userExists := true
	if errors.Is(err, repository.ErrNotFound) {
		userExists = false

		firstName, lastName := convert.EmailToNameParts(opts.Email)

		user, err = s.userRepo.Create(ctx, repository.CreateUserOpts{
			Username:  xid.New().String(),
			FirstName: firstName,
			LastName:  lastName,
			Email:     opts.Email,
			Password:  password.UnusablePassword,
			Status:    model.UserStatusPending,
			Links:     make([]string, 0),
			Languages: make([]model.Language, 0),
		})
		if err != nil {
			return errors.Join(ErrOrganizationMemberInvite, err)
		}
	}

	if userExists {
		if user.Status != model.UserStatusActive && user.Status != model.UserStatusPending {
			return errors.Join(ErrOrganizationMemberInvite, ErrOrganizationMemberInvalidStatus)
		}
	}

	hasPermission, err := s.permissionService.Has(ctx, user.ID, orgID, model.ActionOrganizationRead)
	if err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}
	if hasPermission {
		return errors.Join(ErrOrganizationMemberInvite, ErrOrganizationMemberAlreadyExists)
	}

	organization, err := s.organizationRepo.Get(ctx, orgID, repository.OrganizationDetailProjection())
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

	if !opts.RoleID.IsNil() {
		tokenData["role_id"] = opts.RoleID.String()
	}

	public, secret, err := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
	if err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	if _, err := s.userTokenRepo.Create(ctx, repository.CreateUserTokenOpts{
		UserID:  user.ID,
		SentTo:  opts.Email,
		Token:   secret,
		Context: model.UserTokenContextInvite,
	}); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	token := public

	if err := s.emailService.SendOrganizationInvitationEmail(ctx, organization.ID, organization.Name, email.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, token); err != nil {
		return errors.Join(ErrOrganizationMemberInvite, err)
	}

	// Send notification to the invited user
	if s.notificationService != nil {
		notificationTitle := fmt.Sprintf("You've been invited to join %s", organization.Name)
		notificationDescription := fmt.Sprintf("You have been invited to join the organization %s. Click the link in your email to accept the invitation.", organization.Name)

		if _, err := s.notificationService.Create(ctx, CreateNotificationOpts{
			Title:       notificationTitle,
			Description: notificationDescription,
			Recipient:   user.ID,
		}); err != nil {
			s.logger.Warn(ctx, "failed to send notification for invitation",
				log.WithError(err),
				log.WithUserID(user.ID.String()))
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

	if !s.permissionService.CtxUserHas(ctx, orgID, model.ActionOrganizationMembersManage) {
		return errors.Join(ErrOrganizationInviteRevoke, ErrNoPermission)
	}

	// Get user to verify they exist and check status
	user, err := s.userRepo.Get(ctx, userID, repository.UserDetailProjection())
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
		organizations, err := s.organizationRepo.List(
			ctx,
			userID,
			repository.CursorPage{Size: 1},
			repository.OrganizationListProjection(),
		)
		if err != nil {
			s.logger.Warn(ctx, "failed to check user organization membership during invitation revocation",
				log.WithError(err),
				log.WithUserID(userID.String()))
			return nil
		}

		if len(organizations.Items) == 0 {
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

func (s *organizationService) AcceptInvitation(ctx context.Context, orgID model.ID, opts AcceptOrganizationInvitationOpts) error {
	ctx, span := s.tracer.Start(ctx, "service.organizationService/AcceptInvitation")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if err := opts.Validate(); err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	kind, _, tokenData := auth.SplitToken(opts.Token)

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

	if !auth.IsTokenMatching(confirmation.Token, opts.Token) {
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

	user, err := s.userRepo.Get(ctx, userID, repository.UserDetailProjection())
	if err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if user.Status != model.UserStatusPending && user.Status != model.UserStatusActive {
		return errors.Join(ErrOrganizationInviteAccept, errors.New("user account is not in a valid state to accept invitations"))
	}

	if user.Status == model.UserStatusPending {
		if opts.Password == "" {
			return errors.Join(ErrOrganizationInviteAccept, errors.New("password is required for pending users"))
		}

		hashedPassword := password.HashPassword(opts.Password)

		if _, err := s.userRepo.Update(ctx, userID, repository.UpdateUserOpts{
			Status:   optional.Some(model.UserStatusActive),
			Password: optional.Some(hashedPassword),
		}); err != nil {
			return errors.Join(ErrOrganizationInviteAccept, err)
		}
	}

	if err := s.organizationRepo.RemoveInvitation(ctx, orgID, userID); err != nil {
		s.logger.Warn(ctx, "failed to remove invitation edge during acceptance",
			log.WithError(err),
			log.WithUserID(userID.String()),
			slog.String("organization_id", orgID.String()))
	}

	if _, err := s.organizationRepo.Get(ctx, orgID, repository.OrganizationDetailProjection()); err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if err := s.organizationRepo.AddMember(ctx, orgID, userID); err != nil {
		return errors.Join(ErrOrganizationInviteAccept, err)
	}

	if roleIDStr, ok := tokenData["role_id"].(string); ok && roleIDStr != "" {
		roleID, err := model.NewIDFromString(roleIDStr, model.ResourceTypeRole.String())
		if err == nil && !roleID.IsNil() {
			if err := s.permissionService.GrantRole(ctx, userID, orgID, roleID); err != nil {
				s.logger.Warn(ctx, "failed to grant role during invitation acceptance",
					log.WithError(err),
					log.WithUserID(userID.String()),
					slog.String("organization_id", orgID.String()),
					slog.String("role_id", roleID.String()))
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

	if svc.roleRepo == nil {
		return nil, ErrNoRoleRepository
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
