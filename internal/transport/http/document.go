package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// DocumentController is a controller for document endpoints.
type DocumentController interface {
	V1ProjectsDocumentsGet(ctx context.Context, request api.V1ProjectsDocumentsGetRequestObject) (api.V1ProjectsDocumentsGetResponseObject, error)
	V1ProjectsDocumentsCreate(ctx context.Context, request api.V1ProjectsDocumentsCreateRequestObject) (api.V1ProjectsDocumentsCreateResponseObject, error)
	V1ProjectsDocumentsRelate(ctx context.Context, request api.V1ProjectsDocumentsRelateRequestObject) (api.V1ProjectsDocumentsRelateResponseObject, error)
	V1ProjectsDocumentsUnrelate(ctx context.Context, request api.V1ProjectsDocumentsUnrelateRequestObject) (api.V1ProjectsDocumentsUnrelateResponseObject, error)
	V1NamespacesDocumentsGet(ctx context.Context, request api.V1NamespacesDocumentsGetRequestObject) (api.V1NamespacesDocumentsGetResponseObject, error)
	V1NamespacesDocumentsCreate(ctx context.Context, request api.V1NamespacesDocumentsCreateRequestObject) (api.V1NamespacesDocumentsCreateResponseObject, error)
	V1OrganizationsDocumentsGet(ctx context.Context, request api.V1OrganizationsDocumentsGetRequestObject) (api.V1OrganizationsDocumentsGetResponseObject, error)
	V1OrganizationsDocumentsCreate(ctx context.Context, request api.V1OrganizationsDocumentsCreateRequestObject) (api.V1OrganizationsDocumentsCreateResponseObject, error)
	V1IssuesDocumentsGet(ctx context.Context, request api.V1IssuesDocumentsGetRequestObject) (api.V1IssuesDocumentsGetResponseObject, error)
	V1IssuesDocumentsCreate(ctx context.Context, request api.V1IssuesDocumentsCreateRequestObject) (api.V1IssuesDocumentsCreateResponseObject, error)
	V1IssuesDocumentsRelate(ctx context.Context, request api.V1IssuesDocumentsRelateRequestObject) (api.V1IssuesDocumentsRelateResponseObject, error)
	V1IssuesDocumentsUnrelate(ctx context.Context, request api.V1IssuesDocumentsUnrelateRequestObject) (api.V1IssuesDocumentsUnrelateResponseObject, error)
	V1DocumentGet(ctx context.Context, request api.V1DocumentGetRequestObject) (api.V1DocumentGetResponseObject, error)
	V1DocumentUpdate(ctx context.Context, request api.V1DocumentUpdateRequestObject) (api.V1DocumentUpdateResponseObject, error)
	V1DocumentDelete(ctx context.Context, request api.V1DocumentDeleteRequestObject) (api.V1DocumentDeleteResponseObject, error)
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

	page, err := c.documentService.ListRelated(ctx, projectID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ProjectsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ProjectsDocumentsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectsDocumentsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectsDocumentsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectsDocumentsGet200JSONResponse(partialDocumentPageToDTO(page)), nil
}

func (c *documentController) V1ProjectsDocumentsCreate(ctx context.Context, request api.V1ProjectsDocumentsCreateRequestObject) (api.V1ProjectsDocumentsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectsDocumentsCreate")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectsDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1ProjectsDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	doc, err := c.documentService.Create(ctx, projectID, createDocumentOpts(request.Body.Title, request.Body.Excerpt, request.Body.Content))
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ProjectsDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ProjectsDocumentsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectsDocumentsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectsDocumentsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectsDocumentsCreate201JSONResponse(documentToDTO(doc)), nil
}

func (c *documentController) V1ProjectsDocumentsRelate(ctx context.Context, request api.V1ProjectsDocumentsRelateRequestObject) (api.V1ProjectsDocumentsRelateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectsDocumentsRelate")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectsDocumentsRelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	documentID, err := model.NewIDFromString(request.DocumentId, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1ProjectsDocumentsRelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.documentService.Relate(ctx, documentID, projectID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ProjectsDocumentsRelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ProjectsDocumentsRelate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectsDocumentsRelate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectsDocumentsRelate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectsDocumentsRelate204Response{}, nil
}

func (c *documentController) V1ProjectsDocumentsUnrelate(ctx context.Context, request api.V1ProjectsDocumentsUnrelateRequestObject) (api.V1ProjectsDocumentsUnrelateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectsDocumentsUnrelate")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectsDocumentsUnrelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	documentID, err := model.NewIDFromString(request.DocumentId, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1ProjectsDocumentsUnrelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.documentService.Unrelate(ctx, documentID, projectID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ProjectsDocumentsUnrelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ProjectsDocumentsUnrelate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectsDocumentsUnrelate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectsDocumentsUnrelate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectsDocumentsUnrelate204Response{}, nil
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

	filter, err := libraryListFilterFromParams(request.Params.FolderId, request.Params.All)
	if err != nil {
		return api.V1NamespacesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.documentService.ListLibrary(ctx, namespaceID, filter, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesDocumentsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesDocumentsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesDocumentsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1NamespacesDocumentsGet200JSONResponse(partialDocumentPageToDTO(page)), nil
}

func (c *documentController) V1NamespacesDocumentsCreate(ctx context.Context, request api.V1NamespacesDocumentsCreateRequestObject) (api.V1NamespacesDocumentsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesDocumentsCreate")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1NamespacesDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	doc, err := c.documentService.Create(ctx, namespaceID, createDocumentOpts(request.Body.Title, request.Body.Excerpt, request.Body.Content))
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesDocumentsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesDocumentsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesDocumentsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1NamespacesDocumentsCreate201JSONResponse(documentToDTO(doc)), nil
}

func (c *documentController) V1OrganizationsDocumentsGet(ctx context.Context, request api.V1OrganizationsDocumentsGetRequestObject) (api.V1OrganizationsDocumentsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsDocumentsGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	filter, err := libraryListFilterFromParams(request.Params.FolderId, request.Params.All)
	if err != nil {
		return api.V1OrganizationsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.documentService.ListLibrary(ctx, organizationID, filter, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationsDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationsDocumentsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationsDocumentsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationsDocumentsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationsDocumentsGet200JSONResponse(partialDocumentPageToDTO(page)), nil
}

func (c *documentController) V1OrganizationsDocumentsCreate(ctx context.Context, request api.V1OrganizationsDocumentsCreateRequestObject) (api.V1OrganizationsDocumentsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsDocumentsCreate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1OrganizationsDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	doc, err := c.documentService.Create(ctx, organizationID, createDocumentOpts(request.Body.Title, request.Body.Excerpt, request.Body.Content))
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationsDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationsDocumentsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationsDocumentsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationsDocumentsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationsDocumentsCreate201JSONResponse(documentToDTO(doc)), nil
}

func (c *documentController) V1IssuesDocumentsGet(ctx context.Context, request api.V1IssuesDocumentsGetRequestObject) (api.V1IssuesDocumentsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssuesDocumentsGet")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssuesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1IssuesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.documentService.ListRelated(ctx, issueID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssuesDocumentsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssuesDocumentsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssuesDocumentsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssuesDocumentsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssuesDocumentsGet200JSONResponse(partialDocumentPageToDTO(page)), nil
}

func (c *documentController) V1IssuesDocumentsCreate(ctx context.Context, request api.V1IssuesDocumentsCreateRequestObject) (api.V1IssuesDocumentsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssuesDocumentsCreate")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssuesDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1IssuesDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	doc, err := c.documentService.Create(ctx, issueID, createDocumentOpts(request.Body.Title, request.Body.Excerpt, request.Body.Content))
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssuesDocumentsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssuesDocumentsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssuesDocumentsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssuesDocumentsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssuesDocumentsCreate201JSONResponse(documentToDTO(doc)), nil
}

func (c *documentController) V1IssuesDocumentsRelate(ctx context.Context, request api.V1IssuesDocumentsRelateRequestObject) (api.V1IssuesDocumentsRelateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssuesDocumentsRelate")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssuesDocumentsRelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	documentID, err := model.NewIDFromString(request.DocumentId, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1IssuesDocumentsRelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.documentService.Relate(ctx, documentID, issueID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssuesDocumentsRelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssuesDocumentsRelate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssuesDocumentsRelate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssuesDocumentsRelate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssuesDocumentsRelate204Response{}, nil
}

func (c *documentController) V1IssuesDocumentsUnrelate(ctx context.Context, request api.V1IssuesDocumentsUnrelateRequestObject) (api.V1IssuesDocumentsUnrelateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssuesDocumentsUnrelate")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssuesDocumentsUnrelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	documentID, err := model.NewIDFromString(request.DocumentId, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1IssuesDocumentsUnrelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.documentService.Unrelate(ctx, documentID, issueID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssuesDocumentsUnrelate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssuesDocumentsUnrelate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssuesDocumentsUnrelate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssuesDocumentsUnrelate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssuesDocumentsUnrelate204Response{}, nil
}

func (c *documentController) V1DocumentGet(ctx context.Context, request api.V1DocumentGetRequestObject) (api.V1DocumentGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1DocumentGet")
	defer span.End()

	documentID, err := model.NewIDFromString(request.Id, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1DocumentGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	doc, err := c.documentService.Get(ctx, documentID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1DocumentGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1DocumentGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1DocumentGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1DocumentGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1DocumentGet200JSONResponse(documentToDTO(doc)), nil
}

func (c *documentController) V1DocumentUpdate(ctx context.Context, request api.V1DocumentUpdateRequestObject) (api.V1DocumentUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1DocumentUpdate")
	defer span.End()

	documentID, err := model.NewIDFromString(request.Id, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1DocumentUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1DocumentUpdate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	opts, err := updateDocumentJSONRequestBodyToUpdateDocumentOpts(request.Body)
	if err != nil {
		return api.V1DocumentUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	doc, err := c.documentService.Update(ctx, documentID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1DocumentUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1DocumentUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1DocumentUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1DocumentUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1DocumentUpdate200JSONResponse(documentToDTO(doc)), nil
}

func (c *documentController) V1DocumentDelete(ctx context.Context, request api.V1DocumentDeleteRequestObject) (api.V1DocumentDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1DocumentDelete")
	defer span.End()

	documentID, err := model.NewIDFromString(request.Id, model.ResourceTypeDocument.String())
	if err != nil {
		return api.V1DocumentDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.documentService.Delete(ctx, documentID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1DocumentDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1DocumentDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1DocumentDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1DocumentDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1DocumentDelete204Response{}, nil
}

func createDocumentOpts(title string, excerpt api.Optional[string], content api.Optional[string]) service.CreateDocumentOpts {
	opts := service.CreateDocumentOpts{
		Title: title,
	}

	if excerpt.Defined && excerpt.Value != nil {
		opts.Excerpt = *excerpt.Value
	}

	if content.Defined && content.Value != nil {
		opts.Content = []byte(*content.Value)
	}

	return opts
}

func updateDocumentJSONRequestBodyToUpdateDocumentOpts(body *api.V1DocumentUpdateJSONRequestBody) (service.UpdateDocumentOpts, error) {
	opts := service.UpdateDocumentOpts{}

	if body.Title.Defined {
		opts.Title = body.Title
	}
	if body.Excerpt.Defined {
		opts.Excerpt = body.Excerpt
	}
	if body.Content.Defined {
		if body.Content.Value != nil {
			opts.Content = optional.Some([]byte(*body.Content.Value))
		} else {
			opts.Content = optional.Null[[]byte]()
		}
	}
	if body.LibraryId.Defined && body.LibraryId.Value != nil {
		libraryID, err := model.NewIDFromString(*body.LibraryId.Value, model.ResourceTypeOrganization.String())
		if err != nil {
			return service.UpdateDocumentOpts{}, err
		}
		opts.LibraryID = optional.Some(libraryID)
	}
	if body.FolderId.Defined {
		if body.FolderId.Value == nil {
			opts.FolderID = optional.Null[model.ID]()
		} else {
			folderID, err := model.NewIDFromString(*body.FolderId.Value, model.ResourceTypeFolder.String())
			if err != nil {
				return service.UpdateDocumentOpts{}, err
			}
			opts.FolderID = optional.Some(folderID)
		}
	}

	return opts, nil
}

func libraryListFilterFromParams(folderID *string, all *bool) (service.LibraryListFilter, error) {
	filter := service.LibraryListFilter{}
	if all != nil && *all {
		filter.All = true
		return filter, nil
	}
	if folderID != nil && *folderID != "" {
		id, err := model.NewIDFromString(*folderID, model.ResourceTypeFolder.String())
		if err != nil {
			return filter, err
		}
		filter.FolderID = &id
	}
	return filter, nil
}

func documentLibraryToDTO(library service.DocumentLibrary) api.DocumentLibrary {
	return api.DocumentLibrary{
		Id:   library.ID.String(),
		Type: api.DocumentLibraryType(library.Type.String()),
		Name: library.Name,
	}
}

func documentFolderToDTO(folder *service.DocumentFolder) *api.DocumentFolder {
	if folder == nil {
		return nil
	}
	dto := &api.DocumentFolder{
		Id:   folder.ID.String(),
		Name: folder.Name,
	}
	if folder.ParentID != nil {
		parentID := folder.ParentID.String()
		dto.ParentId = &parentID
	}
	return dto
}

func documentRelationsToDTO(relations []service.DocumentRelation) []api.DocumentRelation {
	out := make([]api.DocumentRelation, len(relations))
	for i, relation := range relations {
		out[i] = api.DocumentRelation{
			Id:   relation.ID.String(),
			Type: api.DocumentRelationType(relation.Type.String()),
			Name: relation.Name,
		}
	}
	return out
}

func documentToDTO(document *service.Document) api.Document {
	createdAt := time.Time{}
	if document.CreatedAt != nil {
		createdAt = *document.CreatedAt
	}

	labels := partialLabelsToDTO(document.Labels)
	if labels == nil {
		labels = make([]api.PartialLabel, 0)
	}

	relations := documentRelationsToDTO(document.Relations)
	if relations == nil {
		relations = make([]api.DocumentRelation, 0)
	}

	d := api.Document{
		Id:              document.ID.String(),
		Title:           document.Title,
		Content:         string(document.Content),
		CreatedBy:       partialUserToDTO(&document.CreatedBy),
		Library:         documentLibraryToDTO(document.Library),
		Folder:          documentFolderToDTO(document.Folder),
		Relations:       relations,
		Labels:          labels,
		CommentCount:    document.CommentCount,
		AttachmentCount: document.AttachmentCount,
		CreatedAt:       createdAt,
		UpdatedAt:       document.UpdatedAt,
	}

	if document.Excerpt != "" {
		d.Excerpt = &document.Excerpt
	}

	return d
}

func partialDocumentToDTO(document *service.PartialDocument) api.PartialDocument {
	nd := api.PartialDocument{
		Id:        document.ID.String(),
		Title:     document.Title,
		CreatedBy: partialUserToDTO(&document.CreatedBy),
		CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt,
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
