package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// TodoService serves the business logic of interacting with todos in the
// system.
type TodoService interface {
	// Create creates a new todo item. Users can create todos for each other
	// if they are related in some way. If the creator and owner are not
	// related, an error is returned.
	Create(ctx context.Context, todo *model.Todo) error
	// Get returns a todo by its ID. If the todo does not exist, an error is
	// returned.
	Get(ctx context.Context, id model.ID) (*model.Todo, error)
	// GetAll returns all todos for the authenticated user. If the completed
	// parameter is set to true, only completed todos are returned. If the
	// completed parameter is set to false, only incomplete todos are
	// returned. If the completed parameter is nil, all todos are returned.
	GetAll(ctx context.Context, offset, limit int, completed *bool) ([]*model.Todo, error)
	// Update updates a todo by its ID. The patch parameter is a map of
	// fields to update. If the todo does not exist, an error is returned.
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Todo, error)
	// Delete deletes a todo by its ID. If the todo does not exist, an error
	// is returned.
	Delete(ctx context.Context, id model.ID) error
}

// todoService is the concrete implementation of the TodoService interface.
type todoService struct {
	*baseService
}

func (s *todoService) Create(ctx context.Context, todo *model.Todo) error {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return license.ErrLicenseExpired
	}

	if err := todo.Validate(); err != nil {
		return errors.Join(ErrTodoCreate, err)
	}

	if todo.CreatedBy != todo.OwnedBy {
		hasRelation, err := s.permissionService.HasAnyRelation(ctx, todo.CreatedBy, todo.OwnedBy)
		if err != nil {
			return errors.Join(ErrTodoCreate, err)
		}
		if !hasRelation {
			return ErrNoPermission
		}
	}

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return errors.Join(ErrTodoCreate, err)
	}

	return nil
}

func (s *todoService) Get(ctx context.Context, id model.ID) (*model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrTodoGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, ErrNoPermission
	}

	todo, err := s.todoRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrTodoGet, err)
	}

	return todo, nil
}

func (s *todoService) GetAll(ctx context.Context, offset, limit int, completed *bool) ([]*model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/GetAll")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, ErrNoUser
	}

	todos, err := s.todoRepo.GetByOwner(ctx, userID, offset, limit, completed)
	if err != nil {
		return nil, errors.Join(ErrTodoGetAll, err)
	}

	return todos, nil
}

func (s *todoService) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, license.ErrLicenseExpired
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrTodoUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, ErrNoPermission
	}

	todo, err := s.todoRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, errors.Join(ErrTodoUpdate, err)
	}

	return todo, nil
}

func (s *todoService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.todoService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return license.ErrLicenseExpired
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrTodoDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return ErrNoPermission
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
