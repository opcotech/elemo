//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	testConfig "github.com/opcotech/elemo/internal/testutil/config"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

var (
	neo4jDBConf = &testConfig.Conf.GraphDatabase
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
