package cli

import (
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/opcotech/elemo/internal/transport/asynq"
)

// startWorkerCmd represents the start command
var startWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the worker",
	Long: `Starts the worker processes and listens for prometheus metrics on the
configured port.`,
	Run: func(cmd *cobra.Command, args []string) {
		initTracer("worker")

		logger.Info("starting worker", zap.Any("version", versionInfo))

		if _, err := parseLicense(&cfg.License); err != nil {
			logger.Fatal("failed to parse license", zap.Error(err))
		}

		asynq.SetRateLimiter(cfg.Worker.RateLimit, cfg.Worker.RateLimitBurst)

		worker, err := asynq.NewWorker(
			asynq.WithWorkerConfig(&cfg.Worker),
			asynq.WithWorkerLogger(logger.Named("worker")),
			asynq.WithWorkerTracer(tracer),
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

func startWorkerServer(worker *asynq.Worker) error {
	return worker.Start()
}

func startWorkerMetricsServer() error {
	router, err := asynq.NewWorkerMetricsServer(&cfg.WorkerMetricsServer, tracer)
	if err != nil {
		logger.Fatal("failed to initialize metrics router", zap.Error(err))
	}

	logger.Info("starting HTTP metrics server", zap.String("address", cfg.MetricsServer.Address))
	s := &http.Server{
		Addr:              cfg.WorkerMetricsServer.Address,
		Handler:           router,
		ReadTimeout:       cfg.WorkerMetricsServer.ReadTimeout * time.Second,
		ReadHeaderTimeout: cfg.WorkerMetricsServer.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WorkerMetricsServer.WriteTimeout * time.Second,
	}

	return s.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
}

func startWorkerServers(worker *asynq.Worker) {
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
