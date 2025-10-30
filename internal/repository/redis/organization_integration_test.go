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

type CachedOrganizationRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser         *model.User
	organization     *model.Organization
	organizationRepo *redis.CachedOrganizationRepository
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.organizationRepo, _ = redis.NewCachedOrganizationRepository(s.OrganizationRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.organization = testModel.NewOrganization()

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.organizationRepo.Create(context.Background(), s.testUser.ID, s.organization))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeOrganization), s.organization.ID)
	s.Assert().NotNil(s.organization.CreatedAt)
	s.Assert().Nil(s.organization.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	original, err := s.OrganizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	usingCache, err := s.organizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.organizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, testModel.NewOrganization()))

	originalOrganizations, err := s.OrganizationRepo.GetAll(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheOrganizations, err := s.organizationRepo.GetAll(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalOrganizations, usingCacheOrganizations)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedOrganizations, err := s.organizationRepo.GetAll(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheOrganizations), len(cachedOrganizations))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	patch := map[string]any{
		"name":  "new name",
		"email": testutil.GenerateEmail(10),
	}

	organization, err := s.organizationRepo.Update(context.Background(), s.organization.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.organization.ID, organization.ID)
	s.Assert().Equal(patch["name"], organization.Name)
	s.Assert().Equal(patch["email"], organization.Email)
	s.Assert().Equal(s.organization.Logo, organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().Equal(s.organization.Namespaces, organization.Namespaces)
	s.Assert().Equal(s.organization.Teams, organization.Teams)
	s.Assert().Equal([]model.ID{s.testUser.ID}, organization.Members)
	s.Assert().WithinDuration(*s.organization.CreatedAt, *organization.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(organization.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	_, err := s.organizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.organizationRepo.Delete(context.Background(), s.organization.ID))

	_, err = s.organizationRepo.Get(context.Background(), s.organization.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedOrganizationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedOrganizationRepositoryIntegrationTestSuite))
}
