package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewClient(t *testing.T) {
	type args struct {
		conf *config.CacheDatabaseConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create new redis client",
			args: args{
				conf: &config.CacheDatabaseConfig{
					RedisConfig: config.RedisConfig{
						Host:         "localhost",
						Port:         6379,
						Username:     "default",
						Password:     "redisSecret",
						Database:     0,
						IsSecure:     false,
						DialTimeout:  10,
						ReadTimeout:  10,
						WriteTimeout: 10,
						PoolSize:     10,
					},
					MaxIdleConnections:    10,
					MinIdleConnections:    10,
					ConnectionMaxIdleTime: 10,
					ConnectionMaxLifetime: 10,
				},
			},
		},
		{
			name: "create new redis client with no config",
			args: args{
				conf: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(tt.args.conf)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithDatabaseClient(t *testing.T) {
	type args struct {
		client redis.UniversalClient
	}
	tests := []struct {
		name    string
		args    args
		want    redis.UniversalClient
		wantErr error
	}{
		{
			name: "create new option with client",
			args: args{
				client: new(mock.RedisClient),
			},
			want: new(mock.RedisClient),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: repository.ErrNoClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(Database)
			err := WithClient(tt.args.client)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.client)
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
		client redis.UniversalClient
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
				client: new(mock.RedisClient),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			want: &Database{
				client: new(mock.RedisClient),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "create new database with nil client",
			args: args{
				client: nil,
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			wantErr: repository.ErrNoClient,
		},
		{
			name: "create new database with nil logger",
			args: args{
				client: new(mock.RedisClient),
				logger: nil,
				tracer: new(mock.Tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			args: args{
				client: new(mock.RedisClient),
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
				WithClient(tt.args.client),
				WithDatabaseLogger(tt.args.logger),
				WithDatabaseTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db)
		})
	}
}

func TestDatabase_GetClient(t *testing.T) {
	t.Parallel()

	client := new(mock.RedisClient)

	db := &Database{
		client: client,
	}

	require.Equal(t, client, db.GetClient())
}

func TestDatabase_Close(t *testing.T) {
	t.Parallel()

	client := new(mock.RedisClient)
	client.On("Close").Return(nil)

	db := &Database{
		client: client,
	}

	require.NoError(t, db.Close())
}

func TestDatabase_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		client func(ctx context.Context) redis.UniversalClient
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
				client: func(ctx context.Context) redis.UniversalClient {
					p := new(mock.RedisClient)
					p.On("Ping", ctx).Return(&redis.StatusCmd{})
					return p
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := &Database{
				client: tt.fields.client(tt.args.ctx),
			}
			err := db.Ping(tt.args.ctx)
			if !tt.wantErr && err != nil {
				require.Error(t, err)
			}
		})
	}
}
