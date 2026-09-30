package queue

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestNewSystemHealthCheckTask(t *testing.T) {
	tests := []struct {
		name    string
		want    *asynq.Task
		wantErr error
	}{
		{
			name: "create new task",
			want: asynq.NewTask(TaskTypeSystemHealthCheck.String(),
				[]byte(`{"message":"healthy"}`),
				asynq.Timeout(DefaultTaskTimeout),
				asynq.Retention(DefaultTaskRetention)),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewSystemHealthCheckTask()
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSystemLicenseExpiryTask(t *testing.T) {
	t.Parallel()

	got, err := NewSystemLicenseExpiryTask()
	assert.NoError(t, err)

	want := asynq.NewTask(
		TaskTypeSystemLicenseExpiry.String(),
		nil,
		asynq.TaskID(licenseExpiryReminderTaskID),
		asynq.Timeout(30*time.Second),
		asynq.Retention(LicenseExpiryReminderInterval),
		asynq.Unique(LicenseExpiryReminderInterval),
		asynq.Queue(MessageQueueHighPriority),
	)
	assert.Equal(t, want, got)
}
