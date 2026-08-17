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

// IssueController is a controller for issue endpoints.
type IssueController interface {
	V1ProjectsIssuesCreate(ctx context.Context, request api.V1ProjectsIssuesCreateRequestObject) (api.V1ProjectsIssuesCreateResponseObject, error)
	V1ProjectsIssuesGet(ctx context.Context, request api.V1ProjectsIssuesGetRequestObject) (api.V1ProjectsIssuesGetResponseObject, error)
	V1NamespacesIssuesGet(ctx context.Context, request api.V1NamespacesIssuesGetRequestObject) (api.V1NamespacesIssuesGetResponseObject, error)
	V1UsersIssuesGet(ctx context.Context, request api.V1UsersIssuesGetRequestObject) (api.V1UsersIssuesGetResponseObject, error)
	V1NamespacesIssuesKeyGet(ctx context.Context, request api.V1NamespacesIssuesKeyGetRequestObject) (api.V1NamespacesIssuesKeyGetResponseObject, error)
	V1IssueGet(ctx context.Context, request api.V1IssueGetRequestObject) (api.V1IssueGetResponseObject, error)
	V1IssueUpdate(ctx context.Context, request api.V1IssueUpdateRequestObject) (api.V1IssueUpdateResponseObject, error)
	V1IssueDelete(ctx context.Context, request api.V1IssueDeleteRequestObject) (api.V1IssueDeleteResponseObject, error)
	V1IssueRelationsGet(ctx context.Context, request api.V1IssueRelationsGetRequestObject) (api.V1IssueRelationsGetResponseObject, error)
	V1IssueRelationsCreate(ctx context.Context, request api.V1IssueRelationsCreateRequestObject) (api.V1IssueRelationsCreateResponseObject, error)
	V1IssueRelationUpdate(ctx context.Context, request api.V1IssueRelationUpdateRequestObject) (api.V1IssueRelationUpdateResponseObject, error)
	V1IssueRelationDelete(ctx context.Context, request api.V1IssueRelationDeleteRequestObject) (api.V1IssueRelationDeleteResponseObject, error)
}

// issueController is the concrete implementation of IssueController.
type issueController struct {
	*baseController
}

func (c *issueController) V1ProjectsIssuesCreate(ctx context.Context, request api.V1ProjectsIssuesCreateRequestObject) (api.V1ProjectsIssuesCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectsIssuesCreate")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectsIssuesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1ProjectsIssuesCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	opts, err := createIssueJSONRequestBodyToCreateIssueOpts(request.Body)
	if err != nil {
		return api.V1ProjectsIssuesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	issue, err := c.issueService.Create(ctx, projectID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ProjectsIssuesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ProjectsIssuesCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectsIssuesCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectsIssuesCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectsIssuesCreate201JSONResponse(issueToDTO(issue)), nil
}

func (c *issueController) V1ProjectsIssuesGet(ctx context.Context, request api.V1ProjectsIssuesGetRequestObject) (api.V1ProjectsIssuesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectsIssuesGet")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectsIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1ProjectsIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.issueService.List(ctx, projectID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1ProjectsIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1ProjectsIssuesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectsIssuesGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectsIssuesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	issuesDTO := make([]api.PartialIssue, len(page.Items))
	for i, issue := range page.Items {
		issuesDTO[i] = partialIssueToDTO(issue)
	}

	return api.V1ProjectsIssuesGet200JSONResponse{
		Items:    issuesDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *issueController) V1NamespacesIssuesGet(ctx context.Context, request api.V1NamespacesIssuesGetRequestObject) (api.V1NamespacesIssuesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesIssuesGet")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1NamespacesIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.issueService.ListByNamespace(ctx, namespaceID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesIssuesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesIssuesGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesIssuesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	issuesDTO := make([]api.PartialIssue, len(page.Items))
	for i, issue := range page.Items {
		issuesDTO[i] = partialIssueToDTO(issue)
	}

	return api.V1NamespacesIssuesGet200JSONResponse{
		Items:    issuesDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *issueController) V1UsersIssuesGet(ctx context.Context, request api.V1UsersIssuesGetRequestObject) (api.V1UsersIssuesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UsersIssuesGet")
	defer span.End()

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1UsersIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1UsersIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.issueService.ListByUser(ctx, userID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1UsersIssuesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1UsersIssuesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1UsersIssuesGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1UsersIssuesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	issuesDTO := make([]api.PartialIssue, len(page.Items))
	for i, issue := range page.Items {
		issuesDTO[i] = partialIssueToDTO(issue)
	}

	return api.V1UsersIssuesGet200JSONResponse{
		Items:    issuesDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *issueController) V1IssueGet(ctx context.Context, request api.V1IssueGetRequestObject) (api.V1IssueGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueGet")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	issue, err := c.issueService.Get(ctx, issueID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssueGet200JSONResponse(issueToDTO(issue)), nil
}

func (c *issueController) V1NamespacesIssuesKeyGet(ctx context.Context, request api.V1NamespacesIssuesKeyGetRequestObject) (api.V1NamespacesIssuesKeyGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesIssuesKeyGet")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesIssuesKeyGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if _, _, err := model.ParseIssueKey(request.Key); err != nil {
		return api.V1NamespacesIssuesKeyGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	issue, err := c.issueService.GetByKey(ctx, namespaceID, request.Key)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesIssuesKeyGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesIssuesKeyGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesIssuesKeyGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesIssuesKeyGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1NamespacesIssuesKeyGet200JSONResponse(issueToDTO(issue)), nil
}

func (c *issueController) V1IssueUpdate(ctx context.Context, request api.V1IssueUpdateRequestObject) (api.V1IssueUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueUpdate")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1IssueUpdate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(request.Body)
	if err != nil {
		return api.V1IssueUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	issue, err := c.issueService.Update(ctx, issueID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssueUpdate200JSONResponse(issueToDTO(issue)), nil
}

func (c *issueController) V1IssueDelete(ctx context.Context, request api.V1IssueDeleteRequestObject) (api.V1IssueDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueDelete")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.issueService.Delete(ctx, issueID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssueDelete204Response{}, nil
}

func (c *issueController) V1IssueRelationsGet(ctx context.Context, request api.V1IssueRelationsGetRequestObject) (api.V1IssueRelationsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueRelationsGet")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1IssueRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.issueService.ListRelations(ctx, issueID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueRelationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueRelationsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueRelationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	items := make([]api.IssueRelation, len(page.Items))
	for i, relation := range page.Items {
		items[i] = issueRelationToDTO(relation)
	}

	return api.V1IssueRelationsGet200JSONResponse{
		Items:    items,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *issueController) V1IssueRelationsCreate(ctx context.Context, request api.V1IssueRelationsCreateRequestObject) (api.V1IssueRelationsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueRelationsCreate")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1IssueRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	relatedID, err := model.NewIDFromString(request.Body.RelatedId, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	kind, err := model.IssueRelationKindString(string(request.Body.Kind))
	if err != nil {
		return api.V1IssueRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	relation, err := c.issueService.AddRelation(ctx, issueID, relatedID, kind)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueRelationsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueRelationsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueRelationsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssueRelationsCreate201JSONResponse(issueRelationToDTO(relation)), nil
}

func (c *issueController) V1IssueRelationUpdate(ctx context.Context, request api.V1IssueRelationUpdateRequestObject) (api.V1IssueRelationUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueRelationUpdate")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueRelationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	relationID, err := model.NewIDFromString(request.RelationId, model.ResourceTypeIssueRelation.String())
	if err != nil {
		return api.V1IssueRelationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if request.Body == nil {
		return api.V1IssueRelationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("request body is required"))}, nil
	}

	kind, err := model.IssueRelationKindString(string(request.Body.Kind))
	if err != nil {
		return api.V1IssueRelationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	relation, err := c.issueService.UpdateRelation(ctx, issueID, relationID, kind)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueRelationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueRelationUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueRelationUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueRelationUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssueRelationUpdate200JSONResponse(issueRelationToDTO(relation)), nil
}

func (c *issueController) V1IssueRelationDelete(ctx context.Context, request api.V1IssueRelationDeleteRequestObject) (api.V1IssueRelationDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1IssueRelationDelete")
	defer span.End()

	issueID, err := model.NewIDFromString(request.Id, model.ResourceTypeIssue.String())
	if err != nil {
		return api.V1IssueRelationDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	relationID, err := model.NewIDFromString(request.RelationId, model.ResourceTypeIssueRelation.String())
	if err != nil {
		return api.V1IssueRelationDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.issueService.RemoveRelation(ctx, issueID, relationID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1IssueRelationDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1IssueRelationDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1IssueRelationDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1IssueRelationDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1IssueRelationDelete204Response{}, nil
}

// NewIssueController creates a new IssueController.
func NewIssueController(opts ...ControllerOption) (IssueController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &issueController{
		baseController: c,
	}

	if controller.issueService == nil {
		return nil, ErrNoIssueService
	}

	return controller, nil
}

func createIssueJSONRequestBodyToCreateIssueOpts(body *api.V1ProjectsIssuesCreateJSONRequestBody) (service.CreateIssueOpts, error) {
	kind, err := model.IssueKindString(string(body.Kind))
	if err != nil {
		return service.CreateIssueOpts{}, err
	}

	opts := service.CreateIssueOpts{
		Kind:  kind,
		Title: body.Title,
	}

	if body.Parent != nil && *body.Parent != "" {
		parentID, err := model.NewIDFromString(*body.Parent, model.ResourceTypeIssue.String())
		if err != nil {
			return service.CreateIssueOpts{}, err
		}
		opts.Parent = &parentID
	}

	if body.Description.Defined && body.Description.Value != nil {
		opts.Description = *body.Description.Value
	}

	if body.Status != nil {
		status, err := model.IssueStatusString(string(*body.Status))
		if err != nil {
			return service.CreateIssueOpts{}, err
		}
		opts.Status = status
	}

	if body.Priority != nil {
		priority, err := model.IssuePriorityString(string(*body.Priority))
		if err != nil {
			return service.CreateIssueOpts{}, err
		}
		opts.Priority = priority
	}

	if body.Resolution != nil {
		resolution, err := model.IssueResolutionString(string(*body.Resolution))
		if err != nil {
			return service.CreateIssueOpts{}, err
		}
		opts.Resolution = resolution
	}

	if body.Links != nil {
		opts.Links = issueLinksFromAPI(*body.Links)
	}

	if body.DueDate != nil {
		opts.DueDate = body.DueDate
	}

	if body.StartDate != nil {
		opts.StartDate = body.StartDate
	}

	return opts, nil
}

func updateIssueJSONRequestBodyToUpdateIssueOpts(body *api.V1IssueUpdateJSONRequestBody) (service.UpdateIssueOpts, error) {
	opts := service.UpdateIssueOpts{}

	if body.Kind != nil {
		kind, err := model.IssueKindString(string(*body.Kind))
		if err != nil {
			return service.UpdateIssueOpts{}, err
		}
		opts.Kind = optional.Some(kind)
	}

	if body.Title.Defined {
		opts.Title = body.Title
	}

	if body.Description.Defined {
		opts.Description = body.Description
	}

	if body.Status != nil {
		status, err := model.IssueStatusString(string(*body.Status))
		if err != nil {
			return service.UpdateIssueOpts{}, err
		}
		opts.Status = optional.Some(status)
	}

	if body.Priority != nil {
		priority, err := model.IssuePriorityString(string(*body.Priority))
		if err != nil {
			return service.UpdateIssueOpts{}, err
		}
		opts.Priority = optional.Some(priority)
	}

	if body.Resolution != nil {
		resolution, err := model.IssueResolutionString(string(*body.Resolution))
		if err != nil {
			return service.UpdateIssueOpts{}, err
		}
		opts.Resolution = optional.Some(resolution)
	}

	if body.Links.Defined {
		if body.Links.Value == nil {
			opts.Links = optional.Null[[]model.IssueLink]()
		} else {
			opts.Links = optional.Some(issueLinksFromAPI(*body.Links.Value))
		}
	}

	if body.DueDate.Defined {
		opts.DueDate = body.DueDate
	}

	if body.StartDate.Defined {
		opts.StartDate = body.StartDate
	}

	if body.Assignees.Defined {
		assignees := make([]model.ID, 0)
		if body.Assignees.Value != nil {
			assignees = make([]model.ID, 0, len(*body.Assignees.Value))
			for _, assignee := range *body.Assignees.Value {
				id, err := model.NewIDFromString(assignee, model.ResourceTypeUser.String())
				if err != nil {
					return service.UpdateIssueOpts{}, err
				}
				assignees = append(assignees, id)
			}
		}
		opts.Assignees = optional.Some(assignees)
	}

	if body.Reviewers.Defined {
		reviewers := make([]model.ID, 0)
		if body.Reviewers.Value != nil {
			reviewers = make([]model.ID, 0, len(*body.Reviewers.Value))
			for _, reviewer := range *body.Reviewers.Value {
				id, err := model.NewIDFromString(reviewer, model.ResourceTypeUser.String())
				if err != nil {
					return service.UpdateIssueOpts{}, err
				}
				reviewers = append(reviewers, id)
			}
		}
		opts.Reviewers = optional.Some(reviewers)
	}

	if body.Labels.Defined {
		labels := make([]model.ID, 0)
		if body.Labels.Value != nil {
			labels = make([]model.ID, 0, len(*body.Labels.Value))
			for _, label := range *body.Labels.Value {
				id, err := model.NewIDFromString(label, model.ResourceTypeLabel.String())
				if err != nil {
					return service.UpdateIssueOpts{}, err
				}
				labels = append(labels, id)
			}
		}
		opts.Labels = optional.Some(labels)
	}

	if body.Parent.Defined {
		if body.Parent.Value == nil {
			opts.Parent = optional.Null[model.ID]()
		} else {
			parentID, err := model.NewIDFromString(*body.Parent.Value, model.ResourceTypeIssue.String())
			if err != nil {
				return service.UpdateIssueOpts{}, err
			}
			opts.Parent = optional.Some(parentID)
		}
	}

	return opts, nil
}

func issueToDTO(issue *service.Issue) api.Issue {
	createdAt := time.Time{}
	if issue.CreatedAt != nil {
		createdAt = *issue.CreatedAt
	}

	i := api.Issue{
		Id:              issue.ID.String(),
		Key:             issue.Key,
		NumericId:       int(issue.NumericID),
		Kind:            api.IssueKind(issue.Kind.String()),
		Title:           issue.Title,
		Status:          api.IssueStatus(issue.Status.String()),
		Priority:        api.IssuePriority(issue.Priority.String()),
		Resolution:      api.IssueResolution(issue.Resolution.String()),
		ReportedBy:      partialUserToDTO(issue.ReportedBy),
		Assignees:       assignmentUsersByKind(issue.Assignments, model.AssignmentKindAssignee),
		Reviewers:       assignmentUsersByKind(issue.Assignments, model.AssignmentKindReviewer),
		Labels:          partialLabelsToDTO(issue.Labels),
		Project:         partialProjectToDTO(issue.Project),
		Namespace:       partialNamespaceToDTO(issue.Namespace),
		CommentCount:    issue.CommentCount,
		DocumentCount:   issue.DocumentCount,
		AttachmentCount: issue.AttachmentCount,
		WatcherCount:    issue.WatcherCount,
		RelationCount:   issue.RelationCount,
		Links:           issueLinksToAPI(issue.Links),
		DueDate:         issue.DueDate,
		StartDate:       issue.StartDate,
		CreatedAt:       createdAt,
		UpdatedAt:       issue.UpdatedAt,
	}

	if issue.Description != "" {
		i.Description = &issue.Description
	}

	if issue.Parent != nil {
		parent := partialIssueToDTO(issue.Parent)
		i.Parent = &parent
	}

	if i.Links == nil {
		i.Links = make([]api.IssueLink, 0)
	}

	return i
}

func issueLinksFromAPI(links []api.IssueLink) []model.IssueLink {
	out := make([]model.IssueLink, len(links))
	for i, link := range links {
		out[i] = model.IssueLink{URL: link.Url, Label: link.Label}
	}
	return out
}

func issueLinksToAPI(links []model.IssueLink) []api.IssueLink {
	out := make([]api.IssueLink, len(links))
	for i, link := range links {
		out[i] = api.IssueLink{Url: link.URL, Label: link.Label}
	}
	return out
}

func issueRelationToDTO(relation *service.IssueRelation) api.IssueRelation {
	createdAt := time.Time{}
	if relation.CreatedAt != nil {
		createdAt = *relation.CreatedAt
	}

	related := api.PartialIssue{}
	if relation.Related != nil {
		related = partialIssueToDTO(relation.Related)
	}

	return api.IssueRelation{
		Id:        relation.ID.String(),
		Kind:      api.IssueRelationKind(relation.Kind.String()),
		Direction: api.IssueRelationDirection(relation.Direction.String()),
		Related:   related,
		CreatedAt: createdAt,
	}
}
