package mock

import (
	"context"

	"github.com/go-redis/cache/v9"
	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/pkg/convert"
)

type CacheRepository struct {
	mock.Mock
}

func (c *CacheRepository) Set(item *cache.Item) error {
	args := c.Called(item)
	return args.Error(0)
}

func (c *CacheRepository) Get(ctx context.Context, key string, dst any) error {
	args := c.Called(ctx, key, dst)
	convert.MustAnyToAny(args.Get(0), dst)
	return args.Error(1)
}

func (c *CacheRepository) Delete(ctx context.Context, key string) error {
	args := c.Called(ctx, key)
	return args.Error(0)
}
