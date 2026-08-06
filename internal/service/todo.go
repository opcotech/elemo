package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// Todo represents a todo returned by the service.
type Todo struct {
	ID          model.ID
	Title       string
	Description string
	Priority    model.TodoPriority
	Completed   bool
	OwnedBy     model.ID
	CreatedBy   model.ID
	DueDate     *time.Time
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// CreateTodoOpts holds the data required to create a todo.
type CreateTodoOpts struct {
	Title       string             `json:"title" validate:"required,min=3,max=120"`
	Description string             `json:"description" validate:"omitempty,min=10,max=500"`
	Priority    model.TodoPriority `json:"priority" validate:"required,min=1,max=4"`
	Completed   bool               `json:"completed"`
	OwnedBy     model.ID           `json:"owned_by" validate:"required"`
	CreatedBy   model.ID           `json:"created_by" validate:"required"`
	DueDate     *time.Time         `json:"due_date" validate:"omitempty"`
}

// Validate validates the create options.
func (o *CreateTodoOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidTodoDetails, err)
	}
	if err := o.OwnedBy.Validate(); err != nil {
		return errors.Join(model.ErrInvalidTodoDetails, err)
	}
	if err := o.CreatedBy.Validate(); err != nil {
		return errors.Join(model.ErrInvalidTodoDetails, err)
	}
	return nil
}

// UpdateTodoOpts holds the fields that can be updated on a todo.
// Undefined fields (Defined == false) are left unchanged.
type UpdateTodoOpts struct {
	Title       optional.Optional[string]
	Description optional.Optional[string]
	Priority    optional.Optional[model.TodoPriority]
	Completed   optional.Optional[bool]
	DueDate     optional.Optional[time.Time]
}

// TodoService serves the business logic of interacting with todos in the
// system.
type TodoService interface {
	// Create creates a new todo item. Users can create todos for each other
	// if they are related in some way. If the creator and owner are not
	// related, an error is returned.
	Create(ctx context.Context, opts CreateTodoOpts) (*Todo, error)
	// Get returns a todo by its ID. If the todo does not exist, an error is
	// returned.
	Get(ctx context.Context, id model.ID) (*Todo, error)
	// GetAll returns all todos for the authenticated user. If the completed
	// parameter is set to true, only completed todos are returned. If the
	// completed parameter is set to false, only incomplete todos are
	// returned. If the completed parameter is nil, all todos are returned.
	GetAll(ctx context.Context, offset, limit int, completed *bool) ([]*Todo, error)
	// Update updates a todo by its ID. If the todo does not exist, an error
	// is returned.
	Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error)
	// Delete deletes a todo by its ID. If the todo does not exist, an error
	// is returned.
	Delete(ctx context.Context, id model.ID) error
}

// todoService is the concrete implementation of the TodoService interface.
type todoService struct {
	*baseService
}

func todoFromRepository(t *repository.Todo) *Todo {
	if t == nil {
		return nil
	}
	return &Todo{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Completed:   t.Completed,
		OwnedBy:     t.OwnedBy,
		CreatedBy:   t.CreatedBy,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func todosFromRepository(todos []*repository.Todo) []*Todo {
	out := make([]*Todo, len(todos))
	for i, t := range todos {
		out[i] = todoFromRepository(t)
	}
	return out
}

func (s *todoService) Create(ctx context.Context, opts CreateTodoOpts) (*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrTodoCreate, license.ErrLicenseExpired)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrTodoCreate, err)
	}

	if opts.CreatedBy != opts.OwnedBy {
		hasRelation, err := s.permissionService.HasAnyRelation(ctx, opts.CreatedBy, opts.OwnedBy)
		if err != nil {
			return nil, errors.Join(ErrTodoCreate, err)
		}
		if !hasRelation {
			return nil, errors.Join(ErrTodoCreate, ErrNoPermission)
		}
	}

	todo, err := s.todoRepo.Create(ctx, repository.CreateTodoOpts{
		Title:       opts.Title,
		Description: opts.Description,
		Priority:    opts.Priority,
		Completed:   opts.Completed,
		OwnedBy:     opts.OwnedBy,
		CreatedBy:   opts.CreatedBy,
		DueDate:     opts.DueDate,
	})
	if err != nil {
		return nil, errors.Join(ErrTodoCreate, err)
	}

	return todoFromRepository(todo), nil
}

func (s *todoService) Get(ctx context.Context, id model.ID) (*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrTodoGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, errors.Join(ErrTodoGet, ErrNoPermission)
	}

	todo, err := s.todoRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrTodoGet, err)
	}

	return todoFromRepository(todo), nil
}

func (s *todoService) GetAll(ctx context.Context, offset, limit int, completed *bool) ([]*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/GetAll")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrTodoGetAll, ErrNoUser)
	}

	todos, err := s.todoRepo.GetByOwner(ctx, userID, offset, limit, completed)
	if err != nil {
		return nil, errors.Join(ErrTodoGetAll, err)
	}

	return todosFromRepository(todos), nil
}

func (s *todoService) Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrTodoUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrTodoUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrTodoUpdate, ErrNoPermission)
	}

	todo, err := s.todoRepo.Update(ctx, id, repository.UpdateTodoOpts{
		Title:       opts.Title,
		Description: opts.Description,
		Priority:    opts.Priority,
		Completed:   opts.Completed,
		DueDate:     opts.DueDate,
	})
	if err != nil {
		return nil, errors.Join(ErrTodoUpdate, err)
	}

	return todoFromRepository(todo), nil
}

func (s *todoService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrTodoDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrTodoDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrTodoDelete, ErrNoPermission)
	}

	if err := s.todoRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrTodoDelete, err)
	}

	return nil
}

// NewTodoService returns a new instance of the TodoService interface.
func NewTodoService(opts ...Option) (TodoService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &todoService{
		baseService: s,
	}

	if svc.todoRepo == nil {
		return nil, ErrNoTodoRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}
