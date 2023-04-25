package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// NewClient creates a new Redis client.
func NewClient(conf *config.CacheDatabaseConfig) (redis.UniversalClient, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	db, err := strconv.Atoi(conf.Database)
	if err != nil {
		return nil, config.ErrInvalidConfig
	}

	return redis.NewClient(&redis.Options{
		Addr:                  "",
		Dialer:                nil,
		OnConnect:             nil,
		Username:              conf.Username,
		Password:              conf.Password,
		CredentialsProvider:   nil,
		DB:                    db,
		MaxRetries:            3,
		DialTimeout:           conf.DialTimeout * time.Second,
		ReadTimeout:           conf.ReadTimeout * time.Second,
		WriteTimeout:          conf.WriteTimeout * time.Second,
		ContextTimeoutEnabled: true,
		PoolSize:              conf.PoolSize,
		MinIdleConns:          conf.MinIdleConnections,
		MaxIdleConns:          conf.MaxIdleConnections,
		ConnMaxIdleTime:       conf.ConnectionMaxIdleTime,
		ConnMaxLifetime:       conf.ConnectionMaxLifetime,
	}), nil
}

// DatabaseOption configures a Redis database.
type DatabaseOption func(*Database) error

// WithDatabaseClient sets the client for a Redis database.
func WithDatabaseClient(client redis.UniversalClient) DatabaseOption {
	return func(db *Database) error {
		if client == nil {
			return repository.ErrNoClient
		}

		db.client = client

		return nil
	}
}

// WithDatabaseLogger sets the logger for a Neo4j database.
func WithDatabaseLogger(logger log.Logger) DatabaseOption {
	return func(db *Database) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		db.logger = logger

		return nil
	}
}

// WithDatabaseTracer sets the tracer for a Neo4j database.
func WithDatabaseTracer(tracer trace.Tracer) DatabaseOption {
	return func(db *Database) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		db.tracer = tracer

		return nil
	}
}

// Database represents a Redis database, wrapping a redis connection.
type Database struct {
	client redis.UniversalClient `validate:"required"`
	logger log.Logger            `validate:"required"`
	tracer trace.Tracer          `validate:"required"`
}

// Ping checks the database connection.
func (db *Database) Ping(ctx context.Context) error {
	return db.client.Ping(ctx).Err()
}

// Close closes the database connection.
func (db *Database) Close() error {
	return db.client.Close()
}

// NewDatabase creates a new Redis database.
func NewDatabase(opts ...DatabaseOption) (*Database, error) {
	db := &Database{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(db); err != nil {
			return nil, err
		}
	}

	if err := validate.Struct(db); err != nil {
		return nil, errors.Join(repository.ErrInvalidDatabase, err)
	}

	return db, nil
}
