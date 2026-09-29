package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/model"
)

const (
	LicenseExpiryReminderInterval = 24 * time.Hour
	LicenseExpiryReminderWindow   = 7 * 24 * time.Hour
	licenseExpiryReminderTaskID   = "system:license_expiry:daily"
)

// HealthCheckTaskPayload is the payload for the health check task.
type HealthCheckTaskPayload struct {
	Message string `json:"message"`
}

// NewSystemHealthCheckTask creates a new health check task.
func NewSystemHealthCheckTask() (*asynq.Task, error) {
	payload, _ := json.Marshal(HealthCheckTaskPayload{Message: model.HealthStatusHealthy.String()})
	return asynq.NewTask(
		TaskTypeSystemHealthCheck.String(),
		payload,
		asynq.Timeout(DefaultTaskTimeout),
		asynq.Retention(DefaultTaskRetention),
	), nil
}

// NewSystemLicenseExpiryTask creates a periodic License expiration
// reminder task. Task ID retention prevents successful reminders from being
// enqueued more than once per day; Unique also suppresses concurrent copies.
func NewSystemLicenseExpiryTask() (*asynq.Task, error) {
	return asynq.NewTask(
		TaskTypeSystemLicenseExpiry.String(),
		nil,
		asynq.TaskID(licenseExpiryReminderTaskID),
		asynq.Timeout(30*time.Second),
		asynq.Retention(LicenseExpiryReminderInterval),
		asynq.Unique(LicenseExpiryReminderInterval),
		asynq.Queue(MessageQueueHighPriority),
	), nil
}
