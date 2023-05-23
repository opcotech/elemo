package http

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
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
}

func (c *todoController) V1TodosCreate(ctx context.Context, request api.V1TodosCreateRequestObject) (api.V1TodosCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodosCreate")
	defer span.End()

	createdBy, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1TodosCreate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	ownerID, err := model.NewIDFromString(request.Body.OwnedBy, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1TodosCreate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	todo, err := createTodoJSONRequestBodyToTodo(request.Body, ownerID, createdBy)
	if err != nil {
		return api.V1TodosCreate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.todoService.Create(ctx, todo); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1TodosCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1TodosCreate500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
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
		return api.V1TodoGet400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	todo, err := c.todoService.Get(ctx, todoID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1TodoGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1TodoGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1TodoGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1TodoGet200JSONResponse(todoToDTO(todo)), nil
}

func (c *todoController) V1TodosGet(ctx context.Context, request api.V1TodosGetRequestObject) (api.V1TodosGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodosGet")
	defer span.End()

	todos, err := c.todoService.GetAll(ctx,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
		request.Params.Completed,
	)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1TodosGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1TodosGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	todosDTO := make([]api.Todo, len(todos))
	for i, todo := range todos {
		todosDTO[i] = todoToDTO(todo)
	}

	return api.V1TodosGet200JSONResponse(todosDTO), nil
}

func (c *todoController) V1TodoUpdate(ctx context.Context, request api.V1TodoUpdateRequestObject) (api.V1TodoUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1TodoUpdate")
	defer span.End()

	todoID, err := model.NewIDFromString(request.Id, model.ResourceTypeTodo.String())
	if err != nil {
		return api.V1TodoUpdate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return api.V1TodoUpdate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if dueDate, ok := patch["due_date"]; ok && dueDate != nil {
		if dueDate == "" {
			patch["due_date"] = nil
		} else {
			patch["due_date"], err = time.Parse(time.RFC3339, dueDate.(string))
			if err != nil {
				return api.V1TodoUpdate400JSONResponse{N400JSONResponse: badRequest}, nil
			}
		}
	}

	todo, err := c.todoService.Update(ctx, todoID, patch)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1TodoUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1TodoUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1TodoUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
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
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1TodoDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1TodoDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1TodoDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1TodoDelete204Response{}, nil
}

// NewTodoController creates a new TodoController.
func NewTodoController(opts ...ControllerOption) (TodoController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &todoController{
		baseController: c,
	}

	if controller.todoService == nil {
		return nil, ErrNoTodoService
	}

	if controller.userService == nil {
		return nil, ErrNoTodoService
	}

	return controller, nil
}

func createTodoJSONRequestBodyToTodo(body *api.V1TodosCreateJSONRequestBody, ownedBy, createdBy model.ID) (*model.Todo, error) {
	todo, err := model.NewTodo(body.Title, ownedBy, createdBy)
	if err != nil {
		return nil, err
	}

	todo.Description = pkg.GetDefaultPtr(body.Description, "")

	if body.DueDate != nil && *body.DueDate != "" {
		dueDate, err := time.Parse(time.RFC3339, *body.DueDate)
		if err != nil {
			return nil, err
		}

		todo.DueDate = &dueDate
	}

	if err := todo.Priority.UnmarshalText([]byte(body.Priority)); err != nil {
		return nil, err
	}

	return todo, nil
}

func todoToDTO(todo *model.Todo) api.Todo {
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
