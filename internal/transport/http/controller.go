package http

import (
	authServer "github.com/go-oauth2/oauth2/v4/server"

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
func WithTracer(tracer tracing.Tracer) ControllerOption {
	return func(c *baseController) error {
		if tracer == nil {
			return ErrNoTracer
		}

		c.tracer = tracer

		return nil
	}
}

// WithAuthProvider sets the authServer provider for the controller.
func WithAuthProvider(authProvider *authServer.Server) ControllerOption {
	return func(c *baseController) error {
		if authProvider == nil {
			return ErrNoAuthProvider
		}

		c.authProvider = authProvider

		return nil
	}
}

// WithPermissionService sets the permission service for the controller.
func WithPermissionService(permissionService service.PermissionService) ControllerOption {
	return func(c *baseController) error {
		if permissionService == nil {
			return ErrNoPermissionService
		}

		c.permissionService = permissionService

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

// WithLicenseService sets the license service for the controller.
func WithLicenseService(licenseService service.LicenseService) ControllerOption {
	return func(c *baseController) error {
		if licenseService == nil {
			return ErrNoLicenseService
		}

		c.licenseService = licenseService

		return nil
	}
}

// WithOrganizationService sets the organization service for the controller.
func WithOrganizationService(organizationService service.OrganizationService) ControllerOption {
	return func(c *baseController) error {
		if organizationService == nil {
			return ErrNoOrganizationService
		}

		c.organizationService = organizationService

		return nil
	}
}

// WithRoleService sets the role service for the controller.
func WithRoleService(roleService service.RoleService) ControllerOption {
	return func(c *baseController) error {
		if roleService == nil {
			return ErrNoRoleService
		}

		c.roleService = roleService

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

// WithTodoService sets the todo service for the controller.
func WithTodoService(todoService service.TodoService) ControllerOption {
	return func(c *baseController) error {
		if todoService == nil {
			return ErrNoTodoService
		}

		c.todoService = todoService

		return nil
	}
}

// WithNotificationService sets the notification service for the controller.
func WithNotificationService(notificationService service.NotificationService) ControllerOption {
	return func(c *baseController) error {
		if notificationService == nil {
			return ErrNoTodoService
		}

		c.notificationService = notificationService

		return nil
	}
}

// baseController defines the dependencies that are required to be injected
// into a controller.
type baseController struct {
	conf   config.ServerConfig
	logger log.Logger
	tracer tracing.Tracer

	authProvider *authServer.Server

	organizationService service.OrganizationService
	roleService         service.RoleService
	userService         service.UserService
	todoService         service.TodoService
	systemService       service.SystemService
	licenseService      service.LicenseService
	permissionService   service.PermissionService
	notificationService service.NotificationService
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
