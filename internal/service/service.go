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

// WithRoleRepository sets the role repository for the
// baseService.
func WithRoleRepository(roleRepo repository.RoleRepository) Option {
	return func(s *baseService) error {
		if roleRepo == nil {
			return ErrNoRoleRepository
		}

		s.roleRepo = roleRepo
		return nil
	}
}

// WithTeamRepository sets the team repository for the
// baseService.
func WithTeamRepository(teamRepo repository.TeamRepository) Option {
	return func(s *baseService) error {
		if teamRepo == nil {
			return ErrNoTeamRepository
		}

		s.teamRepo = teamRepo
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

// WithNamespaceRepository sets the namespace repository for the baseService.
func WithNamespaceRepository(namespaceRepo repository.NamespaceRepository) Option {
	return func(s *baseService) error {
		if namespaceRepo == nil {
			return ErrNoNamespaceRepository
		}

		s.namespaceRepo = namespaceRepo
		return nil
	}
}

// WithProjectRepository sets the project repository for the baseService.
func WithProjectRepository(projectRepo repository.ProjectRepository) Option {
	return func(s *baseService) error {
		if projectRepo == nil {
			return ErrNoProjectRepository
		}

		s.projectRepo = projectRepo
		return nil
	}
}

// WithIssueRepository sets the issue repository for the baseService.
func WithIssueRepository(issueRepo repository.IssueRepository) Option {
	return func(s *baseService) error {
		if issueRepo == nil {
			return ErrNoIssueRepository
		}

		s.issueRepo = issueRepo
		return nil
	}
}

// WithAssignmentRepository sets the assignment repository for the baseService.
func WithAssignmentRepository(assignmentRepo repository.AssignmentRepository) Option {
	return func(s *baseService) error {
		if assignmentRepo == nil {
			return ErrNoAssignmentRepository
		}

		s.assignmentRepo = assignmentRepo
		return nil
	}
}

// WithLabelRepository sets the label repository for the baseService.
func WithLabelRepository(labelRepo repository.LabelRepository) Option {
	return func(s *baseService) error {
		if labelRepo == nil {
			return ErrNoLabelRepository
		}

		s.labelRepo = labelRepo
		return nil
	}
}

// WithDocumentRepository sets the document repository for the baseService.
func WithDocumentRepository(documentRepo repository.DocumentRepository) Option {
	return func(s *baseService) error {
		if documentRepo == nil {
			return ErrNoDocumentRepository
		}

		s.documentRepo = documentRepo
		return nil
	}
}

// WithFolderRepository sets the folder repository for the baseService.
func WithFolderRepository(folderRepo repository.FolderRepository) Option {
	return func(s *baseService) error {
		if folderRepo == nil {
			return ErrNoFolderRepository
		}

		s.folderRepo = folderRepo
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
			return ErrNoNotificationService
		}

		s.notificationService = notificationService
		return nil
	}
}

// WithSearchService sets the search service for the baseService.
func WithSearchService(searchService SearchService) Option {
	return func(s *baseService) error {
		if searchService == nil {
			return ErrNoSearchService
		}

		s.searchService = searchService
		return nil
	}
}

// WithSearchTaskEnqueuer sets the queue client used to schedule search index tasks.
func WithSearchTaskEnqueuer(enqueuer SearchTaskEnqueuer) Option {
	return func(s *baseService) error {
		if enqueuer == nil {
			return ErrNoSearchTaskEnqueuer
		}

		s.searchTaskEnqueuer = enqueuer
		return nil
	}
}

// WithEmailService sets the email service for the baseService.
func WithEmailService(emailService EmailService) Option {
	return func(s *baseService) error {
		if emailService == nil {
			return ErrNoEmailService
		}

		s.emailService = emailService
		return nil
	}
}

// WithStaticFileService sets the static file service for the baseService.
func WithStaticFileService(staticFileService StaticFileService) Option {
	return func(s *baseService) error {
		if staticFileService == nil {
			return ErrNoStaticFileService
		}

		s.staticFileService = staticFileService
		return nil
	}
}

// baseService defines the dependencies that are required to interact with the
// core functionality.
type baseService struct {
	logger log.Logger
	tracer tracing.Tracer

	organizationRepo repository.OrganizationRepository
	namespaceRepo    repository.NamespaceRepository
	projectRepo      repository.ProjectRepository
	issueRepo        repository.IssueRepository
	assignmentRepo   repository.AssignmentRepository
	labelRepo        repository.LabelRepository
	documentRepo     repository.DocumentRepository
	folderRepo       repository.FolderRepository
	roleRepo         repository.RoleRepository
	teamRepo         repository.TeamRepository
	todoRepo         repository.TodoRepository
	userRepo         repository.UserRepository
	userTokenRepo    repository.UserTokenRepository

	licenseService      LicenseService
	permissionService   PermissionService
	notificationService NotificationService
	searchService       SearchService
	searchTaskEnqueuer  SearchTaskEnqueuer
	emailService        EmailService
	staticFileService   StaticFileService
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
