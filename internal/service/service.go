package service

import (
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

var (
	ErrNoLogger        = errors.New("no logger provided")         // no logger provided
	ErrNoTracer        = errors.New("no tracer provided")         // no tracer provided
	ErrNoSystemService = errors.New("no system service provided") // no system service provided
)

// Option defines a configuration option for the service.
type Option func(*baseService) error

// WithLogger sets the logger for the baseService.
func WithLogger(logger log.Logger) Option {
	return func(s *baseService) error {
		if logger == nil {
			return ErrNoLogger
		}

		s.logger = logger
		return nil
	}
}

// WithTracer sets the tracer for the baseService.
func WithTracer(tracer trace.Tracer) Option {
	return func(s *baseService) error {
		if tracer == nil {
			return ErrNoTracer
		}

		s.tracer = tracer
		return nil
	}
}

// WithSystemService sets the system baseService for the baseService.
func WithSystemService(systemService SystemService) Option {
	return func(s *baseService) error {
		if systemService == nil {
			return ErrNoSystemService
		}

		s.systemService = systemService
		return nil
	}
}

// baseService defines the dependencies that are required to interact with the
// core functionality.
type baseService struct {
	logger log.Logger
	tracer trace.Tracer

	systemService SystemService
}

// newService creates a new baseService and defines the default values. Those
// options that are unique to a specific service are defined in the  concrete
// baseService implementation's constructor. For an example see NewSystemService.
func newService(opts ...Option) (*baseService, error) {
	s := &baseService{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}
