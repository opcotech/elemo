package metrics

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PGPoolStats is a snapshot of Postgres pool occupancy.
type PGPoolStats struct {
	Acquired float64
	Idle     float64
	Max      float64
}

var pgStatsFunc atomic.Value // func() PGPoolStats

// RedisPoolStats is a snapshot of Redis connection pool occupancy.
type RedisPoolStats struct {
	Total    float64
	Idle     float64
	Timeouts float64
}

var redisStatsFunc atomic.Value // func() RedisPoolStats

func init() {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "pg",
		Name:      "pool_acquired",
		Help:      "Postgres connections currently acquired from the pool.",
	},
		func() float64 {
			return loadPGStats().Acquired
		})
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "pg",
		Name:      "pool_idle",
		Help:      "Idle Postgres connections in the pool.",
	},
		func() float64 {
			return loadPGStats().Idle
		})
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "pg",
		Name:      "pool_max",
		Help:      "Maximum Postgres connections allowed in the pool.",
	}, func() float64 {
		return loadPGStats().Max
	})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "redis",
		Name:      "pool_total",
		Help:      "Total Redis connections in the pool.",
	}, func() float64 {
		return loadRedisStats().Total
	})
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "redis",
		Name:      "pool_idle",
		Help:      "Idle Redis connections in the pool.",
	}, func() float64 {
		return loadRedisStats().Idle
	})
	promauto.NewCounterFunc(prometheus.CounterOpts{
		Namespace: "elemo",
		Subsystem: "redis",
		Name:      "pool_timeouts_total",
		Help:      "Times a request timed out waiting for a Redis connection.",
	}, func() float64 {
		return loadRedisStats().Timeouts
	})
}

// SetPGStatsFunc registers the function used to scrape Postgres pool gauges.
func SetPGStatsFunc(fn func() PGPoolStats) {
	if fn != nil {
		pgStatsFunc.Store(fn)
	}
}

func loadPGStats() PGPoolStats {
	fn, ok := pgStatsFunc.Load().(func() PGPoolStats)
	if !ok || fn == nil {
		return PGPoolStats{}
	}
	return fn()
}

// SetRedisStatsFunc registers the function used to scrape Redis pool metrics.
func SetRedisStatsFunc(fn func() RedisPoolStats) {
	if fn != nil {
		redisStatsFunc.Store(fn)
	}
}

func loadRedisStats() RedisPoolStats {
	fn, ok := redisStatsFunc.Load().(func() RedisPoolStats)
	if !ok || fn == nil {
		return RedisPoolStats{}
	}
	return fn()
}
