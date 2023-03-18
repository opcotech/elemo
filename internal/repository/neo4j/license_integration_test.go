package neo4j_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type LicenseRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
}

func (s *LicenseRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *LicenseRepositoryIntegrationTestSuite) SetupTest() {
	testUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), testUser))

	testOrg := testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), testUser.ID, testOrg))

	testDoc := testModel.NewDocument(testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), testUser.ID, testDoc))

	testNamespace := testModel.NewNamespace()
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), testOrg.ID, testNamespace))

	testProject := testModel.NewProject()
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), testNamespace.ID, testProject))

	testRole := testModel.NewRole()
	s.Require().NoError(s.RoleRepo.Create(context.Background(), testUser.ID, testProject.ID, testRole))
}

func (s *LicenseRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *LicenseRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *LicenseRepositoryIntegrationTestSuite) TestActiveUserCount() {
	count, err := s.LicenseRepo.ActiveUserCount(context.Background())
	s.Require().NoError(err)

	s.Assert().Equal(1, count)
}

func (s *LicenseRepositoryIntegrationTestSuite) TestActiveOrganizationCount() {
	count, err := s.LicenseRepo.ActiveOrganizationCount(context.Background())
	s.Require().NoError(err)

	s.Assert().Equal(1, count)
}

func (s *LicenseRepositoryIntegrationTestSuite) TestDocumentCount() {
	count, err := s.LicenseRepo.DocumentCount(context.Background())
	s.Require().NoError(err)

	s.Assert().Equal(1, count)
}

func (s *LicenseRepositoryIntegrationTestSuite) TestNamespaceCount() {
	count, err := s.LicenseRepo.NamespaceCount(context.Background())
	s.Require().NoError(err)

	s.Assert().Equal(1, count)
}

func (s *LicenseRepositoryIntegrationTestSuite) TestProjectCount() {
	count, err := s.LicenseRepo.ProjectCount(context.Background())
	s.Require().NoError(err)

	s.Assert().Equal(1, count)
}

func (s *LicenseRepositoryIntegrationTestSuite) TestRoleCount() {
	count, err := s.LicenseRepo.RoleCount(context.Background())
	s.Require().NoError(err)

	s.Assert().Equal(1, count)
}

func TestLicenseRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(LicenseRepositoryIntegrationTestSuite))
}
