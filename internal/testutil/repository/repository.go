package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"
)

// NewNeo4jDatabase creates a new Neo4j database connection for testing.
func NewNeo4jDatabase(t *testing.T, conf *config.GraphDatabaseConfig) (*neo4j.Database, func(ctx context.Context) error) {
	driver, err := neo4j.NewDriver(conf)
	require.NoError(t, err)

	db, err := neo4j.NewDatabase(
		neo4j.WithDriver(driver),
		neo4j.WithDatabaseName(conf.Database),
	)
	require.NoError(t, err)

	err = db.Ping(context.Background())
	require.NoError(t, err)

	return db, db.Close
}

// CleanupNeo4jStore deletes all nodes and relationships in the database.
func CleanupNeo4jStore(t *testing.T, ctx context.Context, db *neo4j.Database) {
	_, err := db.GetWriteSession(ctx).Run(ctx, "MATCH (n) DETACH DELETE n", nil)
	require.NoError(t, err)
}

// NewPGDatabase creates a new PostgreSQL database connection for testing.
func NewPGDatabase(t *testing.T, conf *config.RelationalDatabaseConfig) (*pg.Database, func() error) {
	pool, err := pg.NewPool(context.Background(), conf)
	require.NoError(t, err)

	db, err := pg.NewDatabase(
		pg.WithDatabasePool(pool),
	)
	require.NoError(t, err)

	err = db.Ping(context.Background())
	require.NoError(t, err)

	return db, db.Close
}
