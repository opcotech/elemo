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

func init() {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "pg",
		Name:      "pool_acquired",
		Help:      "Postgres connections currently acquired from the pool.",
	}, func() float64 {
		return loadPGStats().Acquired
	})
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "elemo",
		Subsystem: "pg",
		Name:      "pool_idle",
		Help:      "Idle Postgres connections in the pool.",
	}, func() float64 {
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
