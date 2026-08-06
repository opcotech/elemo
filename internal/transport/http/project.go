package http

import (
	"context"

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

func (c *projectController) V1NamespacesProjectsCreate(_ context.Context, _ api.V1NamespacesProjectsCreateRequestObject) (api.V1NamespacesProjectsCreateResponseObject, error) {
	panic("not implemented")
}

func (c *projectController) V1NamespacesProjectsGet(_ context.Context, _ api.V1NamespacesProjectsGetRequestObject) (api.V1NamespacesProjectsGetResponseObject, error) {
	panic("not implemented")
}

func (c *projectController) V1ProjectGet(_ context.Context, _ api.V1ProjectGetRequestObject) (api.V1ProjectGetResponseObject, error) {
	panic("not implemented")
}

func (c *projectController) V1ProjectUpdate(_ context.Context, _ api.V1ProjectUpdateRequestObject) (api.V1ProjectUpdateResponseObject, error) {
	panic("not implemented")
}

func (c *projectController) V1ProjectDelete(_ context.Context, _ api.V1ProjectDeleteRequestObject) (api.V1ProjectDeleteResponseObject, error) {
	panic("not implemented")
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

	return controller, nil
}
