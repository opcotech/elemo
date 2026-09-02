package queue

import (
	"github.com/hibiken/asynq"
)

// NewCustomFieldReconcileTask creates a reconciliation task for hybrid writes.
func NewCustomFieldReconcileTask() (*asynq.Task, error) {
	return asynq.NewTask(
		TaskTypeCustomFieldReconcile.String(),
		nil,
		asynq.Timeout(DefaultTaskTimeout),
		asynq.Retention(DefaultTaskRetention),
		asynq.Queue(MessageQueueLowPriority),
	), nil
}
