package container

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/opcotech/elemo/internal/config"
)

var (
	neo4jContainerRequest = func(name string) testcontainers.GenericContainerRequest {
		return testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "neo4j:5.26",
				Name:         name + "-neo4j",
				ExposedPorts: []string{"7687/tcp"},
				WaitingFor:   wait.ForLog("Started."),
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
				Image:        "postgres:17.5",
				Name:         name + "-pg",
				ExposedPorts: []string{"5432/tcp"},
				WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
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
				Image:        "redis:8.0",
				Name:         name + "-redis",
				ExposedPorts: []string{"6379/tcp"},
				WaitingFor:   wait.ForLog("* Ready to accept connections"),
			},
			Started: true,
			Reuse:   true,
		}
	}

	localStackContainerRequest = func(name string) testcontainers.GenericContainerRequest {
		return testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:latest",
				Name:  name + "-localstack",
				ExposedPorts: []string{
					"4566/tcp",
				},
				WaitingFor: wait.ForLog("Ready."),
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
		Port:                  port.Int(),
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
		Host:           host,
		Port:           port.Int(),
		Username:       "elemo",
		Password:       "pgsecret",
		Database:       "elemo",
		MaxConnections: 100,
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
			Port:     port.Int(),
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
		BaseEndpoint:    fmt.Sprintf("http://%s:%d", host, port.Int()),
	}

	return container, conf
}
