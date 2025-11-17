package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// CacheBackend represents a cache backend.
//
//go:generate mockgen -destination=../../testutil/mock/universalclient_gen.go -package=mock -mock_names "UniversalClient=UniversalClient" github.com/redis/go-redis/v9 UniversalClient
//go:generate mockgen -source=cache.go -destination=../../testutil/mock/cachebackend_gen.go -package=mock -mock_names "CacheBackend=CacheBackend"
type CacheBackend interface {
	Set(item *cache.Item) error
	Get(ctx context.Context, key string, dst any) error
	Delete(ctx context.Context, key string) error
}

// RepositoryOption configures a baseRepository for a Neo4j baseRepository.
type RepositoryOption func(*baseRepository) error

// WithDatabase sets the baseRepository for a baseRepository.
func WithDatabase(db *Database) RepositoryOption {
	return func(r *baseRepository) error {
		if db == nil {
			return repository.ErrNoDriver
		}
		r.db = db

		return nil
	}
}

// WithRepositoryLogger sets the logger for a baseRepository.
func WithRepositoryLogger(logger log.Logger) RepositoryOption {
	return func(r *baseRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger

		return nil
	}
}

// WithRepositoryTracer sets the tracer for a baseRepository.
func WithRepositoryTracer(tracer tracing.Tracer) RepositoryOption {
	return func(r *baseRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer

		return nil
	}
}

// baseRepository represents a baseRepository for a Neo4j baseRepository.
type baseRepository struct {
	db     *Database      `validate:"required"`
	cache  CacheBackend   `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

func (r *baseRepository) Set(ctx context.Context, key string, val any) error {
	ctx, span := r.tracer.Start(ctx, "repository.redis.baseRepository/Set")
	defer span.End()

	item := &cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
	}

	if err := r.cache.Set(item); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(repository.ErrCacheWrite, err)
	}

	return nil
}

func (r *baseRepository) Get(ctx context.Context, key string, dst any) error {
	ctx, span := r.tracer.Start(ctx, "repository.redis.baseRepository/Get")
	defer span.End()

	if err := r.cache.Get(ctx, key, dst); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(repository.ErrCacheRead, err)
	}

	return nil
}

func (r *baseRepository) Delete(ctx context.Context, key string) error {
	ctx, span := r.tracer.Start(ctx, "repository.redis.baseRepository/Delete")
	defer span.End()

	if err := r.cache.Delete(ctx, key); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return errors.Join(repository.ErrCacheDelete, err)
	}

	return nil
}

func (r *baseRepository) DeletePattern(ctx context.Context, pattern string) error {
	ctx, span := r.tracer.Start(ctx, "repository.redis.baseRepository/DeletePattern")
	defer span.End()

	keys, err := r.db.GetClient().Keys(ctx, pattern).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	for _, key := range keys {
		if err := r.cache.Delete(ctx, key); err != nil && !errors.Is(err, cache.ErrCacheMiss) {
			return errors.Join(repository.ErrCacheDelete, err)
		}
	}

	return nil
}

// newBaseRepository creates a new baseRepository for a Neo4j baseRepository.
func newBaseRepository(opts ...RepositoryOption) (*baseRepository, error) {
	r := &baseRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	r.cache = cache.New(&cache.Options{
		Redis:      r.db.GetClient(),
		LocalCache: nil, // turn off the local cache as it is buggy
	})

	if err := validate.Struct(r); err != nil {
		return nil, errors.Join(repository.ErrInvalidRepository, err)
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
