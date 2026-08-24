package repository

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsCredentials "github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/go-redis/cache/v9"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	neo4jConfig "github.com/neo4j/neo4j-go-driver/v6/neo4j/config"
	"github.com/redis/go-redis/v9"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/metrics"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	EdgeKindAssignedTo    EdgeKind = iota + 1 // ASSIGNED_TO
	EdgeKindBelongsTo                         // BELONGS_TO
	EdgeKindCommented                         // COMMENTED
	EdgeKindCreated                           // CREATED
	EdgeKindHasAttachment                     // HAS_ATTACHMENT
	EdgeKindHasComment                        // HAS_COMMENT
	EdgeKindHasLabel                          // HAS_LABEL
	EdgeKindHasNamespace                      // HAS_NAMESPACE
	EdgeKindHasPermission                     // HAS_PERMISSION
	EdgeKindHasProject                        // HAS_PROJECT
	EdgeKindHasTeam                           // HAS_TEAM
	EdgeKindInvited                           // INVITED
	EdgeKindInvitedTo                         // INVITED_TO
	EdgeKindKindOf                            // KIND_OF
	EdgeKindMemberOf                          // MEMBER_OF
	EdgeKindRelatedTo                         // RELATED_TO
	EdgeKindSpeaks                            // SPEAKS
	EdgeKindWatches                           // WATCHES
	EdgeKindScopedTo                          // SCOPED_TO
	EdgeKindLocatedIn                         // LOCATED_IN
	EdgeKindInScopeOf                         // IN_SCOPE_OF
	EdgeKindGranted                           // GRANTED
	EdgeKindDefinesRole                       // DEFINES_ROLE
)

var (
	ErrCacheDelete = errors.New("failed to delete cache") // cache cannot be deleted
	ErrCacheRead   = errors.New("failed to read cache")   // cache cannot be read
	ErrCacheWrite  = errors.New("failed to write cache")  // cache cannot be written

	// ErrNoMoreRecords is returned by neo4j.Result.Next() when there are no
	// more records to be read, and the result has been fully consumed, but
	// we are still trying to read more.
	ErrNoMoreRecords = &neo4j.UsageError{
		Message: "Result contains no more records",
	}

	ErrInvalidCursor     = errors.New("invalid cursor")          // the page cursor is invalid
	ErrInvalidPageSize   = errors.New("invalid page size")       // the page size is out of bounds
	ErrUnsupportedOrder  = errors.New("unsupported order")       // the order field/direction is not allowed
	ErrUnsupportedFilter = errors.New("unsupported filter")      // the filter is not allowed
	ErrInvalidProjection = errors.New("invalid projection")      // the projection selection is invalid
	ErrQueryCompile      = errors.New("failed to compile query") // the query plan could not be compiled
)

// EdgeKind is the kind of relation between two entities.
//
//go:generate go tool enumer -type=EdgeKind -transform=noop -linecomment -output=repository_edge_kind_gen.go
type EdgeKind uint8

// boltLogger implements Neo4j's logger interface.
type boltLogger struct {
	logger log.Logger
}

func (l *boltLogger) LogClientMessage(ctx string, msg string, args ...any) {
	l.logger.Debug(context.Background(), msg, log.WithDetails(ctx), log.WithValue(args))
}

func (l *boltLogger) LogServerMessage(ctx string, msg string, args ...any) {
	l.logger.Debug(context.Background(), msg, log.WithDetails(ctx), log.WithValue(args))
}

// NewNeo4jDriver creates a new Neo4j driver.
func NewNeo4jDriver(conf *config.GraphDatabaseConfig) (neo4j.Driver, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	driver, err := neo4j.NewDriver(conf.ConnectionURL(), neo4j.BasicAuth(conf.Username, conf.Password, ""), func(c *neo4jConfig.Config) {
		c.MaxTransactionRetryTime = conf.MaxTransactionRetryTime * time.Second
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

// Neo4jDatabaseOption configures a Neo4j database.
type Neo4jDatabaseOption func(*Neo4jDatabase)

// WithNeo4jDriver sets the driver for a Neo4j database.
func WithNeo4jDriver(driver neo4j.Driver) Neo4jDatabaseOption {
	return func(db *Neo4jDatabase) {
		db.driver = driver
	}
}

// WithNeo4jDatabaseName sets the name for a Neo4j database.
func WithNeo4jDatabaseName(name string) Neo4jDatabaseOption {
	return func(db *Neo4jDatabase) {
		db.name = name
	}
}

// WithNeo4jDatabaseLogger sets the logger for a Neo4j database.
func WithNeo4jDatabaseLogger(logger log.Logger) Neo4jDatabaseOption {
	return func(db *Neo4jDatabase) {
		db.logger = logger
	}
}

// WithNeo4jDatabaseTracer sets the tracer for a Neo4j database.
func WithNeo4jDatabaseTracer(tracer tracing.Tracer) Neo4jDatabaseOption {
	return func(db *Neo4jDatabase) {
		db.tracer = tracer
	}
}

// Neo4jDatabase represents a Neo4j database, wrapping a Neo4j driver.
type Neo4jDatabase struct {
	driver neo4j.Driver   `validate:"required"`
	name   string         `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// ReadSession returns a "read" session.
func (db *Neo4jDatabase) ReadSession(ctx context.Context) neo4j.Session {
	return db.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: db.name,
		FetchSize:    neo4j.FetchDefault,
	})
}

// WriteSession returns a "write" session.
func (db *Neo4jDatabase) WriteSession(ctx context.Context) neo4j.Session {
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
func (db *Neo4jDatabase) Ping(ctx context.Context) error {
	return db.driver.VerifyConnectivity(ctx)
}

// Close closes the database connections.
func (db *Neo4jDatabase) Close(ctx context.Context) error {
	return db.driver.Close(ctx)
}

// NewNeo4jDatabase creates a new Neo4j database.
func NewNeo4jDatabase(opts ...Neo4jDatabaseOption) (*Neo4jDatabase, error) {
	db := &Neo4jDatabase{
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

// Neo4jRepositoryOption configures a neo4jBaseRepository for a Neo4j neo4jBaseRepository.
type Neo4jRepositoryOption func(*neo4jBaseRepository) error

// WithNeo4jDatabase sets the neo4jBaseRepository for a neo4jBaseRepository.
func WithNeo4jDatabase(db *Neo4jDatabase) Neo4jRepositoryOption {
	return func(r *neo4jBaseRepository) error {
		if db == nil {
			return ErrNoDriver
		}
		r.db = db

		return nil
	}
}

// WithNeo4jRepositoryLogger sets the logger for a neo4jBaseRepository.
func WithNeo4jRepositoryLogger(logger log.Logger) Neo4jRepositoryOption {
	return func(r *neo4jBaseRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger

		return nil
	}
}

// WithNeo4jRepositoryTracer sets the tracer for a neo4jBaseRepository.
func WithNeo4jRepositoryTracer(tracer tracing.Tracer) Neo4jRepositoryOption {
	return func(r *neo4jBaseRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer

		return nil
	}
}

// neo4jBaseRepository represents a neo4jBaseRepository for a Neo4j neo4jBaseRepository.
type neo4jBaseRepository struct {
	db     *Neo4jDatabase `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// newNeo4jRepository creates a new neo4jBaseRepository for a Neo4j neo4jBaseRepository.
func newNeo4jRepository(opts ...Neo4jRepositoryOption) (*neo4jBaseRepository, error) {
	r := &neo4jBaseRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := validate.Struct(r); err != nil {
		return nil, errors.Join(ErrInvalidRepository, err)
	}

	return r, nil
}

// Neo4jPropertyGetter is an interface for getting properties from a node or
// relationship.
type Neo4jPropertyGetter interface {
	GetProperties() map[string]any
}

func observeUncompiledNeo4j(started time.Time, err error) {
	metrics.ObserveNeo4jQuery(metrics.Neo4jUncompiledQuery, time.Since(started), 0, err)
}

// Neo4jExecuteAndConsumeResult executes a query and consumes its result.
func Neo4jExecuteAndConsumeResult(ctx context.Context, tx neo4j.ManagedTransaction, query string, params map[string]any) (err error) {
	started := time.Now()
	defer func() { observeUncompiledNeo4j(started, err) }()

	result, err := tx.Run(ctx, query, params)
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

// Neo4jExecuteWriteAndConsume executes a query and consumes its result.
func Neo4jExecuteWriteAndConsume(ctx context.Context, db *Neo4jDatabase, query string, params map[string]any) error {
	session := db.WriteSession(ctx)
	defer func(ctx context.Context, sess neo4j.Session) {
		err := sess.Close(ctx)
		if err != nil {
			log.Error(ctx, err)
		}
	}(ctx, session)

	_, err := neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (any, error) {
		err := Neo4jExecuteAndConsumeResult(ctx, tx, query, params)
		return new(struct{}), err
	})

	return err
}

// Neo4jExecuteReadAndReadSingle executes a query and reads a single result.
func Neo4jExecuteReadAndReadSingle[T any](ctx context.Context, db *Neo4jDatabase, query string, params map[string]any, reader func(record *neo4j.Record) (*T, error)) (*T, error) {
	session := db.ReadSession(ctx)
	defer func(ctx context.Context, sess neo4j.Session) {
		err := sess.Close(ctx)
		if err != nil {
			log.Error(ctx, err)
		}
	}(ctx, session)

	return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (res *T, err error) {
		started := time.Now()
		defer func() { observeUncompiledNeo4j(started, err) }()

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		res, err = neo4j.SingleT(ctx, result, reader)
		if err != nil {
			if errors.As(err, &ErrNoMoreRecords) {
				err = ErrNotFound
			}
			return nil, err
		}

		return res, nil
	})
}

// Neo4jExecuteWriteAndReadSingle executes a query and reads a single result.
func Neo4jExecuteWriteAndReadSingle[T any](ctx context.Context, db *Neo4jDatabase, query string, params map[string]any, reader func(record *neo4j.Record) (*T, error)) (*T, error) {
	session := db.WriteSession(ctx)
	defer func(ctx context.Context, sess neo4j.Session) {
		err := sess.Close(ctx)
		if err != nil {
			log.Error(ctx, err)
		}
	}(ctx, session)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (res *T, err error) {
		started := time.Now()
		defer func() { observeUncompiledNeo4j(started, err) }()

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		res, err = neo4j.SingleT(ctx, result, reader)
		if err != nil {
			if errors.As(err, &ErrNoMoreRecords) {
				err = ErrNotFound
			}
			return nil, err
		}

		return res, nil
	})
}

// Neo4jExecuteReadAndReadAll executes a query and reads all results.
func Neo4jExecuteReadAndReadAll[T any](ctx context.Context, db *Neo4jDatabase, query string, params map[string]any, reader func(record *neo4j.Record) (T, error)) ([]T, error) {
	session := db.ReadSession(ctx)
	defer func(ctx context.Context, sess neo4j.Session) {
		err := sess.Close(ctx)
		if err != nil {
			log.Error(ctx, err)
		}
	}(ctx, session)

	return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (res []T, err error) {
		started := time.Now()
		defer func() { observeUncompiledNeo4j(started, err) }()

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		res = make([]T, 0)
		for result.Next(ctx) {
			r, err := reader(result.Record())
			if err != nil {
				return nil, err
			}

			res = append(res, r)
		}

		if result.Err() != nil {
			if errors.As(result.Err(), &ErrNoMoreRecords) {
				return nil, ErrNotFound
			}
			return nil, result.Err()
		}

		return res, nil
	})
}

// Neo4jExecuteWriteAndReadAll executes a query and reads all results.
func Neo4jExecuteWriteAndReadAll[T any](ctx context.Context, db *Neo4jDatabase, query string, params map[string]any, reader func(record *neo4j.Record) (T, error)) ([]T, error) {
	session := db.WriteSession(ctx)
	defer func(ctx context.Context, sess neo4j.Session) {
		err := sess.Close(ctx)
		if err != nil {
			log.Error(ctx, err)
		}
	}(ctx, session)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (res []T, err error) {
		started := time.Now()
		defer func() { observeUncompiledNeo4j(started, err) }()

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		res = make([]T, 0)
		for result.Next(ctx) {
			r, err := reader(result.Record())
			if err != nil {
				return nil, err
			}

			res = append(res, r)
		}

		if result.Err() != nil {
			if errors.As(result.Err(), &ErrNoMoreRecords) {
				return nil, ErrNotFound
			}
			return nil, result.Err()
		}

		return res, nil
	})
}

// PGPool defines the interface for a database connection pool.
//
//go:generate go tool mockgen -destination=mock/mock_pgpool_gen.go -package=mockrepo . PGPool
//go:generate go tool mockgen -destination=mock/mock_pgx_gen.go -package=mockrepo github.com/jackc/pgx/v5 Row,Rows
type PGPool interface {
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
func NewPool(ctx context.Context, conf *config.RelationalDatabaseConfig) (PGPool, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	poolConf, err := pgxpool.ParseConfig(conf.ConnectionURL())
	if err != nil {
		return nil, errors.Join(ErrInvalidPool, err)
	}

	if conf.MaxConnectionLifetime > 0 {
		poolConf.MaxConnLifetime = conf.MaxConnectionLifetime * time.Second
	}

	if conf.MaxConnectionIdleTime > 0 {
		poolConf.MaxConnIdleTime = conf.MaxConnectionIdleTime * time.Second
	}

	poolConf.MaxConns = int32(conf.MaxConnections) // nolint:gosec
	poolConf.MinConns = int32(conf.MinConnections) // nolint:gosec

	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, errors.Join(ErrInvalidPool, err)
	}

	instrumented := newInstrumentedPGPool(pool)
	registerPGPoolStats(instrumented)
	return instrumented, nil
}

// PGDatabaseOption configures a Postgres database.
type PGDatabaseOption func(*PGDatabase) error

// WithDatabasePool sets the pool for a Postgres database.
func WithDatabasePool(pool PGPool) PGDatabaseOption {
	return func(db *PGDatabase) error {
		if pool == nil {
			return ErrNoPool
		}

		db.pool = pool

		return nil
	}
}

// WithPGDatabaseLogger sets the logger for a Neo4j database.
func WithPGDatabaseLogger(logger log.Logger) PGDatabaseOption {
	return func(db *PGDatabase) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		db.logger = logger

		return nil
	}
}

// WithPGDatabaseTracer sets the tracer for a Neo4j database.
func WithPGDatabaseTracer(tracer tracing.Tracer) PGDatabaseOption {
	return func(db *PGDatabase) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		db.tracer = tracer

		return nil
	}
}

// PGDatabase represents a Postgres database, wrapping a postgres connection.
type PGDatabase struct {
	pool   PGPool         `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// Ping checks the database connection.
func (db *PGDatabase) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Pool returns the database pool.
func (db *PGDatabase) Pool() PGPool {
	return db.pool
}

// Close closes the database connection.
func (db *PGDatabase) Close() error {
	db.pool.Close()
	return nil
}

// NewPGDatabase creates a new Postgres database.
func NewPGDatabase(opts ...PGDatabaseOption) (*PGDatabase, error) {
	db := &PGDatabase{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

type PGRepositoryOption func(*pgBaseRepository) error

// WithPGDatabase sets the pgBaseRepository for a pgBaseRepository.
func WithPGDatabase(db *PGDatabase) PGRepositoryOption {
	return func(r *pgBaseRepository) error {
		if db == nil {
			return ErrNoDriver
		}
		r.db = db

		return nil
	}
}

// WithPGRepositoryLogger sets the logger for a pgBaseRepository.
func WithPGRepositoryLogger(logger log.Logger) PGRepositoryOption {
	return func(r *pgBaseRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger

		return nil
	}
}

// WithPGRepositoryTracer sets the tracer for a pgBaseRepository.
func WithPGRepositoryTracer(tracer tracing.Tracer) PGRepositoryOption {
	return func(r *pgBaseRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer

		return nil
	}
}

// pgBaseRepository represents a pgBaseRepository for a Neo4j pgBaseRepository.
type pgBaseRepository struct {
	db     *PGDatabase    `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// newPGRepository creates a new pgBaseRepository for a Postgres pgBaseRepository.
func newPGRepository(opts ...PGRepositoryOption) (*pgBaseRepository, error) {
	r := &pgBaseRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := validate.Struct(r); err != nil {
		return nil, errors.Join(ErrInvalidRepository, err)
	}

	return r, nil
}

// NewRedisClient creates a new Redis client.
func NewRedisClient(conf *config.CacheDatabaseConfig) (redis.UniversalClient, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	// TODO: use URL parsing + extend options
	options := &redis.Options{
		Addr:                  fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Username:              conf.Username,
		Password:              conf.Password,
		DB:                    conf.Database,
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

// RedisDatabaseOption configures a Redis database.
type RedisDatabaseOption func(*RedisDatabase) error

// WithRedisClient sets the client for a Redis database.
func WithRedisClient(client redis.UniversalClient) RedisDatabaseOption {
	return func(db *RedisDatabase) error {
		if client == nil {
			return ErrNoClient
		}

		db.client = client

		return nil
	}
}

// WithRedisDatabaseLogger sets the logger for a Neo4j database.
func WithRedisDatabaseLogger(logger log.Logger) RedisDatabaseOption {
	return func(db *RedisDatabase) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		db.logger = logger

		return nil
	}
}

// WithRedisDatabaseTracer sets the tracer for a Neo4j database.
func WithRedisDatabaseTracer(tracer tracing.Tracer) RedisDatabaseOption {
	return func(db *RedisDatabase) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		db.tracer = tracer

		return nil
	}
}

// RedisDatabase represents a Redis database, wrapping a redis connection.
type RedisDatabase struct {
	client redis.UniversalClient `validate:"required"`
	logger log.Logger            `validate:"required"`
	tracer tracing.Tracer        `validate:"required"`
}

// Client returns the database client.
func (db *RedisDatabase) Client() redis.UniversalClient {
	return db.client
}

// Ping checks the database connection.
func (db *RedisDatabase) Ping(ctx context.Context) error {
	return db.client.Ping(ctx).Err()
}

// Close closes the database connection.
func (db *RedisDatabase) Close() error {
	return db.client.Close()
}

// NewRedisDatabase creates a new Redis database.
func NewRedisDatabase(opts ...RedisDatabaseOption) (*RedisDatabase, error) {
	db := &RedisDatabase{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

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

// CacheBackend represents a cache backend.
//
//go:generate go tool mockgen -destination=mock/mock_universalclient_gen.go -package=mockrepo github.com/redis/go-redis/v9 UniversalClient
//go:generate go tool mockgen -destination=mock/mock_cachebackend_gen.go -package=mockrepo . CacheBackend
type CacheBackend interface {
	Set(item *cache.Item) error
	Get(ctx context.Context, key string, dst any) error
	Delete(ctx context.Context, key string) error
}

// RedisRepositoryOption configures a redisBaseRepository for a Neo4j redisBaseRepository.
type RedisRepositoryOption func(*redisBaseRepository) error

// WithRedisDatabase sets the redisBaseRepository for a redisBaseRepository.
func WithRedisDatabase(db *RedisDatabase) RedisRepositoryOption {
	return func(r *redisBaseRepository) error {
		if db == nil {
			return ErrNoDriver
		}
		r.db = db

		return nil
	}
}

// WithRedisRepositoryLogger sets the logger for a redisBaseRepository.
func WithRedisRepositoryLogger(logger log.Logger) RedisRepositoryOption {
	return func(r *redisBaseRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger

		return nil
	}
}

// WithRedisRepositoryTracer sets the tracer for a redisBaseRepository.
func WithRedisRepositoryTracer(tracer tracing.Tracer) RedisRepositoryOption {
	return func(r *redisBaseRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer

		return nil
	}
}

// WithCacheBackend sets the cache backend for a redisBaseRepository.
// When omitted, newRedisBaseRepository creates a go-redis/cache backend.
func WithCacheBackend(backend CacheBackend) RedisRepositoryOption {
	return func(r *redisBaseRepository) error {
		if backend == nil {
			return ErrNoCacheBackend
		}
		r.cache = backend

		return nil
	}
}

// redisBaseRepository represents a redisBaseRepository for a Neo4j redisBaseRepository.
type redisBaseRepository struct {
	db     *RedisDatabase `validate:"required"`
	cache  CacheBackend   `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

func (r *redisBaseRepository) Set(ctx context.Context, key string, val any) error {
	ctx, span := r.tracer.Start(ctx, "repository.redisBaseRepository/Set")
	defer span.End()

	item := &cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
	}

	if err := r.cache.Set(item); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(ErrCacheWrite, err)
	}

	return nil
}

func (r *redisBaseRepository) SetWithTTL(ctx context.Context, key string, val any, ttl time.Duration) error {
	ctx, span := r.tracer.Start(ctx, "repository.redisBaseRepository/Set")
	defer span.End()

	item := &cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
		TTL:   ttl,
	}

	if err := r.cache.Set(item); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(ErrCacheWrite, err)
	}

	return nil
}

func (r *redisBaseRepository) Get(ctx context.Context, key string, dst any) error {
	ctx, span := r.tracer.Start(ctx, "repository.redisBaseRepository/Get")
	defer span.End()

	if err := r.cache.Get(ctx, key, dst); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(ErrCacheRead, err)
	}

	return nil
}

func (r *redisBaseRepository) Delete(ctx context.Context, key string) error {
	ctx, span := r.tracer.Start(ctx, "repository.redisBaseRepository/Delete")
	defer span.End()

	if err := r.cache.Delete(ctx, key); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(ErrCacheDelete, err)
	}

	return nil
}

func (r *redisBaseRepository) DeletePattern(ctx context.Context, pattern string) error {
	ctx, span := r.tracer.Start(ctx, "repository.redisBaseRepository/DeletePattern")
	defer span.End()

	keys, err := r.db.Client().Keys(ctx, pattern).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	for _, key := range keys {
		if err := r.cache.Delete(ctx, key); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
			return errors.Join(ErrCacheDelete, err)
		}
	}

	return nil
}

// newRedisBaseRepository creates a new redisBaseRepository for a Neo4j redisBaseRepository.
func newRedisBaseRepository(opts ...RedisRepositoryOption) (*redisBaseRepository, error) {
	r := &redisBaseRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if r.cache == nil {
		r.cache = cache.New(&cache.Options{
			Redis:      r.db.Client(),
			LocalCache: nil, // turn off the local cache as it is buggy
		})
	}

	if err := validate.Struct(r); err != nil {
		return nil, errors.Join(ErrInvalidRepository, err)
	}

	return r, nil
}

// composeCacheKey composes a key using a prefix.
func composeCacheKey(params ...any) string {
	sep := ":"

	key := make([]string, len(params))
	for i, param := range params {
		if param != nil {
			switch p := param.(type) {
			case []string:
				key[i] = strings.Join(p, sep)
			default:
				key[i] = fmt.Sprintf("%v", param)
			}
		}
	}
	return strings.Join(key, sep)
}

//go:generate go tool mockgen -destination=mock/mock_s3client_gen.go -package=mockrepo . S3Client
type S3Client interface {
	CreateBucket(ctx context.Context, params *awsS3.CreateBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.CreateBucketOutput, error)
	HeadBucket(ctx context.Context, params *awsS3.HeadBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.HeadBucketOutput, error)
	DeleteBucket(ctx context.Context, params *awsS3.DeleteBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteBucketOutput, error)
	PutObject(ctx context.Context, params *awsS3.PutObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *awsS3.GetObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.GetObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *awsS3.ListObjectsV2Input, optFns ...func(*awsS3.Options)) (*awsS3.ListObjectsV2Output, error)
	DeleteObject(ctx context.Context, params *awsS3.DeleteObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteObjectOutput, error)
}

// NewS3Client creates a new S3 storage client.
func NewS3Client(ctx context.Context, conf *config.S3StorageConfig) (S3Client, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	sdkConfig, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Join(ErrInvalidConfig, err)
	}

	if conf.BaseEndpoint != "" {
		sdkConfig.BaseEndpoint = &conf.BaseEndpoint
	}

	return awsS3.NewFromConfig(sdkConfig, func(o *awsS3.Options) {
		o.UsePathStyle = true
		o.Region = conf.Region
		o.Credentials = awsCredentials.NewStaticCredentialsProvider(
			conf.AccessKeyID,
			conf.SecretAccessKey,
			"",
		)
	}), nil
}

// S3StorageOption configures a Postgres database.
type S3StorageOption func(*S3Storage) error

// WithStorageClient sets the S3 client on the S3Storage.
func WithStorageClient(client S3Client) S3StorageOption {
	return func(storage *S3Storage) error {
		if client == nil {
			return ErrNoClient
		}

		storage.client = client
		return nil
	}
}

// WithStorageBucket sets the S3 bucket on the S3Storage.
func WithStorageBucket(bucket string) S3StorageOption {
	return func(storage *S3Storage) error {
		if bucket == "" {
			return ErrNoBucket
		}

		storage.bucket = bucket
		return nil
	}
}

// WithStorageLogger sets the logger for a Neo4j database.
func WithStorageLogger(logger log.Logger) S3StorageOption {
	return func(storage *S3Storage) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		storage.logger = logger
		return nil
	}
}

// WithStorageTracer sets the tracer for a Neo4j database.
func WithStorageTracer(tracer tracing.Tracer) S3StorageOption {
	return func(storage *S3Storage) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		storage.tracer = tracer
		return nil
	}
}

// S3Storage defines the interface for S3 storage.
type S3Storage struct {
	client S3Client       `validate:"required"`
	bucket string         `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// Ping checks the database connection.
func (s *S3Storage) Ping(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: &s.bucket})
	return err
}

// Client returns the S3 client.
func (s *S3Storage) Client() S3Client {
	return s.client
}

// NewStorage creates a new Postgres database.
func NewStorage(opts ...S3StorageOption) (*S3Storage, error) {
	storage := &S3Storage{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(storage); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

type S3RepositoryOption func(*s3BaseRepository) error

// WithS3Storage sets the s3BaseRepository for a s3BaseRepository.
func WithS3Storage(storage *S3Storage) S3RepositoryOption {
	return func(r *s3BaseRepository) error {
		if storage == nil {
			return ErrNoDriver
		}
		r.storage = storage

		return nil
	}
}

// WithS3RepositoryLogger sets the logger for a s3BaseRepository.
func WithS3RepositoryLogger(logger log.Logger) S3RepositoryOption {
	return func(r *s3BaseRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger

		return nil
	}
}

// WithS3RepositoryTracer sets the tracer for a s3BaseRepository.
func WithS3RepositoryTracer(tracer tracing.Tracer) S3RepositoryOption {
	return func(r *s3BaseRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer

		return nil
	}
}

// s3BaseRepository represents an S3 static file storage.
type s3BaseRepository struct {
	storage *S3Storage
	logger  log.Logger
	tracer  tracing.Tracer
}

// newS3BaseRepository creates a new s3BaseRepository.
func newS3BaseRepository(opts ...S3RepositoryOption) (*s3BaseRepository, error) {
	r := &s3BaseRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func isNotFoundError(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey"
}
