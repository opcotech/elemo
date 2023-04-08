package pg

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewPool(t *testing.T) {
	type args struct {
		ctx  context.Context
		conf *config.RelationalDatabaseConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create new PG pool",
			args: args{
				ctx: context.Background(),
				conf: &config.RelationalDatabaseConfig{
					Host:           "localhost",
					Port:           5432,
					Username:       "postgres",
					Password:       "postgres",
					Database:       "postgres",
					MaxConnections: 10,
				},
			},
		},
		{
			name: "create new PG pool with invalid mac connections",
			args: args{
				ctx: context.Background(),
				conf: &config.RelationalDatabaseConfig{
					Host:           "localhost",
					Port:           5432,
					Username:       "postgres",
					Password:       "postgres",
					Database:       "postgres",
					MaxConnections: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "create new PG pool with invalid config",
			args: args{
				ctx:  context.Background(),
				conf: &config.RelationalDatabaseConfig{},
			},
			wantErr: true,
		},
		{
			name: "create new PG pool with nil config",
			args: args{
				ctx:  context.Background(),
				conf: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewPool(tt.args.ctx, tt.args.conf)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWithDatabasePool(t *testing.T) {
	type args struct {
		pool Pool
	}
	tests := []struct {
		name    string
		args    args
		want    Pool
		wantErr error
	}{
		{
			name: "create new option with pool",
			args: args{
				pool: new(mock.PGPool),
			},
			want: new(mock.PGPool),
		},
		{
			name: "create new option with nil pool",
			args: args{
				pool: nil,
			},
			wantErr: repository.ErrNoPool,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(Database)
			err := WithDatabasePool(tt.args.pool)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.pool)
		})
	}
}

func TestWithDatabaseLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(Database)
			err := WithDatabaseLogger(tt.args.logger)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.logger)
		})
	}
}

func TestWithDatabaseTracer(t *testing.T) {
	type args struct {
		tracer trace.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    trace.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(Database)
			err := WithDatabaseTracer(tt.args.tracer)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.tracer)
		})
	}
}

func TestNewDatabase(t *testing.T) {
	type args struct {
		pool   Pool
		logger log.Logger
		tracer trace.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *Database
		wantErr error
	}{
		{
			name: "create new database",
			args: args{
				pool:   new(mock.PGPool),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			want: &Database{
				pool:   new(mock.PGPool),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "create new database with nil pool",
			args: args{
				pool:   nil,
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			wantErr: repository.ErrNoPool,
		},
		{
			name: "create new database with nil logger",
			args: args{
				pool:   new(mock.PGPool),
				logger: nil,
				tracer: new(mock.Tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			args: args{
				pool:   new(mock.PGPool),
				logger: new(mock.Logger),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, err := NewDatabase(
				WithDatabasePool(tt.args.pool),
				WithDatabaseLogger(tt.args.logger),
				WithDatabaseTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db)
		})
	}
}

func TestDatabase_Close(t *testing.T) {
	t.Parallel()

	pool := new(mock.PGPool)
	pool.On("Close").Return(nil)

	db := &Database{
		pool: pool,
	}

	require.NoError(t, db.Close())
}

func TestDatabase_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		pool func(ctx context.Context) Pool
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "ping database",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				pool: func(ctx context.Context) Pool {
					p := new(mock.PGPool)
					p.On("Ping", ctx).Return(nil)
					return p
				},
			},
		},
		{
			name: "ping database with error",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				pool: func(ctx context.Context) Pool {
					p := new(mock.PGPool)
					p.On("Ping", ctx).Return(errors.New("error"))
					return p
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := &Database{
				pool: tt.fields.pool(tt.args.ctx),
			}
			err := db.Ping(tt.args.ctx)
			if !tt.wantErr && err != nil {
				require.Error(t, err)
			}
		})
	}
}
