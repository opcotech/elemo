package http

import (
	"context"
	"errors"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
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
	V1OrganizationMemberRemove(ctx context.Context, request api.V1OrganizationMemberRemoveRequestObject) (api.V1OrganizationMemberRemoveResponseObject, error)
	V1OrganizationRolesCreate(ctx context.Context, request api.V1OrganizationRolesCreateRequestObject) (api.V1OrganizationRolesCreateResponseObject, error)
	V1OrganizationRoleGet(ctx context.Context, request api.V1OrganizationRoleGetRequestObject) (api.V1OrganizationRoleGetResponseObject, error)
	V1OrganizationRolesGet(ctx context.Context, request api.V1OrganizationRolesGetRequestObject) (api.V1OrganizationRolesGetResponseObject, error)
	V1OrganizationRoleUpdate(ctx context.Context, request api.V1OrganizationRoleUpdateRequestObject) (api.V1OrganizationRoleUpdateResponseObject, error)
	V1OrganizationRoleDelete(ctx context.Context, request api.V1OrganizationRoleDeleteRequestObject) (api.V1OrganizationRoleDeleteResponseObject, error)
	V1OrganizationRoleMembersGet(ctx context.Context, request api.V1OrganizationRoleMembersGetRequestObject) (api.V1OrganizationRoleMembersGetResponseObject, error)
	V1OrganizationRoleMembersAdd(ctx context.Context, request api.V1OrganizationRoleMembersAddRequestObject) (api.V1OrganizationRoleMembersAddResponseObject, error)
	V1OrganizationRoleMemberRemove(ctx context.Context, request api.V1OrganizationRoleMemberRemoveRequestObject) (api.V1OrganizationRoleMemberRemoveResponseObject, error)
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

	organization, err := createOrganizationJSONRequestBodyToOrganization(request.Body)
	if err != nil {
		return api.V1OrganizationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.organizationService.Create(ctx, ownedBy, organization); err != nil {
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

	organizations, err := c.organizationService.GetAll(ctx,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
	)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	organizationsDTO := make([]api.Organization, len(organizations))
	for i, organization := range organizations {
		organizationsDTO[i] = organizationToDTO(organization)
	}

	return api.V1OrganizationsGet200JSONResponse(organizationsDTO), nil
}

func (c *organizationController) V1OrganizationUpdate(ctx context.Context, request api.V1OrganizationUpdateRequestObject) (api.V1OrganizationUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationUpdate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return api.V1OrganizationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	organization, err := c.organizationService.Update(ctx, organizationID, patch)
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

	if err := c.organizationService.Delete(ctx, organizationID, pkg.GetDefaultPtr(request.Params.Force, false)); err != nil {
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

	users, err := c.organizationService.GetMembers(ctx, organizationID)
	if err != nil {
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

	membersDTO := make([]api.OrganizationMember, len(users))
	for i, member := range users {
		membersDTO[i] = organizationMemberToDTO(member)
	}

	return api.V1OrganizationMembersGet200JSONResponse(membersDTO), nil
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

	role, err := model.NewRole(request.Body.Name)
	if err != nil {
		return api.V1OrganizationRolesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	role.Description = pkg.GetDefaultPtr(request.Body.Description, "")

	if err := c.roleService.Create(ctx, ownedBy, organizationID, role); err != nil {
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

	roles, err := c.roleService.GetAllBelongsTo(ctx,
		organizationID,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
	)
	if err != nil {
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

	rolesDTO := make([]api.Role, len(roles))
	for i, role := range roles {
		rolesDTO[i] = roleToDTO(role)
	}

	return api.V1OrganizationRolesGet200JSONResponse(rolesDTO), nil
}

func (c *organizationController) V1OrganizationRoleGet(ctx context.Context, request api.V1OrganizationRoleGetRequestObject) (api.V1OrganizationRoleGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeUser.String())
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

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationRoleUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return api.V1OrganizationRoleUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	role, err := c.roleService.Update(ctx, roleID, organizationID, patch)
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

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeUser.String())
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

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationRoleMembersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	users, err := c.roleService.GetMembers(ctx, roleID, organizationID)
	if err != nil {
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

	usersDTO := make([]api.User, len(users))
	for i, user := range users {
		usersDTO[i] = userToDTO(user)
	}

	return api.V1OrganizationRoleMembersGet200JSONResponse(usersDTO), nil
}

func (c *organizationController) V1OrganizationRoleMembersAdd(ctx context.Context, request api.V1OrganizationRoleMembersAddRequestObject) (api.V1OrganizationRoleMembersAddResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationRoleMembersAdd")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationRoleMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationRoleMembersAdd400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
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

	roleID, err := model.NewIDFromString(request.RoleId, model.ResourceTypeUser.String())
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

func createOrganizationJSONRequestBodyToOrganization(body *api.V1OrganizationsCreateJSONRequestBody) (*model.Organization, error) {
	organization, err := model.NewOrganization(body.Name, string(body.Email))
	if err != nil {
		return nil, err
	}

	if body.Website != nil {
		organization.Website = *body.Website
	}

	if body.Logo != nil {
		organization.Logo = *body.Logo
	}

	return organization, nil
}

func organizationToDTO(organization *model.Organization) api.Organization {
	o := api.Organization{
		Id:         organization.ID.String(),
		Email:      oapiTypes.Email(organization.Email),
		Name:       organization.Name,
		Logo:       &organization.Logo,
		Website:    &organization.Website,
		Status:     api.OrganizationStatus(organization.Status.String()),
		Members:    make([]api.Id, len(organization.Members)),
		Namespaces: make([]api.Id, len(organization.Namespaces)),
		Teams:      make([]api.Id, len(organization.Teams)),
		CreatedAt:  *organization.CreatedAt,
		UpdatedAt:  organization.UpdatedAt,
	}

	for i, member := range organization.Members {
		o.Members[i] = api.Id(member.String())
	}

	for i, namespace := range organization.Namespaces {
		o.Namespaces[i] = api.Id(namespace.String())
	}

	for i, team := range organization.Teams {
		o.Teams[i] = api.Id(team.String())
	}

	return o
}

func organizationMemberToDTO(member *model.OrganizationMember) api.OrganizationMember {
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
