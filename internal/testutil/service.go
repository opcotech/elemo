package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
)

// NewSystemService creates a new SystemService for testing.
func NewSystemService(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig, pgDBConf *config.RelationalDatabaseConfig) service.SystemService {
	neo4jDB, _ := NewNeo4jDatabase(t, neo4jDBConf)
	pgDB, _ := NewPGDatabase(t, pgDBConf)

	s, err := service.NewSystemService(
		map[model.HealthCheckComponent]service.Pingable{
			model.HealthCheckComponentGraphDB:      neo4jDB,
			model.HealthCheckComponentRelationalDB: pgDB,
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
