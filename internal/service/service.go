package service

import (
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
	logger log.Logger
	tracer tracing.Tracer
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
