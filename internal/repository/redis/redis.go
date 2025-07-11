package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

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

	// TODO: use URL parsing + extend options
	options := &redis.Options{
		Addr:                  fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Username:              conf.Username,
		Password:              conf.Password,
		DB:                    conf.RedisConfig.Database,
		MaxRetries:            3,
		DialTimeout:           conf.DialTimeout * time.Second,
		ReadTimeout:           conf.ReadTimeout * time.Second,
		WriteTimeout:          conf.WriteTimeout * time.Second,
		ContextTimeoutEnabled: true,
		PoolSize:              conf.PoolSize,
		MinIdleConns:          conf.MinIdleConnections,
		MaxIdleConns:          conf.MaxIdleConnections,
		ConnMaxIdleTime:       conf.ConnectionMaxIdleTime * time.Second,
		ConnMaxLifetime:       conf.ConnectionMaxLifetime * time.Second,
	}

	if conf.IsSecure {
		options.TLSConfig = &tls.Config{
			ServerName: conf.Host,
			MinVersion: tls.VersionTLS12,
		}
	}

	return redis.NewClient(options), nil
}

// DatabaseOption configures a Redis database.
type DatabaseOption func(*Database) error

// WithClient sets the client for a Redis database.
func WithClient(client redis.UniversalClient) DatabaseOption {
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
func WithDatabaseTracer(tracer tracing.Tracer) DatabaseOption {
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
	tracer tracing.Tracer        `validate:"required"`
}

// Client returns the database client.
func (db *Database) Client() redis.UniversalClient {
	return db.client
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
