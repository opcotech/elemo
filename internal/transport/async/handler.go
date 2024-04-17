package async

import (
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/service"
)

// TaskHandlerOption is a function that can be used to configure a task handler.
type TaskHandlerOption func(*baseTaskHandler) error

// WithTaskEmailService sets the email service for the worker.
func WithTaskEmailService(emailService service.EmailService) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if emailService == nil {
			return ErrNoEmailService
		}

		t.emailService = emailService
		return nil
	}
}

// WithTaskLogger sets the logger for the task handler.
func WithTaskLogger(logger log.Logger) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		t.logger = logger

		return nil
	}
}

// WithTaskTracer sets the tracer for the task handler.
func WithTaskTracer(tracer tracing.Tracer) TaskHandlerOption {
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
func newBaseTaskHandler(opts ...TaskHandlerOption) (*baseTaskHandler, error) {
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
