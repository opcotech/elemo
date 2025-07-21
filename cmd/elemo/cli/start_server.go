package cli

import (
	"context"
	"net/http"
	"sync"
	"time"

	authStore "github.com/gabor-boros/go-oauth2-pg"
	authManager "github.com/go-oauth2/oauth2/v4/manage"
	authServer "github.com/go-oauth2/oauth2/v4/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

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
			logger.Fatal("failed to parse license", zap.Error(err))
		}

		cacheDB, err := initCacheDatabase()
		if err != nil {
			logger.Fatal("failed to initialize cache database", zap.Error(err))
		}

		graphDB, err := initGraphDatabase()
		if err != nil {
			logger.Fatal("failed to initialize graph database", zap.Error(err))
		}

		relDB, relDBPool, err := initRelationalDatabase()
		if err != nil {
			logger.Fatal("failed to initialize relational database", zap.Error(err))
		}

		messageQueue, err := queue.NewClient(
			queue.WithClientConfig(&cfg.Worker),
			queue.WithClientLogger(logger.Named("message_queue")),
			queue.WithClientTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize message queue", zap.Error(err))
		}
		defer func(messageQueue *queue.Client) {
			err := messageQueue.Close(context.Background())
			if err != nil {
				logger.Error("failed to close message queue", zap.Error(err))
			}
		}(messageQueue)

		licenseRepo, err := neo4j.NewLicenseRepository(
			neo4j.WithDatabase(graphDB),
			neo4j.WithRepositoryLogger(logger.Named("license_repository")),
			neo4j.WithRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize license repository", zap.Error(err))
		}

		var permissionRepo repository.PermissionRepository
		{
			repo, err := neo4j.NewPermissionRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("permission_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize permission repository", zap.Error(err))
			}

			permissionRepo, err = redis.NewCachedPermissionRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_permission_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize cached permission repository", zap.Error(err))
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
				logger.Fatal("failed to initialize organization repository", zap.Error(err))
			}

			organizationRepo, err = redis.NewCachedOrganizationRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_organization_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize cached organization repository", zap.Error(err))
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
				logger.Fatal("failed to initialize role repository", zap.Error(err))
			}

			roleRepo, err = redis.NewCachedRoleRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_role_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize cached role repository", zap.Error(err))
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
				logger.Fatal("failed to initialize user repository", zap.Error(err))
			}

			userRepo, err = redis.NewCachedUserRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_user_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize cached user repository", zap.Error(err))
			}
		}

		var todoRepo repository.TodoRepository
		{
			repo, err := neo4j.NewTodoRepository(
				neo4j.WithDatabase(graphDB),
				neo4j.WithRepositoryLogger(logger.Named("todo_repository")),
				neo4j.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize todo repository", zap.Error(err))
			}

			todoRepo, err = redis.NewCachedTodoRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_todo_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize cached todo repository", zap.Error(err))
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
				logger.Fatal("failed to initialize notification repository", zap.Error(err))
			}

			notificationRepo, err = redis.NewCachedNotificationRepository(
				repo,
				redis.WithDatabase(cacheDB),
				redis.WithRepositoryLogger(logger.Named("cached_notification_repository")),
				redis.WithRepositoryTracer(tracer),
			)
			if err != nil {
				logger.Fatal("failed to initialize cached notification repository", zap.Error(err))
			}
		}

		notificationService, err := service.NewNotificationService(
			notificationRepo,
			service.WithLogger(logger.Named("notification_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize notification service", zap.Error(err))
		}

		permissionService, err := service.NewPermissionService(
			permissionRepo,
			service.WithLogger(logger.Named("permission_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize permission service", zap.Error(err))
		}

		licenseService, err := service.NewLicenseService(
			license,
			licenseRepo,
			service.WithPermissionService(permissionService),
			service.WithLogger(logger.Named("license_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize license service", zap.Error(err))
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
			logger.Fatal("failed to initialize system service", zap.Error(err))
		}

		organizationService, err := service.NewOrganizationService(
			service.WithOrganizationRepository(organizationRepo),
			service.WithUserRepository(userRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("organization_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize organization service", zap.Error(err))
		}

		roleService, err := service.NewRoleService(
			service.WithRoleRepository(roleRepo),
			service.WithUserRepository(userRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("role_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize role service", zap.Error(err))
		}

		userService, err := service.NewUserService(
			service.WithUserRepository(userRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("user_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize user service", zap.Error(err))
		}

		todoService, err := service.NewTodoService(
			service.WithTodoRepository(todoRepo),
			service.WithPermissionService(permissionService),
			service.WithLicenseService(licenseService),
			service.WithLogger(logger.Named("todo_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize todo service", zap.Error(err))
		}

		authProvider, err := initAuthProvider(relDBPool)
		if err != nil {
			logger.Fatal("failed to initialize auth server", zap.Error(err))
		}

		httpServer, err := elemoHttp.NewServer(
			elemoHttp.WithConfig(cfg.Server),
			elemoHttp.WithAuthProvider(authProvider),
			elemoHttp.WithOrganizationService(organizationService),
			elemoHttp.WithRoleService(roleService),
			elemoHttp.WithUserService(userService),
			elemoHttp.WithTodoService(todoService),
			elemoHttp.WithSystemService(systemService),
			elemoHttp.WithLicenseService(licenseService),
			elemoHttp.WithPermissionService(permissionService),
			elemoHttp.WithNotificationService(notificationService),
			elemoHttp.WithLogger(logger.Named("http_server")),
			elemoHttp.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize http server", zap.Error(err))
		}

		systemLicenseExpiryTask, err := queue.NewSystemLicenseExpiryTask(license)
		if err != nil {
			logger.Fatal("failed to initialize system license expiry task", zap.Error(err))
		}

		taskScheduler, err := queue.NewScheduler(
			queue.WithSchedulerTask("@every 1m", systemLicenseExpiryTask),
			queue.WithSchedulerConfig(&cfg.Worker),
			queue.WithSchedulerLogger(logger.Named("task_scheduler")),
			queue.WithSchedulerTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize scheduler", zap.Error(err))
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
		logger.Fatal("failed to initialize http router", zap.Error(err))
	}

	logger.Info("starting HTTP server", zap.String("address", cfg.Server.Address))
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
	logger.Info("starting task scheduler")
	return scheduler.Start()
}

func startHTTPMetricsServer() error {
	router, err := elemoHttp.NewMetricsServer(&cfg.MetricsServer, tracer)
	if err != nil {
		logger.Fatal("failed to initialize metrics router", zap.Error(err))
	}

	logger.Info("starting HTTP metrics server", zap.String("address", cfg.MetricsServer.Address))
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
		logger.Fatal("failed to start task scheduler", zap.Error(err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startHTTPServer(server)
		logger.Fatal("failed to start HTTP server", zap.Error(err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startHTTPMetricsServer()
		logger.Fatal("failed to start HTTP metrics server", zap.Error(err))
		wg.Done()
	}(wg)

	wg.Wait()
}
