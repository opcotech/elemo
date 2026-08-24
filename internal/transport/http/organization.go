package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

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
	V1OrganizationTeamsGet(ctx context.Context, request api.V1OrganizationTeamsGetRequestObject) (api.V1OrganizationTeamsGetResponseObject, error)
	V1OrganizationTeamsCreate(ctx context.Context, request api.V1OrganizationTeamsCreateRequestObject) (api.V1OrganizationTeamsCreateResponseObject, error)
	V1OrganizationTeamGet(ctx context.Context, request api.V1OrganizationTeamGetRequestObject) (api.V1OrganizationTeamGetResponseObject, error)
	V1OrganizationTeamUpdate(ctx context.Context, request api.V1OrganizationTeamUpdateRequestObject) (api.V1OrganizationTeamUpdateResponseObject, error)
	V1OrganizationTeamDelete(ctx context.Context, request api.V1OrganizationTeamDeleteRequestObject) (api.V1OrganizationTeamDeleteResponseObject, error)
	V1OrganizationTeamMembersGet(ctx context.Context, request api.V1OrganizationTeamMembersGetRequestObject) (api.V1OrganizationTeamMembersGetResponseObject, error)
	V1OrganizationTeamMembersAdd(ctx context.Context, request api.V1OrganizationTeamMembersAddRequestObject) (api.V1OrganizationTeamMembersAddResponseObject, error)
	V1OrganizationTeamMemberRemove(ctx context.Context, request api.V1OrganizationTeamMemberRemoveRequestObject) (api.V1OrganizationTeamMemberRemoveResponseObject, error)
}

// organizationController is the concrete implementation of OrganizationController.
type organizationController struct {
	*baseController
	organizationService service.OrganizationService
	roleService         service.RoleService
	teamService         service.TeamService
	userService         service.UserService
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
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1OrganizationsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1OrganizationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationMembersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationMembersGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationMembersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationMembersAdd403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationMembersAdd404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationMembersAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
			switch classifyServiceError(inviteErr) {
			case http.StatusBadRequest:
				return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(inviteErr)}, nil
			case http.StatusForbidden:
				return api.V1OrganizationMembersInvite403JSONResponse{N403JSONResponse: permissionDenied}, nil
			case http.StatusNotFound:
				return api.V1OrganizationMembersInvite404JSONResponse{N404JSONResponse: notFound}, nil
			default:
				return api.V1OrganizationMembersInvite500JSONResponse{N500JSONResponse: api.N500JSONResponse{
					Message: inviteErr.Error(),
				}}, nil
			}
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
			switch classifyServiceError(inviteErr) {
			case http.StatusBadRequest:
				return api.V1OrganizationMembersInvite400JSONResponse{N400JSONResponse: formatBadRequest(inviteErr)}, nil
			case http.StatusForbidden:
				return api.V1OrganizationMembersInvite403JSONResponse{N403JSONResponse: permissionDenied}, nil
			case http.StatusNotFound:
				return api.V1OrganizationMembersInvite404JSONResponse{N404JSONResponse: notFound}, nil
			default:
				return api.V1OrganizationMembersInvite500JSONResponse{N500JSONResponse: api.N500JSONResponse{
					Message: inviteErr.Error(),
				}}, nil
			}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationMemberRemove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationMemberRemove404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationMemberRemove500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationMemberInviteRevoke403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationMemberInviteRevoke404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationMemberInviteRevoke500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationMembersAccept400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusNotFound:
			return api.V1OrganizationMembersAccept404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationMembersAccept500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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

	opts, err := createRoleJSONRequestBodyToCreateRoleOpts(request.Body)
	if err != nil {
		return api.V1OrganizationRolesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	role, err := c.roleService.Create(ctx, ownedBy, organizationID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationRolesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationRolesCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationRolesCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationRolesCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationRolesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationRolesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationRolesGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationRolesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationRoleGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationRoleGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationRoleGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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

	opts, err := updateRoleJSONRequestBodyToUpdateRoleOpts(request.Body)
	if err != nil {
		return api.V1OrganizationRoleUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	role, err := c.roleService.Update(ctx, roleID, organizationID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationRoleUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationRoleUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationRoleUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationRoleDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationRoleDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationRoleDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationRoleDelete204Response{}, nil
}

func (c *organizationController) V1OrganizationTeamsCreate(ctx context.Context, request api.V1OrganizationTeamsCreateRequestObject) (api.V1OrganizationTeamsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamsCreate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts := createTeamJSONRequestBodyToCreateTeamOpts(request.Body)

	team, err := c.teamService.Create(ctx, organizationID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationTeamsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationTeamsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationTeamsCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: team.ID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationTeamsGet(ctx context.Context, request api.V1OrganizationTeamsGetRequestObject) (api.V1OrganizationTeamsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamsGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationTeamsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teams, err := c.teamService.ListBelongsTo(ctx, organizationID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationTeamsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationTeamsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	teamsDTO := make([]api.Team, len(teams.Items))
	for i, team := range teams.Items {
		teamsDTO[i] = teamToDTO(team)
	}

	return api.V1OrganizationTeamsGet200JSONResponse{
		Items:    teamsDTO,
		PageInfo: pageInfoToDTO(teams.PageInfo),
	}, nil
}

func (c *organizationController) V1OrganizationTeamGet(ctx context.Context, request api.V1OrganizationTeamGetRequestObject) (api.V1OrganizationTeamGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teamID, err := model.NewIDFromString(request.TeamId, model.ResourceTypeTeam.String())
	if err != nil {
		return api.V1OrganizationTeamGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	team, err := c.teamService.Get(ctx, teamID, organizationID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationTeamGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationTeamGet200JSONResponse(teamToDTO(team)), nil
}

func (c *organizationController) V1OrganizationTeamUpdate(ctx context.Context, request api.V1OrganizationTeamUpdateRequestObject) (api.V1OrganizationTeamUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamUpdate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teamID, err := model.NewIDFromString(request.TeamId, model.ResourceTypeTeam.String())
	if err != nil {
		return api.V1OrganizationTeamUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts := updateTeamJSONRequestBodyToUpdateTeamOpts(request.Body)

	team, err := c.teamService.Update(ctx, teamID, organizationID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationTeamUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationTeamUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationTeamUpdate200JSONResponse(teamToDTO(team)), nil
}

func (c *organizationController) V1OrganizationTeamDelete(ctx context.Context, request api.V1OrganizationTeamDeleteRequestObject) (api.V1OrganizationTeamDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamDelete")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teamID, err := model.NewIDFromString(request.TeamId, model.ResourceTypeTeam.String())
	if err != nil {
		return api.V1OrganizationTeamDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.teamService.Delete(ctx, teamID, organizationID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationTeamDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationTeamDelete204Response{}, nil
}

func (c *organizationController) V1OrganizationTeamMembersGet(ctx context.Context, request api.V1OrganizationTeamMembersGetRequestObject) (api.V1OrganizationTeamMembersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamMembersGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teamID, err := model.NewIDFromString(request.TeamId, model.ResourceTypeTeam.String())
	if err != nil {
		return api.V1OrganizationTeamMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationTeamMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	users, err := c.teamService.ListMembers(ctx, teamID, organizationID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationTeamMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationTeamMembersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamMembersGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamMembersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	usersDTO := make([]api.User, len(users.Items))
	for i, user := range users.Items {
		usersDTO[i] = userToDTO(user)
	}

	return api.V1OrganizationTeamMembersGet200JSONResponse{
		Items:    usersDTO,
		PageInfo: pageInfoToDTO(users.PageInfo),
	}, nil
}

func (c *organizationController) V1OrganizationTeamMembersAdd(ctx context.Context, request api.V1OrganizationTeamMembersAddRequestObject) (api.V1OrganizationTeamMembersAddResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamMembersAdd")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teamID, err := model.NewIDFromString(request.TeamId, model.ResourceTypeTeam.String())
	if err != nil {
		return api.V1OrganizationTeamMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body == nil || request.Body.UserId == "" {
		return api.V1OrganizationTeamMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(fmt.Errorf("user_id is required"))}, nil
	}

	userID, err := model.NewIDFromString(request.Body.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationTeamMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.teamService.AddMember(ctx, teamID, userID, organizationID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationTeamMembersAdd403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamMembersAdd404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamMembersAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationTeamMembersAdd201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: userID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationTeamMemberRemove(ctx context.Context, request api.V1OrganizationTeamMemberRemoveRequestObject) (api.V1OrganizationTeamMemberRemoveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationTeamMemberRemove")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationTeamMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	teamID, err := model.NewIDFromString(request.TeamId, model.ResourceTypeTeam.String())
	if err != nil {
		return api.V1OrganizationTeamMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationTeamMemberRemove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.teamService.RemoveMember(ctx, teamID, userID, organizationID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1OrganizationTeamMemberRemove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationTeamMemberRemove404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationTeamMemberRemove500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationTeamMemberRemove204Response{}, nil
}

// NewOrganizationController creates a new OrganizationController.
func NewOrganizationController(
	organizationService service.OrganizationService,
	roleService service.RoleService,
	teamService service.TeamService,
	userService service.UserService,
	opts ...ControllerOption,
) (OrganizationController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if organizationService == nil {
		return nil, ErrNoOrganizationService
	}

	if roleService == nil {
		return nil, ErrNoRoleService
	}

	if teamService == nil {
		return nil, ErrNoTeamService
	}

	if userService == nil {
		return nil, ErrNoUserService
	}

	return &organizationController{
		baseController:      c,
		organizationService: organizationService,
		roleService:         roleService,
		teamService:         teamService,
		userService:         userService,
	}, nil
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
