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

type NamespaceRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateNamespaceOpts
}

func (s *NamespaceRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *NamespaceRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestCreate() {
	ns, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNamespace), ns.ID)
	s.Assert().NotNil(ns.CreatedAt)
	s.Assert().Nil(ns.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)

	ns, err := s.NamespaceRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, ns.ID)
	s.Assert().Equal(s.createOpts.Name, ns.Name)
	s.Assert().Equal(s.createOpts.Description, ns.Description)
	s.Assert().WithinDuration(*created.CreatedAt, *ns.CreatedAt, 100*time.Millisecond)
	s.Require().Len(ns.Projects, 1)
	s.Assert().Equal(project.ID, ns.Projects[0].ID)
	s.Assert().Equal(project.Key, ns.Projects[0].Key)
	s.Assert().Equal(project.Name, ns.Projects[0].Name)
	s.Require().Len(ns.Documents, 1)
	s.Assert().Equal(doc.ID, ns.Documents[0].ID)
	s.Assert().Equal(doc.Name, ns.Documents[0].Name)
	s.Assert().Equal(doc.Excerpt, ns.Documents[0].Excerpt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestGetAll() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	_, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)

	namespaces, err := s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 3)

	var withRelated *repository.Namespace
	for _, ns := range namespaces {
		if ns.ID == created.ID {
			withRelated = ns
			break
		}
	}
	s.Require().NotNil(withRelated)
	s.Require().Len(withRelated.Projects, 1)
	s.Assert().Equal(project.ID, withRelated.Projects[0].ID)
	s.Require().Len(withRelated.Documents, 1)
	s.Assert().Equal(doc.ID, withRelated.Documents[0].ID)

	namespaces, err = s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 2)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Update(context.Background(), created.ID, repository.UpdateNamespaceOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", ns.Name)
	s.Assert().Equal("new description", ns.Description)
	s.Assert().NotNil(ns.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.NamespaceRepo.Delete(context.Background(), created.ID))
	_, err = s.NamespaceRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestNamespaceRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceRepositoryIntegrationTestSuite))
}

type CachedNamespaceRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser      *repository.User
	testOrg       *repository.Organization
	createOpts    repository.CreateNamespaceOpts
	namespaceRepo *repository.RedisCachedNamespaceRepository
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.namespaceRepo, _ = repository.NewCachedNamespaceRepository(s.NamespaceRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestCreate() {
	ns, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(ns.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.NamespaceRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	usingCache, err := s.namespaceRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.namespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	ns, err := s.namespaceRepo.Update(context.Background(), created.ID, repository.UpdateNamespaceOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", ns.Name)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.namespaceRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.namespaceRepo.Delete(context.Background(), created.ID))
	_, err = s.namespaceRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedNamespaceRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedNamespaceRepositoryIntegrationTestSuite))
}
