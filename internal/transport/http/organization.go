package http

import (
	"context"
	"errors"
	"fmt"
	"strings"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// isOrganizationScopedResource checks if a resource type is organization-scoped.
// Organization-scoped resources are: Organization, Namespace, Document, Project, Role.
func isOrganizationScopedResource(resourceType model.ResourceType) bool {
	switch resourceType {
	case model.ResourceTypeOrganization,
		model.ResourceTypeNamespace,
		model.ResourceTypeDocument,
		model.ResourceTypeFolder,
		model.ResourceTypeProject,
		model.ResourceTypeRole:
		return true
	default:
		return false
	}
}

// OrganizationController is a controller for organization endpoints.
type OrganizationController interface {
	V1OrganizationsGet(ctx context.Context, request api.V1OrganizationsGetRequestObject) (api.V1OrganizationsGetResponseObject, error)
	V1OrganizationsCreate(ctx context.Context, request api.V1OrganizationsCreateRequestObject) (api.V1OrganizationsCreateResponseObject, error)
	V1OrganizationDelete(ctx context.Context, request api.V1OrganizationDeleteRequestObject) (api.V1OrganizationDeleteResponseObject, error)
	V1OrganizationGet(ctx context.Context, request api.V1OrganizationGetRequestObject) (api.V1OrganizationGetResponseObject, error)
	V1OrganizationUpdate(ctx context.Context, request api.V1OrganizationUpdateRequestObject) (api.V1OrganizationUpdateResponseObject, error)
	V1OrganizationMembersGet(ctx context.Context, request api.V1OrganizationMembersGetRequestObject) (api.V1OrganizationMembersGetResponseObject, error)
	V1OrganizationMembersAdd(ctx context.Context, request api.V1OrganizationMembersAddRequestObject) (api.V1OrganizationMembersAddResponseObject, error)
	V1OrganizationMembersInvite(ctx context.Context, request api.V1OrganizationMembersInviteRequestObject) (api.V1OrganizationMembersInviteResponseObject, error)
	V1OrganizationMembersAccept(ctx context.Context, request api.V1OrganizationMembersAcceptRequestObject) (api.V1OrganizationMembersAcceptResponseObject, error)
	V1OrganizationMemberRemove(ctx context.Context, request api.V1OrganizationMemberRemoveRequestObject) (api.V1OrganizationMemberRemoveResponseObject, error)
	V1OrganizationMemberInviteRevoke(ctx context.Context, request api.V1OrganizationMemberInviteRevokeRequestObject) (api.V1OrganizationMemberInviteRevokeResponseObject, error)
	V1OrganizationRolesCreate(ctx context.Context, request api.V1OrganizationRolesCreateRequestObject) (api.V1OrganizationRolesCreateResponseObject, error)
	V1OrganizationRoleGet(ctx context.Context, request api.V1OrganizationRoleGetRequestObject) (api.V1OrganizationRoleGetResponseObject, error)
	V1OrganizationRolesGet(ctx context.Context, request api.V1OrganizationRolesGetRequestObject) (api.V1OrganizationRolesGetResponseObject, error)
	V1OrganizationRoleUpdate(ctx context.Context, request api.V1OrganizationRoleUpdateRequestObject) (api.V1OrganizationRoleUpdateResponseObject, error)
	V1OrganizationRoleDelete(ctx context.Context, request api.V1OrganizationRoleDeleteRequestObject) (api.V1OrganizationRoleDeleteResponseObject, error)
	V1OrganizationRoleMembersGet(ctx context.Context, request api.V1OrganizationRoleMembersGetRequestObject) (api.V1OrganizationRoleMembersGetResponseObject, error)
	V1OrganizationRoleMembersAdd(ctx context.Context, request api.V1OrganizationRoleMembersAddRequestObject) (api.V1OrganizationRoleMembersAddResponseObject, error)
	V1OrganizationRoleMemberRemove(ctx context.Context, request api.V1OrganizationRoleMemberRemoveRequestObject) (api.V1OrganizationRoleMemberRemoveResponseObject, error)
	V1OrganizationRolePermissionsGet(ctx context.Context, request api.V1OrganizationRolePermissionsGetRequestObject) (api.V1OrganizationRolePermissionsGetResponseObject, error)
	V1OrganizationRolePermissionAdd(ctx context.Context, request api.V1OrganizationRolePermissionAddRequestObject) (api.V1OrganizationRolePermissionAddResponseObject, error)
	V1OrganizationRolePermissionRemove(ctx context.Context, request api.V1OrganizationRolePermissionRemoveRequestObject) (api.V1OrganizationRolePermissionRemoveResponseObject, error)
}

// organizationController is the concrete implementation of OrganizationController.
type organizationController struct {
	*baseController
}

func (c *organizationController) V1OrganizationsCreate(ctx context.Context, request api.V1OrganizationsCreateRequestObject) (api.V1OrganizationsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsCreate")
	defer span.End()

	ownedBy, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1OrganizationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	opts := createOrganizationJSONRequestBodyToCreateOrganizationOpts(request.Body)

	organization, err := c.organizationService.Create(ctx, ownedBy, opts)
	if err != nil {
		if errors.Is(err, model.ErrInvalidOrganizationDetails) {
			return api.V1OrganizationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationsCreate500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	return api.V1OrganizationsCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: organization.ID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationGet(ctx context.Context, request api.V1OrganizationGetRequestObject) (api.V1OrganizationGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	organization, err := c.organizationService.Get(ctx, organizationID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationGet200JSONResponse(organizationToDTO(organization)), nil
}

func (c *organizationController) V1OrganizationsGet(ctx context.Context, request api.V1OrganizationsGetRequestObject) (api.V1OrganizationsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsGet")
	defer span.End()

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.organizationService.List(ctx, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1OrganizationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	organizationsDTO := make([]api.Organization, len(page.Items))
	for i, organization := range page.Items {
		organizationsDTO[i] = organizationToDTO(organization)
	}

	return api.V1OrganizationsGet200JSONResponse{
		Items:    organizationsDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *organizationController) V1OrganizationUpdate(ctx context.Context, request api.V1OrganizationUpdateRequestObject) (api.V1OrganizationUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationUpdate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts, err := updateOrganizationJSONRequestBodyToUpdateOrganizationOpts(request.Body)
	if err != nil {
		return api.V1OrganizationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	organization, err := c.organizationService.Update(ctx, organizationID, opts)
	if err != nil {
		if isNotFoundError(err) {
			return api.V1OrganizationUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationUpdate200JSONResponse(organizationToDTO(organization)), nil
}

func (c *organizationController) V1OrganizationDelete(ctx context.Context, request api.V1OrganizationDeleteRequestObject) (api.V1OrganizationDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationDelete")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.organizationService.Delete(ctx, organizationID, pkg.DefaultPtr(request.Params.Force, false)); err != nil {
		if isNotFoundError(err) {
			return api.V1OrganizationDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationDelete204Response{}, nil
}

func (c *organizationController) V1OrganizationMembersGet(ctx context.Context, request api.V1OrganizationMembersGetRequestObject) (api.V1OrganizationMembersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	users, err := c.organizationService.ListMembers(ctx, organizationID, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1OrganizationMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMembersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationMembersGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMembersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	membersDTO := make([]api.OrganizationMember, len(users.Items))
	for i, member := range users.Items {
		membersDTO[i] = organizationMemberToDTO(member)
	}

	return api.V1OrganizationMembersGet200JSONResponse{
		Items:    membersDTO,
		PageInfo: pageInfoToDTO(users.PageInfo),
	}, nil
}

func (c *organizationController) V1OrganizationMembersAdd(ctx context.Context, request api.V1OrganizationMembersAddRequestObject) (api.V1OrganizationMembersAddResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersAdd")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.organizationService.AddMember(ctx, organizationID, userID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMembersAdd403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationMembersAdd404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMembersAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationMembersAdd201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: userID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationMembersInvite(ctx context.Context, request api.V1OrganizationMembersInviteRequestObject) (api.V1OrganizationMembersInviteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersInvite")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body == nil || request.Body.Email == "" {
		return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("email is required"))}, nil
	}

	email := string(request.Body.Email)

	// Parse optional role ID
	var roleID model.ID
	if request.Body.RoleId != nil && *request.Body.RoleId != "" {
		var err error
		roleID, err = model.NewIDFromString(*request.Body.RoleId, model.ResourceTypeRole.String())
		if err != nil {
			return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
	}

	inviteOpts := service.InviteOrganizationMemberOpts{
		Email:  email,
		RoleID: roleID,
	}

	// Get user by email to return their ID (will be created if doesn't exist)
	user, err := c.userService.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return api.V1OrganizationMembersInvite500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	var userID model.ID
	if errors.Is(err, repository.ErrNotFound) {
		// User will be created by InviteMember, we'll get it after
		// For now, we'll call InviteMember and then get the user
		inviteErr := c.organizationService.InviteMember(ctx, organizationID, inviteOpts)
		if inviteErr != nil {
			if errors.Is(inviteErr, service.ErrNoPermission) {
				return api.V1OrganizationMembersInvite403JSONResponse{N403JSONResponse: permissionDenied}, nil
			}
			if errors.Is(inviteErr, service.ErrOrganizationMemberAlreadyExists) {
				return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(service.ErrOrganizationMemberAlreadyExists)}, nil
			}
			if errors.Is(inviteErr, service.ErrOrganizationMemberInvalidStatus) {
				return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(service.ErrOrganizationMemberInvalidStatus)}, nil
			}
			if isNotFoundError(inviteErr) {
				return api.V1OrganizationMembersInvite404JSONResponse{N404JSONResponse: notFound}, nil
			}
			return api.V1OrganizationMembersInvite500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: inviteErr.Error(),
			}}, nil
		}

		// Get the user after invitation
		user, err = c.userService.GetByEmail(ctx, email)
		if err != nil {
			return api.V1OrganizationMembersInvite500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
		userID = user.ID
	} else {
		userID = user.ID
		// Invite existing user
		inviteErr := c.organizationService.InviteMember(ctx, organizationID, inviteOpts)
		if inviteErr != nil {
			if errors.Is(inviteErr, service.ErrNoPermission) {
				return api.V1OrganizationMembersInvite403JSONResponse{N403JSONResponse: permissionDenied}, nil
			}
			if errors.Is(inviteErr, service.ErrOrganizationMemberAlreadyExists) {
				return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(service.ErrOrganizationMemberAlreadyExists)}, nil
			}
			if errors.Is(inviteErr, service.ErrOrganizationMemberInvalidStatus) {
				return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(service.ErrOrganizationMemberInvalidStatus)}, nil
			}
			if isNotFoundError(inviteErr) {
				return api.V1OrganizationMembersInvite404JSONResponse{N404JSONResponse: notFound}, nil
			}
			return api.V1OrganizationMembersInvite500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: inviteErr.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationMembersInvite201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: userID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationMemberRemove(ctx context.Context, request api.V1OrganizationMemberRemoveRequestObject) (api.V1OrganizationMemberRemoveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMemberRemove")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.organizationService.RemoveMember(ctx, organizationID, userID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMemberRemove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationMemberRemove404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMemberRemove500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationMemberRemove204Response{}, nil
}

func (c *organizationController) V1OrganizationMemberInviteRevoke(ctx context.Context, request api.V1OrganizationMemberInviteRevokeRequestObject) (api.V1OrganizationMemberInviteRevokeResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMemberInviteRevoke")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMemberInviteRevoke400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationMemberInviteRevoke400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.organizationService.RevokeInvitation(ctx, organizationID, userID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMemberInviteRevoke403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationMemberInviteRevoke404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMemberInviteRevoke500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationMemberInviteRevoke204Response{}, nil
}

func (c *organizationController) V1OrganizationMembersAccept(ctx context.Context, request api.V1OrganizationMembersAcceptRequestObject) (api.V1OrganizationMembersAcceptResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersAccept")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersAccept400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body == nil || request.Body.Token == "" {
		return api.V1OrganizationMembersAccept400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("token is required"))}, nil
	}

	token := string(request.Body.Token)
	userPassword := ""
	if request.Body.Password != nil {
		userPassword = string(*request.Body.Password)
	}

	if err := c.organizationService.AcceptInvitation(ctx, organizationID, service.AcceptOrganizationInvitationOpts{
		Token:    token,
		Password: userPassword,
	}); err != nil {
		if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrExpiredToken) {
			return api.V1OrganizationMembersAccept400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationMembersAccept404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMembersAccept500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationMembersAccept204Response{}, nil
}

func (c *organizationController) V1OrganizationRolesCreate(ctx context.Context, request api.V1OrganizationRolesCreateRequestObject) (api.V1OrganizationRolesCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRolesCreate")
	defer span.End()

	ownedBy, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1OrganizationRolesCreate400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRolesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts := createRoleJSONRequestBodyToCreateRoleOpts(request.Body)

	role, err := c.roleService.Create(ctx, ownedBy, organizationID, opts)
	if err != nil {
		if errors.Is(err, model.ErrInvalidRoleDetails) {
			return api.V1OrganizationRolesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRolesCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRolesCreate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRolesCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRolesCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: role.ID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationRolesGet(ctx context.Context, request api.V1OrganizationRolesGetRequestObject) (api.V1OrganizationRolesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRolesGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRolesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationRolesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roles, err := c.roleService.ListBelongsTo(ctx, organizationID, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1OrganizationRolesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRolesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRolesGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRolesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	rolesDTO := make([]api.Role, len(roles.Items))
	for i, role := range roles.Items {
		rolesDTO[i] = roleToDTO(role)
	}

	return api.V1OrganizationRolesGet200JSONResponse{
		Items:    rolesDTO,
		PageInfo: pageInfoToDTO(roles.PageInfo),
	}, nil
}

func (c *organizationController) V1OrganizationRoleGet(ctx context.Context, request api.V1OrganizationRoleGetRequestObject) (api.V1OrganizationRoleGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRoleGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	role, err := c.roleService.Get(ctx, roleID, organizationID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRoleGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRoleGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRoleGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRoleGet200JSONResponse(roleToDTO(role)), nil
}

func (c *organizationController) V1OrganizationRoleUpdate(ctx context.Context, request api.V1OrganizationRoleUpdateRequestObject) (api.V1OrganizationRoleUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleUpdate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRoleUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts := updateRoleJSONRequestBodyToUpdateRoleOpts(request.Body)

	role, err := c.roleService.Update(ctx, roleID, organizationID, opts)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRoleUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRoleUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRoleUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRoleUpdate200JSONResponse(roleToDTO(role)), nil
}

func (c *organizationController) V1OrganizationRoleDelete(ctx context.Context, request api.V1OrganizationRoleDeleteRequestObject) (api.V1OrganizationRoleDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleDelete")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRoleDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.roleService.Delete(ctx, roleID, organizationID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRoleDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRoleDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRoleDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRoleDelete204Response{}, nil
}

func (c *organizationController) V1OrganizationRoleMembersGet(ctx context.Context, request api.V1OrganizationRoleMembersGetRequestObject) (api.V1OrganizationRoleMembersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleMembersGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRoleMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationRoleMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	users, err := c.roleService.ListMembers(ctx, roleID, organizationID, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1OrganizationRoleMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRoleMembersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRoleMembersGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRoleMembersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	usersDTO := make([]api.User, len(users.Items))
	for i, user := range users.Items {
		usersDTO[i] = userToDTO(user)
	}

	return api.V1OrganizationRoleMembersGet200JSONResponse{
		Items:    usersDTO,
		PageInfo: pageInfoToDTO(users.PageInfo),
	}, nil
}

func (c *organizationController) V1OrganizationRoleMembersAdd(ctx context.Context, request api.V1OrganizationRoleMembersAddRequestObject) (api.V1OrganizationRoleMembersAddResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleMembersAdd")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRoleMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body == nil || request.Body.UserId == "" {
		return api.V1OrganizationRoleMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(fmt.Errorf("user_id is required"))}, nil
	}

	userID, err := model.NewIDFromString(request.Body.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationRoleMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.roleService.AddMember(ctx, roleID, userID, organizationID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRoleMembersAdd403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRoleMembersAdd404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRoleMembersAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRoleMembersAdd201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: userID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationRoleMemberRemove(ctx context.Context, request api.V1OrganizationRoleMemberRemoveRequestObject) (api.V1OrganizationRoleMemberRemoveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleMemberRemove")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRoleMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationRoleMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.roleService.RemoveMember(ctx, roleID, userID, organizationID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRoleMemberRemove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRoleMemberRemove404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRoleMemberRemove500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRoleMemberRemove204Response{}, nil
}

func (c *organizationController) V1OrganizationRolePermissionsGet(ctx context.Context, request api.V1OrganizationRolePermissionsGetRequestObject) (api.V1OrganizationRolePermissionsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRolePermissionsGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRolePermissionsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRolePermissionsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	permissions, err := c.roleService.GetPermissions(ctx, roleID, organizationID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRolePermissionsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRolePermissionsGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRolePermissionsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	permissionsDTO := make([]api.Permission, len(permissions))
	for i, perm := range permissions {
		permissionsDTO[i] = permissionToDTO(perm)
	}

	return api.V1OrganizationRolePermissionsGet200JSONResponse(permissionsDTO), nil
}

func (c *organizationController) V1OrganizationRolePermissionAdd(ctx context.Context, request api.V1OrganizationRolePermissionAddRequestObject) (api.V1OrganizationRolePermissionAddResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRolePermissionAdd")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body == nil {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(fmt.Errorf("request body is required"))}, nil
	}

	// Parse target string in format "ResourceType:id"
	parts := strings.Split(request.Body.Target, ":")
	if len(parts) != 2 {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(fmt.Errorf("invalid target format, expected ResourceType:id"))}, nil
	}

	targetID, err := model.NewIDFromString(parts[1], parts[0])
	if err != nil {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	var kind model.PermissionKind
	if err := kind.UnmarshalText([]byte(string(request.Body.Kind))); err != nil {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if !isOrganizationScopedResource(targetID.Type) {
		return api.V1OrganizationRolePermissionAdd400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidResourceType)}, nil
	}

	if err := c.roleService.AddPermission(ctx, roleID, organizationID, targetID, kind); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRolePermissionAdd403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRolePermissionAdd404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRolePermissionAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	permissions, err := c.permissionService.GetBySubjectAndTarget(ctx, roleID, targetID)
	if err != nil {
		return api.V1OrganizationRolePermissionAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: fmt.Errorf("failed to retrieve created permission: %w", err).Error(),
		}}, nil
	}

	var createdPermID string
	for _, p := range permissions {
		if p.Kind == kind {
			createdPermID = p.ID.String()
			break
		}
	}

	if createdPermID == "" {
		return api.V1OrganizationRolePermissionAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: "permission was created but could not be retrieved",
		}}, nil
	}

	return api.V1OrganizationRolePermissionAdd201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: createdPermID,
	}}, nil
}

func (c *organizationController) V1OrganizationRolePermissionRemove(ctx context.Context, request api.V1OrganizationRolePermissionRemoveRequestObject) (api.V1OrganizationRolePermissionRemoveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRolePermissionRemove")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRolePermissionRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeRole.String())
	if err != nil {
		return api.V1OrganizationRolePermissionRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	permissionID, err := model.NewIDFromString(request.PermissionId, model.ResourceTypePermission.String())
	if err != nil {
		return api.V1OrganizationRolePermissionRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.roleService.RemovePermission(ctx, roleID, organizationID, permissionID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationRolePermissionRemove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationRolePermissionRemove404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationRolePermissionRemove500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationRolePermissionRemove204Response{}, nil
}

// NewOrganizationController creates a new OrganizationController.
func NewOrganizationController(opts ...ControllerOption) (OrganizationController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &organizationController{
		baseController: c,
	}

	if controller.organizationService == nil {
		return nil, ErrNoOrganizationService
	}

	if controller.roleService == nil {
		return nil, ErrNoRoleService
	}

	return controller, nil
}

func createOrganizationJSONRequestBodyToCreateOrganizationOpts(body *api.V1OrganizationsCreateJSONRequestBody) service.CreateOrganizationOpts {
	opts := service.CreateOrganizationOpts{
		Name:  body.Name,
		Email: string(body.Email),
	}

	if body.Website != nil {
		opts.Website = *body.Website
	}

	if body.Logo != nil {
		opts.Logo = *body.Logo
	}

	return opts
}

func updateOrganizationJSONRequestBodyToUpdateOrganizationOpts(body *api.V1OrganizationUpdateJSONRequestBody) (service.UpdateOrganizationOpts, error) {
	opts := service.UpdateOrganizationOpts{}

	if body.Name != nil {
		opts.Name = optional.Some(*body.Name)
	}
	if body.Email != nil {
		opts.Email = optional.Some(string(*body.Email))
	}
	if body.Logo.Defined {
		opts.Logo = body.Logo
	}
	if body.Website.Defined {
		opts.Website = body.Website
	}
	if body.Status != nil {
		var status model.OrganizationStatus
		if err := status.UnmarshalText([]byte(string(*body.Status))); err != nil {
			return service.UpdateOrganizationOpts{}, err
		}
		opts.Status = optional.Some(status)
	}

	return opts, nil
}

func organizationToDTO(organization *service.Organization) api.Organization {
	o := api.Organization{
		Id:             organization.ID.String(),
		Email:          oapiTypes.Email(organization.Email),
		Name:           organization.Name,
		Logo:           &organization.Logo,
		Website:        &organization.Website,
		Status:         api.OrganizationStatus(organization.Status.String()),
		MemberCount:    organization.MemberCount,
		TeamCount:      organization.TeamCount,
		NamespaceCount: organization.NamespaceCount,
		DocumentCount:  organization.DocumentCount,
		CreatedAt:      *organization.CreatedAt,
		UpdatedAt:      organization.UpdatedAt,
	}

	return o
}

func organizationMemberToDTO(member *service.OrganizationMember) api.OrganizationMember {
	return api.OrganizationMember{
		Id:        member.ID.String(),
		FirstName: member.FirstName,
		LastName:  member.LastName,
		Email:     oapiTypes.Email(member.Email),
		Picture:   member.Picture,
		Status:    api.UserStatus(member.Status.String()),
		Roles:     member.Roles,
	}
}
