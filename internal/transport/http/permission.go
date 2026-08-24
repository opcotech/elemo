package http

import (
	"context"
	"net/http"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// PermissionController is a controller for grant endpoints.
type PermissionController interface {
	V1PermissionsCreate(ctx context.Context, request api.V1PermissionsCreateRequestObject) (api.V1PermissionsCreateResponseObject, error)
	V1PermissionGet(ctx context.Context, request api.V1PermissionGetRequestObject) (api.V1PermissionGetResponseObject, error)
	V1PermissionDelete(ctx context.Context, request api.V1PermissionDeleteRequestObject) (api.V1PermissionDeleteResponseObject, error)
	V1PermissionResourceGet(ctx context.Context, request api.V1PermissionResourceGetRequestObject) (api.V1PermissionResourceGetResponseObject, error)
}

// permissionController is the concrete implementation of PermissionController.
type permissionController struct {
	*baseController
	permissionService service.PermissionService
}

func (c *permissionController) V1PermissionsCreate(ctx context.Context, request api.V1PermissionsCreateRequestObject) (api.V1PermissionsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionsCreate")
	defer span.End()

	opts, err := createGrantJSONRequestBodyToCreateGrantOpts(request.Body)
	if err != nil {
		return api.V1PermissionsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	grant, err := c.permissionService.CtxUserCreate(ctx, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PermissionsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PermissionsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1PermissionsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1PermissionsCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: grant.ID.String(),
	}}, nil
}

func (c *permissionController) V1PermissionGet(ctx context.Context, request api.V1PermissionGetRequestObject) (api.V1PermissionGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionGet")
	defer span.End()

	grantID, err := model.NewIDFromString(request.Id, model.ResourceTypePermission.String())
	if err != nil {
		return api.V1PermissionGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	grant, err := c.permissionService.Get(ctx, grantID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusNotFound:
			return api.V1PermissionGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PermissionGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1PermissionGet200JSONResponse(grantToDTO(grant)), nil
}

func (c *permissionController) V1PermissionDelete(ctx context.Context, request api.V1PermissionDeleteRequestObject) (api.V1PermissionDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionDelete")
	defer span.End()

	grantID, err := model.NewIDFromString(request.Id, model.ResourceTypePermission.String())
	if err != nil {
		return api.V1PermissionDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.permissionService.CtxUserDelete(ctx, grantID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PermissionDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PermissionDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PermissionDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1PermissionDelete204Response{}, nil
}

func (c *permissionController) V1PermissionResourceGet(ctx context.Context, request api.V1PermissionResourceGetRequestObject) (api.V1PermissionResourceGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionResourceGet")
	defer span.End()

	id, err := model.ParseCompositeID(request.ResourceId)
	if err != nil {
		return api.V1PermissionResourceGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	actions, err := c.permissionService.CtxUserEffectiveActions(ctx, id)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PermissionResourceGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PermissionResourceGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PermissionResourceGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PermissionResourceGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1PermissionResourceGet200JSONResponse{
		Actions: actionStringsOrEmpty(actions),
	}, nil
}

// NewPermissionController creates a new PermissionController.
func NewPermissionController(permissionService service.PermissionService, opts ...ControllerOption) (PermissionController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if permissionService == nil {
		return nil, ErrNoPermissionService
	}

	return &permissionController{
		baseController:    c,
		permissionService: permissionService,
	}, nil
}

func createGrantJSONRequestBodyToCreateGrantOpts(body *api.V1PermissionsCreateJSONRequestBody) (service.CreateGrantOpts, error) {
	if body == nil {
		return service.CreateGrantOpts{}, model.ErrInvalidGrant
	}

	principal, err := model.NewIDFromString(body.Principal.Id, string(body.Principal.ResourceType))
	if err != nil {
		return service.CreateGrantOpts{}, err
	}

	scope, err := model.NewIDFromString(body.Scope.Id, string(body.Scope.ResourceType))
	if err != nil {
		return service.CreateGrantOpts{}, err
	}

	opts := service.CreateGrantOpts{
		Principal: principal,
		Scope:     scope,
	}

	if body.RoleId != nil && *body.RoleId != "" {
		roleID, err := model.NewIDFromString(*body.RoleId, model.ResourceTypeRole.String())
		if err != nil {
			return service.CreateGrantOpts{}, err
		}
		opts.RoleID = &roleID
	}

	if body.Actions != nil {
		actions, err := model.ParseActions(*body.Actions)
		if err != nil {
			return service.CreateGrantOpts{}, err
		}
		opts.Actions = actions
	}

	return opts, nil
}

func grantToDTO(grant *service.Grant) api.Grant {
	dto := api.Grant{
		Id:            grant.ID.String(),
		Principal:     grant.Principal.String(),
		PrincipalType: api.ResourceType(grant.Principal.Type.String()),
		Scope:         grant.Scope.String(),
		ScopeType:     api.ResourceType(grant.Scope.Type.String()),
		Actions:       actionStringsOrEmpty(grant.Actions),
		CreatedAt:     *grant.CreatedAt,
		UpdatedAt:     grant.UpdatedAt,
	}

	if grant.RoleID != nil {
		id := grant.RoleID.String()
		dto.RoleId = &id
	}

	return dto
}

func actionStringsOrEmpty(actions []model.Action) []api.Action {
	out := model.ActionStrings(actions)
	if out == nil {
		return []api.Action{}
	}
	return out
}
