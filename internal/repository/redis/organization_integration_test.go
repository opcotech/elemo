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

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestAddInvitation() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	invitedUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), invitedUser))

	// Cache the organization first
	_, err := s.organizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	// Add invitation - should clear cache
	s.Require().NoError(s.organizationRepo.AddInvitation(context.Background(), s.organization.ID, invitedUser.ID))

	// Verify cache is cleared
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	// Verify invitation exists
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
	s.Assert().Equal(invitedUser.ID, invitations[0].ID)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestRemoveInvitation() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	invitedUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), invitedUser))

	// Add invitation first
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), s.organization.ID, invitedUser.ID))

	// Cache the organization
	_, err := s.organizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	// Remove invitation - should clear cache
	s.Require().NoError(s.organizationRepo.RemoveInvitation(context.Background(), s.organization.ID, invitedUser.ID))

	// Verify cache is cleared
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	// Verify invitation is removed
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(invitations, 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestGetInvitations() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	// Initially no invitations
	invitations, err := s.organizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(invitations, 0)

	// Add multiple invitations
	invitedUser1 := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), invitedUser1))
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), s.organization.ID, invitedUser1.ID))

	invitedUser2 := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), invitedUser2))
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), s.organization.ID, invitedUser2.ID))

	// Get invitations - should delegate to underlying repo (no caching)
	invitations, err = s.organizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 2)

	invitedIDs := make([]model.ID, len(invitations))
	for i, inv := range invitations {
		invitedIDs[i] = inv.ID
	}
	s.Assert().ElementsMatch([]model.ID{invitedUser1.ID, invitedUser2.ID}, invitedIDs)

	// Verify no cache keys were created (GetInvitations doesn't cache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedOrganizationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedOrganizationRepositoryIntegrationTestSuite))
}
