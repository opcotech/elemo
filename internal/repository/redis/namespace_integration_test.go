package redis_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/redis"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type CachedNamespaceRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser      *model.User
	testOrg       *model.Organization
	namespace     *model.Namespace
	namespaceRepo *redis.CachedNamespaceRepository
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.namespaceRepo, _ = redis.NewCachedNamespaceRepository(s.NamespaceRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.namespace = testModel.NewNamespace()

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.namespaceRepo.Create(context.Background(), s.testOrg.ID, s.namespace))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNamespace), s.namespace.ID)
	s.Assert().NotNil(s.namespace.CreatedAt)
	s.Assert().Nil(s.namespace.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, s.namespace))

	original, err := s.NamespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Require().NoError(err)

	usingCache, err := s.namespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.namespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, s.namespace))
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, testModel.NewNamespace()))

	originalNamespaces, err := s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheNamespaces, err := s.namespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalNamespaces, usingCacheNamespaces)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedNamespaces, err := s.namespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheNamespaces), len(cachedNamespaces))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, s.namespace))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	namespace, err := s.namespaceRepo.Update(context.Background(), s.namespace.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.namespace.ID, namespace.ID)
	s.Assert().Equal(patch["name"], namespace.Name)
	s.Assert().Equal(patch["description"], namespace.Description)
	s.Assert().Equal(s.namespace.Projects, namespace.Projects)
	s.Assert().Equal(s.namespace.Documents, namespace.Documents)
	s.Assert().WithinDuration(*s.namespace.CreatedAt, *namespace.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(namespace.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, s.namespace))

	_, err := s.namespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.namespaceRepo.Delete(context.Background(), s.namespace.ID))

	_, err = s.namespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedNamespaceRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CachedNamespaceRepositoryIntegrationTestSuite))
}
