package queue

import (
	"time"

	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

// SchedulerOption is a function that can be used to configure an async
// scheduler.
type SchedulerOption func(*Scheduler) error

// WithSchedulerConfig sets the config for the scheduler.
func WithSchedulerConfig(conf *config.WorkerConfig) SchedulerOption {
	return func(w *Scheduler) error {
		if conf == nil {
			return config.ErrNoConfig
		}

		w.conf = conf
		return nil
	}
}

// WithSchedulerTask registers a task to be scheduled at a later time.
func WithSchedulerTask(schedule string, task *asynq.Task) SchedulerOption {
	return func(s *Scheduler) error {
		if schedule == "" {
			return ErrNoSchedule
		}

		if task == nil {
			return ErrNoTask
		}

		s.tasks[task] = schedule
		return nil
	}
}

// WithSchedulerLogger sets the logger for the scheduler.
func WithSchedulerLogger(logger log.Logger) SchedulerOption {
	return func(s *Scheduler) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		s.logger = logger

		return nil
	}
}

// WithSchedulerTracer sets the tracer for the scheduler.
func WithSchedulerTracer(tracer tracing.Tracer) SchedulerOption {
	return func(s *Scheduler) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		s.tracer = tracer

		return nil
	}
}

// Scheduler is a task scheduler that schedules tasks to be processed at a
// later time.
type Scheduler struct {
	conf   *config.WorkerConfig
	logger log.Logger
	tracer tracing.Tracer

	scheduler *asynq.Scheduler

	tasks map[*asynq.Task]string
}

// Start starts the scheduler.
func (s *Scheduler) Start() error {
	return s.scheduler.Run()
}

// Shutdown shuts down the scheduler.
func (s *Scheduler) Shutdown() {
	s.scheduler.Shutdown()
}

// NewScheduler returns a new task scheduler.
func NewScheduler(opts ...SchedulerOption) (*Scheduler, error) {
	s := &Scheduler{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
		tasks:  make(map[*asynq.Task]string),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	logLevel := asynq.InfoLevel
	if s.conf.LogLevel != "" {
		if err := logLevel.Set(s.conf.LogLevel); err != nil {
			return nil, log.ErrInvalidLogLevel
		}
	}

	s.scheduler = asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:         s.conf.Broker.Address(),
			Username:     s.conf.Broker.Username,
			Password:     s.conf.Broker.Password,
			DB:           s.conf.Broker.Database,
			DialTimeout:  s.conf.Broker.DialTimeout * time.Second,
			ReadTimeout:  s.conf.Broker.ReadTimeout * time.Second,
			WriteTimeout: s.conf.Broker.WriteTimeout * time.Second,
			PoolSize:     s.conf.Broker.PoolSize,
		},
		&asynq.SchedulerOpts{
			Logger:   log.NewSimpleLogger(s.logger),
			LogLevel: logLevel,
			Location: time.UTC,
		},
	)

	for task, schedule := range s.tasks {
		if _, err := s.scheduler.Register(schedule, task); err != nil {
			return nil, err
		}
	}

	return s, nil
}
