//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository/neo4j"
)

var (
	neo4jDBConf = &config.DatabaseConfig{
		URL:                          "neo4j://localhost:7687", // "neo4j+s://2b7a4bbd.databases.neo4j.io",
		Username:                     "neo4j",
		Password:                     "neo4jsecret", // "SPnwwBrw4K-EhqmJQ5sRIzt7XnBs3mFeZVo_xATLh1g",
		Name:                         "neo4j",       // "neo4j",
		MaxTransactionRetryTime:      1,
		MaxConnectionPoolSize:        100,
		MaxConnectionLifetime:        1 * time.Hour,
		ConnectionAcquisitionTimeout: 1 * time.Minute,
		SocketConnectTimeout:         1 * time.Minute,
		SocketKeepalive:              true,
		FetchSize:                    0,
	}
)

// newNeo4jDatabase creates a new Neo4j database connection for testing.
func newNeo4jDatabase(t *testing.T) (*neo4j.Database, func(ctx context.Context) error) {
	driver, err := neo4j.NewDriver(neo4jDBConf)
	require.NoError(t, err)

	db, err := neo4j.NewDatabase(
		neo4j.WithDriver(driver),
		neo4j.WithDatabaseName(neo4jDBConf.Name),
	)
	require.NoError(t, err)

	err = db.Ping(context.Background())
	require.NoError(t, err)

	return db, db.Close
}

// cleanupNeo4jStore deletes all nodes and relationships in the database.
func cleanupNeo4jStore(t *testing.T, ctx context.Context, db *neo4j.Database) {
	_, err := db.GetWriteSession(ctx).Run(ctx, "MATCH (n) DETACH DELETE n", nil)
	require.NoError(t, err)
}

func TestNewNeo4jStore(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	require.NotNil(t, db)
}

func TestNeo4jStore_GetReadSession(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	session := db.GetReadSession(ctx)
	require.NotNil(t, session)
}

func TestNeo4jStore_GetWriteSession(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	session := db.GetWriteSession(ctx)
	require.NotNil(t, session)
}

func TestNeo4jStore_Ping(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	require.NoError(t, db.Ping(ctx))
}

func TestNeo4jStore_Close(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	require.NoError(t, closer(ctx))

	err := db.Ping(ctx)
	require.Error(t, err)
}
