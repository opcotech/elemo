package neo4j_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/testutil"
)

type Neo4jRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
}

func (s *Neo4jRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *Neo4jRepositoryIntegrationTestSuite) SetupTest() {}

func (s *Neo4jRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *Neo4jRepositoryIntegrationTestSuite) TestGetReadSession() {
	session := s.Neo4jDB.GetReadSession(context.Background())
	s.Require().NotNil(session)
}

func (s *Neo4jRepositoryIntegrationTestSuite) TestGetWriteSession() {
	session := s.Neo4jDB.GetWriteSession(context.Background())
	s.Require().NotNil(session)
}

func (s *Neo4jRepositoryIntegrationTestSuite) TestPing() {
	err := s.Neo4jDB.Ping(context.Background())
	s.Require().NoError(err)
}

func (s *Neo4jRepositoryIntegrationTestSuite) Test_Z_Close() { // The test suite is run in alphabetical order, so we need to run this test last.
	err := s.Neo4jDB.Close(context.Background())
	s.Require().NoError(err)
}

func TestNeo4jRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(Neo4jRepositoryIntegrationTestSuite))
}
