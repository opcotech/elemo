package model

const (
	TodoPriorityNormal    TodoPriority = iota + 1 // normal
	TodoPriorityImportant                         // important
	TodoPriorityUrgent                            // urgent
	TodoPriorityCritical                          // critical
)

// TodoPriority represents the priority of the Todo item.
//
//go:generate go tool enumer -type=TodoPriority -text -transform=noop -linecomment -output=todo_priority_gen.go
type TodoPriority uint8
