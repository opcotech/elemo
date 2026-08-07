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
	TaskTypeSystemHealthCheck   TaskType = iota + 1 // system:health_check
	TaskTypeSystemLicenseExpiry                     // system:license_expiry
)

// TaskType is the type for system tasks.
//
//go:generate go tool enumer -type=TaskType -transform=noop -linecomment -output=queue_task_type_gen.go
type TaskType uint8
