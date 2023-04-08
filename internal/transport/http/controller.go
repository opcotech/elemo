package http

import (
	auth "github.com/go-oauth2/oauth2/v4/server"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/service"
)

const (
	DefaultLimit  = 10 // default limit for pagination
	DefaultOffset = 0  // default offset for pagination
)

// ControllerOption is a function that can be used to configure a controller.
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
func WithTracer(tracer trace.Tracer) ControllerOption {
	return func(c *baseController) error {
		if tracer == nil {
			return ErrNoTracer
		}

		c.tracer = tracer

		return nil
	}
}

// WithAuthProvider sets the auth provider for the controller.
func WithAuthProvider(authProvider *auth.Server) ControllerOption {
	return func(c *baseController) error {
		if authProvider == nil {
			return ErrNoAuthProvider
		}

		c.authProvider = authProvider

		return nil
	}
}

// WithSystemService sets the system service for the controller.
func WithSystemService(systemService service.SystemService) ControllerOption {
	return func(c *baseController) error {
		if systemService == nil {
			return ErrNoSystemService
		}

		c.systemService = systemService

		return nil
	}
}

// WithUserService sets the user service for the controller.
func WithUserService(userService service.UserService) ControllerOption {
	return func(c *baseController) error {
		if userService == nil {
			return ErrNoUserService
		}

		c.userService = userService

		return nil
	}
}

// baseController defines the dependencies that are required to be injected
// into a controller.
type baseController struct {
	conf   config.ServerConfig
	logger log.Logger
	tracer trace.Tracer

	authProvider *auth.Server

	systemService service.SystemService
	userService   service.UserService
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
