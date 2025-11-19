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
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
	"github.com/stretchr/testify/suite"
)

type PermissionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *model.User
	testOrg    *model.Organization
	permission *model.Permission
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupTest() {
	orgOwner := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), orgOwner))

	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), orgOwner.ID, s.testOrg))

	s.permission = testModel.NewPermission(s.testUser.ID, s.testOrg.ID, model.PermissionKindRead)
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypePermission), s.permission.ID)
	s.Assert().NotNil(s.permission.CreatedAt)
	s.Assert().Nil(s.permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permission, err := s.PermissionRepo.Get(context.Background(), s.permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.permission.ID, permission.ID)
	s.Assert().Equal(s.permission.Subject, permission.Subject)
	s.Assert().Equal(s.permission.Target, permission.Target)
	s.Assert().Equal(s.permission.Kind, permission.Kind)
	s.Assert().WithinDuration(*s.permission.CreatedAt, *permission.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasPermission() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	hasPermission, err := s.PermissionRepo.HasPermission(
		context.Background(),
		s.permission.Subject,
		s.permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)
	s.Assert().True(hasPermission)

	hasPermission, err = s.PermissionRepo.HasPermission(
		context.Background(),
		s.permission.Subject,
		s.permission.Target,
		model.PermissionKindDelete,
	)
	s.Require().NoError(err)
	s.Assert().False(hasPermission)

	hasPermission, err = s.PermissionRepo.HasPermission(
		context.Background(),
		s.testUser.ID,
		model.MustNewNilID(model.ResourceTypeOrganization),
		model.PermissionKindCreate,
	)
	s.Require().NoError(err)
	s.Assert().False(hasPermission)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetBySubject() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permissions, err := s.PermissionRepo.GetBySubject(context.Background(), s.permission.Subject)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetByTarget() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permissions, err := s.PermissionRepo.GetByTarget(context.Background(), s.permission.Target)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 2) // the owner and the test user
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetBySubjectAndTarget() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permissions, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), s.permission.Subject, s.permission.Target)
	s.Require().NoError(err)

	s.Assert().Len(permissions, 1)
	s.Assert().Equal(s.permission.ID, permissions[0].ID)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetBySubjectAndTargetSystemLevel() {
	systemTarget := model.MustNewNilID(model.ResourceTypeOrganization)

	permissions, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), s.testUser.ID, systemTarget)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 0)

	s.Require().NoError(testRepo.MakeUserSystemOwner(s.testUser.ID, s.Neo4jDB))

	permissions, err = s.PermissionRepo.GetBySubjectAndTarget(context.Background(), s.testUser.ID, systemTarget)
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(permissions), 1)

	hasAllPermission := false
	for _, perm := range permissions {
		if perm.Kind == model.PermissionKindAll {
			hasAllPermission = true
			break
		}
	}
	s.Assert().True(hasAllPermission, "System owner should have '*' permission")

	for _, perm := range permissions {
		s.Assert().True(perm.Target.IsNil(), "Target should be nil ID for system-level permissions")
		s.Assert().Equal(model.ResourceTypeOrganization, perm.Target.Type)
	}
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetBySubjectAndTargetSystemLevelDirectPermission() {
	systemTarget := model.MustNewNilID(model.ResourceTypeOrganization)

	directPerm := testModel.NewPermission(
		s.testUser.ID,
		systemTarget,
		model.PermissionKindWrite,
	)
	cypher := `
	MATCH (s:` + s.testUser.ID.Label() + ` {id: $subject})
	MATCH (rt:` + model.ResourceTypeResourceType.String() + ` {id: $target_label})
	MERGE (s)-[p:` + repository.EdgeKindHasPermission.String() + ` {id: $id, kind: $kind}]->(rt)
	ON CREATE SET p.created_at = datetime($created_at)
	`
	params := map[string]any{
		"subject":      s.testUser.ID.String(),
		"target_label": systemTarget.Label(),
		"id":           directPerm.ID.String(),
		"kind":         directPerm.Kind.String(),
		"created_at":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	_, err := s.Neo4jDB.GetWriteSession(context.Background()).Run(context.Background(), cypher, params)
	s.Require().NoError(err)

	permissions, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), s.testUser.ID, systemTarget)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	s.Assert().Equal(model.PermissionKindWrite, permissions[0].Kind)
	s.Assert().True(permissions[0].Target.IsNil())
	s.Assert().Equal(model.ResourceTypeOrganization, permissions[0].Target.Type)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	updatedKind := model.PermissionKindDelete
	permission, err := s.PermissionRepo.Update(context.Background(), s.permission.ID, updatedKind)
	s.Require().NoError(err)

	s.Assert().Equal(s.permission.ID, permission.ID)
	s.Assert().Equal(s.permission.Subject, permission.Subject)
	s.Assert().Equal(s.permission.Target, permission.Target)
	s.Assert().Equal(updatedKind, permission.Kind)
	s.Assert().WithinDuration(*s.permission.CreatedAt, *permission.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasAnyRelation() {
	hasRelation, err := s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().False(hasRelation)

	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.testOrg.ID, s.testUser.ID))

	hasRelation, err = s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().True(hasRelation)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasAnyRelationSameUser() {
	// Test that a user always has a relation to themselves
	hasRelation, err := s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testUser.ID)
	s.Require().NoError(err)
	s.Assert().True(hasRelation)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasSystemRole() {
	hasRole, err := s.PermissionRepo.HasSystemRole(
		context.Background(),
		s.testUser.ID,
		model.SystemRoleOwner,
		model.SystemRoleAdmin,
		model.SystemRoleSupport,
	)
	s.Require().NoError(err)
	s.Assert().False(hasRole)

	// Elevate user to system owner
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.testUser.ID, s.Neo4jDB))

	hasRole, err = s.PermissionRepo.HasSystemRole(
		context.Background(),
		s.testUser.ID,
		model.SystemRoleOwner,
		model.SystemRoleAdmin,
		model.SystemRoleSupport,
	)
	s.Require().NoError(err)
	s.Assert().True(hasRole)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	s.Require().NoError(s.PermissionRepo.Delete(context.Background(), s.permission.ID))

	_, err := s.PermissionRepo.Get(context.Background(), s.permission.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestPermissionRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionRepositoryIntegrationTestSuite))
}

type CachedPermissionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser       *model.User
	testOrg        *model.Organization
	permission     *model.Permission
	permissionRepo *repository.RedisCachedPermissionRepository
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.permissionRepo, _ = repository.NewCachedPermissionRepository(s.PermissionRepo, repository.WithRedisDatabase(s.RedisDB))
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

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGetBySubjectAndTarget() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	original, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), s.permission.Subject, s.permission.Target)
	s.Require().NoError(err)

	usingCache, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), s.permission.Subject, s.permission.Target)
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
