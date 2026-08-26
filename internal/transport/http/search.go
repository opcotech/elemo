package http

import (
	"context"
	"net/http"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// SearchController is the controller for the search endpoints.
type SearchController interface {
	V1SearchGet(ctx context.Context, request api.V1SearchGetRequestObject) (api.V1SearchGetResponseObject, error)
}

type searchController struct {
	*baseController
	searchService service.SearchService
}

func (c *searchController) V1SearchGet(ctx context.Context, request api.V1SearchGetRequestObject) (api.V1SearchGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1SearchGet")
	defer span.End()

	query, err := searchQueryFromParams(request.Params)
	if err != nil {
		return api.V1SearchGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.searchService.Search(ctx, query)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1SearchGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		default:
			return api.V1SearchGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	items := make([]api.SearchResult, len(page.Items))
	for i, item := range page.Items {
		items[i] = searchResultToDTO(item)
	}

	return api.V1SearchGet200JSONResponse{
		Items:    items,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func searchQueryFromParams(params api.V1SearchGetParams) (service.SearchQuery, error) {
	q := service.SearchQuery{
		PageSize: service.DefaultSearchPageSize,
	}
	if params.Q != nil {
		q.Text = *params.Q
	}
	if params.PageSize != nil {
		q.PageSize = *params.PageSize
	}
	if params.PageToken != nil {
		q.PageToken = params.PageToken
	}
	if params.Types != nil {
		types := make([]model.ResourceType, 0, len(*params.Types))
		for _, raw := range *params.Types {
			rt, err := model.ResourceTypeString(string(raw))
			if err != nil {
				return service.SearchQuery{}, err
			}
			types = append(types, rt)
		}
		q.Types = types
	}
	if params.OrganizationId != nil && *params.OrganizationId != "" {
		id, err := model.NewIDFromString(*params.OrganizationId, model.ResourceTypeOrganization.String())
		if err != nil {
			return service.SearchQuery{}, err
		}
		q.OrganizationID = &id
	}
	if params.NamespaceId != nil && *params.NamespaceId != "" {
		id, err := model.NewIDFromString(*params.NamespaceId, model.ResourceTypeNamespace.String())
		if err != nil {
			return service.SearchQuery{}, err
		}
		q.NamespaceID = &id
	}
	if params.ProjectId != nil && *params.ProjectId != "" {
		id, err := model.NewIDFromString(*params.ProjectId, model.ResourceTypeProject.String())
		if err != nil {
			return service.SearchQuery{}, err
		}
		q.ProjectID = &id
	}
	return q, nil
}

func searchResultToDTO(result *service.SearchResult) api.SearchResult {
	dto := api.SearchResult{
		Id:        result.ID.String(),
		Type:      api.SearchResultType(result.Type.String()),
		Title:     result.Title,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}
	if result.Subtitle != "" {
		dto.Subtitle = &result.Subtitle
	}
	if result.Key != "" {
		dto.Key = &result.Key
	}
	if result.OrganizationID != nil {
		id := result.OrganizationID.String()
		dto.OrganizationId = &id
	}
	if result.NamespaceID != nil {
		id := result.NamespaceID.String()
		dto.NamespaceId = &id
	}
	if result.ProjectID != nil {
		id := result.ProjectID.String()
		dto.ProjectId = &id
	}
	if result.OrganizationSlug != "" {
		dto.OrganizationSlug = &result.OrganizationSlug
	}
	if result.NamespaceSlug != "" {
		dto.NamespaceSlug = &result.NamespaceSlug
	}
	if dto.CreatedAt.IsZero() {
		dto.CreatedAt = time.Unix(0, 0).UTC()
	}
	return dto
}

// NewSearchController creates a new SearchController.
func NewSearchController(searchService service.SearchService, opts ...ControllerOption) (SearchController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if searchService == nil {
		return nil, ErrNoSearchService
	}

	return &searchController{
		baseController: c,
		searchService:  searchService,
	}, nil
}
