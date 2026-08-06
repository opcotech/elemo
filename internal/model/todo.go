package model

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
