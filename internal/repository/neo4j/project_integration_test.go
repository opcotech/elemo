package neo4j_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type ProjectRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser      *model.User
	testOrg       *model.Organization
	testNamespace *model.Namespace
	project       *model.Project
}

func (s *ProjectRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *ProjectRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testNamespace = testModel.NewNamespace()
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, s.testNamespace))

	s.project = testModel.NewProject()
}

func (s *ProjectRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *ProjectRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *ProjectRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.project))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeProject), s.project.ID)
	s.Assert().NotNil(s.project.CreatedAt)
	s.Assert().Nil(s.project.UpdatedAt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.project))

	project, err := s.ProjectRepo.Get(context.Background(), s.project.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.project.Key, project.Key)
	s.Assert().Equal(s.project.Name, project.Name)
	s.Assert().Equal(s.project.Description, project.Description)
	s.Assert().Equal(s.project.Logo, project.Logo)
	s.Assert().Equal(s.project.Status, project.Status)
	s.Assert().Equal(s.project.Teams, project.Teams)
	s.Assert().Equal(s.project.Documents, project.Documents)
	s.Assert().Equal(s.project.Issues, project.Issues)
	s.Assert().WithinDuration(*s.project.CreatedAt, *project.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(project.UpdatedAt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestGetByKey() {
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.project))

	project, err := s.ProjectRepo.GetByKey(context.Background(), s.project.Key)
	s.Require().NoError(err)

	s.Assert().Equal(s.project.Key, project.Key)
	s.Assert().Equal(s.project.Name, project.Name)
	s.Assert().Equal(s.project.Description, project.Description)
	s.Assert().Equal(s.project.Logo, project.Logo)
	s.Assert().Equal(s.project.Status, project.Status)
	s.Assert().Equal(s.project.Teams, project.Teams)
	s.Assert().Equal(s.project.Documents, project.Documents)
	s.Assert().Equal(s.project.Issues, project.Issues)
	s.Assert().WithinDuration(*s.project.CreatedAt, *project.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(project.UpdatedAt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.project))
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, testModel.NewProject()))
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, testModel.NewProject()))

	projects, err := s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(projects, 3)

	projects, err = s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(projects, 2)

	projects, err = s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(projects, 1)

	projects, err = s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(projects, 0)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.project))

	patch := map[string]any{
		"name":        testutil.GenerateRandomString(10),
		"description": testutil.GenerateRandomString(10),
	}

	project, err := s.ProjectRepo.Update(context.Background(), s.project.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.project.Key, project.Key)
	s.Assert().Equal(patch["name"], project.Name)
	s.Assert().Equal(patch["description"], project.Description)
	s.Assert().Equal(s.project.Logo, project.Logo)
	s.Assert().Equal(s.project.Status, project.Status)
	s.Assert().Equal(s.project.Teams, project.Teams)
	s.Assert().Equal(s.project.Documents, project.Documents)
	s.Assert().Equal(s.project.Issues, project.Issues)
	s.Assert().WithinDuration(*s.project.CreatedAt, *project.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(project.UpdatedAt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.project))

	s.Require().NoError(s.ProjectRepo.Delete(context.Background(), s.project.ID))

	_, err := s.ProjectRepo.Get(context.Background(), s.project.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestProjectRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectRepositoryIntegrationTestSuite))
}
