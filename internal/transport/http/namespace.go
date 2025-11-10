package http

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// NamespaceController is a controller for namespace endpoints.
type NamespaceController interface {
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

func (c *namespaceController) V1OrganizationsNamespacesCreate(ctx context.Context, request api.V1OrganizationsNamespacesCreateRequestObject) (api.V1OrganizationsNamespacesCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsNamespacesCreate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsNamespacesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	namespace, err := createNamespaceJSONRequestBodyToNamespace(request.Body)
	if err != nil {
		return api.V1OrganizationsNamespacesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.namespaceService.Create(ctx, organizationID, namespace); err != nil {
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

	namespaces, err := c.namespaceService.GetAll(ctx, organizationID,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
	)
	if err != nil {
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

	namespacesDTO := make([]api.Namespace, len(namespaces))
	for i, namespace := range namespaces {
		namespacesDTO[i] = namespaceToDTO(namespace)
	}

	return api.V1OrganizationsNamespacesGet200JSONResponse(namespacesDTO), nil
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

	patch, err := api.ConvertRequestToMap(request.Body)
	if err != nil {
		return api.V1NamespaceUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	namespace, err := c.namespaceService.Update(ctx, namespaceID, patch)
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

func createNamespaceJSONRequestBodyToNamespace(body *api.V1OrganizationsNamespacesCreateJSONRequestBody) (*model.Namespace, error) {
	namespace, err := model.NewNamespace(body.Name)
	if err != nil {
		return nil, err
	}

	if body.Description.Defined && body.Description.Value != nil {
		namespace.Description = *body.Description.Value
	}

	return namespace, nil
}

func namespaceProjectToDTO(project *model.NamespaceProject) api.NamespaceProject {
	np := api.NamespaceProject{
		Id:     project.ID.String(),
		Key:    project.Key,
		Name:   project.Name,
		Status: api.ProjectStatus(project.Status.String()),
	}

	if project.Description != "" {
		np.Description = &project.Description
	}

	if project.Logo != "" {
		np.Logo = &project.Logo
	}

	return np
}

func namespaceDocumentToDTO(document *model.NamespaceDocument) api.NamespaceDocument {
	nd := api.NamespaceDocument{
		Id:        document.ID.String(),
		Name:      document.Name,
		CreatedBy: document.CreatedBy.String(),
		CreatedAt: document.CreatedAt,
	}

	if document.Excerpt != "" {
		nd.Excerpt = &document.Excerpt
	}

	return nd
}

func namespaceToDTO(namespace *model.Namespace) api.Namespace {
	n := api.Namespace{
		Id:        namespace.ID.String(),
		Name:      namespace.Name,
		Projects:  make([]api.NamespaceProject, len(namespace.Projects)),
		Documents: make([]api.NamespaceDocument, len(namespace.Documents)),
		CreatedAt: *namespace.CreatedAt,
		UpdatedAt: namespace.UpdatedAt,
	}

	if namespace.Description != "" {
		n.Description = &namespace.Description
	}

	for i, project := range namespace.Projects {
		n.Projects[i] = namespaceProjectToDTO(project)
	}

	for i, document := range namespace.Documents {
		n.Documents[i] = namespaceDocumentToDTO(document)
	}

	return n
}
