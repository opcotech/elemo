package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

var (
	ErrInvalidDatabase = errors.New("invalid database") // the database is invalid
	ErrInvalidPool     = errors.New("invalid pool")     // the pool is invalid
	ErrNoPool          = errors.New("no pool")          // the pool is nil
	ErrNoLogger        = errors.New("no logger")        // the logger is nil
	ErrNoTracer        = errors.New("no tracer")        // the tracer is nil
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

	pool, err := pgxpool.New(ctx, conf.ConnectionURL())
	if err != nil {
		return nil, errors.Join(ErrInvalidPool, err)
	}

	return pool, nil
}

// DatabaseOption configures a Postgres database.
type DatabaseOption func(*Database) error

// WithDatabasePool sets the pool for a Postgres database.
func WithDatabasePool(pool Pool) DatabaseOption {
	return func(db *Database) error {
		if pool == nil {
			return ErrNoPool
		}

		db.pool = pool

		return nil
	}
}

// WithDatabaseLogger sets the logger for a Neo4j database.
func WithDatabaseLogger(logger log.Logger) DatabaseOption {
	return func(db *Database) error {
		if logger == nil {
			return ErrNoLogger
		}

		db.logger = logger

		return nil
	}
}

// WithDatabaseTracer sets the tracer for a Neo4j database.
func WithDatabaseTracer(tracer trace.Tracer) DatabaseOption {
	return func(db *Database) error {
		if tracer == nil {
			return ErrNoTracer
		}

		db.tracer = tracer

		return nil
	}
}

// Database represents a Postgres database, wrapping a postgres connection.
type Database struct {
	pool   Pool         `validate:"required"`
	logger log.Logger   `validate:"required"`
	tracer trace.Tracer `validate:"required"`
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
	db := &Database{}

	for _, opt := range opts {
		if err := opt(db); err != nil {
			return nil, err
		}
	}

	if err := validate.Struct(db); err != nil {
		return nil, errors.Join(ErrInvalidDatabase, err)
	}

	return db, nil
}
