package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/opcotech/elemo/internal/pkg/metrics"
)

const (
	pgOpExec     = "exec"
	pgOpQuery    = "query"
	pgOpQueryRow = "query_row"
)

type instrumentedPGPool struct {
	inner PGPool
}

func newInstrumentedPGPool(inner PGPool) PGPool {
	return &instrumentedPGPool{inner: inner}
}

// AsPgxPool returns the concrete pgx pool, walking past instrumentation
// wrappers. Fosite stores require *pgxpool.Pool.
func AsPgxPool(pool PGPool) (*pgxpool.Pool, bool) {
	switch p := pool.(type) {
	case *pgxpool.Pool:
		return p, true
	case *instrumentedPGPool:
		return AsPgxPool(p.inner)
	default:
		return nil, false
	}
}

func (p *instrumentedPGPool) Close() { p.inner.Close() }

func (p *instrumentedPGPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.inner.Acquire(ctx)
}

func (p *instrumentedPGPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	return p.inner.AcquireFunc(ctx, f)
}

func (p *instrumentedPGPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	return p.inner.AcquireAllIdle(ctx)
}

func (p *instrumentedPGPool) Reset() { p.inner.Reset() }

func (p *instrumentedPGPool) Config() *pgxpool.Config { return p.inner.Config() }

func (p *instrumentedPGPool) Stat() *pgxpool.Stat { return p.inner.Stat() }

func (p *instrumentedPGPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	started := time.Now()
	tag, err := p.inner.Exec(ctx, sql, arguments...)
	metrics.ObservePGQuery(pgOpExec, time.Since(started), err)
	return tag, err
}

func (p *instrumentedPGPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	started := time.Now()
	rows, err := p.inner.Query(ctx, sql, args...)
	metrics.ObservePGQuery(pgOpQuery, time.Since(started), err)
	return rows, err
}

func (p *instrumentedPGPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &instrumentedPGRow{
		inner:   p.inner.QueryRow(ctx, sql, args...),
		started: time.Now(),
	}
}

func (p *instrumentedPGPool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.inner.SendBatch(ctx, b)
}

func (p *instrumentedPGPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.inner.Begin(ctx)
}

func (p *instrumentedPGPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.inner.BeginTx(ctx, txOptions)
}

func (p *instrumentedPGPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return p.inner.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *instrumentedPGPool) Ping(ctx context.Context) error {
	return p.inner.Ping(ctx)
}

type instrumentedPGRow struct {
	inner   pgx.Row
	started time.Time
}

func (r *instrumentedPGRow) Scan(dest ...any) error {
	err := r.inner.Scan(dest...)
	metrics.ObservePGQuery(pgOpQueryRow, time.Since(r.started), err)
	return err
}

func registerPGPoolStats(pool PGPool) {
	metrics.SetPGStatsFunc(func() metrics.PGPoolStats {
		stat := pool.Stat()
		if stat == nil {
			return metrics.PGPoolStats{}
		}
		return metrics.PGPoolStats{
			Acquired: float64(stat.AcquiredConns()),
			Idle:     float64(stat.IdleConns()),
			Max:      float64(stat.MaxConns()),
		}
	})
}
