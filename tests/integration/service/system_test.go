//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil"
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

	pgDBConf = &config.RelationalDatabaseConfig{
		Host:                  "localhost",
		Port:                  5432,
		Username:              "elemo",
		Password:              "pgsecret",
		Database:              "elemo",
		IsSecure:              false,
		MaxConnections:        10,
		MaxConnectionLifetime: 10 * time.Minute,
		MaxConnectionIdleTime: 10 * time.Minute,
		MinConnections:        1,
	}
)

func TestSystemService_GetHeartbeat(t *testing.T) {
	s := testutil.NewSystemService(t, neo4jDBConf, pgDBConf)
	require.NoError(t, s.GetHeartbeat(context.Background()))
}

func TestSystemService_GetVersion(t *testing.T) {
	versionInfo := &model.VersionInfo{
		Version:   "0.0.1",
		Commit:    "1234567890abcdef1234567890abcdef12345678",
		Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).String(),
		GoVersion: "1.20.0",
	}

	s := testutil.NewSystemService(t, neo4jDBConf, pgDBConf)

	got := s.GetVersion(context.Background())
	require.Equal(t, versionInfo, got)
}

func TestSystemService_Healthcheck(t *testing.T) {
	s := testutil.NewSystemService(t, neo4jDBConf, pgDBConf)

	health, err := s.GetHealth(context.Background())
	require.NoError(t, err)

	assert.Len(t, health, 2)
	assert.Equal(t, model.HealthStatusHealthy, health[model.HealthCheckComponentGraphDB])
	assert.Equal(t, model.HealthStatusHealthy, health[model.HealthCheckComponentRelationalDB])
}
