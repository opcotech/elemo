package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateTodoOpts creates repository.CreateTodoOpts for tests.
func NewCreateTodoOpts(owner, creator model.ID) repository.CreateTodoOpts {
	return repository.CreateTodoOpts{
		Title:       pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Priority:    model.TodoPriorityNormal,
		Completed:   false,
		OwnedBy:     owner,
		CreatedBy:   creator,
		DueDate:     convert.ToPointer(time.Now().UTC().Add(24 * time.Hour)),
	}
}

// NewRepositoryTodo creates a repository.Todo for mock returns.
func NewRepositoryTodo(owner, creator model.ID) *repository.Todo {
	opts := NewCreateTodoOpts(owner, creator)
	return &repository.Todo{
		ID:          model.MustNewID(model.ResourceTypeTodo),
		Title:       opts.Title,
		Description: opts.Description,
		Priority:    opts.Priority,
		Completed:   opts.Completed,
		OwnedBy:     opts.OwnedBy,
		CreatedBy:   opts.CreatedBy,
		DueDate:     opts.DueDate,
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}
