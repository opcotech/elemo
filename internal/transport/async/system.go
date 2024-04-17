package async

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"

	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/queue"
)

// SystemHealthCheckTaskHandler is the health check task. The health check task is used to
// check the health of the async worker. If the async worker is unhealthy, the
// task won't be processed.
type SystemHealthCheckTaskHandler struct {
	*baseTaskHandler
}

// ProcessTask unmarshals the task payload and returns an error if the task
// payload is invalid. Otherwise, it returns nil, indicating that the task has
// been processed successfully.
func (h *SystemHealthCheckTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	_, span := h.tracer.Start(ctx, "transport.asynq.SystemHealthCheckTaskHandler/ProcessTask")
	defer span.End()

	var payload queue.HealthCheckTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	return nil
}

// NewSystemHealthCheckTaskHandler creates a new health check task handler.
func NewSystemHealthCheckTaskHandler(opts ...TaskHandlerOption) (*SystemHealthCheckTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}

	return &SystemHealthCheckTaskHandler{h}, nil
}

// SystemLicenseExpiryTaskHandler is the license expiry check task. If the
// license is about to expire, it sends an email to the licensee.
type SystemLicenseExpiryTaskHandler struct {
	*baseTaskHandler
}

// ProcessTask unmarshals the task payload and checks if the license is about
// to expire. If the license is about to expire, it sends an email to the
// licensee. Otherwise, it skips the task.
func (h *SystemLicenseExpiryTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	ctx, span := h.tracer.Start(ctx, "transport.asynq.SystemLicenseExpiryTaskHandler/ProcessTask")
	defer span.End()

	var payload queue.LicenseExpiryTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	// If the license is not about to expire, skip the task.
	if payload.LicenseExpiresAt.After(time.Now().Add(7 * 24 * time.Hour)) {
		return nil
	}

	return h.emailService.SendSystemLicenseExpiryEmail(
		ctx,
		payload.LicenseID,
		payload.LicenseEmail,
		payload.LicenseOrganization,
		payload.LicenseExpiresAt,
	)
}

// NewSystemLicenseExpiryTaskHandler creates a new license expiry check task handler.
func NewSystemLicenseExpiryTaskHandler(opts ...TaskHandlerOption) (*SystemLicenseExpiryTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}

	if h.emailService == nil {
		return nil, ErrNoEmailService
	}

	return &SystemLicenseExpiryTaskHandler{h}, nil
}
