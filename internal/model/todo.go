package model

const (
	TodoPriorityNormal    TodoPriority = iota + 1 // the todo is normal
	TodoPriorityImportant                         // the todo is important
	TodoPriorityUrgent                            // the todo is urgent
	TodoPriorityCritical                          // the todo is critical
)

// TodoPriority represents the priority of the Todo item.
//
//go:generate go tool enumer -type=TodoPriority -trimprefix=TodoPriority -text -transform=snake -output=todo_priority_gen.go
type TodoPriority uint8
