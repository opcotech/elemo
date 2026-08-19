package http

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// NamespaceController is a controller for namespace endpoints.
type NamespaceController interface {
	V1NamespacesGet(ctx context.Context, request api.V1NamespacesGetRequestObject) (api.V1NamespacesGetResponseObject, error)
	V1OrganizationsNamespacesCreate(ctx context.Context, request api.V1OrganizationsNamespacesCreateRequestObject) (api.V1OrganizationsNamespacesCreateResponseObject, error)
	V1OrganizationsNamespacesGet(ctx context.Context, request api.V1OrganizationsNamespacesGetRequestObject) (api.V1OrganizationsNamespacesGetResponseObject, error)
	V1NamespaceGet(ctx context.Context, request api.V1NamespaceGetRequestObject) (api.V1NamespaceGetResponseObject, error)
	V1NamespaceUpdate(ctx context.Context, request api.V1NamespaceUpdateRequestObject) (api.V1NamespaceUpdateResponseObject, error)
	V1NamespaceDelete(ctx context.Context, request api.V1NamespaceDeleteRequestObject) (api.V1NamespaceDeleteResponseObject, error)
}

// namespaceController is the concrete implementation of NamespaceController.
type namespaceController struct {
	*baseController
}

func (c *namespaceController) V1NamespacesGet(ctx context.Context, request api.V1NamespacesGetRequestObject) (api.V1NamespacesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesGet")
	defer span.End()

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1NamespacesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.namespaceService.ListAccessible(ctx, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1NamespacesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		return api.V1NamespacesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	namespacesDTO := make([]api.AccessibleNamespace, len(page.Items))
	for i, namespace := range page.Items {
		namespacesDTO[i] = accessibleNamespaceToDTO(namespace)
	}

	return api.V1NamespacesGet200JSONResponse{
		Items:    namespacesDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *namespaceController) V1OrganizationsNamespacesCreate(ctx context.Context, request api.V1OrganizationsNamespacesCreateRequestObject) (api.V1OrganizationsNamespacesCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsNamespacesCreate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsNamespacesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts := createNamespaceJSONRequestBodyToCreateNamespaceOpts(request.Body)

	namespace, err := c.namespaceService.Create(ctx, organizationID, opts)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsNamespacesCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationsNamespacesCreate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationsNamespacesCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationsNamespacesCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: namespace.ID.String(),
	}}, nil
}

func (c *namespaceController) V1OrganizationsNamespacesGet(ctx context.Context, request api.V1OrganizationsNamespacesGetRequestObject) (api.V1OrganizationsNamespacesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsNamespacesGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsNamespacesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationsNamespacesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.namespaceService.List(ctx, organizationID, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1OrganizationsNamespacesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsNamespacesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1OrganizationsNamespacesGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationsNamespacesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	namespacesDTO := make([]api.Namespace, len(page.Items))
	for i, namespace := range page.Items {
		namespacesDTO[i] = namespaceToDTO(namespace)
	}

	return api.V1OrganizationsNamespacesGet200JSONResponse{
		Items:    namespacesDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *namespaceController) V1NamespaceGet(ctx context.Context, request api.V1NamespaceGetRequestObject) (api.V1NamespaceGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespaceGet")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespaceGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	namespace, err := c.namespaceService.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NamespaceGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NamespaceGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NamespaceGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NamespaceGet200JSONResponse(namespaceToDTO(namespace)), nil
}

func (c *namespaceController) V1NamespaceUpdate(ctx context.Context, request api.V1NamespaceUpdateRequestObject) (api.V1NamespaceUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespaceUpdate")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespaceUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts := updateNamespaceJSONRequestBodyToUpdateNamespaceOpts(request.Body)

	namespace, err := c.namespaceService.Update(ctx, namespaceID, opts)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NamespaceUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NamespaceUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NamespaceUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NamespaceUpdate200JSONResponse(namespaceToDTO(namespace)), nil
}

func (c *namespaceController) V1NamespaceDelete(ctx context.Context, request api.V1NamespaceDeleteRequestObject) (api.V1NamespaceDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespaceDelete")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespaceDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.namespaceService.Delete(ctx, namespaceID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NamespaceDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NamespaceDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NamespaceDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NamespaceDelete204Response{}, nil
}

// NewNamespaceController creates a new NamespaceController.
func NewNamespaceController(opts ...ControllerOption) (NamespaceController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &namespaceController{
		baseController: c,
	}

	if controller.namespaceService == nil {
		return nil, ErrNoNamespaceService
	}

	return controller, nil
}

func createNamespaceJSONRequestBodyToCreateNamespaceOpts(body *api.V1OrganizationsNamespacesCreateJSONRequestBody) service.CreateNamespaceOpts {
	opts := service.CreateNamespaceOpts{
		Name: body.Name,
	}

	if body.Description.Defined && body.Description.Value != nil {
		opts.Description = *body.Description.Value
	}

	return opts
}

func updateNamespaceJSONRequestBodyToUpdateNamespaceOpts(body *api.V1NamespaceUpdateJSONRequestBody) service.UpdateNamespaceOpts {
	opts := service.UpdateNamespaceOpts{}

	if body.Name != nil {
		opts.Name = optional.Some(*body.Name)
	}
	if body.Description.Defined {
		opts.Description = body.Description
	}

	return opts
}

func namespaceToDTO(namespace *service.Namespace) api.Namespace {
	n := api.Namespace{
		Id:            namespace.ID.String(),
		Name:          namespace.Name,
		ProjectCount:  namespace.ProjectCount,
		DocumentCount: namespace.DocumentCount,
		CreatedAt:     *namespace.CreatedAt,
		UpdatedAt:     namespace.UpdatedAt,
	}

	if namespace.Description != "" {
		n.Description = &namespace.Description
	}

	return n
}

func accessibleNamespaceToDTO(namespace *service.AccessibleNamespace) api.AccessibleNamespace {
	n := api.AccessibleNamespace{
		Id:            namespace.ID.String(),
		Name:          namespace.Name,
		ProjectCount:  namespace.ProjectCount,
		DocumentCount: namespace.DocumentCount,
		CreatedAt:     *namespace.CreatedAt,
		UpdatedAt:     namespace.UpdatedAt,
		Organization: api.PartialOrganization{
			Id:   namespace.Organization.ID.String(),
			Name: namespace.Organization.Name,
		},
	}

	if namespace.Description != "" {
		n.Description = &namespace.Description
	}

	return n
}
