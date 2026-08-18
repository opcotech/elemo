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

// FolderController is a controller for folder endpoints.
type FolderController interface {
	V1OrganizationsFoldersGet(ctx context.Context, request api.V1OrganizationsFoldersGetRequestObject) (api.V1OrganizationsFoldersGetResponseObject, error)
	V1OrganizationsFoldersCreate(ctx context.Context, request api.V1OrganizationsFoldersCreateRequestObject) (api.V1OrganizationsFoldersCreateResponseObject, error)
	V1NamespacesFoldersGet(ctx context.Context, request api.V1NamespacesFoldersGetRequestObject) (api.V1NamespacesFoldersGetResponseObject, error)
	V1NamespacesFoldersCreate(ctx context.Context, request api.V1NamespacesFoldersCreateRequestObject) (api.V1NamespacesFoldersCreateResponseObject, error)
	V1FolderGet(ctx context.Context, request api.V1FolderGetRequestObject) (api.V1FolderGetResponseObject, error)
	V1FolderUpdate(ctx context.Context, request api.V1FolderUpdateRequestObject) (api.V1FolderUpdateResponseObject, error)
	V1FolderDelete(ctx context.Context, request api.V1FolderDeleteRequestObject) (api.V1FolderDeleteResponseObject, error)
}

type folderController struct {
	*baseController
}

func (c *folderController) V1OrganizationsFoldersGet(ctx context.Context, request api.V1OrganizationsFoldersGetRequestObject) (api.V1OrganizationsFoldersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsFoldersGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1OrganizationsFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	parentID, err := optionalFolderIDFromQuery(request.Params.ParentId)
	if err != nil {
		return api.V1OrganizationsFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.folderService.List(ctx, organizationID, parentID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationsFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationsFoldersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationsFoldersGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationsFoldersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationsFoldersGet200JSONResponse(folderPageToDTO(page)), nil
}

func (c *folderController) V1OrganizationsFoldersCreate(ctx context.Context, request api.V1OrganizationsFoldersCreateRequestObject) (api.V1OrganizationsFoldersCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsFoldersCreate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationsFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1OrganizationsFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	opts, err := createFolderOptsFromBody(request.Body.Name, request.Body.ParentId)
	if err != nil {
		return api.V1OrganizationsFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	folder, err := c.folderService.Create(ctx, organizationID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1OrganizationsFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1OrganizationsFoldersCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1OrganizationsFoldersCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1OrganizationsFoldersCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1OrganizationsFoldersCreate201JSONResponse(folderToDTO(folder)), nil
}

func (c *folderController) V1NamespacesFoldersGet(ctx context.Context, request api.V1NamespacesFoldersGetRequestObject) (api.V1NamespacesFoldersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesFoldersGet")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1NamespacesFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	parentID, err := optionalFolderIDFromQuery(request.Params.ParentId)
	if err != nil {
		return api.V1NamespacesFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.folderService.List(ctx, namespaceID, parentID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesFoldersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesFoldersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesFoldersGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesFoldersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1NamespacesFoldersGet200JSONResponse(folderPageToDTO(page)), nil
}

func (c *folderController) V1NamespacesFoldersCreate(ctx context.Context, request api.V1NamespacesFoldersCreateRequestObject) (api.V1NamespacesFoldersCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesFoldersCreate")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1NamespacesFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	opts, err := createFolderOptsFromBody(request.Body.Name, request.Body.ParentId)
	if err != nil {
		return api.V1NamespacesFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	folder, err := c.folderService.Create(ctx, namespaceID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesFoldersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesFoldersCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesFoldersCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesFoldersCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1NamespacesFoldersCreate201JSONResponse(folderToDTO(folder)), nil
}

func (c *folderController) V1FolderGet(ctx context.Context, request api.V1FolderGetRequestObject) (api.V1FolderGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1FolderGet")
	defer span.End()

	folderID, err := model.NewIDFromString(request.Id, model.ResourceTypeFolder.String())
	if err != nil {
		return api.V1FolderGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	folder, err := c.folderService.Get(ctx, folderID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1FolderGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1FolderGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1FolderGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1FolderGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1FolderGet200JSONResponse(folderToDTO(folder)), nil
}

func (c *folderController) V1FolderUpdate(ctx context.Context, request api.V1FolderUpdateRequestObject) (api.V1FolderUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1FolderUpdate")
	defer span.End()

	folderID, err := model.NewIDFromString(request.Id, model.ResourceTypeFolder.String())
	if err != nil {
		return api.V1FolderUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1FolderUpdate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	opts, err := updateFolderOptsFromBody(request.Body)
	if err != nil {
		return api.V1FolderUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	folder, err := c.folderService.Update(ctx, folderID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1FolderUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1FolderUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1FolderUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1FolderUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1FolderUpdate200JSONResponse(folderToDTO(folder)), nil
}

func (c *folderController) V1FolderDelete(ctx context.Context, request api.V1FolderDeleteRequestObject) (api.V1FolderDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1FolderDelete")
	defer span.End()

	folderID, err := model.NewIDFromString(request.Id, model.ResourceTypeFolder.String())
	if err != nil {
		return api.V1FolderDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.folderService.Delete(ctx, folderID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1FolderDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1FolderDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1FolderDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1FolderDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1FolderDelete204Response{}, nil
}

func optionalFolderIDFromQuery(raw *string) (*model.ID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := model.NewIDFromString(*raw, model.ResourceTypeFolder.String())
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func createFolderOptsFromBody(name string, parentID *string) (service.CreateFolderOpts, error) {
	opts := service.CreateFolderOpts{Name: name}
	if parentID == nil || *parentID == "" {
		return opts, nil
	}
	id, err := model.NewIDFromString(*parentID, model.ResourceTypeFolder.String())
	if err != nil {
		return service.CreateFolderOpts{}, err
	}
	opts.ParentID = &id
	return opts, nil
}

func updateFolderOptsFromBody(body *api.V1FolderUpdateJSONRequestBody) (service.UpdateFolderOpts, error) {
	opts := service.UpdateFolderOpts{}
	if body.Name.Defined {
		opts.Name = body.Name
	}
	if body.ParentId.Defined {
		if body.ParentId.Value == nil {
			opts.ParentID = optional.Null[model.ID]()
		} else {
			parentID, err := model.NewIDFromString(*body.ParentId.Value, model.ResourceTypeFolder.String())
			if err != nil {
				return service.UpdateFolderOpts{}, err
			}
			opts.ParentID = optional.Some(parentID)
		}
	}
	return opts, nil
}

func folderToDTO(folder *service.Folder) api.Folder {
	createdAt := time.Time{}
	if folder.CreatedAt != nil {
		createdAt = *folder.CreatedAt
	}

	return api.Folder{
		Id:        folder.ID.String(),
		Name:      folder.Name,
		Library:   documentLibraryToDTO(folder.Library),
		Parent:    documentFolderToDTO(folder.Parent),
		CreatedBy: partialUserToDTO(&folder.CreatedBy),
		CreatedAt: createdAt,
		UpdatedAt: folder.UpdatedAt,
	}
}

func folderPageToDTO(page service.Page[*service.Folder]) api.FolderPage {
	items := make([]api.Folder, len(page.Items))
	for i, folder := range page.Items {
		items[i] = folderToDTO(folder)
	}
	return api.FolderPage{
		Items:    items,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}
}

// NewFolderController creates a new FolderController.
func NewFolderController(opts ...ControllerOption) (FolderController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &folderController{
		baseController: c,
	}

	if controller.folderService == nil {
		return nil, ErrNoFolderService
	}

	return controller, nil
}
