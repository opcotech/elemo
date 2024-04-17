package cli

import (
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/async"
)

// startWorkerCmd represents the start command
var startWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the worker",
	Long: `Starts the worker processes and listens for prometheus metrics on the
configured port.`,
	Run: func(_ *cobra.Command, _ []string) {
		initTracer("worker")

		if _, err := parseLicense(&cfg.License); err != nil {
			logger.Fatal("failed to parse license", zap.Error(err))
		}

		smtpClient, err := initSMTPClient(&cfg.SMTP)
		if err != nil {
			logger.Fatal("failed to initialize SMTP client", zap.Error(err))
		}

		emailService, err := service.NewEmailService(
			smtpClient,
			cfg.Template.Directory,
			&cfg.SMTP,
			service.WithLogger(logger.Named("email_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize email service", zap.Error(err))
		}
		_ = emailService

		systemHealthCheckHandler, err := async.NewSystemHealthCheckTaskHandler(
			async.WithTaskLogger(logger.Named("system_health_check_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize system health check task handler", zap.Error(err))
		}

		systemLicenseExpiryTaskHandler, err := async.NewSystemLicenseExpiryTaskHandler(
			async.WithTaskEmailService(emailService),
			async.WithTaskLogger(logger.Named("system_license_expiry_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to initialize system license expiry task handler", zap.Error(err))
		}

		async.SetRateLimiter(cfg.Worker.RateLimit, cfg.Worker.RateLimitBurst)
		worker, err := async.NewWorker(
			async.WithWorkerTaskHandler(queue.TaskTypeSystemHealthCheck, systemHealthCheckHandler),
			async.WithWorkerTaskHandler(queue.TaskTypeSystemLicenseExpiry, systemLicenseExpiryTaskHandler),
			async.WithWorkerConfig(&cfg.Worker),
			async.WithWorkerLogger(logger.Named("worker")),
			async.WithWorkerTracer(tracer),
		)
		if err != nil {
			logger.Fatal("failed to create worker", zap.Error(err))
		}

		startWorkerServers(worker)
	},
}

func init() {
	startCmd.AddCommand(startWorkerCmd)
}

func startWorkerServer(worker *async.Worker) error {
	logger.Info("starting worker server")
	return worker.Start()
}

func startWorkerMetricsServer() error {
	router, err := async.NewWorkerMetricsServer(&cfg.WorkerMetricsServer, tracer)
	if err != nil {
		logger.Fatal("failed to initialize metrics router", zap.Error(err))
	}

	logger.Info("starting worker metrics server", zap.String("address", cfg.MetricsServer.Address))
	s := &http.Server{
		Addr:              cfg.WorkerMetricsServer.Address,
		Handler:           router,
		ReadTimeout:       cfg.WorkerMetricsServer.ReadTimeout * time.Second,
		ReadHeaderTimeout: cfg.WorkerMetricsServer.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WorkerMetricsServer.WriteTimeout * time.Second,
	}

	return s.ListenAndServeTLS(cfg.WorkerMetricsServer.TLS.CertFile, cfg.WorkerMetricsServer.TLS.KeyFile)
}

func startWorkerServers(worker *async.Worker) {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		err := startWorkerServer(worker)
		logger.Fatal("failed to start async worker", zap.Error(err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startWorkerMetricsServer()
		logger.Fatal("failed to start worker metrics server", zap.Error(err))
		wg.Done()
	}(wg)

	wg.Wait()
}
