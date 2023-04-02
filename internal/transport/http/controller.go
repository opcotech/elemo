package http

import (
	"errors"

	auth "github.com/go-oauth2/oauth2/v4/server"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/service"
)

const (
	DefaultLimit  = 10 // default limit for pagination
	DefaultOffset = 0  // default offset for pagination
)

var (
	ErrNoLogger        = errors.New("no logger provided")         // no logger provided
	ErrNoTracer        = errors.New("no tracer provided")         // no tracer provided
	ErrNoAuthProvider  = errors.New("no auth provider provided")  // no auth provider provided
	ErrNoSystemService = errors.New("no system service provided") // no system service provided
)

// ControllerOption is a function that can be used to configure a controller.
type ControllerOption func(*baseController) error

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

// baseController defines the dependencies that are required to be injected
// into a controller.
type baseController struct {
	logger log.Logger
	tracer trace.Tracer

	authProvider *auth.Server

	systemService service.SystemService
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
