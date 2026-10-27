package repository

import (
	"github.com/redis/go-redis/v9"

	"github.com/opcotech/elemo/internal/pkg/metrics"
)

func registerRedisPoolStats(client redis.UniversalClient) {
	metrics.SetRedisStatsFunc(func() metrics.RedisPoolStats {
		stats := client.PoolStats()
		if stats == nil {
			return metrics.RedisPoolStats{}
		}
		return metrics.RedisPoolStats{
			Total:    float64(stats.TotalConns),
			Idle:     float64(stats.IdleConns),
			Timeouts: float64(stats.Timeouts),
		}
	})
}
