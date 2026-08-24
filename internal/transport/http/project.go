package http

import (
	"context"
	"net/http"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// ProjectController is a controller for project endpoints.
type ProjectController interface {
	V1NamespacesProjectsCreate(ctx context.Context, request api.V1NamespacesProjectsCreateRequestObject) (api.V1NamespacesProjectsCreateResponseObject, error)
	V1NamespacesProjectsGet(ctx context.Context, request api.V1NamespacesProjectsGetRequestObject) (api.V1NamespacesProjectsGetResponseObject, error)
	V1ProjectGet(ctx context.Context, request api.V1ProjectGetRequestObject) (api.V1ProjectGetResponseObject, error)
	V1ProjectUpdate(ctx context.Context, request api.V1ProjectUpdateRequestObject) (api.V1ProjectUpdateResponseObject, error)
	V1ProjectDelete(ctx context.Context, request api.V1ProjectDeleteRequestObject) (api.V1ProjectDeleteResponseObject, error)
}

// projectController is the concrete implementation of ProjectController.
type projectController struct {
	*baseController
	projectService service.ProjectService
}

func (c *projectController) V1NamespacesProjectsCreate(ctx context.Context, request api.V1NamespacesProjectsCreateRequestObject) (api.V1NamespacesProjectsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesProjectsCreate")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesProjectsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts, err := createProjectJSONRequestBodyToCreateProjectOpts(request.Body)
	if err != nil {
		return api.V1NamespacesProjectsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	project, err := c.projectService.Create(ctx, namespaceID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1NamespacesProjectsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesProjectsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesProjectsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1NamespacesProjectsCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: project.ID.String(),
	}}, nil
}

func (c *projectController) V1NamespacesProjectsGet(ctx context.Context, request api.V1NamespacesProjectsGetRequestObject) (api.V1NamespacesProjectsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NamespacesProjectsGet")
	defer span.End()

	namespaceID, err := model.NewIDFromString(request.Id, model.ResourceTypeNamespace.String())
	if err != nil {
		return api.V1NamespacesProjectsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1NamespacesProjectsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.projectService.List(ctx, namespaceID, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1NamespacesProjectsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1NamespacesProjectsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1NamespacesProjectsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1NamespacesProjectsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	projectsDTO := make([]api.Project, len(page.Items))
	for i, project := range page.Items {
		projectsDTO[i] = projectToDTO(project)
	}

	return api.V1NamespacesProjectsGet200JSONResponse{
		Items:    projectsDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *projectController) V1ProjectGet(ctx context.Context, request api.V1ProjectGetRequestObject) (api.V1ProjectGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectGet")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	project, err := c.projectService.Get(ctx, projectID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1ProjectGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectGet200JSONResponse(projectToDTO(project)), nil
}

func (c *projectController) V1ProjectUpdate(ctx context.Context, request api.V1ProjectUpdateRequestObject) (api.V1ProjectUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectUpdate")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts, err := updateProjectJSONRequestBodyToUpdateProjectOpts(request.Body)
	if err != nil {
		return api.V1ProjectUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	project, err := c.projectService.Update(ctx, projectID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1ProjectUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectUpdate200JSONResponse(projectToDTO(project)), nil
}

func (c *projectController) V1ProjectDelete(ctx context.Context, request api.V1ProjectDeleteRequestObject) (api.V1ProjectDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1ProjectDelete")
	defer span.End()

	projectID, err := model.NewIDFromString(request.Id, model.ResourceTypeProject.String())
	if err != nil {
		return api.V1ProjectDelete400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.projectService.Delete(ctx, projectID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1ProjectDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1ProjectDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1ProjectDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1ProjectDelete204Response{}, nil
}

// NewProjectController creates a new ProjectController.
func NewProjectController(projectService service.ProjectService, opts ...ControllerOption) (ProjectController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if projectService == nil {
		return nil, ErrNoProjectService
	}

	return &projectController{
		baseController: c,
		projectService: projectService,
	}, nil
}

func createProjectJSONRequestBodyToCreateProjectOpts(body *api.V1NamespacesProjectsCreateJSONRequestBody) (service.CreateProjectOpts, error) {
	opts := service.CreateProjectOpts{
		Key:  body.Key,
		Name: body.Name,
	}

	if body.Description.Defined && body.Description.Value != nil {
		opts.Description = *body.Description.Value
	}

	if body.Logo != nil {
		opts.Logo = *body.Logo
	}

	if body.Status != nil {
		status, err := model.ProjectStatusString(string(*body.Status))
		if err != nil {
			return service.CreateProjectOpts{}, err
		}
		opts.Status = status
	}

	return opts, nil
}

func updateProjectJSONRequestBodyToUpdateProjectOpts(body *api.V1ProjectUpdateJSONRequestBody) (service.UpdateProjectOpts, error) {
	opts := service.UpdateProjectOpts{}

	if body.Key != nil {
		opts.Key = optional.Some(*body.Key)
	}
	if body.Name != nil {
		opts.Name = optional.Some(*body.Name)
	}
	if body.Description.Defined {
		opts.Description = body.Description
	}
	if body.Logo.Defined {
		opts.Logo = body.Logo
	}
	if body.Status != nil {
		status, err := model.ProjectStatusString(string(*body.Status))
		if err != nil {
			return service.UpdateProjectOpts{}, err
		}
		opts.Status = optional.Some(status)
	}

	return opts, nil
}

func assignmentUsersByKind(assignments []service.PartialAssignee, kind model.AssignmentKind) []api.PartialUser {
	users := make([]api.PartialUser, 0, len(assignments))
	for _, assignment := range assignments {
		if assignment.Kind != kind {
			continue
		}
		user := api.PartialUser{
			Id:        assignment.ID.String(),
			FirstName: assignment.FirstName,
			LastName:  assignment.LastName,
		}
		if assignment.Picture != "" {
			picture := assignment.Picture
			user.Picture = &picture
		}
		users = append(users, user)
	}
	return users
}

func partialUserToDTO(user *service.PartialUser) api.PartialUser {
	if user == nil {
		return api.PartialUser{}
	}
	dto := api.PartialUser{
		Id:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	if user.Picture != "" {
		picture := user.Picture
		dto.Picture = &picture
	}
	return dto
}

func partialUserPtrToDTO(user *service.PartialUser) *api.PartialUser {
	if user == nil {
		return nil
	}
	dto := partialUserToDTO(user)
	return &dto
}

func partialLabelsToDTO(labels []service.PartialLabel) []api.PartialLabel {
	out := make([]api.PartialLabel, len(labels))
	for i, label := range labels {
		out[i] = api.PartialLabel{
			Id:   label.ID.String(),
			Name: label.Name,
		}
	}
	return out
}

func partialProjectToDTO(project *service.PartialProject) *api.PartialProject {
	if project == nil {
		return nil
	}
	dto := api.PartialProject{
		Id:     project.ID.String(),
		Key:    project.Key,
		Name:   project.Name,
		Status: api.ProjectStatus(project.Status.String()),
	}
	if project.Description != "" {
		dto.Description = &project.Description
	}
	if project.Logo != "" {
		dto.Logo = &project.Logo
	}
	return &dto
}

func partialNamespaceToDTO(namespace *service.PartialNamespace) *api.PartialNamespace {
	if namespace == nil {
		return nil
	}
	return &api.PartialNamespace{
		Id:   namespace.ID.String(),
		Name: namespace.Name,
	}
}

func partialIssueToDTO(issue *service.PartialIssue) api.PartialIssue {
	createdAt := time.Time{}
	if issue.CreatedAt != nil {
		createdAt = *issue.CreatedAt
	}

	ni := api.PartialIssue{
		Id:        issue.ID.String(),
		Key:       issue.Key,
		NumericId: int(issue.NumericID),
		Kind:      api.IssueKind(issue.Kind.String()),
		Title:     issue.Title,
		Status:    api.IssueStatus(issue.Status.String()),
		Priority:  api.IssuePriority(issue.Priority.String()),
		Assignees: assignmentUsersByKind(issue.Assignments, model.AssignmentKindAssignee),
		Reviewers: assignmentUsersByKind(issue.Assignments, model.AssignmentKindReviewer),
		Labels:    partialLabelsToDTO(issue.Labels),
		Project:   partialProjectToDTO(issue.Project),
		Namespace: partialNamespaceToDTO(issue.Namespace),
		DueDate:   issue.DueDate,
		StartDate: issue.StartDate,
		CreatedAt: createdAt,
		UpdatedAt: issue.UpdatedAt,
	}

	if issue.ReportedBy != nil {
		ni.ReportedBy = partialUserPtrToDTO(issue.ReportedBy)
	}

	if issue.Description != "" {
		ni.Description = &issue.Description
	}

	if issue.Parent != nil {
		parent := partialIssueToDTO(issue.Parent)
		ni.Parent = &parent
	}

	return ni
}

func projectToDTO(project *service.Project) api.Project {
	teams := project.Teams
	if teams == nil {
		teams = []model.ID{}
	}

	p := api.Project{
		Id:            project.ID.String(),
		Key:           project.Key,
		Name:          project.Name,
		Status:        api.ProjectStatus(project.Status.String()),
		Teams:         make([]string, len(teams)),
		DocumentCount: project.DocumentCount,
		IssueCount:    project.IssueCount,
		CreatedAt:     *project.CreatedAt,
		UpdatedAt:     project.UpdatedAt,
	}

	if project.Description != "" {
		p.Description = &project.Description
	}

	if project.Logo != "" {
		p.Logo = &project.Logo
	}

	for i, team := range teams {
		p.Teams[i] = team.String()
	}

	return p
}
