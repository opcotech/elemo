package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"
)

const BootstrapScript = `
// ============================================================================
// Overview
//
// This script creates the initial database schema for the system. It should be
// run once when the system is first installed.
//
// Some resources are system resources, which means they are not created by
// users. They are created by this script and have intentionally invalid IDs to
// prevent users from reading or writing them.
// ============================================================================

// ============================================================================
// Create system resource types
// ============================================================================
UNWIND [
  'Attachment',
  'Comment',
  'Document',
  'Issue',
  'Label',
  'Namespace',
  'Organization',
  'Project',
  'Role',
  'Todo',
  'User'
] AS rt
MERGE (t:ResourceType {id: rt, system: true})
ON CREATE SET t.created_at = datetime();

// ============================================================================
// Create system roles to manage resources
// ============================================================================

// Create roles
UNWIND ['Owner', 'Admin', 'Support'] AS r
MERGE (sr:Role {id: r, name: r, system: true})
ON CREATE SET sr.created_at = datetime();

// Create role bindings
UNWIND [
  'Attachment',
  'Comment',
  'Document',
  'Issue',
  'Label',
  'Namespace',
  'Organization',
  'Project',
  'Role',
  'Todo',
  'User'
] AS t
UNWIND [
  ['Owner', '*'],
  ['Admin', 'create', 'read', 'write'],
  ['Support', 'read', 'write']
] AS bindings
WITH t, bindings[0] AS role, bindings[1..] AS permissions
MATCH (rt:ResourceType {id: t}), (r:Role {id: role})
WITH rt, r, permissions
UNWIND permissions AS permission
MERGE (r)-[p:HAS_PERMISSION {kind: permission}]->(rt)
ON CREATE SET p.created_at = datetime()`

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

	statements := strings.Split(BootstrapScript, ";")

	for _, statement := range statements {
		_, err = db.GetWriteSession(context.Background()).Run(context.Background(), statement, nil)
		require.NoError(t, err)
	}

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
