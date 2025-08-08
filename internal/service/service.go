package service

import (
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
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
func WithTracer(tracer tracing.Tracer) Option {
	return func(s *baseService) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		s.tracer = tracer
		return nil
	}
}

// WithOrganizationRepository sets the organization repository for the
// baseService.
func WithOrganizationRepository(organizationRepo repository.OrganizationRepository) Option {
	return func(s *baseService) error {
		if organizationRepo == nil {
			return ErrNoOrganizationRepository
		}

		s.organizationRepo = organizationRepo
		return nil
	}
}

// WithRoleRepository sets the organization repository for the
// baseService.
func WithRoleRepository(roleRepo repository.RoleRepository) Option {
	return func(s *baseService) error {
		if roleRepo == nil {
			return ErrNoOrganizationRepository
		}

		s.roleRepo = roleRepo
		return nil
	}
}

// WithUserRepository sets the user repository for the baseService.
func WithUserRepository(userRepo repository.UserRepository) Option {
	return func(s *baseService) error {
		if userRepo == nil {
			return ErrNoUserRepository
		}

		s.userRepo = userRepo
		return nil
	}
}

// WithUserTokenRepository sets the user token repository for the baseService.
func WithUserTokenRepository(userTokenRepo repository.UserTokenRepository) Option {
	return func(s *baseService) error {
		if userTokenRepo == nil {
			return ErrNoUserTokenRepository
		}

		s.userTokenRepo = userTokenRepo
		return nil
	}
}

// WithTodoRepository sets the todo repository for the baseService.
func WithTodoRepository(todoRepo repository.TodoRepository) Option {
	return func(s *baseService) error {
		if todoRepo == nil {
			return ErrNoTodoRepository
		}

		s.todoRepo = todoRepo
		return nil
	}
}

// WithLicenseService sets the license service for the baseService.
func WithLicenseService(licenseService LicenseService) Option {
	return func(s *baseService) error {
		if licenseService == nil {
			return ErrNoLicenseService
		}

		s.licenseService = licenseService
		return nil
	}
}

// WithPermissionService sets the permission service for the baseService.
func WithPermissionService(permissionService PermissionService) Option {
	return func(s *baseService) error {
		if permissionService == nil {
			return ErrNoPermissionService
		}

		s.permissionService = permissionService
		return nil
	}
}

// WithNotificationService sets the notification service for the baseService.
func WithNotificationService(notificationService NotificationService) Option {
	return func(s *baseService) error {
		if notificationService == nil {
			return ErrNoPermissionService
		}

		s.notificationService = notificationService
		return nil
	}
}

// baseService defines the dependencies that are required to interact with the
// core functionality.
type baseService struct {
	logger log.Logger
	tracer tracing.Tracer

	organizationRepo repository.OrganizationRepository
	roleRepo         repository.RoleRepository
	todoRepo         repository.TodoRepository
	userRepo         repository.UserRepository
	userTokenRepo    repository.UserTokenRepository

	licenseService      LicenseService
	permissionService   PermissionService
	notificationService NotificationService
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
