package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

var (
	ErrInvalidDatabase   = errors.New("invalid database")   // the database is invalid
	ErrInvalidDriver     = errors.New("invalid driver")     // the driver is invalid
	ErrInvalidRepository = errors.New("invalid repository") // the repository is invalid
	ErrMalformedResult   = errors.New("malformed result")   // the result is malformed
)

// boltLogger implements Neo4j's logger interface.
type boltLogger struct {
	logger log.Logger
}

func (l *boltLogger) LogClientMessage(context string, msg string, args ...any) {
	l.logger.Debug(msg, log.WithDetails(context), log.WithValue(args))
}

func (l *boltLogger) LogServerMessage(context string, msg string, args ...any) {
	l.logger.Debug(msg, log.WithDetails(context), log.WithValue(args))
}

// NewDriver creates a new Neo4j driver.
func NewDriver(conf *config.GraphDatabaseConfig) (neo4j.DriverWithContext, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	driver, err := neo4j.NewDriverWithContext(conf.ConnectionURL(), neo4j.BasicAuth(conf.Username, conf.Password, ""), func(c *neo4j.Config) {
		c.MaxTransactionRetryTime = conf.MaxTransactionRetryTime
		c.MaxConnectionPoolSize = conf.MaxConnectionPoolSize
		c.MaxConnectionLifetime = conf.MaxConnectionLifetime * time.Second
		c.ConnectionAcquisitionTimeout = conf.ConnectionAcquisitionTimeout * time.Second
		c.SocketConnectTimeout = conf.SocketConnectTimeout * time.Second
		c.SocketKeepalive = conf.SocketKeepalive
		c.FetchSize = conf.FetchSize
	})
	if err != nil {
		return nil, errors.Join(ErrInvalidDriver, err)
	}

	return driver, nil
}

// DatabaseOption configures a Neo4j database.
type DatabaseOption func(*Database)

// WithDriver sets the driver for a Neo4j database.
func WithDriver(driver neo4j.DriverWithContext) DatabaseOption {
	return func(db *Database) {
		db.driver = driver
	}
}

// WithDatabaseName sets the name for a Neo4j database.
func WithDatabaseName(name string) DatabaseOption {
	return func(db *Database) {
		db.name = name
	}
}

// WithDatabaseLogger sets the logger for a Neo4j database.
func WithDatabaseLogger(logger log.Logger) DatabaseOption {
	return func(db *Database) {
		db.logger = logger
	}
}

// WithDatabaseTracer sets the tracer for a Neo4j database.
func WithDatabaseTracer(tracer trace.Tracer) DatabaseOption {
	return func(db *Database) {
		db.tracer = tracer
	}
}

// Database represents a Neo4j database, wrapping a Neo4j driver.
type Database struct {
	driver       neo4j.DriverWithContext  `validate:"required"`
	name         string                   `validate:"required"`
	logger       log.Logger               `validate:"required"`
	tracer       trace.Tracer             `validate:"required"`
	readSession  neo4j.SessionWithContext `validate:"required"`
	writeSession neo4j.SessionWithContext `validate:"required"`
}

// GetReadSession returns a "read" session.
func (db *Database) GetReadSession(ctx context.Context) neo4j.SessionWithContext {
	return db.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: db.name,
		FetchSize:    neo4j.FetchDefault,
	})
}

// GetWriteSession returns a "write" session.
func (db *Database) GetWriteSession(ctx context.Context) neo4j.SessionWithContext {
	return db.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: db.name,
		FetchSize:    neo4j.FetchDefault,
		BoltLogger: &boltLogger{
			logger: db.logger,
		},
	})
}

// Ping verifies the connection to the database.
func (db *Database) Ping(ctx context.Context) error {
	return db.driver.VerifyConnectivity(ctx)
}

// Close closes the database connections.
func (db *Database) Close(ctx context.Context) error {
	return db.driver.Close(ctx)
}

// NewDatabase creates a new Neo4j database.
func NewDatabase(opts ...DatabaseOption) (*Database, error) {
	db := &Database{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		opt(db)
	}

	if err := validate.Struct(db); err != nil {
		return nil, errors.Join(ErrInvalidDatabase, err)
	}

	return db, nil
}

// RepositoryOption configures a repository for a Neo4j repository.
type RepositoryOption func(*repository)

// WithDatabase sets the repository for a repository.
func WithDatabase(db *Database) RepositoryOption {
	return func(r *repository) {
		r.db = db
	}
}

// WithRepositoryLogger sets the logger for a repository.
func WithRepositoryLogger(logger log.Logger) RepositoryOption {
	return func(r *repository) {
		r.logger = logger
	}
}

// WithRepositoryTracer sets the tracer for a repository.
func WithRepositoryTracer(tracer trace.Tracer) RepositoryOption {
	return func(r *repository) {
		r.tracer = tracer
	}
}

// repository represents a repository for a Neo4j repository.
type repository struct {
	db     *Database    `validate:"required"`
	logger log.Logger   `validate:"required"`
	tracer trace.Tracer `validate:"required"`
}

// newRepository creates a new repository for a Neo4j repository.
func newRepository(opts ...RepositoryOption) (*repository, error) {
	r := &repository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		opt(r)
	}

	if err := validate.Struct(r); err != nil {
		return nil, errors.Join(ErrInvalidRepository, err)
	}

	return r, nil
}

// PropertyGetter is an interface for getting properties from a node or
// relationship.
type PropertyGetter interface {
	GetProperties() map[string]any
}

// ScanIntoStruct parses a struct from a neo4j node or relationship.
func ScanIntoStruct(n PropertyGetter, dst any, exclude []string) error {
	props := make(map[string]any)

	for k, v := range n.GetProperties() {
		props[k] = v
	}

	for _, e := range exclude {
		delete(props, e)
	}

	return convert.AnyToAny(props, dst)
}

// ParseValueFromRecord parses a value from a neo4j record.
func ParseValueFromRecord[T neo4j.RecordValue](record *neo4j.Record, key string) (T, error) {
	var zero T

	value, _, err := neo4j.GetRecordValue[T](record, key)
	if err != nil {
		return zero, errors.Join(ErrMalformedResult, err)
	}

	return value, nil
}

// ParseIDsFromRecord parses a list of IDs from a neo4j record.
func ParseIDsFromRecord(record *neo4j.Record, key, label string) ([]model.ID, error) {
	val, err := ParseValueFromRecord[[]any](record, key)
	if err != nil {
		return nil, err
	}

	ids := make([]model.ID, len(val))
	for i, p := range val {
		id, err := model.NewIDFromString(p.(string), label)
		if err != nil {
			return nil, err
		}

		ids[i] = id
	}

	return ids, nil
}

// ExecuteAndConsumeResult executes a query and consumes its result.
func ExecuteAndConsumeResult(ctx context.Context, tx neo4j.ManagedTransaction, query string, params map[string]any) error {
	result, err := tx.Run(ctx, query, params)
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

// ExecuteWriteAndConsume executes a query and consumes its result.
func ExecuteWriteAndConsume(ctx context.Context, db *Database, query string, params map[string]any) error {
	_, err := neo4j.ExecuteWrite(ctx, db.GetWriteSession(ctx), func(tx neo4j.ManagedTransaction) (any, error) {
		err := ExecuteAndConsumeResult(ctx, tx, query, params)
		return new(struct{}), err
	})

	return err
}

// ExecuteReadAndReadSingle executes a query and reads a single result.
func ExecuteReadAndReadSingle[T any](ctx context.Context, db *Database, query string, params map[string]any, reader func(record *neo4j.Record) (*T, error)) (*T, error) {
	return neo4j.ExecuteWrite(ctx, db.GetReadSession(ctx), func(tx neo4j.ManagedTransaction) (*T, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return neo4j.SingleTWithContext(ctx, result, reader)
	})
}

// ExecuteWriteAndReadSingle executes a query and reads a single result.
func ExecuteWriteAndReadSingle[T any](ctx context.Context, db *Database, query string, params map[string]any, reader func(record *neo4j.Record) (*T, error)) (*T, error) {
	return neo4j.ExecuteWrite(ctx, db.GetWriteSession(ctx), func(tx neo4j.ManagedTransaction) (*T, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return neo4j.SingleTWithContext(ctx, result, reader)
	})
}

// ExecuteReadAndReadAll executes a query and reads all results.
func ExecuteReadAndReadAll[T any](ctx context.Context, db *Database, query string, params map[string]any, reader func(record *neo4j.Record) (T, error)) ([]T, error) {
	return neo4j.ExecuteWrite(ctx, db.GetReadSession(ctx), func(tx neo4j.ManagedTransaction) ([]T, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		res := make([]T, 0)
		for result.Next(ctx) {
			r, err := reader(result.Record())
			if err != nil {
				return nil, err
			}

			res = append(res, r)
		}

		if result.Err() != nil {
			return nil, result.Err()
		}

		return res, nil
	})
}

// ExecuteWriteAndReadAll executes a query and reads all results.
func ExecuteWriteAndReadAll[T any](ctx context.Context, db *Database, query string, params map[string]any, reader func(record *neo4j.Record) (T, error)) ([]T, error) {
	return neo4j.ExecuteWrite(ctx, db.GetWriteSession(ctx), func(tx neo4j.ManagedTransaction) ([]T, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		res := make([]T, 0)
		for result.Next(ctx) {
			r, err := reader(result.Record())
			if err != nil {
				return nil, err
			}

			res = append(res, r)
		}

		if result.Err() != nil {
			return nil, result.Err()
		}

		return res, nil
	})
}
