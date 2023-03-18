package repository

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"

	testConfig "github.com/opcotech/elemo/internal/testutil/config"
)

var (
	bootstrapScript, _ = os.ReadFile(testConfig.RootDir + "/assets/queries/bootstrap.cypher")
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

// BootstrapNeo4jDatabase creates the initial database schema for the system.
func BootstrapNeo4jDatabase(ctx context.Context, t *testing.T, db *neo4j.Database) {
	statements := strings.Split(string(bootstrapScript), ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			_, err := db.GetWriteSession(ctx).Run(ctx, statement, nil)
			if err != nil {
				t.Log(statement)
			}
			require.NoError(t, err)
		}
	}
}

// CleanupNeo4jStore deletes all nodes and relationships in the database.
func CleanupNeo4jStore(ctx context.Context, t *testing.T, db *neo4j.Database) {
	_, err := db.GetWriteSession(ctx).Run(ctx, "MATCH (n) WHERE n.system IS NULL OR n.system = false DETACH DELETE n", nil)
	require.NoError(t, err)
}

// NewPgDatabase creates a new PostgreSQL database connection for testing.
func NewPgDatabase(t *testing.T, conf *config.RelationalDatabaseConfig) (*pg.Database, func() error) {
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
