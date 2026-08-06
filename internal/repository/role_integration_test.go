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

type RoleRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateRoleOpts
}

func (s *RoleRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *RoleRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateRoleOpts(s.testUser.ID, s.testOrg.ID)
}

func (s *RoleRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *RoleRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *RoleRepositoryIntegrationTestSuite) TestCreate() {
	role, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeRole), role.ID)
	s.Assert().NotNil(role.CreatedAt)
	s.Assert().Nil(role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	role, err := s.RoleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, role.ID)
	s.Assert().Equal(s.createOpts.Name, role.Name)
	s.Assert().Equal(s.createOpts.Description, role.Description)
	s.Assert().Equal([]model.ID{s.testUser.ID}, role.Members)
	s.Assert().WithinDuration(*created.CreatedAt, *role.CreatedAt, 100*time.Millisecond)
}

func (s *RoleRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.RoleRepo.Create(context.Background(), testModel.NewCreateRoleOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	_, err = s.RoleRepo.Create(context.Background(), testModel.NewCreateRoleOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)

	roles, err := s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(roles, 3)

	roles, err = s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 2)
}

func (s *RoleRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	role, err := s.RoleRepo.Update(context.Background(), created.ID, s.testOrg.ID, repository.UpdateRoleOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", role.Name)
	s.Assert().Equal("new description", role.Description)
	s.Assert().NotNil(role.UpdatedAt)
}

func (s *RoleRepositoryIntegrationTestSuite) TestAddMember() {
	created, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	newUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.Require().NoError(s.RoleRepo.AddMember(context.Background(), created.ID, newUser.ID, s.testOrg.ID))

	role, err := s.RoleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.testUser.ID, newUser.ID}, role.Members)
}

func (s *RoleRepositoryIntegrationTestSuite) TestRemoveMember() {
	created, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	newUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.Require().NoError(s.RoleRepo.AddMember(context.Background(), created.ID, newUser.ID, s.testOrg.ID))
	s.Require().NoError(s.RoleRepo.RemoveMember(context.Background(), created.ID, s.testUser.ID, s.testOrg.ID))

	role, err := s.RoleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{newUser.ID}, role.Members)
}

func (s *RoleRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.RoleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.RoleRepo.Delete(context.Background(), created.ID, s.testOrg.ID))
	_, err = s.RoleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestRoleRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RoleRepositoryIntegrationTestSuite))
}

type CachedRoleRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateRoleOpts
	roleRepo   *repository.RedisCachedRoleRepository
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
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateRoleOpts(s.testUser.ID, s.testOrg.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestCreate() {
	role, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(role.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.RoleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	usingCache, err := s.roleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.RoleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.roleRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestAddMember() {
	created, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	newUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.roleRepo.AddMember(context.Background(), created.ID, newUser.ID, s.testOrg.ID))
	role, err := s.roleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.testUser.ID, newUser.ID}, role.Members)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestRemoveMember() {
	created, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	newUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.roleRepo.AddMember(context.Background(), created.ID, newUser.ID, s.testOrg.ID))
	s.Require().NoError(s.roleRepo.RemoveMember(context.Background(), created.ID, s.testUser.ID, s.testOrg.ID))
	role, err := s.roleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{newUser.ID}, role.Members)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	role, err := s.roleRepo.Update(context.Background(), created.ID, s.testOrg.ID, repository.UpdateRoleOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", role.Name)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedRoleRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.roleRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.roleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.roleRepo.Delete(context.Background(), created.ID, s.testOrg.ID))
	_, err = s.roleRepo.Get(context.Background(), created.ID, s.testOrg.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedRoleRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedRoleRepositoryIntegrationTestSuite))
}
