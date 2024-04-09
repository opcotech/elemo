package asynq

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	elemoHttp "github.com/opcotech/elemo/internal/transport/http"
)

const (
	MessageQueueDefaultPriority = "default" // The default queue name.
	MessageQueueLowPriority     = "low"     // The low priority queue name.
	MessageQueueHighPriority    = "high"    // The high priority queue name.

	PathRoot    = "/"
	PathMetrics = "/metrics"
)

// WorkerOption is a function that can be used to configure an async worker.
type WorkerOption func(*Worker) error

// WithWorkerConfig sets the config for the worker.
func WithWorkerConfig(conf *config.WorkerConfig) WorkerOption {
	return func(w *Worker) error {
		if conf == nil {
			return config.ErrNoConfig
		}

		w.conf = conf
		return nil
	}
}

// WithWorkerTaskHandler sets a task handler for the worker.
func WithWorkerTaskHandler(taskType TaskType, handler asynq.Handler) WorkerOption {
	return func(w *Worker) error {
		if handler == nil {
			return ErrNoTaskHandler
		}

		w.handlers[taskType] = handler
		return nil
	}
}

// WithWorkerLogger sets the logger for the worker.
func WithWorkerLogger(logger log.Logger) WorkerOption {
	return func(w *Worker) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		w.logger = logger

		return nil
	}
}

// WithWorkerTracer sets the tracer for the worker.
func WithWorkerTracer(tracer tracing.Tracer) WorkerOption {
	return func(w *Worker) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		w.tracer = tracer

		return nil
	}
}

// Worker is the async worker.
type Worker struct {
	conf   *config.WorkerConfig
	logger log.Logger
	tracer tracing.Tracer

	*asynq.ServeMux
	server *asynq.Server

	handlers map[TaskType]asynq.Handler
}

// Start starts the async worker.
func (w *Worker) Start() error {
	return w.server.Run(w)
}

// Shutdown gracefully shuts down the async worker.
func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

// NewWorker returns a new async worker. Before creating a worker, the rate
// limiter should be initialized first, otherwise the worker will not be able
// to start and will return an error.
func NewWorker(opts ...WorkerOption) (*Worker, error) {
	w := &Worker{
		logger:   log.DefaultLogger(),
		tracer:   tracing.NoopTracer(),
		handlers: make(map[TaskType]asynq.Handler),
		ServeMux: asynq.NewServeMux(),
	}

	for _, opt := range opts {
		if err := opt(w); err != nil {
			return nil, err
		}
	}
	logLevel := asynq.InfoLevel
	if w.conf.LogLevel != "" {
		if err := logLevel.Set(w.conf.LogLevel); err != nil {
			return nil, log.ErrInvalidLogLevel
		}
	}

	if rateLimiter == nil {
		return nil, ErrNoRateLimiter
	}

	w.server = asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:         w.conf.Broker.Address(),
			Username:     w.conf.Broker.Username,
			Password:     w.conf.Broker.Password,
			DB:           w.conf.Broker.Database,
			DialTimeout:  w.conf.Broker.DialTimeout * time.Second,
			ReadTimeout:  w.conf.Broker.ReadTimeout * time.Second,
			WriteTimeout: w.conf.Broker.WriteTimeout * time.Second,
			PoolSize:     w.conf.Broker.PoolSize,
		},
		asynq.Config{
			Concurrency:              w.conf.Concurrency,
			StrictPriority:           w.conf.StrictPriority,
			ShutdownTimeout:          w.conf.ShutdownTimeout * time.Second,
			HealthCheckInterval:      w.conf.HealthCheckInterval * time.Second,
			DelayedTaskCheckInterval: w.conf.DelayedTaskCheckInterval * time.Second,
			GroupGracePeriod:         w.conf.GroupGracePeriod * time.Second,
			GroupMaxDelay:            w.conf.GroupMaxDelay * time.Second,
			GroupMaxSize:             w.conf.GroupMaxSize,
			Logger:                   log.NewSimpleLogger(w.logger),
			LogLevel:                 logLevel,
			IsFailure: func(err error) bool {
				return !errors.Is(err, ErrRateLimitExceeded)
			},
			Queues: map[string]int{
				MessageQueueHighPriority:    6,
				MessageQueueDefaultPriority: 3,
				MessageQueueLowPriority:     1,
			},
		},
	)

	w.Use(WithMetricsExporter(w.tracer))
	w.Use(WithRateLimiter(w.tracer, rateLimiter))
	w.Use(WithErrorLogger(w.tracer))

	for taskType, handler := range w.handlers {
		w.Handle(taskType.String(), handler)
	}

	return w, nil
}

// NewWorkerMetricsServer creates a new metrics server to export prometheus
// metrics.
func NewWorkerMetricsServer(serverConfig *config.ServerConfig, tracer tracing.Tracer) (http.Handler, error) {
	router := chi.NewRouter()

	if serverConfig.CORS.Enabled {
		router.Use(elemoHttp.WithTracedMiddleware(tracer, cors.Handler(cors.Options{
			AllowedOrigins:   serverConfig.CORS.AllowedOrigins,
			AllowedMethods:   serverConfig.CORS.AllowedMethods,
			AllowedHeaders:   serverConfig.CORS.AllowedHeaders,
			AllowCredentials: serverConfig.CORS.AllowCredentials,
			MaxAge:           serverConfig.CORS.MaxAge,
		})))
	}

	router.Route(PathMetrics, func(r chi.Router) {
		r.Handle(PathRoot, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "transport.http.handler/GetPrometheusMetrics")
			defer span.End()

			promhttp.Handler().ServeHTTP(w, r.WithContext(ctx))
		}))
	})

	return router, nil
}
