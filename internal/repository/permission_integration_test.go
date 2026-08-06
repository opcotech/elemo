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

	testUser   *repository.User
	testOrg    *repository.Organization
	permission *repository.Permission
	createOpts repository.CreatePermissionOpts
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupTest() {
	orgOwner, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(orgOwner.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreatePermissionOpts(s.testUser.ID, s.testOrg.ID, model.PermissionKindRead)
	s.permission = nil
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionRepositoryIntegrationTestSuite) TestCreate() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypePermission), permission.ID)
	s.Assert().NotNil(permission.CreatedAt)
	s.Assert().Nil(permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGet() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	got, err := s.PermissionRepo.Get(context.Background(), permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(permission.ID, got.ID)
	s.Assert().Equal(permission.Subject, got.Subject)
	s.Assert().Equal(permission.Target, got.Target)
	s.Assert().Equal(permission.Kind, got.Kind)
	s.Assert().WithinDuration(*permission.CreatedAt, *got.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(got.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasPermission() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	hasPermission, err := s.PermissionRepo.HasPermission(
		context.Background(),
		permission.Subject,
		permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)
	s.Assert().True(hasPermission)

	hasPermission, err = s.PermissionRepo.HasPermission(
		context.Background(),
		permission.Subject,
		permission.Target,
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
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	permissions, err := s.PermissionRepo.GetBySubject(context.Background(), permission.Subject)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetByTarget() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	permissions, err := s.PermissionRepo.GetByTarget(context.Background(), permission.Target)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 2) // the owner and the test user
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetBySubjectAndTarget() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	permissions, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), permission.Subject, permission.Target)
	s.Require().NoError(err)

	s.Assert().Len(permissions, 1)
	s.Assert().Equal(permission.ID, permissions[0].ID)
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

	directPerm := testModel.NewRepositoryPermission(
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
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	updatedKind := model.PermissionKindDelete
	updated, err := s.PermissionRepo.Update(context.Background(), permission.ID, updatedKind)
	s.Require().NoError(err)

	s.Assert().Equal(permission.ID, updated.ID)
	s.Assert().Equal(permission.Subject, updated.Subject)
	s.Assert().Equal(permission.Target, updated.Target)
	s.Assert().Equal(updatedKind, updated.Kind)
	s.Assert().WithinDuration(*permission.CreatedAt, *updated.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(updated.UpdatedAt)
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
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.PermissionRepo.Delete(context.Background(), permission.ID))

	_, err = s.PermissionRepo.Get(context.Background(), permission.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestPermissionRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionRepositoryIntegrationTestSuite))
}

type CachedPermissionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser       *repository.User
	testOrg        *repository.Organization
	permission     *repository.Permission
	createOpts     repository.CreatePermissionOpts
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
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreatePermissionOpts(s.testUser.ID, s.testOrg.ID, model.PermissionKindRead)
	s.permission = nil

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestCreate() {
	permission, err := s.permissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypePermission), permission.ID)
	s.Assert().NotNil(permission.CreatedAt)
	s.Assert().Nil(permission.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGet() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	original, err := s.PermissionRepo.Get(context.Background(), permission.ID)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.Get(context.Background(), permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.Get(context.Background(), permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGetBySubject() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	original, err := s.PermissionRepo.GetBySubject(context.Background(), permission.Subject)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.GetBySubject(context.Background(), permission.Subject)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.GetBySubject(context.Background(), permission.Subject)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGetByTarget() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	original, err := s.PermissionRepo.GetByTarget(context.Background(), permission.Target)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.GetByTarget(context.Background(), permission.Target)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.GetByTarget(context.Background(), permission.Target)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestGetBySubjectAndTarget() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	original, err := s.PermissionRepo.GetBySubjectAndTarget(context.Background(), permission.Subject, permission.Target)
	s.Require().NoError(err)

	usingCache, err := s.permissionRepo.GetBySubjectAndTarget(context.Background(), permission.Subject, permission.Target)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	cached, err := s.permissionRepo.GetBySubjectAndTarget(context.Background(), permission.Subject, permission.Target)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestUpdate() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	updatedKind := model.PermissionKindDelete
	updated, err := s.permissionRepo.Update(context.Background(), permission.ID, updatedKind)
	s.Require().NoError(err)

	s.Assert().Equal(permission.ID, updated.ID)
	s.Assert().Equal(permission.Subject, updated.Subject)
	s.Assert().Equal(permission.Target, updated.Target)
	s.Assert().Equal(updatedKind, updated.Kind)
	s.Assert().WithinDuration(*permission.CreatedAt, *updated.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(updated.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestDelete() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	_, err = s.permissionRepo.Get(context.Background(), permission.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)

	s.Require().NoError(s.permissionRepo.Delete(context.Background(), permission.ID))

	_, err = s.permissionRepo.Get(context.Background(), permission.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestHasPermission() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	original, err := s.PermissionRepo.HasPermission(
		context.Background(),
		permission.Subject,
		permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)

	cached, err := s.permissionRepo.HasPermission(
		context.Background(),
		permission.Subject,
		permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)

	s.Require().Equal(original, cached)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestHasAnyRelation() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	original, err := s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)

	cached, err := s.permissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)

	s.Require().Equal(original, cached)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedPermissionRepositoryIntegrationTestSuite) TestHasSystemRole() {
	permission, err := s.PermissionRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

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
