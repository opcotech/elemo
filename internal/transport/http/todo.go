package http

import (
	"context"
	"net/http"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// TodoController is the controller for the todo endpoints.
type TodoController interface {
	V1TodosCreate(ctx context.Context, request api.V1TodosCreateRequestObject) (api.V1TodosCreateResponseObject, error)
	V1TodoGet(ctx context.Context, request api.V1TodoGetRequestObject) (api.V1TodoGetResponseObject, error)
	V1TodosGet(ctx context.Context, request api.V1TodosGetRequestObject) (api.V1TodosGetResponseObject, error)
	V1TodoUpdate(ctx context.Context, request api.V1TodoUpdateRequestObject) (api.V1TodoUpdateResponseObject, error)
	V1TodoDelete(ctx context.Context, request api.V1TodoDeleteRequestObject) (api.V1TodoDeleteResponseObject, error)
}

// todoController is the concrete implementation of TodoController.
type todoController struct {
	*baseController
	todoService service.TodoService
}

func (c *todoController) V1TodosCreate(ctx context.Context, request api.V1TodosCreateRequestObject) (api.V1TodosCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodosCreate")
	defer span.End()

	createdBy, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1TodosCreate400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	ownerID, err := model.NewIDFromString(request.Body.OwnedBy, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1TodosCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts, err := createTodoJSONRequestBodyToCreateTodoOpts(request.Body, ownerID, createdBy)
	if err != nil {
		return api.V1TodosCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	todo, err := c.todoService.Create(ctx, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1TodosCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1TodosCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1TodosCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: todo.ID.String(),
	}}, nil
}

func (c *todoController) V1TodoGet(ctx context.Context, request api.V1TodoGetRequestObject) (api.V1TodoGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodoGet")
	defer span.End()

	todoID, err := model.NewIDFromString(request.Id, model.ResourceTypeTodo.String())
	if err != nil {
		return api.V1TodoGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	todo, err := c.todoService.Get(ctx, todoID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1TodoGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1TodoGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1TodoGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1TodoGet200JSONResponse(todoToDTO(todo)), nil
}

func (c *todoController) V1TodosGet(ctx context.Context, request api.V1TodosGetRequestObject) (api.V1TodosGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodosGet")
	defer span.End()

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1TodosGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.todoService.List(ctx, pageParams, request.Params.Completed)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1TodosGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1TodosGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1TodosGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	todosDTO := make([]api.Todo, len(page.Items))
	for i, todo := range page.Items {
		todosDTO[i] = todoToDTO(todo)
	}

	return api.V1TodosGet200JSONResponse{
		Items:    todosDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *todoController) V1TodoUpdate(ctx context.Context, request api.V1TodoUpdateRequestObject) (api.V1TodoUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodoUpdate")
	defer span.End()

	todoID, err := model.NewIDFromString(request.Id, model.ResourceTypeTodo.String())
	if err != nil {
		return api.V1TodoUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	opts, err := updateTodoJSONRequestBodyToUpdateTodoOpts(request.Body)
	if err != nil {
		return api.V1TodoUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	todo, err := c.todoService.Update(ctx, todoID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1TodoUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1TodoUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1TodoUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1TodoUpdate200JSONResponse(todoToDTO(todo)), nil
}

func (c *todoController) V1TodoDelete(ctx context.Context, request api.V1TodoDeleteRequestObject) (api.V1TodoDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodoDelete")
	defer span.End()

	todoID, err := model.NewIDFromString(request.Id, model.ResourceTypeTodo.String())
	if err != nil {
		return api.V1TodoDelete404JSONResponse{N404JSONResponse: notFound}, nil
	}

	if err := c.todoService.Delete(ctx, todoID); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1TodoDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1TodoDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1TodoDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1TodoDelete204Response{}, nil
}

// NewTodoController creates a new TodoController.
func NewTodoController(todoService service.TodoService, opts ...ControllerOption) (TodoController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if todoService == nil {
		return nil, ErrNoTodoService
	}

	controller := &todoController{
		baseController: c,
		todoService:    todoService,
	}

	return controller, nil
}

func createTodoJSONRequestBodyToCreateTodoOpts(body *api.V1TodosCreateJSONRequestBody, ownedBy, createdBy model.ID) (service.CreateTodoOpts, error) {
	opts := service.CreateTodoOpts{
		Title:       body.Title,
		Description: pkg.DefaultPtr(body.Description.Value, ""),
		Priority:    model.TodoPriorityNormal,
		OwnedBy:     ownedBy,
		CreatedBy:   createdBy,
	}

	if body.DueDate.Value != nil {
		opts.DueDate = *body.DueDate.Value
	}

	if err := opts.Priority.UnmarshalText([]byte(body.Priority)); err != nil {
		return service.CreateTodoOpts{}, err
	}

	return opts, nil
}

func updateTodoJSONRequestBodyToUpdateTodoOpts(body *api.V1TodoUpdateJSONRequestBody) (service.UpdateTodoOpts, error) {
	opts := service.UpdateTodoOpts{}

	if body.Title != nil {
		opts.Title = optional.Some(*body.Title)
	}
	if body.Completed != nil {
		opts.Completed = optional.Some(*body.Completed)
	}
	if body.Description.Defined {
		opts.Description = body.Description
	}
	if body.DueDate.Defined {
		opts.DueDate = optionalDueDateFromAPI(body.DueDate)
	}
	if body.Priority != nil {
		var priority model.TodoPriority
		if err := priority.UnmarshalText([]byte(*body.Priority)); err != nil {
			return service.UpdateTodoOpts{}, err
		}
		opts.Priority = optional.Some(priority)
	}

	return opts, nil
}

func optionalDueDateFromAPI(o optional.Optional[*time.Time]) optional.Optional[time.Time] {
	if !o.Defined {
		return optional.None[time.Time]()
	}
	if o.Value == nil || *o.Value == nil {
		return optional.Null[time.Time]()
	}
	return optional.Some(**o.Value)
}

func todoToDTO(todo *service.Todo) api.Todo {
	return api.Todo{
		Id:          todo.ID.String(),
		Title:       todo.Title,
		Completed:   todo.Completed,
		Priority:    api.TodoPriority(todo.Priority.String()),
		Description: todo.Description,
		OwnedBy:     todo.OwnedBy.String(),
		CreatedBy:   todo.CreatedBy.String(),
		DueDate:     todo.DueDate,
		CreatedAt:   *todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}
