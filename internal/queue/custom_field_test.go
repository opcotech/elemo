package queue

import (
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomFieldReconcileTask(t *testing.T) {
	t.Parallel()

	got, err := NewCustomFieldReconcileTask()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, TaskTypeCustomFieldReconcile.String(), got.Type())
	assert.Nil(t, got.Payload())

	want := asynq.NewTask(
		TaskTypeCustomFieldReconcile.String(),
		nil,
		asynq.Timeout(DefaultTaskTimeout),
		asynq.Retention(DefaultTaskRetention),
		asynq.Queue(MessageQueueLowPriority),
	)
	assert.Equal(t, want, got)
}
