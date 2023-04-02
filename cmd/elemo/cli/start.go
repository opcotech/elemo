package cli

import (
	"net/http"
	"sync"
	"time"

	authManager "github.com/go-oauth2/oauth2/v4/manage"
	authServer "github.com/go-oauth2/oauth2/v4/server"
	authStore "github.com/go-oauth2/oauth2/v4/store"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/service"

	"github.com/go-oauth2/oauth2/v4/server"

	elemoHttp "github.com/opcotech/elemo/internal/transport/http"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Starts listening on the specified address.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("starting server", zap.Any("version", versionInfo))

		database, err := initDatabase()
		if err != nil {
			logger.Fatal("failed to initialize database", zap.Error(err))
		}

		systemService, err := service.NewSystemService(
			map[string]service.Pingable{
				"database": database,
			},
			versionInfo,
			service.WithLogger(logger.Named("system_service")),
			service.WithTracer(tracer),
		)

		authProvider, err := initAuthProvider(database)
		if err != nil {
			logger.Fatal("failed to initialize auth server", zap.Error(err))
		}

		httpServer, err := elemoHttp.NewServer(
			elemoHttp.WithAuthProvider(authProvider),
			elemoHttp.WithSystemService(systemService),
			elemoHttp.WithLogger(logger.Named("http_server")),
			elemoHttp.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize http server", zap.Error(err))
		}

		startServers(httpServer)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func initAuthProvider(db *neo4j.Database) (*authServer.Server, error) {
	/*clientStore, err := arangoStore.NewClientStore(
		arangoStore.WithClientStoreDatabase(db),
		arangoStore.WithClientStoreCollection(string(model.CollectionAuthClients)),
	)
	if err != nil {
		return nil, err
	}

	tokenStore, err := arangoStore.NewTokenStore(
		arangoStore.WithTokenStoreDatabase(db),
		arangoStore.WithTokenStoreCollection(string(model.CollectionAuthTokens)),
	)
	if err != nil {
		return nil, err
	}*/

	clientStore := authStore.NewClientStore()
	tokenStore, err := authStore.NewMemoryTokenStore()
	if err != nil {
		return nil, err
	}

	manager := authManager.NewDefaultManager()
	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetInternalErrorHandler(srv.InternalErrorHandler)
	srv.SetResponseErrorHandler(srv.ResponseErrorHandler)

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

	return s.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
}

func startMetricsServer() error {
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

	return s.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
}

func startServers(server elemoHttp.StrictServer) {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		err := startHTTPServer(server)
		logger.Fatal("failed to start HTTP server", zap.Error(err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startMetricsServer()
		logger.Fatal("failed to start HTTP metrics server", zap.Error(err))
		wg.Done()
	}(wg)

	wg.Wait()
}