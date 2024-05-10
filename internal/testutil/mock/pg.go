package mock

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

type PGRow struct {
	mock.Mock
}

func (r *PGRow) Scan(dest ...any) error {
	args := r.Called(dest)
	if args.Get(0) == nil {
		return args.Error(1)
	}
	for i, x := range args.Get(0).([]any) {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(x))
	}

	return args.Error(1)
}

type PGRows struct {
	mock.Mock
}

func (r *PGRows) Close() {
	r.Called()
}

func (r *PGRows) Err() error {
	return r.Called().Error(0)
}

func (r *PGRows) CommandTag() pgconn.CommandTag {
	return r.Called().Get(0).(pgconn.CommandTag)
}

func (r *PGRows) FieldDescriptions() []pgconn.FieldDescription {
	return r.Called().Get(0).([]pgconn.FieldDescription)
}

func (r *PGRows) Next() bool {
	return r.Called().Bool(0)
}

func (r *PGRows) Scan(dest ...any) error {
	args := r.Called(dest)
	if args.Get(0) == nil {
		return args.Error(1)
	}
	for i, x := range args.Get(0).([]any) {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(x))
	}

	return args.Error(1)
}

func (r *PGRows) Values() ([]any, error) {
	args := r.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]any), args.Error(1)
	}
	return nil, args.Error(1)
}

func (r *PGRows) RawValues() [][]byte {
	return r.Called().Get(0).([][]byte)
}

func (r *PGRows) Conn() *pgx.Conn {
	return r.Called().Get(0).(*pgx.Conn)
}

type PGPool struct {
	mock.Mock
}

func (p *PGPool) Close() {
	args := p.Called()
	_ = args.Error(0)
}

func (p *PGPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := p.Called(ctx)
	return args.Get(0).(*pgxpool.Conn), args.Error(1)
}

func (p *PGPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	args := p.Called(ctx, f)
	return args.Error(0)
}

func (p *PGPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	args := p.Called(ctx)
	return args.Get(0).([]*pgxpool.Conn)
}

func (p *PGPool) Reset() {
	args := p.Called()
	_ = args.Error(0)
}

func (p *PGPool) Config() *pgxpool.Config {
	args := p.Called()
	return args.Get(0).(*pgxpool.Config)
}

func (p *PGPool) Stat() *pgxpool.Stat {
	args := p.Called()
	return args.Get(0).(*pgxpool.Stat)
}

func (p *PGPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := p.Called(append([]any{ctx, sql}, arguments...)...)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (p *PGPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	a := p.Called(ctx, sql, args)
	return a.Get(0).(pgx.Rows), a.Error(1)
}

func (p *PGPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	a := p.Called(ctx, sql, args)
	return a.Get(0).(pgx.Row)
}

func (p *PGPool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := p.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (p *PGPool) Begin(ctx context.Context) (pgx.Tx, error) {
	args := p.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (p *PGPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	args := p.Called(ctx, txOptions)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (p *PGPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := p.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (p *PGPool) Ping(ctx context.Context) error {
	args := p.Called(ctx)
	return args.Error(0)
}
