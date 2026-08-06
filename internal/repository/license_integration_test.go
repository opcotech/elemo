package repository_test

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
	testUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	testOrg, err := s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(testUser.ID))
	s.Require().NoError(err)

	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(testOrg.ID, testUser.ID))
	s.Require().NoError(err)

	testNamespace, err := s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(testUser.ID, testOrg.ID))
	s.Require().NoError(err)

	testProject, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(testNamespace.ID))
	s.Require().NoError(err)

	_, err = s.RoleRepo.Create(context.Background(), testModel.NewCreateRoleOpts(testUser.ID, testProject.ID))
	s.Require().NoError(err)
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
	suite.Run(t, new(LicenseRepositoryIntegrationTestSuite))
}
