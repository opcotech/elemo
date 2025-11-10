package cli

import (
	"context"
	"net/http"
	"sync"
	"time"

	"log/slog"

	authStore "github.com/gabor-boros/go-oauth2-pg"
	authManager "github.com/go-oauth2/oauth2/v4/manage"
	authServer "github.com/go-oauth2/oauth2/v4/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"
	"github.com/opcotech/elemo/internal/repository/redis"
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

		licenseRepo, err := neo4j.NewLicenseRepository(
			neo4j.WithDatabase(graphDB),
			neo4j.WithRepositoryLogger(logger.Named("license_repository")),
			neo4j.WithRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize license repository", slog.Any("error", err))
		}

		var permissionRepo repository.PermissionRepository
		{
			repo, err := neo4j.NewPermissionRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("permission_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize permission repository", slog.Any("error", err))
			}

			permissionRepo, err = redis.NewCachedPermissionRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_permission_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached permission repository", slog.Any("error", err))
			}
		}

		var organizationRepo repository.OrganizationRepository
		{
			repo, err := neo4j.NewOrganizationRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("organization_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize organization repository", slog.Any("error", err))
			}

			organizationRepo, err = redis.NewCachedOrganizationRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_organization_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached organization repository", slog.Any("error", err))
			}
		}

		var roleRepo repository.RoleRepository
		{
			repo, err := neo4j.NewRoleRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("role_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize role repository", slog.Any("error", err))
			}

			roleRepo, err = redis.NewCachedRoleRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_role_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached role repository", slog.Any("error", err))
			}
		}

		var userRepo repository.UserRepository
		{
			repo, err := neo4j.NewUserRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("user_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize user repository", slog.Any("error", err))
			}

			userRepo, err = redis.NewCachedUserRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_user_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached user repository", slog.Any("error", err))
			}
		}

		var userTokenRepo repository.UserTokenRepository
		{
			repo, err := pg.NewUserTokenRepository(
				pg.WithDatabase(relDB),
				pg.WithRepositoryLogger(logger.Named("user_token_repository")),
				pg.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize user token repository", slog.Any("error", err))
			}

			userTokenRepo = repo
		}

		var todoRepo repository.TodoRepository
		{
			repo, err := neo4j.NewTodoRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("todo_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize todo repository", slog.Any("error", err))
			}

			todoRepo, err = redis.NewCachedTodoRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_todo_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached todo repository", slog.Any("error", err))
			}
		}

		var notificationRepo repository.NotificationRepository
		{
			repo, err := pg.NewNotificationRepository(
				pg.WithDatabase(relDB),
				pg.WithRepositoryLogger(logger.Named("notification_repository")),
				pg.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize notification repository", slog.Any("error", err))
			}

			notificationRepo, err = redis.NewCachedNotificationRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_notification_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize cached notification repository", slog.Any("error", err))
			}
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
			service.WithLogger(logger.Named("permission_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize permission service", slog.Any("error", err))
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
			service.WithPermissionService(permissionService),
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
			elemoHttp.WithRoleService(roleService),
			elemoHttp.WithUserService(userService),
			elemoHttp.WithTodoService(todoService),
			elemoHttp.WithEmailService(emailService),
			elemoHttp.WithSystemService(systemService),
			elemoHttp.WithLicenseService(licenseService),
			elemoHttp.WithPermissionService(permissionService),
			elemoHttp.WithNotificationService(notificationService),
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

func initAuthProvider(pool pg.Pool) (*authServer.Server, error) {
	storeLogger := &authStoreLogger{
		logger: logger.Named("auth_store"),
	}

	clientStore, err := authStore.NewClientStore(
		authStore.WithClientStoreTable(authStore.DefaultClientStoreTable),
		authStore.WithClientStoreConnPool(pool.(*pgxpool.Pool)),
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
		authStore.WithTokenStoreConnPool(pool.(*pgxpool.Pool)),
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
