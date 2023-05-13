package asynq

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

const (
	TaskTypeSystemHealthCheck TaskType = iota + 1 // Health check task type.
)

var (
	taskTypeValues = map[TaskType]string{
		TaskTypeSystemHealthCheck: "system:health_check",
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
func WithTaskTracer(tracer trace.Tracer) TaskOption {
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
	tracer trace.Tracer
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
