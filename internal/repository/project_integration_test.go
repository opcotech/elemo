package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type ProjectRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser      *repository.User
	testOrg       *repository.Organization
	testNamespace *repository.Namespace
	createOpts    repository.CreateProjectOpts
}

func (s *ProjectRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *ProjectRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testNamespace, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateProjectOpts(s.testNamespace.ID)
}

func (s *ProjectRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *ProjectRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *ProjectRepositoryIntegrationTestSuite) TestCreate() {
	project, err := s.ProjectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeProject), project.ID)
	s.Assert().NotNil(project.CreatedAt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.ProjectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)

	project, err := s.ProjectRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, project.ID)
	s.Assert().Equal(s.createOpts.Key, project.Key)
	s.Assert().Equal(s.createOpts.Name, project.Name)
	s.Assert().Equal(s.createOpts.Description, project.Description)
	s.Assert().WithinDuration(*created.CreatedAt, *project.CreatedAt, 100*time.Millisecond)
	s.Require().Len(project.Documents, 1)
	s.Assert().Equal(doc.ID, project.Documents[0].ID)
	s.Assert().Equal(doc.Name, project.Documents[0].Name)
	s.Assert().Equal(doc.Excerpt, project.Documents[0].Excerpt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestGetByKey() {
	created, err := s.ProjectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)

	project, err := s.ProjectRepo.GetByKey(context.Background(), created.Key)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, project.ID)
	s.Assert().Equal(created.Key, project.Key)
	s.Require().Len(project.Documents, 1)
	s.Assert().Equal(doc.ID, project.Documents[0].ID)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.ProjectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID))
	s.Require().NoError(err)
	_, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID))
	s.Require().NoError(err)

	projects, err := s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(projects, 3)

	projects, err = s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(projects, 2)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.ProjectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Update(context.Background(), created.ID, repository.UpdateProjectOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", project.Name)
	s.Assert().Equal("new description", project.Description)
	s.Assert().NotNil(project.UpdatedAt)
}

func (s *ProjectRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.ProjectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.ProjectRepo.Delete(context.Background(), created.ID))
	_, err = s.ProjectRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestProjectRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectRepositoryIntegrationTestSuite))
}

type CachedProjectRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser      *repository.User
	testOrg       *repository.Organization
	testNamespace *repository.Namespace
	createOpts    repository.CreateProjectOpts
	projectRepo   *repository.RedisCachedProjectRepository
}

func (s *CachedProjectRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.projectRepo, _ = repository.NewCachedProjectRepository(s.ProjectRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedProjectRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testNamespace, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateProjectOpts(s.testNamespace.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TestCreate() {
	project, err := s.projectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(project.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.projectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.ProjectRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	usingCache, err := s.projectRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TestGetByKey() {
	created, err := s.projectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.ProjectRepo.GetByKey(context.Background(), created.Key)
	s.Require().NoError(err)
	usingCache, err := s.projectRepo.GetByKey(context.Background(), created.Key)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.projectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.ProjectRepo.GetAll(context.Background(), s.testNamespace.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.projectRepo.GetAll(context.Background(), s.testNamespace.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.projectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	project, err := s.projectRepo.Update(context.Background(), created.ID, repository.UpdateProjectOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", project.Name)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedProjectRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.projectRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.projectRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.projectRepo.Delete(context.Background(), created.ID))
	_, err = s.projectRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedProjectRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedProjectRepositoryIntegrationTestSuite))
}
