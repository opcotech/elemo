//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	testService "github.com/opcotech/elemo/internal/testutil/service"
)

func TestSystemService_GetHeartbeat(t *testing.T) {
	s := testService.NewSystemService(t, neo4jDBConf, pgDBConf)
	require.NoError(t, s.GetHeartbeat(context.Background()))
}

func TestSystemService_GetVersion(t *testing.T) {
	versionInfo := &model.VersionInfo{
		Version:   "0.0.1",
		Commit:    "1234567890abcdef1234567890abcdef12345678",
		Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).String(),
		GoVersion: "1.20.0",
	}

	s := testService.NewSystemService(t, neo4jDBConf, pgDBConf)

	got := s.GetVersion(context.Background())
	require.Equal(t, versionInfo, got)
}

func TestSystemService_Healthcheck(t *testing.T) {
	s := testService.NewSystemService(t, neo4jDBConf, pgDBConf)

	health, err := s.GetHealth(context.Background())
	require.NoError(t, err)

	assert.Len(t, health, 2)
	assert.Equal(t, model.HealthStatusHealthy, health[model.HealthCheckComponentGraphDB])
	assert.Equal(t, model.HealthStatusHealthy, health[model.HealthCheckComponentRelationalDB])
}
