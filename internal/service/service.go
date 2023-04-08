package service

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

// Option defines a configuration option for the service.
type Option func(*baseService) error

// WithLogger sets the logger for the baseService.
func WithLogger(logger log.Logger) Option {
	return func(s *baseService) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		s.logger = logger
		return nil
	}
}

// WithTracer sets the tracer for the baseService.
func WithTracer(tracer trace.Tracer) Option {
	return func(s *baseService) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		s.tracer = tracer
		return nil
	}
}

// WithPermissionRepository sets the permission repository for the baseService.
func WithPermissionRepository(permissionRepo PermissionRepository) Option {
	return func(s *baseService) error {
		if permissionRepo == nil {
			return ErrNoPermissionRepository
		}

		s.permissionRepo = permissionRepo
		return nil
	}
}

// WithUserRepository sets the user repository for the baseService.
func WithUserRepository(userRepo UserRepository) Option {
	return func(s *baseService) error {
		if userRepo == nil {
			return ErrNoUserRepository
		}

		s.userRepo = userRepo
		return nil
	}
}

// baseService defines the dependencies that are required to interact with the
// core functionality.
type baseService struct {
	logger log.Logger
	tracer trace.Tracer

	userRepo       UserRepository
	permissionRepo PermissionRepository
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
