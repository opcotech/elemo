package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

// NewTodo creates a new Todo instance. It does not create the todo in the db.
func NewTodo(owner, creator model.ID) *model.Todo {
	todo, err := model.NewTodo(pkg.GenerateRandomString(10), owner, creator)
	if err != nil {
		panic(err)
	}

	todo.Description = pkg.GenerateRandomString(10)
	todo.DueDate = convert.ToPointer(time.Now().UTC().Add(24 * time.Hour))

	return todo
}
