package queue

import "time"

const (
	DefaultTaskTimeout   = 5 * time.Second // The default task timeout.
	DefaultTaskRetention = 5 * time.Minute // The default task retention.

	MessageQueueDefaultPriority = "default" // The default queue name.
	MessageQueueLowPriority     = "low"     // The low priority queue name.
	MessageQueueHighPriority    = "high"    // The high priority queue name.
)

const (
	TaskTypeSystemHealthCheck   TaskType = iota + 1 // Health check task type.
	TaskTypeSystemLicenseExpiry                     // License expiry task type.
)

// TaskType is the type for system tasks.
//
//go:generate go tool enumer -type=TaskType -trimprefix=TaskType -transform=snake -output=queue_task_type_gen.go
type TaskType uint8
