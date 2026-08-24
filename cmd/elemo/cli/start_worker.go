package cli

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"

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
			logger.Fatal(context.Background(), "failed to parse license", slog.Any("error", err))
		}

		smtpClient, err := initSMTPClient(&cfg.SMTP)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize SMTP client", slog.Any("error", err))
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
		_ = emailService

		systemHealthCheckHandler, err := async.NewSystemHealthCheckTaskHandler(
			async.WithTaskLogger(logger.Named("system_health_check_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize system health check task handler", slog.Any("error", err))
		}

		systemLicenseExpiryTaskHandler, err := async.NewSystemLicenseExpiryTaskHandler(
			async.WithTaskEmailService(emailService),
			async.WithTaskLogger(logger.Named("system_license_expiry_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize system license expiry task handler", slog.Any("error", err))
		}

		graphDB, searchService, err := initSearchService()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search service", slog.Any("error", err))
		}

		messageQueue, err := queue.NewClient(
			queue.WithClientConfig(&cfg.Worker),
			queue.WithClientLogger(logger.Named("message_queue")),
			queue.WithClientTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize message queue", slog.Any("error", err))
		}

		searchIndexHandler, err := async.NewSearchIndexTaskHandler(
			async.WithTaskSearchService(searchService),
			async.WithTaskGraphDatabase(graphDB),
			async.WithTaskLogger(logger.Named("search_index_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search index task handler", slog.Any("error", err))
		}

		reindexBatchSize := cfg.Search.ReindexBatchSize
		searchReindexHandler, err := async.NewSearchReindexTaskHandler(
			async.WithTaskSearchService(searchService),
			async.WithTaskGraphDatabase(graphDB),
			async.WithTaskQueueClient(messageQueue),
			async.WithTaskReindexBatchSize(reindexBatchSize),
			async.WithTaskLogger(logger.Named("search_reindex_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search reindex task handler", slog.Any("error", err))
		}

		searchReindexBatchHandler, err := async.NewSearchReindexBatchTaskHandler(
			async.WithTaskSearchService(searchService),
			async.WithTaskGraphDatabase(graphDB),
			async.WithTaskLogger(logger.Named("search_reindex_batch_task")),
			async.WithTaskTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search reindex batch task handler", slog.Any("error", err))
		}

		async.SetRateLimiter(cfg.Worker.RateLimit, cfg.Worker.RateLimitBurst)
		worker, err := async.NewWorker(
			async.WithWorkerTaskHandler(queue.TaskTypeSystemHealthCheck, systemHealthCheckHandler),
			async.WithWorkerTaskHandler(queue.TaskTypeSystemLicenseExpiry, systemLicenseExpiryTaskHandler),
			async.WithWorkerTaskHandler(queue.TaskTypeSearchIndex, searchIndexHandler),
			async.WithWorkerTaskHandler(queue.TaskTypeSearchReindex, searchReindexHandler),
			async.WithWorkerTaskHandler(queue.TaskTypeSearchReindexBatch, searchReindexBatchHandler),
			async.WithWorkerConfig(&cfg.Worker),
			async.WithWorkerLogger(logger.Named("worker")),
			async.WithWorkerTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to create worker", slog.Any("error", err))
		}

		startWorkerServers(worker)
	},
}

func init() {
	startCmd.AddCommand(startWorkerCmd)
}

func startWorkerServer(worker *async.Worker) error {
	logger.Info(context.Background(), "starting worker server")
	return worker.Start()
}

func startWorkerMetricsServer() error {
	router, err := async.NewWorkerMetricsServer(&cfg.WorkerMetricsServer, tracer)
	if err != nil {
		logger.Fatal(context.Background(), "failed to initialize metrics router", slog.Any("error", err))
	}

	logger.Info(context.Background(), "starting worker metrics server", slog.String("address", cfg.MetricsServer.Address))
	s := &http.Server{
		Addr:              cfg.WorkerMetricsServer.Address,
		Handler:           router,
		ReadTimeout:       cfg.WorkerMetricsServer.ReadTimeout * time.Second,
		ReadHeaderTimeout: cfg.WorkerMetricsServer.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WorkerMetricsServer.WriteTimeout * time.Second,
	}

	if cfg.WorkerMetricsServer.TLS.CertFile != "" && cfg.WorkerMetricsServer.TLS.KeyFile != "" {
		return s.ListenAndServeTLS(cfg.WorkerMetricsServer.TLS.CertFile, cfg.WorkerMetricsServer.TLS.KeyFile)
	}

	return s.ListenAndServe()
}

func startWorkerServers(worker *async.Worker) {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		err := startWorkerServer(worker)
		logger.Fatal(context.Background(), "failed to start async worker", slog.Any("error", err))
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		err := startWorkerMetricsServer()
		logger.Fatal(context.Background(), "failed to start worker metrics server", slog.Any("error", err))
		wg.Done()
	}(wg)

	wg.Wait()
}
