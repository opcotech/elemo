package queue

import "time"

const (
	DefaultTaskTimeout   = 5 * time.Second // The default task timeout.
	DefaultTaskRetention = 5 * time.Minute // The default task retention.

	MessageQueueDefaultPriority = "default" // The default queue name.
	MessageQueueLowPriority     = "low"     // The low priority queue name.
	MessageQueueHighPriority    = "high"    // The high priority queue name.

	TaskTypeSystemHealthCheck   TaskType = iota + 1 // Health check task type.
	TaskTypeSystemLicenseExpiry                     // License expiry task type.
)

var (
	taskTypeValues = map[TaskType]string{
		TaskTypeSystemHealthCheck:   "system:health_check",
		TaskTypeSystemLicenseExpiry: "system:license_expiry",
	}
)

// TaskType is the type for system tasks.
type TaskType uint8

// String returns the string representation of the system task type.
func (t TaskType) String() string {
	return taskTypeValues[t]
}
