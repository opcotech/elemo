package service

import (
	"context"

	"github.com/opcotech/elemo/internal/pkg/event"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

// Option defines a configuration option for the service runtime.
type Option func(*runtime) error

// WithLogger sets the logger for the service runtime.
func WithLogger(logger log.Logger) Option {
	return func(r *runtime) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		r.logger = logger
		return nil
	}
}

// WithTracer sets the tracer for the service runtime.
func WithTracer(tracer tracing.Tracer) Option {
	return func(r *runtime) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		r.tracer = tracer
		return nil
	}
}

type runtime struct {
	logger   log.Logger
	tracer   tracing.Tracer
	eventBus EventPublisher
}

// EventPublisher is the subset of the in-process bus used by domain services.
type EventPublisher interface {
	Publish(ctx context.Context, event event.Event) error
}

// WithEventBus sets the in-process event publisher. A nil bus is a no-op.
func WithEventBus(bus EventPublisher) Option {
	return func(r *runtime) error {
		r.eventBus = bus
		return nil
	}
}

func newRuntime(opts ...Option) (runtime, error) {
	r := runtime{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(&r); err != nil {
			return runtime{}, err
		}
	}

	return r, nil
}
