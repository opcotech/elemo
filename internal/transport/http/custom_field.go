package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// CustomFieldController is the controller for custom-field endpoints.
type CustomFieldController interface {
	V1CustomFieldsGet(ctx context.Context, request api.V1CustomFieldsGetRequestObject) (api.V1CustomFieldsGetResponseObject, error)
	V1CustomFieldsCreate(ctx context.Context, request api.V1CustomFieldsCreateRequestObject) (api.V1CustomFieldsCreateResponseObject, error)
	V1CustomFieldsSearch(ctx context.Context, request api.V1CustomFieldsSearchRequestObject) (api.V1CustomFieldsSearchResponseObject, error)
	V1CustomFieldGet(ctx context.Context, request api.V1CustomFieldGetRequestObject) (api.V1CustomFieldGetResponseObject, error)
	V1CustomFieldUpdate(ctx context.Context, request api.V1CustomFieldUpdateRequestObject) (api.V1CustomFieldUpdateResponseObject, error)
	V1CustomFieldDelete(ctx context.Context, request api.V1CustomFieldDeleteRequestObject) (api.V1CustomFieldDeleteResponseObject, error)
	V1CustomFieldArchive(ctx context.Context, request api.V1CustomFieldArchiveRequestObject) (api.V1CustomFieldArchiveResponseObject, error)
	V1ResourceCustomFieldsGet(ctx context.Context, request api.V1ResourceCustomFieldsGetRequestObject) (api.V1ResourceCustomFieldsGetResponseObject, error)
	V1ResourceCustomFieldValuePut(ctx context.Context, request api.V1ResourceCustomFieldValuePutRequestObject) (api.V1ResourceCustomFieldValuePutResponseObject, error)
	V1ResourceCustomFieldValueDelete(ctx context.Context, request api.V1ResourceCustomFieldValueDeleteRequestObject) (api.V1ResourceCustomFieldValueDeleteResponseObject, error)
}

type customFieldController struct {
	*baseController
	customFieldService service.CustomFieldService
}

func (c *customFieldController) V1CustomFieldsGet(
	ctx context.Context,
	request api.V1CustomFieldsGetRequestObject,
) (api.V1CustomFieldsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldsGet")
	defer span.End()

	scopeType, err := resourceTypeFromAPI(request.Params.ScopeType)
	if err != nil {
		return api.V1CustomFieldsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	scope, err := model.NewIDFromString(request.Params.ScopeId, scopeType.String())
	if err != nil {
		return api.V1CustomFieldsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	target, err := resourceTypeFromAPI(request.Params.TargetType)
	if err != nil {
		return api.V1CustomFieldsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	includeArchived := derefBool(request.Params.IncludeArchived)

	defs, err := c.customFieldService.ListDefinitions(ctx, scope, target, includeArchived)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1CustomFieldsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1CustomFieldsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	out := make([]api.CustomFieldDefinition, len(defs))
	for i, def := range defs {
		dto, err := customFieldDefinitionToDTO(def)
		if err != nil {
			return api.V1CustomFieldsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
		out[i] = dto
	}
	return api.V1CustomFieldsGet200JSONResponse(out), nil
}

func (c *customFieldController) V1CustomFieldsCreate(
	ctx context.Context,
	request api.V1CustomFieldsCreateRequestObject,
) (api.V1CustomFieldsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldsCreate")
	defer span.End()

	opts, err := createCustomFieldOptsFromAPI(request.Body)
	if err != nil {
		return api.V1CustomFieldsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	def, err := c.customFieldService.CreateDefinition(ctx, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusConflict:
			return api.V1CustomFieldsCreate409JSONResponse{N409JSONResponse: api.N409JSONResponse{Message: err.Error()}}, nil
		default:
			return api.V1CustomFieldsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	dto, err := customFieldDefinitionToDTO(def)
	if err != nil {
		return api.V1CustomFieldsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
	}
	return api.V1CustomFieldsCreate201JSONResponse(dto), nil
}

func (c *customFieldController) V1CustomFieldsSearch(
	ctx context.Context,
	request api.V1CustomFieldsSearchRequestObject,
) (api.V1CustomFieldsSearchResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldsSearch")
	defer span.End()

	query, err := customFieldSearchQueryFromAPI(request.Body)
	if err != nil {
		return api.V1CustomFieldsSearch400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	ids, err := c.customFieldService.Search(ctx, query)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldsSearch400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldsSearch403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1CustomFieldsSearch404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1CustomFieldsSearch500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.Composite()
	}
	return api.V1CustomFieldsSearch200JSONResponse{ResourceIds: out}, nil
}

func (c *customFieldController) V1CustomFieldGet(
	ctx context.Context,
	request api.V1CustomFieldGetRequestObject,
) (api.V1CustomFieldGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldGet")
	defer span.End()

	id, err := definitionIDFromAPI(string(request.Id))
	if err != nil {
		return api.V1CustomFieldGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	def, err := c.customFieldService.GetDefinition(ctx, id)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1CustomFieldGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1CustomFieldGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	dto, err := customFieldDefinitionToDTO(def)
	if err != nil {
		return api.V1CustomFieldGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
	}
	return api.V1CustomFieldGet200JSONResponse(dto), nil
}

func (c *customFieldController) V1CustomFieldUpdate(
	ctx context.Context,
	request api.V1CustomFieldUpdateRequestObject,
) (api.V1CustomFieldUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldUpdate")
	defer span.End()

	id, err := definitionIDFromAPI(string(request.Id))
	if err != nil {
		return api.V1CustomFieldUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	opts, err := updateCustomFieldOptsFromAPI(request.Body)
	if err != nil {
		return api.V1CustomFieldUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	def, err := c.customFieldService.UpdateDefinition(ctx, id, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1CustomFieldUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1CustomFieldUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	dto, err := customFieldDefinitionToDTO(def)
	if err != nil {
		return api.V1CustomFieldUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
	}
	return api.V1CustomFieldUpdate200JSONResponse(dto), nil
}

func (c *customFieldController) V1CustomFieldDelete(
	ctx context.Context,
	request api.V1CustomFieldDeleteRequestObject,
) (api.V1CustomFieldDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldDelete")
	defer span.End()

	id, err := definitionIDFromAPI(string(request.Id))
	if err != nil {
		return api.V1CustomFieldDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.customFieldService.DeleteDefinition(ctx, id); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1CustomFieldDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1CustomFieldDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1CustomFieldDelete204Response{}, nil
}

func (c *customFieldController) V1CustomFieldArchive(
	ctx context.Context,
	request api.V1CustomFieldArchiveRequestObject,
) (api.V1CustomFieldArchiveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1CustomFieldArchive")
	defer span.End()

	id, err := definitionIDFromAPI(string(request.Id))
	if err != nil {
		return api.V1CustomFieldArchive400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	def, err := c.customFieldService.ArchiveDefinition(ctx, id)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1CustomFieldArchive400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1CustomFieldArchive403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1CustomFieldArchive404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1CustomFieldArchive500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	dto, err := customFieldDefinitionToDTO(def)
	if err != nil {
		return api.V1CustomFieldArchive500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
	}
	return api.V1CustomFieldArchive200JSONResponse(dto), nil
}

func (c *customFieldController) V1ResourceCustomFieldsGet(
	ctx context.Context,
	request api.V1ResourceCustomFieldsGetRequestObject,
) (api.V1ResourceCustomFieldsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ResourceCustomFieldsGet")
	defer span.End()

	resourceID, err := resourceIDFromAPI(request.ResourceType, string(request.Id))
	if err != nil {
		return api.V1ResourceCustomFieldsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	entries, err := c.customFieldService.ListEffective(ctx, resourceID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ResourceCustomFieldsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ResourceCustomFieldsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ResourceCustomFieldsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ResourceCustomFieldsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	out := make([]api.CustomFieldEntry, len(entries))
	for i, entry := range entries {
		dto, err := customFieldEntryToDTO(entry)
		if err != nil {
			return api.V1ResourceCustomFieldsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
		out[i] = dto
	}
	return api.V1ResourceCustomFieldsGet200JSONResponse(out), nil
}

func (c *customFieldController) V1ResourceCustomFieldValuePut(
	ctx context.Context,
	request api.V1ResourceCustomFieldValuePutRequestObject,
) (api.V1ResourceCustomFieldValuePutResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ResourceCustomFieldValuePut")
	defer span.End()

	if request.Body == nil {
		return api.V1ResourceCustomFieldValuePut400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}
	resourceID, err := resourceIDFromAPI(request.ResourceType, string(request.Id))
	if err != nil {
		return api.V1ResourceCustomFieldValuePut400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	definitionID, err := definitionIDFromAPI(request.DefinitionId)
	if err != nil {
		return api.V1ResourceCustomFieldValuePut400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	value, err := customFieldValueFromAPI(*request.Body)
	if err != nil {
		return api.V1ResourceCustomFieldValuePut400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.customFieldService.SetValue(ctx, resourceID, definitionID, value); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ResourceCustomFieldValuePut400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ResourceCustomFieldValuePut403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ResourceCustomFieldValuePut404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ResourceCustomFieldValuePut500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1ResourceCustomFieldValuePut204Response{}, nil
}

func (c *customFieldController) V1ResourceCustomFieldValueDelete(
	ctx context.Context,
	request api.V1ResourceCustomFieldValueDeleteRequestObject,
) (api.V1ResourceCustomFieldValueDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ResourceCustomFieldValueDelete")
	defer span.End()

	resourceID, err := resourceIDFromAPI(request.ResourceType, string(request.Id))
	if err != nil {
		return api.V1ResourceCustomFieldValueDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	definitionID, err := definitionIDFromAPI(request.DefinitionId)
	if err != nil {
		return api.V1ResourceCustomFieldValueDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.customFieldService.DeleteValue(ctx, resourceID, definitionID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ResourceCustomFieldValueDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ResourceCustomFieldValueDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ResourceCustomFieldValueDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ResourceCustomFieldValueDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1ResourceCustomFieldValueDelete204Response{}, nil
}

// NewCustomFieldController creates a CustomFieldController.
func NewCustomFieldController(
	customFieldService service.CustomFieldService,
	opts ...ControllerOption,
) (CustomFieldController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}
	if customFieldService == nil {
		return nil, ErrNoCustomFieldService
	}
	return &customFieldController{
		baseController:     c,
		customFieldService: customFieldService,
	}, nil
}
