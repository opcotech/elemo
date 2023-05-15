package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

// TodoController is the controller for the todo endpoints.
type TodoController interface {
	GetTodos(ctx context.Context, request gen.GetTodosRequestObject) (gen.GetTodosResponseObject, error)
	CreateTodo(ctx context.Context, request gen.CreateTodoRequestObject) (gen.CreateTodoResponseObject, error)
	DeleteTodo(ctx context.Context, request gen.DeleteTodoRequestObject) (gen.DeleteTodoResponseObject, error)
	GetTodo(ctx context.Context, request gen.GetTodoRequestObject) (gen.GetTodoResponseObject, error)
	UpdateTodo(ctx context.Context, request gen.UpdateTodoRequestObject) (gen.UpdateTodoResponseObject, error)
}

// todoController is the concrete implementation of TodoController.
type todoController struct {
	*baseController
}

func (c *todoController) CreateTodo(ctx context.Context, request gen.CreateTodoRequestObject) (gen.CreateTodoResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/CreateTodo")
	defer span.End()

	createdBy, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return gen.CreateTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	ownerID, err := model.NewIDFromString(request.Body.OwnedBy, model.ResourceTypeUser.String())
	if err != nil {
		return gen.CreateTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	todo, err := createTodoJSONRequestBodyToTodo(request.Body, ownerID, createdBy)
	if err != nil {
		return gen.CreateTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.todoService.Create(ctx, todo); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.CreateTodo401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		return gen.CreateTododefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.CreateTodo201JSONResponse{
		TodoId: todo.ID.String(),
	}, nil
}

func (c *todoController) GetTodo(ctx context.Context, request gen.GetTodoRequestObject) (gen.GetTodoResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetTodo")
	defer span.End()

	todoID, err := model.NewIDFromString(request.TodoId, model.ResourceTypeTodo.String())
	if err != nil {
		return gen.GetTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	todo, err := c.todoService.Get(ctx, todoID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.GetTodo401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.GetTodo404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.GetTododefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.GetTodo200JSONResponse(todoToDTO(todo)), nil
}

func (c *todoController) GetTodos(ctx context.Context, request gen.GetTodosRequestObject) (gen.GetTodosResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetTodos")
	defer span.End()

	todos, err := c.todoService.GetAll(ctx,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
		request.Params.Completed,
	)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.GetTodos401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.GetTodos404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.GetTodosdefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	todosDTO := make([]gen.Todo, len(todos))
	for i, todo := range todos {
		todosDTO[i] = todoToDTO(todo)
	}

	return gen.GetTodos200JSONResponse(todosDTO), nil
}

func (c *todoController) UpdateTodo(ctx context.Context, request gen.UpdateTodoRequestObject) (gen.UpdateTodoResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/UpdateTodo")
	defer span.End()

	todoID, err := model.NewIDFromString(request.TodoId, model.ResourceTypeTodo.String())
	if err != nil {
		return gen.UpdateTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return gen.UpdateTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	todo, err := c.todoService.Update(ctx, todoID, patch)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.UpdateTodo401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.UpdateTodo404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.UpdateTododefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.UpdateTodo200JSONResponse(todoToDTO(todo)), nil
}

func (c *todoController) DeleteTodo(ctx context.Context, request gen.DeleteTodoRequestObject) (gen.DeleteTodoResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/DeleteTodo")
	defer span.End()

	todoID, err := model.NewIDFromString(request.TodoId, model.ResourceTypeTodo.String())
	if err != nil {
		return gen.DeleteTodo400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.todoService.Delete(ctx, todoID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.DeleteTodo401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.DeleteTodo404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.DeleteTododefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.DeleteTodo204JSONResponse{}, nil
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

func createTodoJSONRequestBodyToTodo(body *gen.CreateTodoJSONRequestBody, ownedBy, createdBy model.ID) (*model.Todo, error) {
	todo := &model.Todo{
		ID:          model.MustNewNilID(model.ResourceTypeTodo),
		Title:       body.Title,
		Description: pkg.GetDefaultPtr(body.Description, ""),
		Completed:   body.Completed,
		OwnedBy:     ownedBy,
		CreatedBy:   createdBy,
		DueDate:     body.DueDate,
	}

	if body.Priority != "" {
		if err := todo.Priority.UnmarshalText([]byte(body.Priority)); err != nil {
			return nil, err
		}
	} else {
		todo.Priority = model.TodoPriorityNormal
	}

	return todo, nil
}

func todoToDTO(todo *model.Todo) gen.Todo {
	return gen.Todo{
		Id:          convert.ToPointer(todo.ID.String()),
		Title:       &todo.Title,
		Completed:   &todo.Completed,
		Priority:    convert.ToPointer(gen.TodoPriority(todo.Priority.String())),
		Description: &todo.Description,
		OwnedBy:     convert.ToPointer(todo.OwnedBy.String()),
		CreatedBy:   convert.ToPointer(todo.CreatedBy.String()),
		DueDate:     todo.DueDate,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}
