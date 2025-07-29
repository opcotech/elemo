package repository

import (
	"context"
	"os"
	"strings"
	"testing"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"
	"github.com/opcotech/elemo/internal/repository/redis"
	"github.com/opcotech/elemo/internal/repository/s3"
	testConfig "github.com/opcotech/elemo/internal/testutil/config"
)

var (
	neo4jBootstrapScript, _ = os.ReadFile(testConfig.RootDir + "/assets/queries/bootstrap.cypher")
	pgBootstrapScript, _    = os.ReadFile(testConfig.RootDir + "/assets/queries/bootstrap.sql")
	s3TestBucketName        = "test-bucket"
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
	statements := strings.Split(string(neo4jBootstrapScript), ";")

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

// CleanupNeo4jStore deletes all nodes and relationships from the database.
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

// BootstrapPgDatabase creates the initial database schema for the system.
func BootstrapPgDatabase(ctx context.Context, t *testing.T, db *pg.Database) {
	statements := strings.Split(string(pgBootstrapScript), ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			_, err := db.GetPool().Exec(ctx, statement)
			if err != nil {
				t.Log(statement)
			}
			require.NoError(t, err)
		}
	}
}

func CleanupPgStore(ctx context.Context, t *testing.T, db *pg.Database) {
	_, err := db.GetPool().Exec(ctx, `
	DO $$ DECLARE table_name text;
	BEGIN
		FOR table_name IN (SELECT tablename FROM pg_tables WHERE schemaname='etl') LOOP
			EXECUTE 'TRUNCATE TABLE etl."' || table_name || '" CASCADE;';
		END LOOP;
	END $$;`)
	require.NoError(t, err)
}

// NewRedisDatabase creates a new Redis database connection for testing.
func NewRedisDatabase(t *testing.T, conf *config.CacheDatabaseConfig) (*redis.Database, func() error) {
	client, err := redis.NewClient(conf)
	require.NoError(t, err)

	db, err := redis.NewDatabase(
		redis.WithClient(client),
	)
	require.NoError(t, err)

	err = db.Ping(context.Background())
	require.NoError(t, err)

	return db, db.Close
}

// CleanupRedisStore deletes all keys from the database.
func CleanupRedisStore(ctx context.Context, t *testing.T, db *redis.Database) {
	err := db.GetClient().FlushDB(ctx).Err()
	require.NoError(t, err)
}

// NewS3Storage creates a new S3 storage for testing with LocalStack.
func NewS3Storage(t *testing.T, conf *config.S3StorageConfig) *s3.Storage {
	client, err := s3.NewClient(context.Background(), conf)
	require.NoError(t, err)

	storage, err := s3.NewStorage(
		s3.WithStorageClient(client),
		s3.WithStorageBucket(s3TestBucketName),
	)
	require.NoError(t, err)

	return storage
}

// BootstrapS3Storage creates the initial bucket.
func BootstrapS3Storage(ctx context.Context, t *testing.T, storage *s3.Storage) {
	_, err := storage.GetClient().CreateBucket(ctx, &awsS3.CreateBucketInput{Bucket: &s3TestBucketName})
	require.NoError(t, err)
}

func CleanupS3Storage(ctx context.Context, t *testing.T, storage *s3.Storage) {
	client := storage.GetClient()

	out, err := client.ListObjectsV2(ctx, &awsS3.ListObjectsV2Input{Bucket: &s3TestBucketName})
	require.NoError(t, err)

	for _, o := range out.Contents {
		_, err := client.DeleteObject(ctx, &awsS3.DeleteObjectInput{Bucket: &s3TestBucketName, Key: o.Key})
		require.NoError(t, err)
	}
}
