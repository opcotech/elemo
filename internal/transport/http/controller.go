package http

import (
	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

type ControllerOption func(*baseController) error

// WithConfig sets the config for the controller.
func WithConfig(conf config.ServerConfig) ControllerOption {
	return func(c *baseController) error {
		c.conf = conf
		return nil
	}
}

// WithLogger sets the logger for the controller.
func WithLogger(logger log.Logger) ControllerOption {
	return func(c *baseController) error {
		if logger == nil {
			return ErrNoLogger
		}

		c.logger = logger

		return nil
	}
}

// WithTracer sets the tracer for the controller.
func WithTracer(tracer tracing.Tracer) ControllerOption {
	return func(c *baseController) error {
		if tracer == nil {
			return ErrNoTracer
		}

		c.tracer = tracer

		return nil
	}
}

// baseController defines the shared logger, tracer, and config injected into
// every controller.
type baseController struct {
	conf   config.ServerConfig
	logger log.Logger
	tracer tracing.Tracer
}

// newController creates a new base controller with the given dependencies
// and default values where applicable.
func newController(opts ...ControllerOption) (*baseController, error) {
	c := &baseController{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}
