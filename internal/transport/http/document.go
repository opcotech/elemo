package http

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// DocumentController is a controller for document list endpoints.
type DocumentController interface {
	V1ProjectsDocumentsGet(ctx context.Context, request api.V1ProjectsDocumentsGetRequestObject) (api.V1ProjectsDocumentsGetResponseObject, error)
	V1NamespacesDocumentsGet(ctx context.Context, request api.V1NamespacesDocumentsGetRequestObject) (api.V1NamespacesDocumentsGetResponseObject, error)
}

type documentController struct {
	*baseController
}

func (c *documentController) V1ProjectsDocumentsGet(ctx context.Context, request api.V1ProjectsDocumentsGetRequestObject) (api.V1ProjectsDocumentsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectsDocumentsGet")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1ProjectsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.documentService.ListBelongsTo(ctx, projectID, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1ProjectsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1ProjectsDocumentsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1ProjectsDocumentsGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1ProjectsDocumentsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1ProjectsDocumentsGet200JSONResponse(partialDocumentPageToDTO(page)), nil
}

func (c *documentController) V1NamespacesDocumentsGet(ctx context.Context, request api.V1NamespacesDocumentsGetRequestObject) (api.V1NamespacesDocumentsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesDocumentsGet")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1NamespacesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.documentService.ListBelongsTo(ctx, namespaceID, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1NamespacesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NamespacesDocumentsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NamespacesDocumentsGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NamespacesDocumentsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NamespacesDocumentsGet200JSONResponse(partialDocumentPageToDTO(page)), nil
}

func partialDocumentToDTO(document *service.PartialDocument) api.PartialDocument {
	nd := api.PartialDocument{
		Id:        document.ID.String(),
		Name:      document.Name,
		CreatedBy: partialUserToDTO(&document.CreatedBy),
		CreatedAt: document.CreatedAt,
	}

	if document.Excerpt != "" {
		nd.Excerpt = &document.Excerpt
	}

	return nd
}

func partialDocumentPageToDTO(page service.Page[*service.PartialDocument]) api.PartialDocumentPage {
	items := make([]api.PartialDocument, len(page.Items))
	for i, document := range page.Items {
		items[i] = partialDocumentToDTO(document)
	}
	return api.PartialDocumentPage{
		Items:    items,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}
}

// NewDocumentController creates a new DocumentController.
func NewDocumentController(opts ...ControllerOption) (DocumentController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &documentController{
		baseController: c,
	}

	if controller.documentService == nil {
		return nil, ErrNoDocumentService
	}

	return controller, nil
}
