package asynq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/model"
)

// HealthCheckTaskPayload is the payload for the health check task.
type HealthCheckTaskPayload struct {
	Message string `json:"message"`
}

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

	var payload HealthCheckTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	return nil
}

// NewSystemHealthCheckTaskHandler creates a new health check task handler.
func NewSystemHealthCheckTaskHandler(opts ...TaskOption) (*SystemHealthCheckTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}

	return &SystemHealthCheckTaskHandler{h}, nil
}

// NewSystemHealthCheckTask creates a new health check task.
func NewSystemHealthCheckTask() (*asynq.Task, error) {
	payload, _ := json.Marshal(HealthCheckTaskPayload{Message: model.HealthStatusHealthy.String()})
	return asynq.NewTask(
		TaskTypeSystemHealthCheck.String(),
		payload,
		asynq.Timeout(5*time.Second),
		asynq.Retention(5*time.Second),
	), nil
}
