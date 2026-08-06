package container

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/opcotech/elemo/internal/config"
)

const startupTimeout = 2 * time.Minute

var (
	neo4jContainerRequest = func(name string) testcontainers.GenericContainerRequest {
		return testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "neo4j:2026.06-community",
				Name:         reusableName(name, "neo4j"),
				ExposedPorts: []string{"7687/tcp"},
				WaitingFor:   wait.ForLog("Bolt enabled on").WithStartupTimeout(startupTimeout),
				Env: map[string]string{
					"NEO4J_AUTH": "neo4j/neo4jsecret",
				},
			},
			Started: true,
			Reuse:   true,
		}
	}

	pgContainerRequest = func(name string) testcontainers.GenericContainerRequest {
		return testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:18.4",
				Name:         reusableName(name, "pg"),
				ExposedPorts: []string{"5432/tcp"},
				WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(startupTimeout),
				Env: map[string]string{
					"POSTGRES_USER":     "elemo",
					"POSTGRES_PASSWORD": "pgsecret",
					"POSTGRES_DB":       "elemo",
				},
			},
			Started: true,
			Reuse:   true,
		}
	}

	redisContainerRequest = func(name string) testcontainers.GenericContainerRequest {
		return testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "redis:8.10",
				Name:         reusableName(name, "redis"),
				ExposedPorts: []string{"6379/tcp"},
				WaitingFor:   wait.ForLog("* Ready to accept connections").WithStartupTimeout(startupTimeout),
			},
			Started: true,
			Reuse:   true,
		}
	}

	localStackContainerRequest = func(name string) testcontainers.GenericContainerRequest {
		return testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:4.14.0",
				Name:  reusableName(name, "localstack"),
				ExposedPorts: []string{
					"4566/tcp",
				},
				WaitingFor: wait.ForLog("Ready.").WithStartupTimeout(startupTimeout),
				Env: map[string]string{
					"DEBUG":                 "1",
					"SERVICES":              "s3",
					"AWS_REGION":            "us-east-1",
					"AWS_ACCESS_KEY_ID":     "aws-access-key",
					"AWS_SECRET_ACCESS_KEY": "aws-secret-key",
				},
			},
			Started: true,
			Reuse:   true,
		}
	}
)

// reusableName derives a package-scoped container name so suites in the same
// test package share one container via testcontainers Reuse instead of
// starting a fresh instance per suite.
func reusableName(name, suffix string) string {
	if i := strings.IndexByte(name, '.'); i > 0 {
		name = name[:i]
	}
	return name + "-" + suffix
}

// NewNeo4jContainer creates a new test container for the Neo4j image.
func NewNeo4jContainer(ctx context.Context, t *testing.T, name string) (testcontainers.Container, *config.GraphDatabaseConfig) {
	container, err := testcontainers.GenericContainer(ctx, neo4jContainerRequest(name))
	if err != nil {
		t.Fatal(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "7687/tcp")
	if err != nil {
		t.Fatal(err)
	}

	conf := &config.GraphDatabaseConfig{
		Host:                  host,
		Port:                  int(port.Num()),
		Username:              "neo4j",
		Password:              "neo4jsecret",
		Database:              "neo4j",
		MaxConnectionPoolSize: 100,
	}

	return container, conf
}

// NewPgContainer creates a new test container for the Postgres image.
func NewPgContainer(ctx context.Context, t *testing.T, name string) (testcontainers.Container, *config.RelationalDatabaseConfig) {
	container, err := testcontainers.GenericContainer(ctx, pgContainerRequest(name))
	if err != nil {
		t.Fatal(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}

	conf := &config.RelationalDatabaseConfig{
		Host:                  host,
		Port:                  int(port.Num()),
		Username:              "elemo",
		Password:              "pgsecret",
		Database:              "elemo",
		MaxConnections:        100,
		MaxConnectionLifetime: 300,
		MaxConnectionIdleTime: 10,
		MinConnections:        5,
	}

	return container, conf
}

// NewRedisContainer creates a new test container for the Redis image.
func NewRedisContainer(ctx context.Context, t *testing.T, name string) (testcontainers.Container, *config.CacheDatabaseConfig) {
	container, err := testcontainers.GenericContainer(ctx, redisContainerRequest(name))
	if err != nil {
		t.Fatal(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatal(err)
	}

	conf := &config.CacheDatabaseConfig{
		RedisConfig: config.RedisConfig{
			Host:     host,
			Port:     int(port.Num()),
			Username: "",
			Password: "",
			Database: 0,
		},
	}

	return container, conf
}

// NewLocalStackContainer creates a new test container for the Postgres image.
func NewLocalStackContainer(ctx context.Context, t *testing.T, name string) (testcontainers.Container, *config.S3StorageConfig) {
	container, err := testcontainers.GenericContainer(ctx, localStackContainerRequest(name))
	if err != nil {
		t.Fatal(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "4566/tcp")
	if err != nil {
		t.Fatal(err)
	}

	conf := &config.S3StorageConfig{
		Region:          "us-east-1",
		AccessKeyID:     "aws-access-key",
		SecretAccessKey: "aws-secret-key",
		BaseEndpoint:    fmt.Sprintf("http://%s:%d", host, int(port.Num())),
	}

	return container, conf
}
