//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

var (
	neo4jDBConf = &config.GraphDatabaseConfig{
		Host:                         "localhost",
		Port:                         7687,
		Username:                     "neo4j",
		Password:                     "neo4jsecret",
		Database:                     "neo4j",
		MaxTransactionRetryTime:      1,
		MaxConnectionPoolSize:        100,
		MaxConnectionLifetime:        1 * time.Hour,
		ConnectionAcquisitionTimeout: 1 * time.Minute,
		SocketConnectTimeout:         1 * time.Minute,
		SocketKeepalive:              true,
		FetchSize:                    0,
	}
)

func TestNewNeo4jStore(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	require.NotNil(t, db)
}

func TestNeo4jStore_GetReadSession(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	session := db.GetReadSession(ctx)
	require.NotNil(t, session)
}

func TestNeo4jStore_GetWriteSession(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	session := db.GetWriteSession(ctx)
	require.NotNil(t, session)
}

func TestNeo4jStore_Ping(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	require.NoError(t, db.Ping(ctx))
}

func TestNeo4jStore_Close(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	require.NoError(t, closer(ctx))

	err := db.Ping(ctx)
	require.Error(t, err)
}
