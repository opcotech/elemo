package asynq

import (
	"time"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/service"
)

const (
	DefaultTaskTimeout   = 5 * time.Second // The default task timeout.
	DefaultTaskRetention = 5 * time.Minute // The default task retention.
)

const (
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

// TaskOption is a function that can be used to configure a task handler.
type TaskOption func(*baseTaskHandler) error

// WithTaskEmailService sets the email service for the worker.
func WithTaskEmailService(emailService service.EmailService) TaskOption {
	return func(t *baseTaskHandler) error {
		if emailService == nil {
			return ErrNoEmailService
		}

		t.emailService = emailService
		return nil
	}
}

// WithTaskLogger sets the logger for the task handler.
func WithTaskLogger(logger log.Logger) TaskOption {
	return func(t *baseTaskHandler) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		t.logger = logger

		return nil
	}
}

// WithTaskTracer sets the tracer for the task handler.
func WithTaskTracer(tracer tracing.Tracer) TaskOption {
	return func(t *baseTaskHandler) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		t.tracer = tracer

		return nil
	}
}

// baseTaskHandler serves as the base type for all task handlers.
type baseTaskHandler struct {
	logger log.Logger
	tracer tracing.Tracer

	emailService service.EmailService
}

// newBaseTaskHandler creates a new base task handler.
func newBaseTaskHandler(opts ...TaskOption) (*baseTaskHandler, error) {
	t := &baseTaskHandler{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}
