package async

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
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

// SystemLicenseExpiryTaskHandler sends License expiration reminders.
type SystemLicenseExpiryTaskHandler struct {
	*baseTaskHandler
}

// ProcessTask evaluates the current entitlement and sends a reminder when the
// verified License is close to expiration, in grace, or expired.
func (h *SystemLicenseExpiryTaskHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	ctx, span := h.tracer.Start(ctx, "transport.asynq.SystemLicenseExpiryTaskHandler/ProcessTask")
	defer span.End()

	recipient := strings.TrimSpace(h.billingEmail)
	address, err := mail.ParseAddress(recipient)
	if err != nil || address.Address != recipient {
		if recipient != "" {
			h.logger.Warn(ctx, "airgap billing email is invalid; license expiration reminder skipped")
		}
		return nil
	}

	status, err := h.entitlementReporter.Status(ctx)
	if err != nil {
		return err
	}
	if !shouldSendLicenseExpiryReminder(status, time.Now().UTC()) {
		return nil
	}

	return h.emailService.SendLicenseExpiryEmail(ctx, recipient, *status.AirGap)
}

func shouldSendLicenseExpiryReminder(status entitlement.Status, now time.Time) bool {
	if status.AirGap == nil {
		return false
	}

	switch status.AirGap.State {
	case license.StateValid:
		return status.AirGap.ExpiresAt != nil &&
			!status.AirGap.ExpiresAt.After(now.Add(queue.LicenseExpiryReminderWindow))
	case license.StateGrace, license.StateExpired:
		return true
	default:
		return false
	}
}

// NewSystemLicenseExpiryTaskHandler creates a license expiration reminder task
// handler.
func NewSystemLicenseExpiryTaskHandler(opts ...TaskHandlerOption) (*SystemLicenseExpiryTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}
	if h.emailService == nil {
		return nil, ErrNoEmailService
	}
	if h.entitlementReporter == nil {
		return nil, ErrNoEntitlementReporter
	}

	return &SystemLicenseExpiryTaskHandler{h}, nil
}
