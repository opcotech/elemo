package http

import (
	"context"
	"errors"
	"strings"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// PermissionController is a controller for system endpoints.
type PermissionController interface {
	V1PermissionsCreate(ctx context.Context, request api.V1PermissionsCreateRequestObject) (api.V1PermissionsCreateResponseObject, error)
	V1PermissionGet(ctx context.Context, request api.V1PermissionGetRequestObject) (api.V1PermissionGetResponseObject, error)
	V1PermissionUpdate(ctx context.Context, request api.V1PermissionUpdateRequestObject) (api.V1PermissionUpdateResponseObject, error)
	V1PermissionDelete(ctx context.Context, request api.V1PermissionDeleteRequestObject) (api.V1PermissionDeleteResponseObject, error)
	V1PermissionResourceGet(ctx context.Context, request api.V1PermissionResourceGetRequestObject) (api.V1PermissionResourceGetResponseObject, error)
	V1PermissionHasRelations(ctx context.Context, request api.V1PermissionHasRelationsRequestObject) (api.V1PermissionHasRelationsResponseObject, error)
	V1PermissionHasSystemRole(ctx context.Context, request api.V1PermissionHasSystemRoleRequestObject) (api.V1PermissionHasSystemRoleResponseObject, error)
}

// permissionController is the concrete implementation of PermissionController.
type permissionController struct {
	*baseController
}

func (c *permissionController) V1PermissionsCreate(ctx context.Context, request api.V1PermissionsCreateRequestObject) (api.V1PermissionsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionsCreate")
	defer span.End()

	permission, err := createPermissionJSONRequestBodyToPermission(request.Body)
	if err != nil {
		return api.V1PermissionsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.permissionService.CtxUserCreate(ctx, permission); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1PermissionsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1PermissionsCreate500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	return api.V1PermissionsCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: permission.ID.String(),
	}}, nil
}

func (c *permissionController) V1PermissionGet(ctx context.Context, request api.V1PermissionGetRequestObject) (api.V1PermissionGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionGet")
	defer span.End()

	permissionID, err := model.NewIDFromString(request.Id, model.ResourceTypePermission.String())
	if err != nil {
		return api.V1PermissionGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	permission, err := c.permissionService.Get(ctx, permissionID)
	if err != nil {
		if isNotFoundError(err) {
			return api.V1PermissionGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1PermissionGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1PermissionGet200JSONResponse(permissionToDTO(permission)), nil
}

func (c *permissionController) V1PermissionUpdate(ctx context.Context, request api.V1PermissionUpdateRequestObject) (api.V1PermissionUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionUpdate")
	defer span.End()

	permissionID, err := model.NewIDFromString(request.Id, model.ResourceTypePermission.String())
	if err != nil {
		return api.V1PermissionUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	var kind model.PermissionKind
	if err := kind.UnmarshalText([]byte(request.Body.Kind)); err != nil {
		return api.V1PermissionUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	permission, err := c.permissionService.CtxUserUpdate(ctx, permissionID, kind)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1PermissionUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1PermissionUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1PermissionUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1PermissionUpdate200JSONResponse(permissionToDTO(permission)), nil
}

func (c *permissionController) V1PermissionDelete(ctx context.Context, request api.V1PermissionDeleteRequestObject) (api.V1PermissionDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionDelete")
	defer span.End()

	permissionID, err := model.NewIDFromString(request.Id, model.ResourceTypePermission.String())
	if err != nil {
		return api.V1PermissionDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.permissionService.CtxUserDelete(ctx, permissionID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1PermissionDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1PermissionDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1PermissionDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1PermissionDelete204Response{}, nil
}

func (c *permissionController) V1PermissionResourceGet(ctx context.Context, request api.V1PermissionResourceGetRequestObject) (api.V1PermissionResourceGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionResourceGet")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1PermissionResourceGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	parts := strings.Split(request.ResourceId, ":")
	if len(parts) != 2 {
		return api.V1PermissionResourceGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	id, err := model.NewIDFromString(parts[1], parts[0])
	if err != nil {
		return api.V1PermissionResourceGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	permissions, err := c.permissionService.GetBySubjectAndTarget(ctx, userID, id)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1PermissionResourceGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1PermissionResourceGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1PermissionResourceGet500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	permissionsDTO := make([]api.Permission, len(permissions))
	for i, permission := range permissions {
		permissionsDTO[i] = permissionToDTO(permission)
	}

	return api.V1PermissionResourceGet200JSONResponse(permissionsDTO), nil
}

func (c *permissionController) V1PermissionHasRelations(ctx context.Context, request api.V1PermissionHasRelationsRequestObject) (api.V1PermissionHasRelationsResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionHasRelations")
	defer span.End()

	parts := strings.Split(request.ResourceId, ":")
	if len(parts) != 2 {
		return api.V1PermissionHasRelations400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	id, err := model.NewIDFromString(parts[1], parts[0])
	if err != nil {
		return api.V1PermissionHasRelations400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	return api.V1PermissionHasRelations200JSONResponse(c.permissionService.CtxUserHasAnyRelation(ctx, id)), nil
}

func (c *permissionController) V1PermissionHasSystemRole(ctx context.Context, request api.V1PermissionHasSystemRoleRequestObject) (api.V1PermissionHasSystemRoleResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionHasSystemRole")
	defer span.End()

	roles := make([]model.SystemRole, len(request.Params.Roles))
	for i, roleName := range request.Params.Roles {
		var role model.SystemRole
		if err := role.UnmarshalText([]byte(roleName)); err != nil {
			return api.V1PermissionHasSystemRole400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		roles[i] = role
	}

	return api.V1PermissionHasSystemRole200JSONResponse(c.permissionService.CtxUserHasSystemRole(ctx, roles...)), nil
}

// NewPermissionController creates a new PermissionController.
func NewPermissionController(opts ...ControllerOption) (PermissionController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &permissionController{
		baseController: c,
	}

	if controller.permissionService == nil {
		return nil, ErrNoSystemService
	}

	return controller, nil
}

func createPermissionJSONRequestBodyToPermission(body *api.V1PermissionsCreateJSONRequestBody) (*model.Permission, error) {
	var kind model.PermissionKind
	if err := kind.UnmarshalText([]byte(body.Kind)); err != nil {
		return nil, err
	}

	subject, err := model.NewIDFromString(body.Subject.Id, string(body.Subject.ResourceType))
	if err != nil {
		return nil, err
	}

	target, err := model.NewIDFromString(body.Target.Id, string(body.Target.ResourceType))
	if err != nil {
		return nil, err
	}

	permission, err := model.NewPermission(subject, target, kind)
	if err != nil {
		return nil, err
	}

	return permission, nil
}

func permissionToDTO(permission *model.Permission) api.Permission {
	return api.Permission{
		Id:        permission.ID.String(),
		Subject:   permission.Subject.String(),
		Target:    permission.Target.String(),
		Kind:      api.PermissionKind(permission.Kind.String()),
		CreatedAt: *permission.CreatedAt,
		UpdatedAt: permission.UpdatedAt,
	}
}
