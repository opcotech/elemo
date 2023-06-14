package http

import (
	"context"

	"github.com/opcotech/elemo/internal/transport/http/api"
)

// PermissionController is a controller for system endpoints.
type PermissionController interface {
	V1PermissionsCreate(ctx context.Context, request api.V1PermissionsCreateRequestObject) (api.V1PermissionsCreateResponseObject, error)
	V1PermissionGet(ctx context.Context, request api.V1PermissionGetRequestObject) (api.V1PermissionGetResponseObject, error)
	V1PermissionUpdate(ctx context.Context, request api.V1PermissionUpdateRequestObject) (api.V1PermissionUpdateResponseObject, error)
	V1PermissionDelete(ctx context.Context, request api.V1PermissionDeleteRequestObject) (api.V1PermissionDeleteResponseObject, error)
	V1PermissionHasRelations(ctx context.Context, request api.V1PermissionHasRelationsRequestObject) (api.V1PermissionHasRelationsResponseObject, error)
	V1PermissionHasSystemRole(ctx context.Context, request api.V1PermissionHasSystemRoleRequestObject) (api.V1PermissionHasSystemRoleResponseObject, error)
	V1PermissionResourceGet(ctx context.Context, request api.V1PermissionResourceGetRequestObject) (api.V1PermissionResourceGetResponseObject, error)
}

// permissionController is the concrete implementation of PermissionController.
type permissionController struct {
	*baseController
}

func (c *permissionController) V1PermissionsCreate(ctx context.Context, request api.V1PermissionsCreateRequestObject) (api.V1PermissionsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionsCreate")
	defer span.End()

	// TODO implement me
	panic("implement me")
}

func (c *permissionController) V1PermissionGet(ctx context.Context, request api.V1PermissionGetRequestObject) (api.V1PermissionGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionGet")
	defer span.End()

	// TODO implement me
	panic("implement me")
}

func (c *permissionController) V1PermissionUpdate(ctx context.Context, request api.V1PermissionUpdateRequestObject) (api.V1PermissionUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionUpdate")
	defer span.End()

	// TODO implement me
	panic("implement me")
}

func (c *permissionController) V1PermissionDelete(ctx context.Context, request api.V1PermissionDeleteRequestObject) (api.V1PermissionDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionDelete")
	defer span.End()

	// TODO implement me
	panic("implement me")
}

func (c *permissionController) V1PermissionHasRelations(ctx context.Context, request api.V1PermissionHasRelationsRequestObject) (api.V1PermissionHasRelationsResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionHasRelations")
	defer span.End()

	// TODO implement me
	panic("implement me")
}

func (c *permissionController) V1PermissionHasSystemRole(ctx context.Context, request api.V1PermissionHasSystemRoleRequestObject) (api.V1PermissionHasSystemRoleResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionHasSystemRole")
	defer span.End()

	// TODO implement me
	panic("implement me")
}

func (c *permissionController) V1PermissionResourceGet(ctx context.Context, request api.V1PermissionResourceGetRequestObject) (api.V1PermissionResourceGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PermissionResourceGet")
	defer span.End()

	// TODO implement me
	panic("implement me")
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
