package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/repository"
	"github.com/opcotech/elemo/internal/transport/asynq"
)

// NewSystemService creates a new SystemService for testing.
func NewSystemService(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig, pgDBConf *config.RelationalDatabaseConfig, workerConf *config.WorkerConfig) service.SystemService {
	neo4jDB, _ := repository.NewNeo4jDatabase(t, neo4jDBConf)
	pgDB, _ := repository.NewPgDatabase(t, pgDBConf)

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(neo4jDB),
	)
	require.NoError(t, err)

	permissionRepo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(neo4jDB),
	)
	require.NoError(t, err)

	permissionSvc, err := service.NewPermissionService(
		permissionRepo,
	)
	require.NoError(t, err)

	licenseSvc, err := service.NewLicenseService(
		testutil.ParseLicense(t),
		licenseRepo,
		service.WithPermissionService(permissionSvc),
	)
	require.NoError(t, err)

	queueClient, err := asynq.NewClient(
		asynq.WithClientConfig(workerConf),
	)
	require.NoError(t, err)

	s, err := service.NewSystemService(
		map[model.HealthCheckComponent]service.Pingable{
			model.HealthCheckComponentGraphDB:      neo4jDB,
			model.HealthCheckComponentRelationalDB: pgDB,
			model.HealthCheckComponentLicense:      licenseSvc,
			model.HealthCheckComponentMessageQueue: queueClient,
		},
		&model.VersionInfo{
			Version:   "0.0.1",
			Commit:    "1234567890abcdef1234567890abcdef12345678",
			Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).String(),
			GoVersion: "1.20.0",
		},
	)
	require.NoError(t, err)

	return s
}
