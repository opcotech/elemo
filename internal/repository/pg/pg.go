package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// Pool defines the interface for a database connection pool.
type Pool interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	Reset()
	Config() *pgxpool.Config
	Stat() *pgxpool.Stat
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
}

// NewPool creates a new Postgres pool.
func NewPool(ctx context.Context, conf *config.RelationalDatabaseConfig) (Pool, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	poolConf, err := pgxpool.ParseConfig(conf.ConnectionURL())
	if err != nil {
		return nil, errors.Join(repository.ErrInvalidPool, err)
	}

	poolConf.MaxConnLifetime = conf.MaxConnectionLifetime * time.Second
	poolConf.MaxConnIdleTime = conf.MaxConnectionIdleTime * time.Second
	poolConf.MaxConns = int32(conf.MaxConnections)
	poolConf.MinConns = int32(conf.MinConnections)

	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, errors.Join(repository.ErrInvalidPool, err)
	}

	return pool, nil
}

// DatabaseOption configures a Postgres database.
type DatabaseOption func(*Database) error

// WithDatabasePool sets the pool for a Postgres database.
func WithDatabasePool(pool Pool) DatabaseOption {
	return func(db *Database) error {
		if pool == nil {
			return repository.ErrNoPool
		}

		db.pool = pool

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

// Database represents a Postgres database, wrapping a postgres connection.
type Database struct {
	pool   Pool           `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// Ping checks the database connection.
func (db *Database) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Close closes the database connection.
func (db *Database) Close() error {
	db.pool.Close()
	return nil
}

// NewDatabase creates a new Postgres database.
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
