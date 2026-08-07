package http

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
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
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NamespacesProjectsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NamespacesProjectsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NamespacesProjectsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
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

	projects, err := c.projectService.GetAll(ctx, namespaceID,
		pkg.DefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.DefaultPtr(request.Params.Limit, DefaultLimit),
	)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NamespacesProjectsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NamespacesProjectsGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NamespacesProjectsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	projectsDTO := make([]api.Project, len(projects))
	for i, project := range projects {
		projectsDTO[i] = projectToDTO(project)
	}

	return api.V1NamespacesProjectsGet200JSONResponse(projectsDTO), nil
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
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1ProjectGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1ProjectGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1ProjectGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
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
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1ProjectUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1ProjectUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1ProjectUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
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
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1ProjectDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1ProjectDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1ProjectDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1ProjectDelete204Response{}, nil
}

// NewProjectController creates a new ProjectController.
func NewProjectController(opts ...ControllerOption) (ProjectController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &projectController{
		baseController: c,
	}

	if controller.projectService == nil {
		return nil, ErrNoProjectService
	}

	return controller, nil
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

func partialProjectToDTO(project *service.PartialProject) api.PartialProject {
	np := api.PartialProject{
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

func projectToDTO(project *service.Project) api.Project {
	p := api.Project{
		Id:        project.ID.String(),
		Key:       project.Key,
		Name:      project.Name,
		Status:    api.ProjectStatus(project.Status.String()),
		Teams:     make([]string, len(project.Teams)),
		Documents: make([]api.PartialDocument, len(project.Documents)),
		Issues:    make([]string, len(project.Issues)),
		CreatedAt: *project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}

	if project.Description != "" {
		p.Description = &project.Description
	}

	if project.Logo != "" {
		p.Logo = &project.Logo
	}

	for i, team := range project.Teams {
		p.Teams[i] = team.String()
	}

	for i, document := range project.Documents {
		p.Documents[i] = partialDocumentToDTO(document)
	}

	for i, issue := range project.Issues {
		p.Issues[i] = issue.String()
	}

	return p
}
