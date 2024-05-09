package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
)

type SystemServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	systemService service.SystemService
	versionInfo   *model.VersionInfo
}

func (s *SystemServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	s.versionInfo = &model.VersionInfo{
		Version:   "1.0.0",
		Commit:    "1234567890",
		Date:      "2023-01-01T00:00:00Z",
		GoVersion: "1.20.0",
	}

	var err error
	s.systemService, err = service.NewSystemService(
		map[model.HealthCheckComponent]service.Pingable{
			model.HealthCheckComponentGraphDB:      s.Neo4jDB,
			model.HealthCheckComponentRelationalDB: s.PostgresDB,
		},
		s.versionInfo,
	)
	s.Require().NoError(err)
}

func (s *SystemServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *SystemServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *SystemServiceIntegrationTestSuite) TestGetHeartbeat() {
	s.Require().NoError(s.systemService.GetHeartbeat(context.Background()))
}

func (s *SystemServiceIntegrationTestSuite) TestGetHealth() {
	health, err := s.systemService.GetHealth(context.Background())
	s.Require().NoError(err)

	s.Require().Equal(map[model.HealthCheckComponent]model.HealthStatus{
		model.HealthCheckComponentGraphDB:      model.HealthStatusHealthy,
		model.HealthCheckComponentRelationalDB: model.HealthStatusHealthy,
	}, health)
}

func (s *SystemServiceIntegrationTestSuite) TestGetVersion() {
	version := s.systemService.GetVersion(context.Background())
	s.Require().Equal(s.versionInfo, version)
}

func TestSystemServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SystemServiceIntegrationTestSuite))
}
