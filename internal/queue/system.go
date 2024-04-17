package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
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

// LicenseExpiryTaskPayload is the payload for the license expiry check task.
type LicenseExpiryTaskPayload struct {
	LicenseID           string
	LicenseEmail        string
	LicenseOrganization string
	LicenseExpiresAt    time.Time
}

// NewSystemLicenseExpiryTask creates a new license expiry check task.
func NewSystemLicenseExpiryTask(l *license.License) (*asynq.Task, error) {
	if l == nil {
		return nil, license.ErrNoLicense
	}

	payload, _ := json.Marshal(LicenseExpiryTaskPayload{
		LicenseID:           l.ID.String(),
		LicenseEmail:        l.Email,
		LicenseOrganization: l.Organization,
		LicenseExpiresAt:    l.ExpiresAt,
	})

	return asynq.NewTask(
		TaskTypeSystemLicenseExpiry.String(),
		payload,
		asynq.Timeout(DefaultTaskTimeout),
		asynq.Queue(MessageQueueHighPriority),
	), nil
}
