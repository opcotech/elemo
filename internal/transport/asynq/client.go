package asynq

import (
	"context"
	"errors"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

// ClientOption is a function that can be used to configure an async worker.
type ClientOption func(*Client) error

// WithClientConfig sets the config for the worker.
func WithClientConfig(conf *config.WorkerConfig) ClientOption {
	return func(w *Client) error {
		w.conf = conf
		return nil
	}
}

// WithClientLogger sets the logger for the worker.
func WithClientLogger(logger log.Logger) ClientOption {
	return func(w *Client) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		w.logger = logger

		return nil
	}
}

// WithClientTracer sets the tracer for the worker.
func WithClientTracer(tracer trace.Tracer) ClientOption {
	return func(w *Client) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		w.tracer = tracer

		return nil
	}
}

// Client is sending async task to the worker for processing.
type Client struct {
	conf      *config.WorkerConfig
	logger    log.Logger
	tracer    trace.Tracer
	client    *asynq.Client
	inspector *asynq.Inspector
}

// Enqueue sends a task to the worker for processing.
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx, span := c.tracer.Start(ctx, "transport.asynq.Client/Enqueue")
	defer span.End()

	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return nil, errors.Join(ErrSendTask, err)
	}
	return info, nil
}

// Ping sends a sample task to the worker and waits for it to finish. If the
// task is not completed within 5 seconds, the task is canceled.
func (c *Client) Ping(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "transport.asynq.Client/Ping")
	defer span.End()

	task, err := NewSystemHealthCheckTask()
	if err != nil {
		return err
	}

	info, err := c.Enqueue(ctx, task)
	if err != nil {
		return err
	}

	for info.State != asynq.TaskStateCompleted {
		if info, err = c.inspector.GetTaskInfo(info.Queue, info.ID); err != nil {
			return errors.Join(ErrReceiveTask, err)
		}
	}

	if info.State != asynq.TaskStateCompleted || info.LastErr != "" {
		return errors.Join(ErrReceiveTask, errors.New(info.LastErr))
	}

	return nil
}

// Close closes the connection with the message broker.
func (c *Client) Close(ctx context.Context) error {
	_, span := c.tracer.Start(ctx, "transport.asynq.Client/Close")
	defer span.End()

	return c.client.Close()
}

// NewClient creates a new client to send async tasks to Worker.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	brokerOpts := asynq.RedisClientOpt{
		Addr:         c.conf.Broker.Address(),
		Username:     c.conf.Broker.Username,
		Password:     c.conf.Broker.Password,
		DB:           c.conf.Broker.Database,
		DialTimeout:  c.conf.Broker.DialTimeout * time.Second,
		ReadTimeout:  c.conf.Broker.ReadTimeout * time.Second,
		WriteTimeout: c.conf.Broker.WriteTimeout * time.Second,
		PoolSize:     c.conf.Broker.PoolSize,
	}

	c.client = asynq.NewClient(brokerOpts)
	c.inspector = asynq.NewInspector(brokerOpts)

	return c, nil
}
