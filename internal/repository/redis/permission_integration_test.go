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

type CachedPermissionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser       *model.User
	testOrg        *model.Organization
	permission     *model.Permission
	permissionRepo *redis.CachedPermissionRepository
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.permissionRepo, _ = redis.NewCachedPermissionRepository(s.PermissionRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.permission = testModel.NewPermission(s.testUser.ID, s.testOrg.ID, model.PermissionKindRead)

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.permissionRepo.Create(context.Background(), s.permission))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypePermission), s.permission.ID)
	s.Assert().NotNil(s.permission.CreatedAt)
	s.Assert().Nil(s.permission.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.Get(context.Background(), s.permission.ID)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.Get(context.Background(), s.permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.Get(context.Background(), s.permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGetBySubject() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.GetBySubject(context.Background(), s.permission.Subject)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.GetBySubject(context.Background(), s.permission.Subject)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.GetBySubject(context.Background(), s.permission.Subject)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGetByTarget() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.GetByTarget(context.Background(), s.permission.Target)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.GetByTarget(context.Background(), s.permission.Target)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.GetByTarget(context.Background(), s.permission.Target)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	updatedKind := model.PermissionKindDelete
	permission, err := s.permissionRepo.Update(context.Background(), s.permission.ID, updatedKind)
	s.Require().NoError(err)

	s.Assert().Equal(s.permission.ID, permission.ID)
	s.Assert().Equal(s.permission.Subject, permission.Subject)
	s.Assert().Equal(s.permission.Target, permission.Target)
	s.Assert().Equal(updatedKind, permission.Kind)
	s.Assert().WithinDuration(*s.permission.CreatedAt, *permission.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(permission.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	_, err := s.permissionRepo.Get(context.Background(), s.permission.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	s.Require().NoError(s.permissionRepo.Delete(context.Background(), s.permission.ID))

	_, err = s.permissionRepo.Get(context.Background(), s.permission.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestHasPermission() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.HasPermission(
		context.Background(),
		s.permission.Subject,
		s.permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)

	cached, err := s.permissionRepo.HasPermission(
		context.Background(),
		s.permission.Subject,
		s.permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)

	s.Require().Equal(original, cached)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestHasAnyRelation() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)

	cached, err := s.permissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Require().Equal(original, cached)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestHasSystemRole() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.HasSystemRole(
		context.Background(),
		s.testUser.ID,
		model.SystemRoleOwner,
		model.SystemRoleAdmin,
		model.SystemRoleSupport,
	)
	s.Require().NoError(err)

	cached, err := s.permissionRepo.HasSystemRole(
		context.Background(),
		s.testUser.ID,
		model.SystemRoleOwner,
		model.SystemRoleAdmin,
		model.SystemRoleSupport,
	)
	s.Require().NoError(err)

	s.Require().Equal(original, cached)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedPermissionRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedPermissionRepositoryIntegrationTestSuite))
}
