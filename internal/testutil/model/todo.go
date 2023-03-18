package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewTodo creates a new Todo instance. It does not create the todo in the db.
func NewTodo(owner, creator model.ID) *model.Todo {
	todo, err := model.NewTodo(testutil.GenerateRandomString(10), owner, creator)
	if err != nil {
		panic(err)
	}

	todo.Description = testutil.GenerateRandomString(10)
	todo.DueDate = convert.ToPointer(time.Now().Add(24 * time.Hour))

	return todo
}
