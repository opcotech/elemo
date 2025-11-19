package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type RoleRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser *model.User
	testOrg  *model.Organization
	role     *model.Role
}

func (s *RoleRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *RoleRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.role = testModel.NewRole()
}

func (s *RoleRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *RoleRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *RoleRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeRole), s.role.ID)
	s.Assert().NotNil(s.role.CreatedAt)
	s.Assert().Nil(s.role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	role, err := s.RoleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.role.ID, role.ID)
	s.Assert().Equal(s.role.Name, role.Name)
	s.Assert().Equal(s.role.Description, role.Description)
	s.Assert().Equal([]model.ID{s.testUser.ID}, role.Members)
	s.Assert().Equal(s.role.Permissions, role.Permissions)
	s.Assert().WithinDuration(*s.role.CreatedAt, *role.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, testModel.NewRole()))
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, testModel.NewRole()))

	roles, err := s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(roles, 3)

	roles, err = s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 2)

	roles, err = s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 1)

	roles, err = s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 0)
}

func (s *RoleRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	role, err := s.RoleRepo.Update(context.Background(), s.role.ID, s.testOrg.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.role.ID, role.ID)
	s.Assert().Equal(patch["name"], role.Name)
	s.Assert().Equal(patch["description"], role.Description)
	s.Assert().Equal([]model.ID{s.testUser.ID}, role.Members)
	s.Assert().Equal(s.role.Permissions, role.Permissions)
	s.Assert().WithinDuration(*s.role.CreatedAt, *role.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	newUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), newUser))

	s.Require().NoError(s.RoleRepo.AddMember(context.Background(), s.role.ID, newUser.ID, s.testOrg.ID))

	role, err := s.RoleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().ElementsMatch([]model.ID{s.testUser.ID, newUser.ID}, role.Members)
	s.Assert().Nil(role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	newUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), newUser))

	s.Require().NoError(s.RoleRepo.AddMember(context.Background(), s.role.ID, newUser.ID, s.testOrg.ID))
	s.Require().NoError(s.RoleRepo.RemoveMember(context.Background(), s.role.ID, s.testUser.ID, s.testOrg.ID))

	role, err := s.RoleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().ElementsMatch([]model.ID{newUser.ID}, role.Members)
	s.Assert().Nil(role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	s.Require().NoError(s.RoleRepo.Delete(context.Background(), s.role.ID, s.testOrg.ID))

	_, err := s.RoleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func TestRoleRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RoleRepositoryIntegrationTestSuite))
}

type CachedRoleRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser *model.User
	testOrg  *model.Organization
	role     *model.Role
	roleRepo *repository.RedisCachedRoleRepository
}

func (s *CachedRoleRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.roleRepo, _ = repository.NewCachedRoleRepository(s.RoleRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedRoleRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.role = testModel.NewRole()

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.roleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeRole), s.role.ID)
	s.Assert().NotNil(s.role.CreatedAt)
	s.Assert().Nil(s.role.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	original, err := s.RoleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	usingCache, err := s.roleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.roleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, testModel.NewRole()))

	originalRoles, err := s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheRoles, err := s.roleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalRoles, usingCacheRoles)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedRoles, err := s.roleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheRoles), len(cachedRoles))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	_, err := s.roleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	_, err = s.roleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 2)

	newUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), newUser))

	s.Require().NoError(s.roleRepo.AddMember(context.Background(), s.role.ID, newUser.ID, s.testOrg.ID))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	_, err := s.roleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	_, err = s.roleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 2)

	newUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), newUser))

	s.Require().NoError(s.roleRepo.AddMember(context.Background(), s.role.ID, newUser.ID, s.testOrg.ID))
	s.Require().NoError(s.roleRepo.RemoveMember(context.Background(), s.role.ID, s.testUser.ID, s.testOrg.ID))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	role, err := s.roleRepo.Update(context.Background(), s.role.ID, s.testOrg.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.role.ID, role.ID)
	s.Assert().Equal(patch["name"], role.Name)
	s.Assert().Equal(patch["description"], role.Description)
	s.Assert().Equal([]model.ID{s.testUser.ID}, role.Members)
	s.Assert().Equal(s.role.Permissions, role.Permissions)
	s.Assert().WithinDuration(*s.role.CreatedAt, *role.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(role.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.role))

	_, err := s.roleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.roleRepo.Delete(context.Background(), s.role.ID, s.testOrg.ID))

	_, err = s.roleRepo.Get(context.Background(), s.role.ID, s.testOrg.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedRoleRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedRoleRepositoryIntegrationTestSuite))
}
