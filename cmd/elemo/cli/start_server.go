package cli

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"log/slog"

	authStore "github.com/gabor-boros/go-oauth2-pg"
	authManager "github.com/go-oauth2/oauth2/v4/manage"
	authServer "github.com/go-oauth2/oauth2/v4/server"
	"github.com/spf13/cobra"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"

	elemoHttp "github.com/opcotech/elemo/internal/transport/http"
)

// startServerCmd represents the start command
var startServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Starts listening on the specified address.`,
	Run: func(_ *cobra.Command, _ []string) {
		initTracer("server")

		license, err := parseLicense(&cfg.License)
		if err != nil {
			logger.Fatal(context.Background(), "failed to parse license", slog.Any("error", err))
		}

		cacheDB, err := initCacheDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize cache database", slog.Any("error", err))
		}

		searchDB, searchRepo, err := initSearchDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search database", slog.Any("error", err))
		}

		graphDB, err := initGraphDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize graph database", slog.Any("error", err))
		}

		relDB, relDBPool, err := initRelationalDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize relational database", slog.Any("error", err))
		}

		smtpClient, err := initSMTPClient(&cfg.SMTP)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize SMTP client", slog.Any("error", err))
		}

		messageQueue, err := queue.NewClient(
			queue.WithClientConfig(&cfg.Worker),
			queue.WithClientLogger(logger.Named("message_queue")),
			queue.WithClientTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize message queue", slog.Any("error", err))
		}
		defer func(messageQueue *queue.Client) {
			err := messageQueue.Close(context.Background())
			if err != nil {
				logger.Error(context.Background(), "failed to close message queue", slog.Any("error", err))
			}
		}(messageQueue)

		s3Client, err := repository.NewS3Client(context.Background(), &cfg.S3Storage)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize S3 client", slog.Any("error", err))
		}

		s3Storage, err := repository.NewStorage(
			repository.WithStorageClient(s3Client),
			repository.WithStorageBucket(cfg.S3Storage.Bucket),
			repository.WithStorageLogger(logger.Named("static_file_storage")),
			repository.WithStorageTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize storage", slog.Any("error", err))
		}

		licenseRepo, err := repository.NewNeo4jLicenseRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("license_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize license repository", slog.Any("error", err))
		}

		var permissionRepo repository.PermissionRepository
		{
			repo, err := repository.NewNeo4jPermissionRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("permission_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize permission repository", slog.Any("error", err))
			}

			permissionRepo, err = repository.NewCachedPermissionRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_permission_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached permission repository", slog.Any("error", err))
			}
		}

		var organizationRepo repository.OrganizationRepository
		{
			repo, err := repository.NewNeo4jOrganizationRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("organization_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize organization repository", slog.Any("error", err))
			}

			organizationRepo, err = repository.NewCachedOrganizationRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_organization_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached organization repository", slog.Any("error", err))
			}
		}

		var roleRepo repository.RoleRepository
		{
			repo, err := repository.NewNeo4jRoleRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("role_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize role repository", slog.Any("error", err))
			}

			roleRepo, err = repository.NewCachedRoleRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_role_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached role repository", slog.Any("error", err))
			}
		}

		var teamRepo repository.TeamRepository
		{
			repo, err := repository.NewNeo4jTeamRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("team_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize team repository", slog.Any("error", err))
			}

			teamRepo, err = repository.NewCachedTeamRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_team_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached team repository", slog.Any("error", err))
			}
		}

		var userRepo repository.UserRepository
		{
			repo, err := repository.NewNeo4jUserRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("user_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize user repository", slog.Any("error", err))
			}

			userRepo, err = repository.NewCachedUserRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_user_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached user repository", slog.Any("error", err))
			}
		}

		var userTokenRepo repository.UserTokenRepository
		{
			repo, err := repository.NewUserTokenRepository(
				repository.WithPGDatabase(relDB),
				repository.WithPGRepositoryLogger(logger.Named("user_token_repository")),
				repository.WithPGRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize user token repository", slog.Any("error", err))
			}

			userTokenRepo = repo
		}

		var todoRepo repository.TodoRepository
		{
			repo, err := repository.NewNeo4jTodoRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("todo_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize todo repository", slog.Any("error", err))
			}

			todoRepo, err = repository.NewCachedTodoRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_todo_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached todo repository", slog.Any("error", err))
			}
		}

		var namespaceRepo repository.NamespaceRepository
		{
			repo, err := repository.NewNeo4jNamespaceRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("namespace_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize namespace repository", slog.Any("error", err))
			}

			namespaceRepo, err = repository.NewCachedNamespaceRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_namespace_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached namespace repository", slog.Any("error", err))
			}
		}

		var projectRepo repository.ProjectRepository
		{
			repo, err := repository.NewNeo4jProjectRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("project_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize project repository", slog.Any("error", err))
			}

			projectRepo, err = repository.NewCachedProjectRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_project_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached project repository", slog.Any("error", err))
			}
		}

		var issueRepo repository.IssueRepository
		{
			repo, err := repository.NewNeo4jIssueRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("issue_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize issue repository", slog.Any("error", err))
			}

			issueRepo, err = repository.NewCachedIssueRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_issue_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached issue repository", slog.Any("error", err))
			}
		}

		var assignmentRepo repository.AssignmentRepository
		{
			repo, err := repository.NewNeo4jAssignmentRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("assignment_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize assignment repository", slog.Any("error", err))
			}

			assignmentRepo, err = repository.NewCachedAssignmentRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_assignment_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached assignment repository", slog.Any("error", err))
			}
		}

		var labelRepo repository.LabelRepository
		{
			repo, err := repository.NewNeo4jLabelRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("label_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize label repository", slog.Any("error", err))
			}

			labelRepo, err = repository.NewCachedLabelRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_label_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached label repository", slog.Any("error", err))
			}
		}

		var documentRepo repository.DocumentRepository
		{
			repo, err := repository.NewNeo4jDocumentRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("document_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize document repository", slog.Any("error", err))
			}

			documentRepo, err = repository.NewCachedDocumentRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_document_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached document repository", slog.Any("error", err))
			}
		}

		var folderRepo repository.FolderRepository
		{
			repo, err := repository.NewNeo4jFolderRepository(
				repository.WithNeo4jDatabase(graphDB),
				repository.WithNeo4jRepositoryLogger(logger.Named("folder_repository")),
				repository.WithNeo4jRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize folder repository", slog.Any("error", err))
			}

			folderRepo, err = repository.NewCachedFolderRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_folder_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached folder repository", slog.Any("error", err))
			}
		}

		var notificationRepo repository.NotificationRepository
		{
			repo, err := repository.NewNotificationRepository(
				repository.WithPGDatabase(relDB),
				repository.WithPGRepositoryLogger(logger.Named("notification_repository")),
				repository.WithPGRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize notification repository", slog.Any("error", err))
			}

			notificationRepo, err = repository.NewCachedNotificationRepository(
				repo,
				repository.WithRedisDatabase(cacheDB),
				repository.WithRedisRepositoryLogger(logger.Named("cached_notification_repository")),
				repository.WithRedisRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached notification repository", slog.Any("error", err))
			}
		}

		var staticFileRepo repository.StaticFileRepository
		{
			repo, err := repository.NewStaticFileRepository(
				repository.WithS3Storage(s3Storage),
				repository.WithS3RepositoryLogger(logger.Named("static_file_repository")),
				repository.WithS3RepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize static file repository", slog.Any("error", err))
			}

			staticFileRepo = repo
		}

		notificationService, err := service.NewNotificationService(
			notificationRepo,
			service.WithLogger(logger.Named("notification_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize notification service", slog.Any("error", err))
		}

		permissionService, err := service.NewPermissionService(
			permissionRepo,
			service.WithRoleRepository(roleRepo),
			service.WithLogger(logger.Named("permission_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize permission service", slog.Any("error", err))
		}

		searchService, err := service.NewSearchService(
			searchRepo,
			service.WithPermissionService(permissionService),
			service.WithSearchTaskEnqueuer(messageQueue),
			service.WithLogger(logger.Named("search_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search service", slog.Any("error", err))
		}

		licenseService, err := service.NewLicenseService(
			license,
			licenseRepo,
			service.WithPermissionService(permissionService),
			service.WithLogger(logger.Named("license_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize license service", slog.Any("error", err))
		}

		systemService, err := service.NewSystemService(
			map[model.HealthCheckComponent]service.Pingable{
				model.HealthCheckComponentCacheDB:      cacheDB,
				model.HealthCheckComponentGraphDB:      graphDB,
				model.HealthCheckComponentRelationalDB: relDB,
				model.HealthCheckComponentLicense:      licenseService,
				model.HealthCheckComponentMessageQueue: messageQueue,
				model.HealthCheckComponentS3Storage:    s3Storage,
				model.HealthCheckComponentSearch:       searchDB,
			},
			versionInfo,
			service.WithLogger(logger.Named("system_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize system service", slog.Any("error", err))
		}

		roleService, err := service.NewRoleService(
			service.WithRoleRepository(roleRepo),
			service.WithUserRepository(userRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithOrganizationRepository(organizationRepo),
			service.WithNotificationService(notificationService),
			service.WithLogger(logger.Named("role_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize role service", slog.Any("error", err))
		}

		teamService, err := service.NewTeamService(
			service.WithTeamRepository(teamRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("team_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize team service", slog.Any("error", err))
		}

		userService, err := service.NewUserService(
			service.WithUserRepository(userRepo),
			service.WithUserTokenRepository(userTokenRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("user_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize user service", slog.Any("error", err))
		}

		todoService, err := service.NewTodoService(
			service.WithTodoRepository(todoRepo),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("todo_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize todo service", slog.Any("error", err))
		}

		emailService, err := service.NewEmailService(
			smtpClient,
			cfg.Template.Directory,
			&cfg.SMTP,
			service.WithLogger(logger.Named("email_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize email service", slog.Any("error", err))
		}

		staticFileService, err := service.NewStaticFileService(
			staticFileRepo,
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("static_file_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize static file service", slog.Any("error", err))
		}

		namespaceService, err := service.NewNamespaceService(
			service.WithNamespaceRepository(namespaceRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("namespace_service")),
			service.WithTracer(tracer),
			service.WithSearchService(searchService),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize namespace service", slog.Any("error", err))
		}

		projectService, err := service.NewProjectService(
			service.WithProjectRepository(projectRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("project_service")),
			service.WithTracer(tracer),
			service.WithSearchService(searchService),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize project service", slog.Any("error", err))
		}

		issueService, err := service.NewIssueService(
			service.WithIssueRepository(issueRepo),
			service.WithAssignmentRepository(assignmentRepo),
			service.WithLabelRepository(labelRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("issue_service")),
			service.WithTracer(tracer),
			service.WithSearchService(searchService),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize issue service", slog.Any("error", err))
		}

		documentService, err := service.NewDocumentService(
			service.WithDocumentRepository(documentRepo),
			service.WithLicenseService(licenseService),
			service.WithPermissionService(permissionService),
			service.WithStaticFileService(staticFileService),
			service.WithLogger(logger.Named("document_service")),
			service.WithTracer(tracer),
			service.WithSearchService(searchService),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize document service", slog.Any("error", err))
		}

		folderService, err := service.NewFolderService(
			service.WithFolderRepository(folderRepo),
			service.WithPermissionService(permissionService),
			service.WithLogger(logger.Named("folder_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize folder service", slog.Any("error", err))
		}

		labelService, err := service.NewLabelService(
			service.WithLabelRepository(labelRepo),
			service.WithLogger(logger.Named("label_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize label service", slog.Any("error", err))
		}

		organizationService, err := service.NewOrganizationService(
			service.WithOrganizationRepository(organizationRepo),
			service.WithUserRepository(userRepo),
			service.WithUserTokenRepository(userTokenRepo),
			service.WithRoleRepository(roleRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithEmailService(emailService),
			service.WithNotificationService(notificationService),
			service.WithLogger(logger.Named("organization_service")),
			service.WithTracer(tracer),
			service.WithSearchService(searchService),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize organization service", slog.Any("error", err))
		}

		authProvider, err := initAuthProvider(relDBPool)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize auth server", slog.Any("error", err))
		}

		httpServer, err := elemoHttp.NewServer(
			elemoHttp.WithConfig(cfg.Server),
			elemoHttp.WithAuthProvider(authProvider),
			elemoHttp.WithOrganizationService(organizationService),
			elemoHttp.WithNamespaceService(namespaceService),
			elemoHttp.WithProjectService(projectService),
			elemoHttp.WithIssueService(issueService),
			elemoHttp.WithDocumentService(documentService),
			elemoHttp.WithFolderService(folderService),
			elemoHttp.WithLabelService(labelService),
			elemoHttp.WithRoleService(roleService),
			elemoHttp.WithTeamService(teamService),
			elemoHttp.WithUserService(userService),
			elemoHttp.WithTodoService(todoService),
			elemoHttp.WithEmailService(emailService),
			elemoHttp.WithSystemService(systemService),
			elemoHttp.WithLicenseService(licenseService),
			elemoHttp.WithPermissionService(permissionService),
			elemoHttp.WithNotificationService(notificationService),
			elemoHttp.WithSearchService(searchService),
			elemoHttp.WithLogger(logger.Named("http_server")),
			elemoHttp.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize http server", slog.Any("error", err))
		}

		systemLicenseExpiryTask, err := queue.NewSystemLicenseExpiryTask(license)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize system license expiry task", slog.Any("error", err))
		}

		taskScheduler, err := queue.NewScheduler(
			queue.WithSchedulerTask("@every 1m", systemLicenseExpiryTask),
			queue.WithSchedulerConfig(&cfg.Worker),
			queue.WithSchedulerLogger(logger.Named("task_scheduler")),
			queue.WithSchedulerTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize scheduler", slog.Any("error", err))
		}

		startHTTPServers(httpServer, taskScheduler)
	},
}

func init() {
	startCmd.AddCommand(startServerCmd)
}

func initAuthProvider(pool repository.PGPool) (*authServer.Server, error) {
	storeLogger := &authStoreLogger{
		logger: logger.Named("auth_store"),
	}

	pgxPool, ok := repository.AsPgxPool(pool)
	if !ok {
		return nil, errors.New("auth store requires a pgx connection pool")
	}

	clientStore, err := authStore.NewClientStore(
		authStore.WithClientStoreTable(authStore.DefaultClientStoreTable),
		authStore.WithClientStoreConnPool(pgxPool),
		authStore.WithClientStoreLogger(storeLogger),
	)
	if err != nil {
		return nil, err
	}

	if err := clientStore.InitTable(context.Background()); err != nil {
		return nil, err
	}

	tokenStore, err := authStore.NewTokenStore(
		authStore.WithTokenStoreTable(authStore.DefaultTokenStoreTable),
		authStore.WithTokenStoreConnPool(pgxPool),
		authStore.WithTokenStoreLogger(storeLogger),
	)
	if err != nil {
		return nil, err
	}

	if err := tokenStore.InitTable(context.Background()); err != nil {
		return nil, err
	}

	manager := authManager.NewDefaultManager()
	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)

	srv := authServer.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(authServer.ClientFormHandler)
	srv.SetInternalErrorHandler(srv.InternalErrorHandler)
	srv.SetResponseErrorHandler(srv.ResponseErrorHandler)
	srv.SetPreRedirectErrorHandler(srv.PreRedirectErrorHandler)

	return srv, nil
}

func startHTTPServer(server elemoHttp.StrictServer) error {
	router, err := elemoHttp.NewRouter(server, &cfg.Server, tracer)
	if err != nil {
		logger.Fatal(context.Background(), "failed to initialize http router", slog.Any("error", err))
	}

	logger.Info(context.Background(), "starting HTTP server", slog.String("address", cfg.Server.Address))
	s := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout * time.Second,
		ReadHeaderTimeout: cfg.Server.ReadTimeout * time.Second,
		WriteTimeout:      cfg.Server.WriteTimeout * time.Second,
	}

	if cfg.Server.TLS.CertFile != "" && cfg.Server.TLS.KeyFile != "" {
		return s.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
	}

	return s.ListenAndServe()
}

func startSchedulerServer(scheduler *queue.Scheduler) error {
	logger.Info(context.Background(), "starting task scheduler")
	return scheduler.Start()
}

func startHTTPMetricsServer() error {
	router, err := elemoHttp.NewMetricsServer(&cfg.MetricsServer, tracer)
	if err != nil {
		logger.Fatal(context.Background(), "failed to initialize metrics router", slog.Any("error", err))
	}

	logger.Info(context.Background(), "starting HTTP metrics server", slog.String("address", cfg.MetricsServer.Address))
	s := &http.Server{
		Addr:              cfg.MetricsServer.Address,
		Handler:           router,
		ReadTimeout:       cfg.MetricsServer.ReadTimeout * time.Second,
		ReadHeaderTimeout: cfg.MetricsServer.ReadTimeout * time.Second,
		WriteTimeout:      cfg.MetricsServer.WriteTimeout * time.Second,
	}

	if cfg.MetricsServer.TLS.CertFile != "" && cfg.MetricsServer.TLS.KeyFile != "" {
		return s.ListenAndServeTLS(cfg.MetricsServer.TLS.CertFile, cfg.MetricsServer.TLS.KeyFile)
	}

	return s.ListenAndServe()
}

func startHTTPServers(server elemoHttp.StrictServer, taskScheduler *queue.Scheduler) {
	wg := new(sync.WaitGroup)
	wg.Add(3)

	go func(wg *sync.WaitGroup) {
		err := startSchedulerServer(taskScheduler)
		logger.Fatal(context.Background(), "failed to start task scheduler", slog.Any("error", err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startHTTPServer(server)
		logger.Fatal(context.Background(), "failed to start HTTP server", slog.Any("error", err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startHTTPMetricsServer()
		logger.Fatal(context.Background(), "failed to start HTTP metrics server", slog.Any("error", err))
		wg.Done()
	}(wg)

	wg.Wait()
}
