package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	TodoIDType = "Todo"
)

const (
	TodoPriorityNormal    TodoPriority = iota + 1 // the todo is normal
	TodoPriorityImportant                         // the todo is important
	TodoPriorityUrgent                            // the todo is urgent
	TodoPriorityCritical                          // the todo is critical
)

var (
	todoPriorityKeys = map[TodoPriority]string{
		TodoPriorityNormal:    "normal",
		TodoPriorityImportant: "important",
		TodoPriorityUrgent:    "urgent",
		TodoPriorityCritical:  "critical",
	}
	todoPriorityValues = map[string]TodoPriority{
		"normal":    TodoPriorityNormal,
		"important": TodoPriorityImportant,
		"urgent":    TodoPriorityUrgent,
		"critical":  TodoPriorityCritical,
	}
)

// TodoPriority represents the priority of the Todo item.
type TodoPriority uint8

// String returns the string representation of the TodoPriority.
func (p TodoPriority) String() string {
	return todoPriorityKeys[p]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (p TodoPriority) MarshalText() (text []byte, err error) {
	if p < 1 || p > 4 {
		return nil, ErrInvalidTodoPriority
	}
	return []byte(p.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *TodoPriority) UnmarshalText(text []byte) error {
	if v, ok := todoPriorityValues[string(text)]; ok {
		*p = v
		return nil
	}
	return ErrInvalidTodoPriority
}

// Todo represents a todo in the system.
type Todo struct {
	ID          ID           `json:"id" validate:"required,dive"`
	Title       string       `json:"title" validate:"required,min=3,max=120"`
	Description string       `json:"description" validate:"omitempty,min=10,max=500"`
	Priority    TodoPriority `json:"priority" validate:"required,min=1,max=4"`
	Completed   bool         `json:"completed"`
	OwnedBy     ID           `json:"owned_by" validate:"required,dive"`
	CreatedBy   ID           `json:"created_by" validate:"required,dive"`
	DueDate     *time.Time   `json:"due_date" validate:"omitempty"`
	CreatedAt   *time.Time   `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time   `json:"updated_at" validate:"omitempty"`
}

func (t *Todo) Validate() error {
	if err := validate.Struct(t); err != nil {
		return errors.Join(ErrInvalidTodoDetails, err)
	}
	if err := t.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidTodoDetails, err)
	}
	if err := t.OwnedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidTodoDetails, err)
	}
	if err := t.CreatedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidTodoDetails, err)
	}
	return nil
}

// NewTodo creates a new todo.
func NewTodo(title string, ownedBy, createdBy ID) (*Todo, error) {
	todo := &Todo{
		ID:        MustNewNilID(TodoIDType),
		Title:     title,
		Priority:  TodoPriorityNormal,
		OwnedBy:   ownedBy,
		CreatedBy: createdBy,
	}

	if err := todo.Validate(); err != nil {
		return nil, err
	}

	return todo, nil
}
